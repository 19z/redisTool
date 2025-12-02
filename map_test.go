package redisTool

import (
	"testing"
)

func TestRedisMap_SetGet(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	rmap := tr.Redis.NewMap("testmap")

	// Test Set
	err := rmap.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Test Get
	value, ok := rmap.Get("key1")
	if !ok {
		t.Fatal("Get() returned false")
	}
	if value != "value1" {
		t.Errorf("Get() = %v, want value1", value)
	}

	// Test Get non-existent key
	_, ok = rmap.Get("nonexistent")
	if ok {
		t.Error("Get() non-existent key should return false")
	}
}

func TestRedisMap_Delete(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	rmap := tr.Redis.NewMap("testmap")

	rmap.Set("key1", "value1")
	rmap.Set("key2", "value2")

	// Test Delete
	err := rmap.Delete("key1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if rmap.Exists("key1") {
		t.Error("key1 should not exist after Delete()")
	}

	// Test Delete multiple keys
	rmap.Set("key3", "value3")
	err = rmap.Delete("key2", "key3")
	if err != nil {
		t.Fatalf("Delete() multiple keys error = %v", err)
	}

	if rmap.Exists("key2") || rmap.Exists("key3") {
		t.Error("keys should not exist after Delete()")
	}

	// Test Delete with no keys
	err = rmap.Delete()
	if err != nil {
		t.Fatalf("Delete() with no keys error = %v", err)
	}
}

func TestRedisMap_Exists(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	rmap := tr.Redis.NewMap("testmap")

	if rmap.Exists("key1") {
		t.Error("Exists() = true, want false")
	}

	rmap.Set("key1", "value1")

	if !rmap.Exists("key1") {
		t.Error("Exists() = false, want true")
	}
}

func TestRedisMap_Length(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	rmap := tr.Redis.NewMap("testmap")

	if rmap.Length() != 0 {
		t.Errorf("Length() = %v, want 0", rmap.Length())
	}

	rmap.Set("key1", "value1")
	rmap.Set("key2", "value2")

	if rmap.Length() != 2 {
		t.Errorf("Length() = %v, want 2", rmap.Length())
	}
}

func TestRedisMap_IsEmpty(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	rmap := tr.Redis.NewMap("testmap")

	if !rmap.IsEmpty() {
		t.Error("IsEmpty() = false, want true")
	}

	if rmap.IsNotEmpty() {
		t.Error("IsNotEmpty() = true, want false")
	}

	rmap.Set("key1", "value1")

	if rmap.IsEmpty() {
		t.Error("IsEmpty() = true, want false")
	}

	if !rmap.IsNotEmpty() {
		t.Error("IsNotEmpty() = false, want true")
	}
}

func TestRedisMap_Clear(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	rmap := tr.Redis.NewMap("testmap")

	rmap.Set("key1", "value1")
	rmap.Set("key2", "value2")

	err := rmap.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if !rmap.IsEmpty() {
		t.Error("Map should be empty after Clear()")
	}
}

func TestRedisMap_ToArray(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	rmap := tr.Redis.NewMap("testmap")

	rmap.Set("key1", "value1")
	rmap.Set("key2", "value2")

	items, err := rmap.ToArray()
	if err != nil {
		t.Fatalf("ToArray() error = %v", err)
	}

	if len(items) != 2 {
		t.Errorf("ToArray() length = %v, want 2", len(items))
	}
}

func TestRedisMap_Keys(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	rmap := tr.Redis.NewMap("testmap")

	rmap.Set("key1", "value1")
	rmap.Set("key2", "value2")

	keys, err := rmap.Keys()
	if err != nil {
		t.Fatalf("Keys() error = %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Keys() length = %v, want 2", len(keys))
	}
}

func TestRedisMap_Iterator(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	rmap := tr.Redis.NewMap("testmap")

	for i := 0; i < 10; i++ {
		rmap.Set("key"+string(rune('0'+i)), i)
	}

	count := 0
	for item := range rmap.Iterator(3) {
		if item.Key == "" {
			t.Error("Iterator item Key is empty")
		}
		count++
	}

	if count != 10 {
		t.Errorf("Iterator count = %v, want 10", count)
	}
}

func TestRedisTypeMap_SetGet(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeMap := NewTypeMap[TestStruct]("testmap", tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	// Test Set
	err := typeMap.Set("key1", s1)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Test Get
	value, ok := typeMap.Get("key1")
	if !ok {
		t.Fatal("Get() returned false")
	}
	if value.Name != "Alice" {
		t.Errorf("Get() Name = %v, want Alice", value.Name)
	}

	// Test Get non-existent key
	_, ok = typeMap.Get("nonexistent")
	if ok {
		t.Error("Get() non-existent key should return false")
	}
}

func TestRedisTypeMap_Delete(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeMap := NewTypeMap[TestStruct]("testmap", tr.Redis)

	typeMap.Set("key1", TestStruct{Name: "Alice", Age: 30})
	typeMap.Set("key2", TestStruct{Name: "Bob", Age: 25})

	err := typeMap.Delete("key1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if typeMap.Exists("key1") {
		t.Error("key1 should not exist after Delete()")
	}
}

func TestRedisTypeMap_ToArray(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeMap := NewTypeMap[TestStruct]("testmap", tr.Redis)

	typeMap.Set("key1", TestStruct{Name: "Alice", Age: 30})
	typeMap.Set("key2", TestStruct{Name: "Bob", Age: 25})

	items, err := typeMap.ToArray()
	if err != nil {
		t.Fatalf("ToArray() error = %v", err)
	}

	if len(items) != 2 {
		t.Errorf("ToArray() length = %v, want 2", len(items))
	}
}

func TestRedisTypeMap_Keys(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeMap := NewTypeMap[TestStruct]("testmap", tr.Redis)

	typeMap.Set("key1", TestStruct{Name: "Alice", Age: 30})
	typeMap.Set("key2", TestStruct{Name: "Bob", Age: 25})

	keys, err := typeMap.Keys()
	if err != nil {
		t.Fatalf("Keys() error = %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Keys() length = %v, want 2", len(keys))
	}
}

func TestRedisTypeMap_Iterator(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeMap := NewTypeMap[TestStruct]("testmap", tr.Redis)

	for i := 0; i < 5; i++ {
		typeMap.Set("key"+string(rune('0'+i)), TestStruct{Name: "User", Age: 20 + i})
	}

	count := 0
	for item := range typeMap.Iterator(2) {
		if item.Key == "" {
			t.Error("Iterator item Key is empty")
		}
		if item.Value.Name != "User" {
			t.Errorf("Iterator item Name = %v, want User", item.Value.Name)
		}
		count++
	}

	if count != 5 {
		t.Errorf("Iterator count = %v, want 5", count)
	}
}

func TestRedisTypeMap_SafeUpset(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeMap := NewTypeMap[TestStruct]("testmap", tr.Redis)

	typeMap.Set("key1", TestStruct{Name: "Alice", Age: 30})

	// Test SafeUpset with existing key
	oldValue, exist, err := typeMap.SafeUpset("key1", TestStruct{Name: "Bob", Age: 25})
	if err != nil {
		t.Fatalf("SafeUpset() error = %v", err)
	}
	if !exist {
		t.Error("SafeUpset() exist = false, want true")
	}
	if oldValue.Name != "Alice" {
		t.Errorf("SafeUpset() old Name = %v, want Alice", oldValue.Name)
	}

	// Verify new value
	value, ok := typeMap.Get("key1")
	if !ok || value.Name != "Bob" {
		t.Errorf("Get() after SafeUpset Name = %v, want Bob", value.Name)
	}

	// Test SafeUpset with non-existent key
	_, exist, _ = typeMap.SafeUpset("key2", TestStruct{Name: "Charlie", Age: 35})
	if exist {
		t.Error("SafeUpset() on non-existent key should return exist=false")
	}
}

func TestRedisNumberMap_SetGet(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	numberMap := tr.Redis.NewNumberMap("testmap")

	// Test Set
	err := numberMap.Set("score1", 100)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Test Get
	value, ok := numberMap.Get("score1")
	if !ok {
		t.Fatal("Get() returned false")
	}
	if value != 100 {
		t.Errorf("Get() = %v, want 100", value)
	}

	// Test Get non-existent key
	_, ok = numberMap.Get("nonexistent")
	if ok {
		t.Error("Get() non-existent key should return false")
	}
}

func TestRedisNumberMap_Increment(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	numberMap := tr.Redis.NewNumberMap("testmap")

	numberMap.Set("score1", 100)

	// Test Increment
	value, err := numberMap.Increment("score1", 10)
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}
	if value != 110 {
		t.Errorf("Increment() = %v, want 110", value)
	}

	// Test Increment on non-existent key
	value, err = numberMap.Increment("score2", 5)
	if err != nil {
		t.Fatalf("Increment() on non-existent key error = %v", err)
	}
	if value != 5 {
		t.Errorf("Increment() on non-existent key = %v, want 5", value)
	}
}

func TestRedisNumberMap_Decrement(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	numberMap := tr.Redis.NewNumberMap("testmap")

	numberMap.Set("score1", 100)

	// Test Decrement
	value, err := numberMap.Decrement("score1", 10)
	if err != nil {
		t.Fatalf("Decrement() error = %v", err)
	}
	if value != 90 {
		t.Errorf("Decrement() = %v, want 90", value)
	}
}

func TestRedisNumberMap_ToArray(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	numberMap := tr.Redis.NewNumberMap("testmap")

	numberMap.Set("score1", 100)
	numberMap.Set("score2", 200)

	items, err := numberMap.ToArray()
	if err != nil {
		t.Fatalf("ToArray() error = %v", err)
	}

	if len(items) != 2 {
		t.Errorf("ToArray() length = %v, want 2", len(items))
	}
}

func TestRedisNumberMap_Operations(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	numberMap := tr.Redis.NewNumberMap("testmap")

	numberMap.Set("score1", 100)

	// Test Length
	if numberMap.Length() != 1 {
		t.Errorf("Length() = %v, want 1", numberMap.Length())
	}

	// Test IsEmpty
	if numberMap.IsEmpty() {
		t.Error("IsEmpty() = true, want false")
	}

	// Test Exists
	if !numberMap.Exists("score1") {
		t.Error("Exists() = false, want true")
	}

	// Test Delete
	err := numberMap.Delete("score1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if numberMap.Exists("score1") {
		t.Error("score1 should not exist after Delete()")
	}

	// Test Clear
	numberMap.Set("score2", 200)
	err = numberMap.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if !numberMap.IsEmpty() {
		t.Error("Map should be empty after Clear()")
	}
}
