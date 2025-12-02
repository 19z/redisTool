package redisTool

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

// Lock 分布式锁
type Lock struct {
	redis  *Redis
	name   string
	token  string
	config LockConfig
	locked bool
}

// NewLock 创建分布式锁
func (r *Redis) NewLock(name string, config ...LockConfig) *Lock {
	cfg := LockConfig{
		WaitTime:           time.Second * 5,
		RetryTime:          time.Second,
		MaxGetLockWaitTime: time.Second * 30,
	}
	
	if len(config) > 0 {
		if config[0].WaitTime > 0 {
			cfg.WaitTime = config[0].WaitTime
		}
		if config[0].RetryTime > 0 {
			cfg.RetryTime = config[0].RetryTime
		}
		if config[0].MaxGetLockWaitTime >= 0 {
			cfg.MaxGetLockWaitTime = config[0].MaxGetLockWaitTime
		}
	}
	
	return &Lock{
		redis:  r,
		name:   r.CreateName(RedisTypeLock_, name),
		token:  uuid.New().String(),
		config: cfg,
		locked: false,
	}
}

// Lock 获取锁
func (l *Lock) Lock() error {
	startTime := time.Now()
	
	for {
		// 尝试获取锁
		if l.tryAcquire() {
			l.locked = true
			return nil
		}
		
		// 检查是否超时
		if l.config.MaxGetLockWaitTime > 0 && time.Since(startTime) >= l.config.MaxGetLockWaitTime {
			return fmt.Errorf("lock timeout: failed to acquire lock within %v", l.config.MaxGetLockWaitTime)
		}
		
		// 如果 MaxGetLockWaitTime 为 0，立即返回
		if l.config.MaxGetLockWaitTime == 0 {
			return fmt.Errorf("lock failed: unable to acquire lock")
		}
		
		// 等待后重试
		time.Sleep(l.config.RetryTime)
	}
}

// TryLock 尝试获取锁
func (l *Lock) TryLock() bool {
	if l.tryAcquire() {
		l.locked = true
		return true
	}
	return false
}

// Unlock 释放锁
func (l *Lock) Unlock() error {
	if !l.locked {
		return nil
	}
	
	// 使用 Lua 脚本确保只删除自己持有的锁
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	
	conn := l.redis.GetConn()
	defer conn.Close()
	
	luaScript := redis.NewScript(1, script)
	result, err := redis.Int(luaScript.Do(conn, l.name, l.token))
	if err != nil {
		return err
	}
	
	if result == 1 {
		l.locked = false
		return nil
	}
	
	return fmt.Errorf("unlock failed: lock not held by this instance")
}

// LockFunc 使用闭包简化锁的使用
func (l *Lock) LockFunc(fn func()) error {
	if err := l.Lock(); err != nil {
		return err
	}
	defer l.Unlock()
	
	fn()
	return nil
}

// TryLockFunc 尝试使用闭包简化锁的使用
func (l *Lock) TryLockFunc(fn func()) bool {
	if !l.TryLock() {
		return false
	}
	defer l.Unlock()
	
	fn()
	return true
}

// IsLocked 检查是否已锁定
func (l *Lock) IsLocked() bool {
	return l.locked
}

// Refresh 刷新锁的过期时间
func (l *Lock) Refresh() error {
	if !l.locked {
		return fmt.Errorf("lock not held")
	}
	
	// 使用 Lua 脚本确保只刷新自己持有的锁
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	
	conn := l.redis.GetConn()
	defer conn.Close()
	
	luaScript := redis.NewScript(1, script)
	result, err := redis.Int(luaScript.Do(conn, l.name, l.token, int(l.config.WaitTime.Milliseconds())))
	if err != nil {
		return err
	}
	
	if result == 0 {
		l.locked = false
		return fmt.Errorf("refresh failed: lock not held by this instance")
	}
	
	return nil
}

// tryAcquire 尝试获取锁
func (l *Lock) tryAcquire() bool {
	conn := l.redis.GetConn()
	defer conn.Close()
	
	// 使用 SET NX PX 命令原子性地设置锁
	result, err := redis.String(conn.Do("SET", l.name, l.token, "NX", "PX", int(l.config.WaitTime.Milliseconds())))
	if err != nil || result != "OK" {
		return false
	}
	
	return true
}

// StartRefreshLoop 启动自动刷新锁的循环
func (l *Lock) StartRefreshLoop() chan struct{} {
	stopCh := make(chan struct{})
	
	go func() {
		ticker := time.NewTicker(l.config.WaitTime / 2)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				if l.locked {
					if err := l.Refresh(); err != nil {
						return
					}
				}
			case <-stopCh:
				return
			}
		}
	}()
	
	return stopCh
}
