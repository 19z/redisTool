package redisTool

import (
	"testing"
)

func TestRedisList_PushPop(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	list := tr.Redis.NewList("testlist")

	// Test Push
	err := list.Push("item1")
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	err = list.Push("item2")
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	// Test Length
	length := list.Length()
	if length != 2 {
		t.Errorf("Length() = %v, want 2", length)
	}

	// Test Pop
	value, ok := list.Pop()
	if !ok {
		t.Fatal("Pop() returned false")
	}
	if value != "item2" {
		t.Errorf("Pop() = %v, want item2", value)
	}

	value, ok = list.Pop()
	if !ok {
		t.Fatal("Pop() returned false")
	}
	if value != "item1" {
		t.Errorf("Pop() = %v, want item1", value)
	}

	// Test Pop on empty list
	_, ok = list.Pop()
	if ok {
		t.Error("Pop() on empty list should return false")
	}
}

func TestRedisList_ShiftUnshift(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	list := tr.Redis.NewList("testlist")

	// Test Unshift
	err := list.Unshift("item1")
	if err != nil {
		t.Fatalf("Unshift() error = %v", err)
	}

	err = list.Unshift("item2")
	if err != nil {
		t.Fatalf("Unshift() error = %v", err)
	}

	// Test Shift
	value, ok := list.Shift()
	if !ok {
		t.Fatal("Shift() returned false")
	}
	if value != "item2" {
		t.Errorf("Shift() = %v, want item2", value)
	}

	value, ok = list.Shift()
	if !ok {
		t.Fatal("Shift() returned false")
	}
	if value != "item1" {
		t.Errorf("Shift() = %v, want item1", value)
	}

	// Test Shift on empty list
	_, ok = list.Shift()
	if ok {
		t.Error("Shift() on empty list should return false")
	}
}

func TestRedisList_Index(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	list := tr.Redis.NewList("testlist")

	list.Push("item0")
	list.Push("item1")
	list.Push("item2")

	// Test Index
	value, ok := list.Index(0)
	if !ok {
		t.Fatal("Index(0) returned false")
	}
	if value != "item0" {
		t.Errorf("Index(0) = %v, want item0", value)
	}

	value, ok = list.Index(2)
	if !ok {
		t.Fatal("Index(2) returned false")
	}
	if value != "item2" {
		t.Errorf("Index(2) = %v, want item2", value)
	}

	// Test Index out of range
	_, ok = list.Index(10)
	if ok {
		t.Error("Index(10) should return false")
	}
}

func TestRedisList_SetAndGet(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	list := tr.Redis.NewList("testlist")

	list.Push("item0")
	list.Push("item1")
	list.Push("item2")

	// Test Set
	err := list.Set(1, "updated")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	value, ok := list.Index(1)
	if !ok || value != "updated" {
		t.Errorf("Index(1) after Set = %v, want updated", value)
	}

	// Test Get
	items, err := list.Get(0, -1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(items) != 3 {
		t.Errorf("Get() length = %v, want 3", len(items))
	}

	// Test Get range
	items, err = list.Get(0, 1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Get() range length = %v, want 2", len(items))
	}
}

func TestRedisList_Delete(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	list := tr.Redis.NewList("testlist")

	list.Push("item1")
	list.Push("item2")
	list.Push("item1")
	list.Push("item3")

	// Test DeleteValue
	err := list.DeleteValue("item1", 1)
	if err != nil {
		t.Fatalf("DeleteValue() error = %v", err)
	}

	length := list.Length()
	if length != 3 {
		t.Errorf("Length() after DeleteValue = %v, want 3", length)
	}

	// Test DeleteIndex
	err = list.DeleteIndex(0)
	if err != nil {
		t.Fatalf("DeleteIndex() error = %v", err)
	}

	// Test DeleteRange (keep only index 0)
	list.Push("item4")
	list.Push("item5")
	err = list.DeleteRange(0, 0)
	if err != nil {
		t.Fatalf("DeleteRange() error = %v", err)
	}

	length = list.Length()
	if length != 1 {
		t.Errorf("Length() after DeleteRange = %v, want 1", length)
	}
}

func TestRedisList_IsEmpty(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	list := tr.Redis.NewList("testlist")

	if !list.IsEmpty() {
		t.Error("IsEmpty() = false, want true")
	}

	if list.IsNotEmpty() {
		t.Error("IsNotEmpty() = true, want false")
	}

	list.Push("item1")

	if list.IsEmpty() {
		t.Error("IsEmpty() = true, want false")
	}

	if !list.IsNotEmpty() {
		t.Error("IsNotEmpty() = false, want true")
	}
}

func TestRedisList_Clear(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	list := tr.Redis.NewList("testlist")

	list.Push("item1")
	list.Push("item2")
	list.Push("item3")

	err := list.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if !list.IsEmpty() {
		t.Error("List should be empty after Clear()")
	}

	if list.Exists() {
		t.Error("List should not exist after Clear()")
	}
}

func TestRedisList_Iterator(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	list := tr.Redis.NewList("testlist")

	for i := 0; i < 10; i++ {
		list.Push(i)
	}

	count := 0
	for range list.Iterator(3) {
		count++
	}

	if count != 10 {
		t.Errorf("Iterator count = %v, want 10", count)
	}
}

func TestRedisTypeList_PushPop(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeList := NewTypeList[TestStruct]("testlist", tr.Redis)

	// Test Push
	err := typeList.Push(TestStruct{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	err = typeList.Push(TestStruct{Name: "Bob", Age: 25})
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	// Test Length
	length := typeList.Length()
	if length != 2 {
		t.Errorf("Length() = %v, want 2", length)
	}

	// Test Pop
	value, ok := typeList.Pop()
	if !ok {
		t.Fatal("Pop() returned false")
	}
	if value.Name != "Bob" {
		t.Errorf("Pop() Name = %v, want Bob", value.Name)
	}

	value, ok = typeList.Pop()
	if !ok {
		t.Fatal("Pop() returned false")
	}
	if value.Name != "Alice" {
		t.Errorf("Pop() Name = %v, want Alice", value.Name)
	}

	// Test Pop on empty list
	_, ok = typeList.Pop()
	if ok {
		t.Error("Pop() on empty list should return false")
	}
}

func TestRedisTypeList_ShiftUnshift(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeList := NewTypeList[TestStruct]("testlist", tr.Redis)

	// Test Unshift
	err := typeList.Unshift(TestStruct{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("Unshift() error = %v", err)
	}

	// Test Shift
	value, ok := typeList.Shift()
	if !ok {
		t.Fatal("Shift() returned false")
	}
	if value.Name != "Alice" {
		t.Errorf("Shift() Name = %v, want Alice", value.Name)
	}
}

func TestRedisTypeList_Index(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeList := NewTypeList[TestStruct]("testlist", tr.Redis)

	typeList.Push(TestStruct{Name: "Alice", Age: 30})
	typeList.Push(TestStruct{Name: "Bob", Age: 25})

	// Test Index
	value, ok := typeList.Index(0)
	if !ok {
		t.Fatal("Index(0) returned false")
	}
	if value.Name != "Alice" {
		t.Errorf("Index(0) Name = %v, want Alice", value.Name)
	}

	// Test Index out of range
	_, ok = typeList.Index(10)
	if ok {
		t.Error("Index(10) should return false")
	}
}

func TestRedisTypeList_SetAndGet(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeList := NewTypeList[TestStruct]("testlist", tr.Redis)

	typeList.Push(TestStruct{Name: "Alice", Age: 30})
	typeList.Push(TestStruct{Name: "Bob", Age: 25})

	// Test Set
	err := typeList.Set(0, TestStruct{Name: "Charlie", Age: 35})
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	value, ok := typeList.Index(0)
	if !ok || value.Name != "Charlie" {
		t.Errorf("Index(0) after Set Name = %v, want Charlie", value.Name)
	}

	// Test Get
	items, err := typeList.Get(0, -1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Get() length = %v, want 2", len(items))
	}
}

func TestRedisTypeList_Delete(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeList := NewTypeList[TestStruct]("testlist", tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	typeList.Push(s1)
	typeList.Push(TestStruct{Name: "Bob", Age: 25})

	// Test DeleteValue
	err := typeList.DeleteValue(s1, 1)
	if err != nil {
		t.Fatalf("DeleteValue() error = %v", err)
	}

	// Test DeleteIndex
	err = typeList.DeleteIndex(0)
	if err != nil {
		t.Fatalf("DeleteIndex() error = %v", err)
	}
}

func TestRedisTypeList_Iterator(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeList := NewTypeList[TestStruct]("testlist", tr.Redis)

	for i := 0; i < 5; i++ {
		typeList.Push(TestStruct{Name: "User", Age: 20 + i})
	}

	count := 0
	for item := range typeList.Iterator(2) {
		if item.Name != "User" {
			t.Errorf("Iterator item Name = %v, want User", item.Name)
		}
		count++
	}

	if count != 5 {
		t.Errorf("Iterator count = %v, want 5", count)
	}
}

func TestRedisTypeList_SafeUpset(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeList := NewTypeList[TestStruct]("testlist", tr.Redis)

	typeList.Push(TestStruct{Name: "Alice", Age: 30})

	// Test SafeUpset
	oldValue, exist, err := typeList.SafeUpset(0, TestStruct{Name: "Bob", Age: 25})
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
	value, ok := typeList.Index(0)
	if !ok || value.Name != "Bob" {
		t.Errorf("Index(0) after SafeUpset Name = %v, want Bob", value.Name)
	}

	// Test SafeUpset on non-existent index
	_, exist, err = typeList.SafeUpset(10, TestStruct{Name: "Charlie", Age: 35})
	if exist {
		t.Error("SafeUpset() on non-existent index should return exist=false")
	}
}

func TestRedisTypeList_Clear(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	typeList := NewTypeList[TestStruct]("testlist", tr.Redis)

	typeList.Push(TestStruct{Name: "Alice", Age: 30})
	typeList.Push(TestStruct{Name: "Bob", Age: 25})

	err := typeList.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if !typeList.IsEmpty() {
		t.Error("List should be empty after Clear()")
	}
}
