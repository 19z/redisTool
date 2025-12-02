package redisTool

import (
	"testing"
)

func TestRedisSet_AddRemove(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	set := tr.Redis.NewSet("testset")

	// Test Add
	err := set.Add("item1", "item2", "item3")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Test Length
	length := set.Length()
	if length != 3 {
		t.Errorf("Length() = %v, want 3", length)
	}

	// Test Exists
	if !set.Exists("item1") {
		t.Error("Exists(item1) = false, want true")
	}

	if set.Exists("item99") {
		t.Error("Exists(item99) = true, want false")
	}

	// Test Remove
	err = set.Remove("item1")
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if set.Exists("item1") {
		t.Error("item1 should not exist after Remove()")
	}

	length = set.Length()
	if length != 2 {
		t.Errorf("Length() after Remove = %v, want 2", length)
	}
}

func TestRedisSet_IsEmpty(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	set := tr.Redis.NewSet("testset")

	if !set.IsEmpty() {
		t.Error("IsEmpty() = false, want true")
	}

	if set.IsNotEmpty() {
		t.Error("IsNotEmpty() = true, want false")
	}

	set.Add("item1")

	if set.IsEmpty() {
		t.Error("IsEmpty() = true, want false")
	}

	if !set.IsNotEmpty() {
		t.Error("IsNotEmpty() = false, want true")
	}
}

func TestRedisSet_Clear(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	set := tr.Redis.NewSet("testset")

	set.Add("item1", "item2", "item3")

	err := set.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if !set.IsEmpty() {
		t.Error("Set should be empty after Clear()")
	}
}

func TestRedisSet_ToArray(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	set := tr.Redis.NewSet("testset")

	set.Add("item1", "item2", "item3")

	items, err := set.ToArray()
	if err != nil {
		t.Fatalf("ToArray() error = %v", err)
	}

	if len(items) != 3 {
		t.Errorf("ToArray() length = %v, want 3", len(items))
	}
}

func TestRedisSet_Iterator(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	set := tr.Redis.NewSet("testset")

	for i := 0; i < 10; i++ {
		set.Add(i)
	}

	count := 0
	for range set.Iterator(3) {
		count++
	}

	if count != 10 {
		t.Errorf("Iterator count = %v, want 10", count)
	}
}

func TestRedisTypeSet_AddRemove(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeSet := NewTypeSet[TestStruct]("testset", tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	s2 := TestStruct{Name: "Bob", Age: 25}

	// Test Add
	err := typeSet.Add(s1, s2)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Test Length
	length := typeSet.Length()
	if length != 2 {
		t.Errorf("Length() = %v, want 2", length)
	}

	// Test Exists
	if !typeSet.Exists(s1) {
		t.Error("Exists(s1) = false, want true")
	}

	// Test Remove
	err = typeSet.Remove(s1)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	length = typeSet.Length()
	if length != 1 {
		t.Errorf("Length() after Remove = %v, want 1", length)
	}
}

func TestRedisTypeSet_ToArray(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeSet := NewTypeSet[TestStruct]("testset", tr.Redis)

	typeSet.Add(
		TestStruct{Name: "Alice", Age: 30},
		TestStruct{Name: "Bob", Age: 25},
	)

	items, err := typeSet.ToArray()
	if err != nil {
		t.Fatalf("ToArray() error = %v", err)
	}

	if len(items) != 2 {
		t.Errorf("ToArray() length = %v, want 2", len(items))
	}
}

func TestRedisTypeSet_Iterator(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeSet := NewTypeSet[TestStruct]("testset", tr.Redis)

	for i := 0; i < 5; i++ {
		typeSet.Add(TestStruct{Name: "User", Age: 20 + i})
	}

	count := 0
	for item := range typeSet.Iterator(2) {
		if item.Name != "User" {
			t.Errorf("Iterator item Name = %v, want User", item.Name)
		}
		count++
	}

	if count != 5 {
		t.Errorf("Iterator count = %v, want 5", count)
	}
}

func TestRedisTypeSet_Clear(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeSet := NewTypeSet[TestStruct]("testset", tr.Redis)

	typeSet.Add(TestStruct{Name: "Alice", Age: 30})

	err := typeSet.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if !typeSet.IsEmpty() {
		t.Error("Set should be empty after Clear()")
	}
}

func TestRedisSet_AddEmpty(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	set := tr.Redis.NewSet("testset")

	// Test Add with no items
	err := set.Add()
	if err != nil {
		t.Fatalf("Add() with no items error = %v", err)
	}

	if !set.IsEmpty() {
		t.Error("Set should be empty after Add() with no items")
	}
}

func TestRedisSet_RemoveEmpty(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	set := tr.Redis.NewSet("testset")
	set.Add("item1")

	// Test Remove with no items
	err := set.Remove()
	if err != nil {
		t.Fatalf("Remove() with no items error = %v", err)
	}

	length := set.Length()
	if length != 1 {
		t.Errorf("Length() after Remove() with no items = %v, want 1", length)
	}
}
