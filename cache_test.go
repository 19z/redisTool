package redisTool

import (
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	// Test Set
	err := cache.Set("key1", s1, time.Minute*5)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Test Get
	value, ok := cache.Get("key1")
	if !ok {
		t.Fatal("Get() returned false")
	}
	if value.Name != "Alice" {
		t.Errorf("Get() Name = %v, want Alice", value.Name)
	}

	// Test Get non-existent key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Get() non-existent key should return false")
	}
}

func TestCache_SetWithDefaultExpire(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	// Test Set with 0 expire (should use default)
	err := cache.Set("key1", s1, 0)
	if err != nil {
		t.Fatalf("Set() with default expire error = %v", err)
	}

	// Verify TTL is set
	ttl, ok := cache.GetTTL("key1")
	if !ok {
		t.Error("GetTTL() should return true for key with TTL")
	}
	if ttl <= 0 {
		t.Errorf("GetTTL() = %v, want > 0", ttl)
	}
}

func TestCache_Delete(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	cache.Set("key1", s1, time.Minute*5)
	cache.Set("key2", s1, time.Minute*5)

	// Test Delete
	err := cache.Delete("key1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if cache.Exists("key1") {
		t.Error("key1 should not exist after Delete()")
	}

	// Test Delete multiple keys
	err = cache.Delete("key2", "nonexistent")
	if err != nil {
		t.Fatalf("Delete() multiple keys error = %v", err)
	}

	// Test Delete with no keys
	err = cache.Delete()
	if err != nil {
		t.Fatalf("Delete() with no keys error = %v", err)
	}
}

func TestCache_Exists(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	if cache.Exists("key1") {
		t.Error("Exists() = true, want false")
	}

	s1 := TestStruct{Name: "Alice", Age: 30}
	cache.Set("key1", s1, time.Minute*5)

	if !cache.Exists("key1") {
		t.Error("Exists() = false, want true")
	}
}

func TestCache_Clear(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	cache.Set("key1", s1, time.Minute*5)
	cache.Set("key2", s1, time.Minute*5)

	// Test Clear
	err := cache.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	length := cache.Length()
	if length != 0 {
		t.Errorf("Length() after Clear = %v, want 0", length)
	}
}

func TestCache_Length(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	if cache.Length() != 0 {
		t.Errorf("Length() = %v, want 0", cache.Length())
	}

	s1 := TestStruct{Name: "Alice", Age: 30}
	cache.Set("key1", s1, time.Minute*5)
	cache.Set("key2", s1, time.Minute*5)

	if cache.Length() != 2 {
		t.Errorf("Length() = %v, want 2", cache.Length())
	}
}

func TestCache_Keys(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	cache.Set("key1", s1, time.Minute*5)
	cache.Set("key2", s1, time.Minute*5)

	keys, err := cache.Keys()
	if err != nil {
		t.Fatalf("Keys() error = %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Keys() length = %v, want 2", len(keys))
	}
}

func TestCache_GetOrSet(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	factoryCalled := false
	factory := func(key string) (TestStruct, time.Duration) {
		factoryCalled = true
		return TestStruct{Name: "Alice", Age: 30}, time.Minute * 5
	}

	// First call should invoke factory
	value := cache.GetOrSet("key1", factory)
	if !factoryCalled {
		t.Error("Factory should have been called")
	}
	if value.Name != "Alice" {
		t.Errorf("GetOrSet() Name = %v, want Alice", value.Name)
	}

	// Second call should not invoke factory
	factoryCalled = false
	value = cache.GetOrSet("key1", factory)
	if factoryCalled {
		t.Error("Factory should not have been called on cache hit")
	}
	if value.Name != "Alice" {
		t.Errorf("GetOrSet() Name = %v, want Alice", value.Name)
	}
}

func TestCache_GetTTL(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	cache.Set("key1", s1, time.Minute*5)

	// Test GetTTL
	ttl, ok := cache.GetTTL("key1")
	if !ok {
		t.Fatal("GetTTL() returned false")
	}
	if ttl <= 0 || ttl > time.Minute*5 {
		t.Errorf("GetTTL() = %v, want between 0 and 5 minutes", ttl)
	}

	// Test GetTTL on non-existent key
	_, ok = cache.GetTTL("nonexistent")
	if ok {
		t.Error("GetTTL() on non-existent key should return false")
	}
}

func TestCache_SetTTL(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	cache.Set("key1", s1, time.Minute*5)

	// Test SetTTL
	err := cache.SetTTL("key1", time.Minute*10)
	if err != nil {
		t.Fatalf("SetTTL() error = %v", err)
	}

	ttl, ok := cache.GetTTL("key1")
	if !ok {
		t.Fatal("GetTTL() returned false after SetTTL")
	}
	if ttl <= time.Minute*5 {
		t.Errorf("GetTTL() after SetTTL = %v, want > 5 minutes", ttl)
	}

	// Test SetTTL on non-existent key (should not error)
	err = cache.SetTTL("nonexistent", time.Minute*5)
	if err != nil {
		t.Fatalf("SetTTL() on non-existent key error = %v", err)
	}
}

func TestCache_ClearExpired(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	
	// Set with very short expire
	cache.Set("key1", s1, time.Millisecond*100)
	// Set with long expire
	cache.Set("key2", s1, time.Minute*10)

	// Fast forward time to expire key1
	tr.FastForward(1)

	// Test ClearExpired
	err := cache.ClearExpired()
	if err != nil {
		t.Fatalf("ClearExpired() error = %v", err)
	}

	// key1 should be gone
	if cache.Exists("key1") {
		t.Error("key1 should not exist after ClearExpired()")
	}

	// key2 should still exist
	if !cache.Exists("key2") {
		t.Error("key2 should exist after ClearExpired()")
	}
}

func TestCache_Expiration(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	
	// Set with very short expire
	cache.Set("key1", s1, time.Millisecond*100)

	// Verify it exists
	if !cache.Exists("key1") {
		t.Error("key1 should exist before expiration")
	}

	// Fast forward time
	tr.FastForward(1)

	// Now it should be expired and Get should return false
	_, ok := cache.Get("key1")
	if ok {
		t.Error("Get() on expired key should return false")
	}

	// Exists should also return false
	if cache.Exists("key1") {
		t.Error("Exists() on expired key should return false")
	}
}

func TestCache_StartAutoCleanup(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	
	// Set with very short expire
	cache.Set("key1", s1, time.Millisecond*100)
	
	// Start auto cleanup with short interval
	cache.StartAutoCleanup(time.Millisecond * 50)

	// Fast forward time
	tr.FastForward(1)

	// Wait a bit for cleanup to run
	time.Sleep(time.Millisecond * 200)

	// key1 should be gone
	if cache.Exists("key1") {
		t.Error("key1 should not exist after auto cleanup")
	}
}

func TestCache_NoExpire(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	cache := NewCache[TestStruct]("testcache", CacheConfig{
		DefaultExpire: 0, // No default expire
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}
	
	// Set without expire
	cache.Set("key1", s1, 0)

	// Verify it exists
	if !cache.Exists("key1") {
		t.Error("key1 should exist")
	}

	// Fast forward time
	tr.FastForward(3600) // 1 hour

	// Should still exist
	value, ok := cache.Get("key1")
	if !ok {
		t.Error("Get() should return true for key without expiration")
	}
	if value.Name != "Alice" {
		t.Errorf("Get() Name = %v, want Alice", value.Name)
	}
}
