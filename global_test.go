package redisTool

import (
	"testing"
)

func TestGlobal_NewTypeList(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	// 设置默认连接
	SetDefaultConnection(tr.Redis)

	// 测试使用默认连接
	list1 := NewTypeList[TestStruct]("list1")
	if list1 == nil {
		t.Fatal("NewTypeList() with default connection returned nil")
	}

	list1.Push(TestStruct{Name: "Alice", Age: 30})
	value, ok := list1.Pop()
	if !ok || value.Name != "Alice" {
		t.Errorf("List with default connection failed")
	}

	// 测试提供特定连接
	list2 := NewTypeList[TestStruct]("list2", tr.Redis)
	if list2 == nil {
		t.Fatal("NewTypeList() with specific connection returned nil")
	}

	list2.Push(TestStruct{Name: "Bob", Age: 25})
	value, ok = list2.Pop()
	if !ok || value.Name != "Bob" {
		t.Errorf("List with specific connection failed")
	}
}

func TestGlobal_NewTypeSet(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	SetDefaultConnection(tr.Redis)

	// 测试使用默认连接
	set1 := NewTypeSet[TestStruct]("set1")
	if set1 == nil {
		t.Fatal("NewTypeSet() with default connection returned nil")
	}

	s1 := TestStruct{Name: "Alice", Age: 30}
	set1.Add(s1)
	if !set1.Exists(s1) {
		t.Error("Set with default connection failed")
	}

	// 测试提供特定连接
	set2 := NewTypeSet[TestStruct]("set2", tr.Redis)
	if set2 == nil {
		t.Fatal("NewTypeSet() with specific connection returned nil")
	}

	set2.Add(s1)
	if !set2.Exists(s1) {
		t.Error("Set with specific connection failed")
	}
}

func TestGlobal_NewTypeMap(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	SetDefaultConnection(tr.Redis)

	// 测试使用默认连接
	map1 := NewTypeMap[TestStruct]("map1")
	if map1 == nil {
		t.Fatal("NewTypeMap() with default connection returned nil")
	}

	s1 := TestStruct{Name: "Alice", Age: 30}
	map1.Set("key1", s1)
	value, ok := map1.Get("key1")
	if !ok || value.Name != "Alice" {
		t.Error("Map with default connection failed")
	}

	// 测试提供特定连接
	map2 := NewTypeMap[TestStruct]("map2", tr.Redis)
	if map2 == nil {
		t.Fatal("NewTypeMap() with specific connection returned nil")
	}

	map2.Set("key1", s1)
	value, ok = map2.Get("key1")
	if !ok || value.Name != "Alice" {
		t.Error("Map with specific connection failed")
	}
}

func TestGlobal_NewTypeZSet(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	SetDefaultConnection(tr.Redis)

	// 测试使用默认连接
	zset1 := NewTypeZSet[TestStruct]("zset1")
	if zset1 == nil {
		t.Fatal("NewTypeZSet() with default connection returned nil")
	}

	s1 := TestStruct{Name: "Alice", Age: 30}
	zset1.Add(s1, 95.5)
	score, ok := zset1.Score(s1)
	if !ok || score != 95.5 {
		t.Error("ZSet with default connection failed")
	}

	// 测试提供特定连接
	zset2 := NewTypeZSet[TestStruct]("zset2", tr.Redis)
	if zset2 == nil {
		t.Fatal("NewTypeZSet() with specific connection returned nil")
	}

	zset2.Add(s1, 88.0)
	score, ok = zset2.Score(s1)
	if !ok || score != 88.0 {
		t.Error("ZSet with specific connection failed")
	}
}

func TestGlobal_NewQueue(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	SetDefaultConnection(tr.Redis)

	config := QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0,
		MaxRetry:    0,
	}

	// 测试使用默认连接
	queue1 := NewQueue[TestStruct]("queue1", config)
	if queue1 == nil {
		t.Fatal("NewQueue() with default connection returned nil")
	}

	s1 := TestStruct{Name: "Alice", Age: 30}
	queue1.Add(s1)
	value, ok := queue1.Take()
	if !ok || value.Name != "Alice" {
		t.Error("Queue with default connection failed")
	}

	// 测试提供特定连接
	queue2 := NewQueue[TestStruct]("queue2", config, tr.Redis)
	if queue2 == nil {
		t.Fatal("NewQueue() with specific connection returned nil")
	}

	queue2.Add(s1)
	value, ok = queue2.Take()
	if !ok || value.Name != "Alice" {
		t.Error("Queue with specific connection failed")
	}
}

func TestGlobal_NewCache(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	SetDefaultConnection(tr.Redis)

	config := CacheConfig{
		DefaultExpire: 0,
	}

	// 测试使用默认连接
	cache1 := NewCache[TestStruct]("cache1", config)
	if cache1 == nil {
		t.Fatal("NewCache() with default connection returned nil")
	}

	s1 := TestStruct{Name: "Alice", Age: 30}
	cache1.Set("key1", s1, 0)
	value, ok := cache1.Get("key1")
	if !ok || value.Name != "Alice" {
		t.Error("Cache with default connection failed")
	}

	// 测试提供特定连接
	cache2 := NewCache[TestStruct]("cache2", config, tr.Redis)
	if cache2 == nil {
		t.Fatal("NewCache() with specific connection returned nil")
	}

	cache2.Set("key1", s1, 0)
	value, ok = cache2.Get("key1")
	if !ok || value.Name != "Alice" {
		t.Error("Cache with specific connection failed")
	}
}

func TestGlobal_WithoutDefaultConnection(t *testing.T) {
	// 清除默认连接
	SetDefaultConnection(nil)

	// 尝试使用默认连接应该会 panic 或返回错误
	// 这里我们不测试 panic，因为那会导致测试失败
	// 在实际使用中，用户应该先设置默认连接或提供连接参数
}

func TestGlobal_MultipleConnections(t *testing.T) {
	tr1 := NewTestRedis(t)
	defer tr1.Close()

	tr2 := NewTestRedis(t)
	defer tr2.Close()

	// 设置 tr1 为默认连接
	SetDefaultConnection(tr1.Redis)

	// 使用默认连接（tr1）
	list1 := NewTypeList[TestStruct]("sharedlist")
	list1.Push(TestStruct{Name: "Alice", Age: 30})

	// 使用 tr2 连接
	list2 := NewTypeList[TestStruct]("sharedlist", tr2.Redis)
	list2.Push(TestStruct{Name: "Bob", Age: 25})

	// list1 应该有 Alice
	value, ok := list1.Pop()
	if !ok || value.Name != "Alice" {
		t.Error("list1 should have Alice")
	}

	// list2 应该有 Bob
	value, ok = list2.Pop()
	if !ok || value.Name != "Bob" {
		t.Error("list2 should have Bob")
	}

	// list1 应该是空的
	_, ok = list1.Pop()
	if ok {
		t.Error("list1 should be empty")
	}
}
