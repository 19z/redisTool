package redisTool

import (
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// TestRedis 测试用的 Redis 实例
type TestRedis struct {
	Redis      *Redis
	MiniRedis  *miniredis.Miniredis
	t          *testing.T
	useReal    bool
}

// NewTestRedis 创建测试用的 Redis 实例
// 通过环境变量 USE_REAL_REDIS=1 可以使用真实 Redis (localhost:16379)
// 通过环境变量 REDIS_ADDR 可以自定义 Redis 地址
func NewTestRedis(t *testing.T) *TestRedis {
	useReal := os.Getenv("USE_REAL_REDIS") == "1"
	redisAddr := os.Getenv("REDIS_ADDR")
	
	if redisAddr == "" {
		redisAddr = "localhost:16379"
	}

	if useReal {
		// 使用真实 Redis
		redis := Builder(redisAddr, "").
			Config(Config{
				Prefix: "test:",
			}).
			Build()

		// 测试连接
		_, err := redis.Do("PING")
		if err != nil {
			t.Fatalf("Failed to connect to real Redis at %s: %v", redisAddr, err)
		}

		// 清空测试数据
		_, _ = redis.Do("FLUSHDB")

		return &TestRedis{
			Redis:   redis,
			t:       t,
			useReal: true,
		}
	}

	// 使用 miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	redis := Builder(mr.Addr(), "").
		Config(Config{
			Prefix: "test:",
		}).
		Build()

	return &TestRedis{
		Redis:     redis,
		MiniRedis: mr,
		t:         t,
		useReal:   false,
	}
}

// Close 关闭测试 Redis
func (tr *TestRedis) Close() {
	if tr.useReal {
		// 清理测试数据
		_, _ = tr.Redis.Do("FLUSHDB")
		tr.Redis.Close()
	} else {
		tr.Redis.Close()
		if tr.MiniRedis != nil {
			tr.MiniRedis.Close()
		}
	}
}

// FlushAll 清空所有数据
func (tr *TestRedis) FlushAll() {
	if tr.useReal {
		_, _ = tr.Redis.Do("FLUSHDB")
	} else {
		tr.MiniRedis.FlushAll()
	}
}

// FastForward 快进时间（仅在 miniredis 模式下有效）
func (tr *TestRedis) FastForward(seconds int) {
	if tr.useReal {
		// 真实 Redis 无法快进时间，只能等待
		time.Sleep(time.Duration(seconds) * time.Second)
	} else {
		tr.MiniRedis.FastForward(time.Duration(seconds) * time.Second)
	}
}
