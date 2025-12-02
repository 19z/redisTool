package redisTool

import (
	"errors"
	"testing"
	"time"
)

func TestQueue_AddTake(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0, // Non-blocking
		MaxRetry:    0,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	// Test Add
	err := queue.Add(s1)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Test Length
	length := queue.Length()
	if length != 1 {
		t.Errorf("Length() = %v, want 1", length)
	}

	// Test Take
	value, ok := queue.Take()
	if !ok {
		t.Fatal("Take() returned false")
	}
	if value.Name != "Alice" {
		t.Errorf("Take() Name = %v, want Alice", value.Name)
	}

	// Test Take on empty queue
	_, ok = queue.Take()
	if ok {
		t.Error("Take() on empty queue should return false")
	}
}

func TestQueue_AddDelayed(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0,
		MaxRetry:    0,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	// Test AddDelayed
	err := queue.AddDelayed(s1, time.Millisecond*100)
	if err != nil {
		t.Fatalf("AddDelayed() error = %v", err)
	}

	// Check delayed queue length
	delayedLength := queue.DelayedLength()
	if delayedLength != 1 {
		t.Errorf("DelayedLength() = %v, want 1", delayedLength)
	}

	// Immediately take should fail (not yet ready)
	_, ok := queue.Take()
	if ok {
		t.Error("Take() should return false before delay expires")
	}

	// Fast forward time
	tr.FastForward(1)

	// Now take should succeed
	value, ok := queue.Take()
	if !ok {
		t.Fatal("Take() after delay should succeed")
	}
	if value.Name != "Alice" {
		t.Errorf("Take() Name = %v, want Alice", value.Name)
	}
}

func TestQueue_Complete(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0,
		MaxRetry:    3,
	})

	s1 := TestStruct{Name: "Alice", Age: 30}

	queue.Add(s1)

	value, ok := queue.Take()
	if !ok {
		t.Fatal("Take() returned false")
	}

	// Test Complete
	err := queue.Complete(value)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	// Check processing queue is empty
	processingLength := queue.ProcessingLength()
	if processingLength != 0 {
		t.Errorf("ProcessingLength() after Complete = %v, want 0", processingLength)
	}
}

func TestQueue_Fail(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	errorHandlerCalled := false
	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0,
		MaxRetry:    3,
		ErrorHandler: func(value interface{}, err error, storage func(value interface{})) time.Duration {
			errorHandlerCalled = true
			return time.Millisecond * 100 // Retry after 100ms
		},
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	queue.Add(s1)

	value, ok := queue.Take()
	if !ok {
		t.Fatal("Take() returned false")
	}

	// Test Fail with error handler
	err := queue.Fail(value, errors.New("test error"))
	if err != nil {
		t.Fatalf("Fail() error = %v", err)
	}

	if !errorHandlerCalled {
		t.Error("ErrorHandler should have been called")
	}
}

func TestQueue_FailWithoutErrorHandler(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0,
		MaxRetry:    3,
		ErrorHandler: nil,
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	queue.Add(s1)

	value, ok := queue.Take()
	if !ok {
		t.Fatal("Take() returned false")
	}

	// Test Fail without error handler
	err := queue.Fail(value, errors.New("test error"))
	if err != nil {
		t.Fatalf("Fail() error = %v", err)
	}
}

func TestQueue_FailNoRetry(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0,
		MaxRetry:    3,
		ErrorHandler: func(value interface{}, err error, storage func(value interface{})) time.Duration {
			return -1 // Don't retry
		},
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	queue.Add(s1)

	value, ok := queue.Take()
	if !ok {
		t.Fatal("Take() returned false")
	}

	// Test Fail with no retry
	err := queue.Fail(value, errors.New("test error"))
	if err != nil {
		t.Fatalf("Fail() error = %v", err)
	}

	// Processing queue should be empty
	processingLength := queue.ProcessingLength()
	if processingLength != 0 {
		t.Errorf("ProcessingLength() after Fail with no retry = %v, want 0", processingLength)
	}
}

func TestQueue_FailImmediateRetry(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0,
		MaxRetry:    3,
		ErrorHandler: func(value interface{}, err error, storage func(value interface{})) time.Duration {
			return 0 // Immediate retry
		},
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	queue.Add(s1)

	value, ok := queue.Take()
	if !ok {
		t.Fatal("Take() returned false")
	}

	// Test Fail with immediate retry
	err := queue.Fail(value, errors.New("test error"))
	if err != nil {
		t.Fatalf("Fail() error = %v", err)
	}

	// Queue should have the item again
	length := queue.Length()
	if length != 1 {
		t.Errorf("Length() after Fail with immediate retry = %v, want 1", length)
	}
}

func TestQueue_MaxLength(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   2,
		MaxWaitTime: 0,
		MaxRetry:    0,
	}, tr.Redis)

	// Add items up to max length
	err := queue.Add(TestStruct{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	err = queue.Add(TestStruct{Name: "Bob", Age: 25})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Try to add one more (should fail)
	err = queue.Add(TestStruct{Name: "Charlie", Age: 35})
	if err == nil {
		t.Error("Add() should error when queue is full")
	}
}

func TestQueue_Clear(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0,
		MaxRetry:    0,
	}, tr.Redis)

	queue.Add(TestStruct{Name: "Alice", Age: 30})
	queue.Add(TestStruct{Name: "Bob", Age: 25})

	// Test Clear
	err := queue.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	length := queue.Length()
	if length != 0 {
		t.Errorf("Length() after Clear = %v, want 0", length)
	}
}

func TestQueue_StartWorker(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	processed := make(chan TestStruct, 1)
	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: time.Millisecond * 100,
		MaxRetry:    0,
	}, tr.Redis)

	// Start worker
	queue.StartWorker(func(value TestStruct) error {
		processed <- value
		return nil
	})

	// Add item
	s1 := TestStruct{Name: "Alice", Age: 30}
	queue.Add(s1)

	// Wait for processing
	select {
	case result := <-processed:
		if result.Name != "Alice" {
			t.Errorf("Worker processed Name = %v, want Alice", result.Name)
		}
	case <-time.After(time.Second):
		t.Fatal("Worker did not process item within timeout")
	}
}

func TestQueue_StartWorkers(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	processed := make(chan TestStruct, 5)
	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: time.Millisecond * 100,
		MaxRetry:    0,
	}, tr.Redis)

	// Start multiple workers
	queue.StartWorkers(3, func(value TestStruct) error {
		processed <- value
		return nil
	})

	// Add items
	for i := 0; i < 5; i++ {
		queue.Add(TestStruct{Name: "User", Age: 20 + i})
	}

	// Wait for processing
	count := 0
	timeout := time.After(time.Second * 2)
	for count < 5 {
		select {
		case <-processed:
			count++
		case <-timeout:
			t.Fatalf("Workers processed %d items, want 5", count)
		}
	}
}

func TestQueue_ErrorHandlerStorage(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	storageCalled := false
	queue := NewQueue[TestStruct]("testqueue", QueueConfig{
		MaxLength:   0,
		MaxWaitTime: 0,
		MaxRetry:    3,
		ErrorHandler: func(value interface{}, err error, storage func(value interface{})) time.Duration {
			if s, ok := value.(TestStruct); ok {
				s.Age++
				storage(s)
				storageCalled = true
			}
			return time.Millisecond * 100
		},
	}, tr.Redis)

	s1 := TestStruct{Name: "Alice", Age: 30}

	queue.Add(s1)

	value, _ := queue.Take()

	// Test Fail with storage modification
	queue.Fail(value, errors.New("test error"))

	if !storageCalled {
		t.Error("Storage function should have been called")
	}
}
