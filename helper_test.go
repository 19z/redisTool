package redisTool

import (
	"testing"
	"time"
)

func TestLastUseTime(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	lastTime := tr.Redis.LastUseTime("key1", false)
	if !lastTime.IsZero() {
		t.Errorf("LastUseTime() first call = %v, want zero", lastTime)
	}

	tr.Redis.LastUseTime("key1", true)
	lastTime = tr.Redis.LastUseTime("key1", false)
	if lastTime.IsZero() {
		t.Error("LastUseTime() after update should not be zero")
	}
}

func TestAcrossMinute(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	if !tr.Redis.AcrossMinute("task1") {
		t.Error("AcrossMinute() first call = false, want true")
	}

	if tr.Redis.AcrossMinute("task1") {
		t.Error("AcrossMinute() second call = true, want false")
	}

	tr.FastForward(61)
	if !tr.Redis.AcrossMinute("task1") {
		t.Error("AcrossMinute() after 61s = false, want true")
	}
}

func TestAcrossTime(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	if !tr.Redis.AcrossTime("task2", time.Second) {
		t.Error("AcrossTime() first call = false, want true")
	}

	if tr.Redis.AcrossTime("task2", time.Second) {
		t.Error("AcrossTime() second call = true, want false")
	}

	tr.FastForward(2)
	if !tr.Redis.AcrossTime("task2", time.Second) {
		t.Error("AcrossTime() after interval = false, want true")
	}
}

func TestCleanSafeTypeMap(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	oldTime := time.Now().Add(-time.Hour * 2)
	tr.Redis.SetLastUseTime("old", oldTime)

	err := tr.Redis.CleanSafeTypeMap(time.Hour)
	if err != nil {
		t.Fatalf("CleanSafeTypeMap() error = %v", err)
	}
}
