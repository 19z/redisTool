package redisTool

import (
	"sync"
	"testing"
	"time"
)

func TestLock_LockUnlock(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock := tr.Redis.NewLock("testlock", LockConfig{
		WaitTime:           time.Second * 5,
		RetryTime:          time.Millisecond * 100,
		MaxGetLockWaitTime: time.Second * 3,
	})

	// Test Lock
	err := lock.Lock()
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}

	if !lock.IsLocked() {
		t.Error("IsLocked() = false, want true")
	}

	// Test Unlock
	err = lock.Unlock()
	if err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	if lock.IsLocked() {
		t.Error("IsLocked() = true, want false after Unlock()")
	}
}

func TestLock_TryLock(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock := tr.Redis.NewLock("testlock")

	// Test TryLock success
	ok := lock.TryLock()
	if !ok {
		t.Error("TryLock() = false, want true")
	}

	if !lock.IsLocked() {
		t.Error("IsLocked() = false, want true")
	}

	// Test TryLock failure (already locked)
	lock2 := tr.Redis.NewLock("testlock")
	ok = lock2.TryLock()
	if ok {
		t.Error("TryLock() on locked resource = true, want false")
	}

	// Unlock and try again
	lock.Unlock()

	ok = lock2.TryLock()
	if !ok {
		t.Error("TryLock() after unlock = false, want true")
	}

	lock2.Unlock()
}

func TestLock_LockTimeout(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock1 := tr.Redis.NewLock("testlock", LockConfig{
		WaitTime:           time.Second * 5,
		RetryTime:          time.Millisecond * 100,
		MaxGetLockWaitTime: time.Millisecond * 500,
	})

	// Acquire first lock
	err := lock1.Lock()
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}
	defer lock1.Unlock()

	// Try to acquire second lock (should timeout)
	lock2 := tr.Redis.NewLock("testlock", LockConfig{
		WaitTime:           time.Second * 5,
		RetryTime:          time.Millisecond * 100,
		MaxGetLockWaitTime: time.Millisecond * 500,
	})

	start := time.Now()
	err = lock2.Lock()
	duration := time.Since(start)

	if err == nil {
		t.Error("Lock() should error on timeout")
	}

	if duration < time.Millisecond*400 {
		t.Errorf("Lock() timeout duration = %v, want >= 400ms", duration)
	}
}

func TestLock_LockImmediateFail(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock1 := tr.Redis.NewLock("testlock")
	lock1.Lock()
	defer lock1.Unlock()

	// Test with MaxGetLockWaitTime = 0 (immediate fail)
	lock2 := tr.Redis.NewLock("testlock", LockConfig{
		WaitTime:           time.Second * 5,
		RetryTime:          time.Millisecond * 100,
		MaxGetLockWaitTime: 0,
	})

	err := lock2.Lock()
	if err == nil {
		t.Error("Lock() with MaxGetLockWaitTime=0 should error immediately")
	}
}

func TestLock_LockFunc(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock := tr.Redis.NewLock("testlock")

	executed := false
	err := lock.LockFunc(func() {
		executed = true
	})

	if err != nil {
		t.Fatalf("LockFunc() error = %v", err)
	}

	if !executed {
		t.Error("LockFunc() did not execute function")
	}

	if lock.IsLocked() {
		t.Error("Lock should be released after LockFunc()")
	}
}

func TestLock_TryLockFunc(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock := tr.Redis.NewLock("testlock")

	// Test TryLockFunc success
	executed := false
	ok := lock.TryLockFunc(func() {
		executed = true
	})

	if !ok {
		t.Error("TryLockFunc() = false, want true")
	}

	if !executed {
		t.Error("TryLockFunc() did not execute function")
	}

	if lock.IsLocked() {
		t.Error("Lock should be released after TryLockFunc()")
	}
}

func TestLock_TryLockFuncFail(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock1 := tr.Redis.NewLock("testlock")
	lock1.Lock()
	defer lock1.Unlock()

	// Test TryLockFunc failure
	lock2 := tr.Redis.NewLock("testlock")
	executed := false
	ok := lock2.TryLockFunc(func() {
		executed = true
	})

	if ok {
		t.Error("TryLockFunc() on locked resource = true, want false")
	}

	if executed {
		t.Error("TryLockFunc() should not execute function on failure")
	}
}

func TestLock_Refresh(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock := tr.Redis.NewLock("testlock", LockConfig{
		WaitTime: time.Second * 2,
	})

	err := lock.Lock()
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}
	defer lock.Unlock()

	// Test Refresh
	err = lock.Refresh()
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	// Test Refresh on unlocked lock
	lock2 := tr.Redis.NewLock("testlock2")
	err = lock2.Refresh()
	if err == nil {
		t.Error("Refresh() on unlocked lock should error")
	}
}

func TestLock_RefreshWrongToken(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock1 := tr.Redis.NewLock("testlock")
	lock1.Lock()

	// Manually unlock to simulate another instance holding the lock
	lock1.Unlock()

	// Another lock acquires it
	lock2 := tr.Redis.NewLock("testlock")
	lock2.Lock()
	defer lock2.Unlock()

	// Try to refresh with lock1 (wrong token)
	lock1.locked = true // Simulate still thinking it's locked
	err := lock1.Refresh()
	if err == nil {
		t.Error("Refresh() with wrong token should error")
	}
}

func TestLock_UnlockWrongToken(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock1 := tr.Redis.NewLock("testlock")
	lock1.Lock()
	lock1.Unlock()

	// Another lock acquires it
	lock2 := tr.Redis.NewLock("testlock")
	lock2.Lock()
	defer lock2.Unlock()

	// Try to unlock with lock1 (wrong token)
	lock1.locked = true // Simulate still thinking it's locked
	err := lock1.Unlock()
	if err == nil {
		t.Error("Unlock() with wrong token should error")
	}
}

func TestLock_UnlockNotLocked(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock := tr.Redis.NewLock("testlock")

	// Test Unlock without Lock
	err := lock.Unlock()
	if err != nil {
		t.Fatalf("Unlock() on unlocked lock error = %v", err)
	}
}

func TestLock_Concurrency(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	var counter int
	var wg sync.WaitGroup
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			lock := tr.Redis.NewLock("counterlock", LockConfig{
				WaitTime:           time.Second * 5,
				RetryTime:          time.Millisecond * 50,
				MaxGetLockWaitTime: time.Second * 3,
			})

			if err := lock.Lock(); err != nil {
				t.Errorf("Lock() error = %v", err)
				return
			}
			defer lock.Unlock()

			// Critical section
			temp := counter
			time.Sleep(time.Millisecond * 10)
			counter = temp + 1
		}()
	}

	wg.Wait()

	if counter != goroutines {
		t.Errorf("Counter = %v, want %v (race condition detected)", counter, goroutines)
	}
}

func TestLock_StartRefreshLoop(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock := tr.Redis.NewLock("testlock", LockConfig{
		WaitTime: time.Millisecond * 500,
	})

	err := lock.Lock()
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}

	// Start refresh loop
	stopCh := lock.StartRefreshLoop()

	// Wait for a few refresh cycles
	time.Sleep(time.Millisecond * 600)

	// Lock should still be held
	if !lock.IsLocked() {
		t.Error("Lock should still be held after refresh")
	}

	// Stop refresh loop
	close(stopCh)

	// Unlock
	lock.Unlock()
}

func TestLock_DefaultConfig(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	// Test with default config
	lock := tr.Redis.NewLock("testlock")

	if lock.config.WaitTime != time.Second*5 {
		t.Errorf("Default WaitTime = %v, want 5s", lock.config.WaitTime)
	}

	if lock.config.RetryTime != time.Second {
		t.Errorf("Default RetryTime = %v, want 1s", lock.config.RetryTime)
	}

	if lock.config.MaxGetLockWaitTime != time.Second*30 {
		t.Errorf("Default MaxGetLockWaitTime = %v, want 30s", lock.config.MaxGetLockWaitTime)
	}
}

func TestLock_PartialConfig(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	// Test with partial config
	lock := tr.Redis.NewLock("testlock", LockConfig{
		WaitTime: time.Second * 10,
	})

	if lock.config.WaitTime != time.Second*10 {
		t.Errorf("WaitTime = %v, want 10s", lock.config.WaitTime)
	}

	// Should use defaults for other fields
	if lock.config.RetryTime != time.Second {
		t.Errorf("RetryTime = %v, want 1s (default)", lock.config.RetryTime)
	}
}

func TestLock_AutoExpire(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lock1 := tr.Redis.NewLock("testlock", LockConfig{
		WaitTime: time.Millisecond * 200, // Very short wait time
	})

	err := lock1.Lock()
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}

	// Don't unlock, let it expire
	tr.FastForward(1)

	// Another lock should be able to acquire
	lock2 := tr.Redis.NewLock("testlock")
	ok := lock2.TryLock()
	if !ok {
		t.Error("TryLock() after expiration = false, want true")
	}
	lock2.Unlock()
}
