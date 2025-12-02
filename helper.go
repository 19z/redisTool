package redisTool

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// LastUseTime 获取上次使用时间
func (r *Redis) LastUseTime(key string, update bool) time.Time {
	safeTypeMapName := r.CreateName(RedisTypeSafeTypeMap_, r.config.SafeTypeMapName)
	
	if update {
		// 使用 Lua 脚本确保原子性：获取旧值并设置新值
		script := `
			local old_value = redis.call('HGET', KEYS[1], ARGV[1])
			redis.call('HSET', KEYS[1], ARGV[1], ARGV[2])
			return old_value
		`
		now := time.Now()
		result, err := r.Do("EVAL", script, 1, safeTypeMapName, key, now.UnixMilli())
		
		var lastTime time.Time
		if err == nil && result != nil {
			if lastTimeMs, err := redis.Int64(result, nil); err == nil && lastTimeMs > 0 {
				lastTime = time.UnixMilli(lastTimeMs)
			}
		}
		return lastTime
	} else {
		// 只读取，不需要原子性
		lastTimeMs, err := redis.Int64(r.Do("HGET", safeTypeMapName, key))
		var lastTime time.Time
		if err == nil && lastTimeMs > 0 {
			lastTime = time.UnixMilli(lastTimeMs)
		}
		return lastTime
	}
}

// AcrossMinute 是否跨越了分钟
func (r *Redis) AcrossMinute(key string) bool {
	return r.AcrossTime(key, time.Minute)
}

// AcrossSecond 是否跨越了秒
func (r *Redis) AcrossSecond(key string) bool {
	return r.AcrossTime(key, time.Second)
}

// AcrossTime 是否跨越了指定的时间间隔
func (r *Redis) AcrossTime(key string, duration time.Duration) bool {
	lastTime := r.LastUseTime(key, false)
	now := time.Now()
	
	// 如果是第一次调用，返回 true 并更新时间
	if lastTime.IsZero() {
		r.LastUseTime(key, true)
		return true
	}
	
	// 计算时间间隔
	lastInterval := lastTime.UnixMilli() / duration.Milliseconds()
	nowInterval := now.UnixMilli() / duration.Milliseconds()
	
	// 如果跨越了时间间隔，更新时间并返回 true
	if nowInterval > lastInterval {
		r.LastUseTime(key, true)
		return true
	}
	
	return false
}

// GetSafeTypeMap 获取安全类型映射
func (r *Redis) GetSafeTypeMap() *RedisTypeMap[int64] {
	return NewTypeMap[int64](r.config.SafeTypeMapName, r)
}

// CleanSafeTypeMap 清理安全类型映射中的过期数据
func (r *Redis) CleanSafeTypeMap(expireDuration time.Duration) error {
	safeTypeMap := r.GetSafeTypeMap()
	
	// 获取所有键值对
	all, err := safeTypeMap.ToArray()
	if err != nil {
		return err
	}
	
	now := time.Now().UnixMilli()
	keysToDelete := make([]string, 0)
	
	// 找出过期的键
	for key, lastTimeMs := range all {
		if now-lastTimeMs > expireDuration.Milliseconds() {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	// 删除过期的键
	if len(keysToDelete) > 0 {
		return safeTypeMap.Delete(keysToDelete...)
	}
	
	return nil
}

// SetLastUseTime 设置上次使用时间
func (r *Redis) SetLastUseTime(key string, t time.Time) error {
	safeTypeMapName := r.CreateName(RedisTypeSafeTypeMap_, r.config.SafeTypeMapName)
	_, err := r.Do("HSET", safeTypeMapName, key, t.UnixMilli())
	return err
}

// DeleteLastUseTime 删除上次使用时间
func (r *Redis) DeleteLastUseTime(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	
	safeTypeMapName := r.CreateName(RedisTypeSafeTypeMap_, r.config.SafeTypeMapName)
	
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, safeTypeMapName)
	for _, key := range keys {
		args = append(args, key)
	}
	
	_, err := r.Do("HDEL", args...)
	return err
}
