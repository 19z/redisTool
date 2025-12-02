package redisTool

import (
	"testing"
)

func TestRedisZSet_AddRemove(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	zset := tr.Redis.NewZSet("testzset")

	// Test Add
	err := zset.Add("item1", 10.5)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	err = zset.Add("item2", 20.5)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Test Length
	length := zset.Length()
	if length != 2 {
		t.Errorf("Length() = %v, want 2", length)
	}

	// Test Score
	score, ok := zset.Score("item1")
	if !ok {
		t.Fatal("Score() returned false")
	}
	if score != 10.5 {
		t.Errorf("Score() = %v, want 10.5", score)
	}

	// Test Score non-existent item
	_, ok = zset.Score("nonexistent")
	if ok {
		t.Error("Score() non-existent item should return false")
	}

	// Test Remove
	err = zset.Remove("item1")
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	length = zset.Length()
	if length != 1 {
		t.Errorf("Length() after Remove = %v, want 1", length)
	}
}

func TestRedisZSet_IncrementScore(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	zset := tr.Redis.NewZSet("testzset")

	zset.Add("item1", 10.0)

	// Test IncrementScore
	newScore, err := zset.IncrementScore("item1", 5.5)
	if err != nil {
		t.Fatalf("IncrementScore() error = %v", err)
	}
	if newScore != 15.5 {
		t.Errorf("IncrementScore() = %v, want 15.5", newScore)
	}

	// Test IncrementScore on non-existent item
	newScore, err = zset.IncrementScore("item2", 10.0)
	if err != nil {
		t.Fatalf("IncrementScore() on non-existent item error = %v", err)
	}
	if newScore != 10.0 {
		t.Errorf("IncrementScore() on non-existent = %v, want 10.0", newScore)
	}
}

func TestRedisZSet_IsEmpty(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	zset := tr.Redis.NewZSet("testzset")

	if !zset.IsEmpty() {
		t.Error("IsEmpty() = false, want true")
	}

	if zset.IsNotEmpty() {
		t.Error("IsNotEmpty() = true, want false")
	}

	zset.Add("item1", 10.0)

	if zset.IsEmpty() {
		t.Error("IsEmpty() = true, want false")
	}

	if !zset.IsNotEmpty() {
		t.Error("IsNotEmpty() = false, want true")
	}
}

func TestRedisZSet_Clear(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	zset := tr.Redis.NewZSet("testzset")

	zset.Add("item1", 10.0)
	zset.Add("item2", 20.0)

	err := zset.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if !zset.IsEmpty() {
		t.Error("ZSet should be empty after Clear()")
	}
}

func TestRedisZSet_RangeByScore(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	zset := tr.Redis.NewZSet("testzset")

	zset.Add("item1", 10.0)
	zset.Add("item2", 20.0)
	zset.Add("item3", 30.0)
	zset.Add("item4", 40.0)

	// Test RangeByScore
	items, err := zset.RangeByScore(15.0, 35.0, false)
	if err != nil {
		t.Fatalf("RangeByScore() error = %v", err)
	}
	if len(items) != 2 {
		t.Errorf("RangeByScore() length = %v, want 2", len(items))
	}
}

func TestRedisZSet_RangeByRank(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	zset := tr.Redis.NewZSet("testzset")

	zset.Add("item1", 10.0)
	zset.Add("item2", 20.0)
	zset.Add("item3", 30.0)

	// Test RangeByRank
	items, err := zset.RangeByRank(0, 1)
	if err != nil {
		t.Fatalf("RangeByRank() error = %v", err)
	}
	if len(items) != 2 {
		t.Errorf("RangeByRank() length = %v, want 2", len(items))
	}
}

func TestRedisZSet_RemoveRangeByScore(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	zset := tr.Redis.NewZSet("testzset")

	zset.Add("item1", 10.0)
	zset.Add("item2", 20.0)
	zset.Add("item3", 30.0)
	zset.Add("item4", 40.0)

	// Test RemoveRangeByScore
	err := zset.RemoveRangeByScore(15.0, 35.0)
	if err != nil {
		t.Fatalf("RemoveRangeByScore() error = %v", err)
	}

	length := zset.Length()
	if length != 2 {
		t.Errorf("Length() after RemoveRangeByScore = %v, want 2", length)
	}
}

func TestRedisZSet_Iterator(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	zset := tr.Redis.NewZSet("testzset")

	for i := 0; i < 10; i++ {
		zset.Add(i, float64(i*10))
	}

	count := 0
	for item := range zset.Iterator(3) {
		if item.Score < 0 {
			t.Errorf("Iterator item Score = %v, want >= 0", item.Score)
		}
		count++
	}

	if count != 10 {
		t.Errorf("Iterator count = %v, want 10", count)
	}
}

func TestRedisTypeZSet_AddRemove(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeZset := NewTypeZSet[TestStruct]("testzset", tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	s2 := TestStruct{Name: "Bob", Age: 25}

	// Test Add
	err := typeZset.Add(s1, 95.5)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	err = typeZset.Add(s2, 88.0)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Test Length
	length := typeZset.Length()
	if length != 2 {
		t.Errorf("Length() = %v, want 2", length)
	}

	// Test Score
	score, ok := typeZset.Score(s1)
	if !ok {
		t.Fatal("Score() returned false")
	}
	if score != 95.5 {
		t.Errorf("Score() = %v, want 95.5", score)
	}

	// Test IncrementScore
	newScore, err := typeZset.IncrementScore(s1, 2.5)
	if err != nil {
		t.Fatalf("IncrementScore() error = %v", err)
	}
	if newScore != 98.0 {
		t.Errorf("IncrementScore() = %v, want 98.0", newScore)
	}

	// Test Remove
	err = typeZset.Remove(s1)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	length = typeZset.Length()
	if length != 1 {
		t.Errorf("Length() after Remove = %v, want 1", length)
	}
}

func TestRedisTypeZSet_RangeByScore(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeZset := NewTypeZSet[TestStruct]("testzset", tr.Redis)

	typeZset.Add(TestStruct{Name: "Alice", Age: 30}, 95.5)
	typeZset.Add(TestStruct{Name: "Bob", Age: 25}, 88.0)
	typeZset.Add(TestStruct{Name: "Charlie", Age: 35}, 92.0)

	// Test RangeByScore
	items, err := typeZset.RangeByScore(90.0, 96.0)
	if err != nil {
		t.Fatalf("RangeByScore() error = %v", err)
	}
	if len(items) != 2 {
		t.Errorf("RangeByScore() length = %v, want 2", len(items))
	}
}

func TestRedisTypeZSet_RangeByRank(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeZset := NewTypeZSet[TestStruct]("testzset", tr.Redis)

	typeZset.Add(TestStruct{Name: "Alice", Age: 30}, 95.5)
	typeZset.Add(TestStruct{Name: "Bob", Age: 25}, 88.0)
	typeZset.Add(TestStruct{Name: "Charlie", Age: 35}, 92.0)

	// Test RangeByRank
	items, err := typeZset.RangeByRank(0, 1)
	if err != nil {
		t.Fatalf("RangeByRank() error = %v", err)
	}
	if len(items) != 2 {
		t.Errorf("RangeByRank() length = %v, want 2", len(items))
	}
}

func TestRedisTypeZSet_RemoveRangeByScore(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeZset := NewTypeZSet[TestStruct]("testzset", tr.Redis)

	typeZset.Add(TestStruct{Name: "Alice", Age: 30}, 95.5)
	typeZset.Add(TestStruct{Name: "Bob", Age: 25}, 88.0)
	typeZset.Add(TestStruct{Name: "Charlie", Age: 35}, 92.0)

	// Test RemoveRangeByScore
	err := typeZset.RemoveRangeByScore(90.0, 93.0)
	if err != nil {
		t.Fatalf("RemoveRangeByScore() error = %v", err)
	}

	length := typeZset.Length()
	if length != 2 {
		t.Errorf("Length() after RemoveRangeByScore = %v, want 2", length)
	}
}

func TestRedisTypeZSet_Iterator(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeZset := NewTypeZSet[TestStruct]("testzset", tr.Redis)

	for i := 0; i < 5; i++ {
		typeZset.Add(TestStruct{Name: "User", Age: 20 + i}, float64(i*10))
	}

	count := 0
	for item := range typeZset.Iterator(2) {
		if item.Value.Name != "User" {
			t.Errorf("Iterator item Name = %v, want User", item.Value.Name)
		}
		if item.Score < 0 {
			t.Errorf("Iterator item Score = %v, want >= 0", item.Score)
		}
		count++
	}

	if count != 5 {
		t.Errorf("Iterator count = %v, want 5", count)
	}
}

func TestRedisTypeZSet_IteratorFilterByScore(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeZset := NewTypeZSet[TestStruct]("testzset", tr.Redis)

	typeZset.Add(TestStruct{Name: "Alice", Age: 30}, 95.5)
	typeZset.Add(TestStruct{Name: "Bob", Age: 25}, 88.0)
	typeZset.Add(TestStruct{Name: "Charlie", Age: 35}, 92.0)
	typeZset.Add(TestStruct{Name: "David", Age: 28}, 85.0)

	count := 0
	for item := range typeZset.IteratorFilterByScore(90.0, 96.0) {
		if item.Score < 90.0 || item.Score > 96.0 {
			t.Errorf("IteratorFilterByScore item Score = %v, want between 90.0 and 96.0", item.Score)
		}
		count++
	}

	if count != 2 {
		t.Errorf("IteratorFilterByScore count = %v, want 2", count)
	}
}

func TestRedisTypeZSet_Clear(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeZset := NewTypeZSet[TestStruct]("testzset", tr.Redis)

	typeZset.Add(TestStruct{Name: "Alice", Age: 30}, 95.5)

	err := typeZset.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if !typeZset.IsEmpty() {
		t.Error("ZSet should be empty after Clear()")
	}
}

func TestRedisZSet_RemoveMultiple(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	zset := tr.Redis.NewZSet("testzset")

	zset.Add("item1", 10.0)
	zset.Add("item2", 20.0)
	zset.Add("item3", 30.0)

	// Test Remove multiple items
	err := zset.Remove("item1", "item2")
	if err != nil {
		t.Fatalf("Remove() multiple items error = %v", err)
	}

	length := zset.Length()
	if length != 1 {
		t.Errorf("Length() after Remove multiple = %v, want 1", length)
	}

	// Test Remove with no items
	err = zset.Remove()
	if err != nil {
		t.Fatalf("Remove() with no items error = %v", err)
	}
}
