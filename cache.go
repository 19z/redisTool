package redisTool

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Cache 缓存
type Cache[T any] struct {
	redis      *Redis
	dataName   string
	expireName string
	config     CacheConfig
}


// Set 设置缓存
func (c *Cache[T]) Set(key string, value T, expire time.Duration) error {
	data, err := c.redis.Serialize(value)
	if err != nil {
		return err
	}
	
	conn := c.redis.GetConn()
	defer conn.Close()
	
	// 设置数据
	if _, err := conn.Do("HSET", c.dataName, key, data); err != nil {
		return err
	}
	
	// 设置过期时间
	if expire > 0 {
		expireTime := float64(time.Now().Add(expire).UnixMilli())
		if _, err := conn.Do("ZADD", c.expireName, expireTime, key); err != nil {
			return err
		}
	} else if c.config.DefaultExpire > 0 {
		expireTime := float64(time.Now().Add(c.config.DefaultExpire).UnixMilli())
		if _, err := conn.Do("ZADD", c.expireName, expireTime, key); err != nil {
			return err
		}
	}
	
	// 概率性清理过期数据（10% 概率）
	if rand.Intn(10) == 0 {
		// 使用 AcrossMinute 判断是否横跨分钟
		cleanupKey := c.dataName + ":cleanup"
		if c.redis.AcrossMinute(cleanupKey) {
			go c.ClearExpired()
		}
	}
	
	return nil
}

// Get 获取缓存
func (c *Cache[T]) Get(key string) (T, bool) {
	var zero T
	
	// 检查是否过期
	if c.isExpired(key) {
		c.Delete(key)
		return zero, false
	}
	
	data, err := redis.Bytes(c.redis.Do("HGET", c.dataName, key))
	if err != nil || len(data) == 0 {
		return zero, false
	}
	
	result := reflect.New(reflect.TypeOf(zero)).Interface()
	if err := c.redis.Deserialize(data, result); err != nil {
		return zero, false
	}
	
	return reflect.ValueOf(result).Elem().Interface().(T), true
}

// GetOrSet 获取或设置缓存
func (c *Cache[T]) GetOrSet(key string, factory func(key string) (T, time.Duration)) T {
	value, ok := c.Get(key)
	if ok {
		return value
	}
	
	value, expire := factory(key)
	c.Set(key, value, expire)
	return value
}

// Delete 删除缓存
func (c *Cache[T]) Delete(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	
	conn := c.redis.GetConn()
	defer conn.Close()
	
	// 删除数据
	dataArgs := make([]interface{}, 0, len(keys)+1)
	dataArgs = append(dataArgs, c.dataName)
	for _, key := range keys {
		dataArgs = append(dataArgs, key)
	}
	if _, err := conn.Do("HDEL", dataArgs...); err != nil {
		return err
	}
	
	// 删除过期时间
	expireArgs := make([]interface{}, 0, len(keys)+1)
	expireArgs = append(expireArgs, c.expireName)
	for _, key := range keys {
		expireArgs = append(expireArgs, key)
	}
	if _, err := conn.Do("ZREM", expireArgs...); err != nil {
		return err
	}
	
	return nil
}

// Exists 判断缓存是否存在
func (c *Cache[T]) Exists(key string) bool {
	if c.isExpired(key) {
		c.Delete(key)
		return false
	}
	
	exists, err := redis.Int(c.redis.Do("HEXISTS", c.dataName, key))
	if err != nil {
		return false
	}
	return exists == 1
}

// Clear 清空缓存
func (c *Cache[T]) Clear() error {
	conn := c.redis.GetConn()
	defer conn.Close()
	
	_, err := conn.Do("DEL", c.dataName, c.expireName)
	return err
}

// ClearExpired 清理过期的缓存
func (c *Cache[T]) ClearExpired() error {
	now := float64(time.Now().UnixMilli())
	
	// 获取过期的键
	keys, err := redis.Strings(c.redis.Do("ZRANGEBYSCORE", c.expireName, 0, now))
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		return c.Delete(keys...)
	}
	
	return nil
}

// Length 获取缓存数量
func (c *Cache[T]) Length() int {
	c.ClearExpired() // 清理过期缓存
	
	length, err := redis.Int(c.redis.Do("HLEN", c.dataName))
	if err != nil {
		return 0
	}
	return length
}

// Keys 获取所有键
func (c *Cache[T]) Keys() ([]string, error) {
	c.ClearExpired() // 清理过期缓存
	
	keys, err := redis.Strings(c.redis.Do("HKEYS", c.dataName))
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// GetTTL 获取剩余生存时间
func (c *Cache[T]) GetTTL(key string) (time.Duration, bool) {
	score, err := redis.Float64(c.redis.Do("ZSCORE", c.expireName, key))
	if err != nil {
		return 0, false
	}
	
	expireTime := time.UnixMilli(int64(score))
	ttl := time.Until(expireTime)
	
	if ttl <= 0 {
		return 0, false
	}
	
	return ttl, true
}

// SetTTL 设置生存时间
func (c *Cache[T]) SetTTL(key string, expire time.Duration) error {
	if !c.Exists(key) {
		return nil
	}
	
	expireTime := float64(time.Now().Add(expire).UnixMilli())
	_, err := c.redis.Do("ZADD", c.expireName, expireTime, key)
	return err
}

// isExpired 判断是否过期
func (c *Cache[T]) isExpired(key string) bool {
	score, err := redis.Float64(c.redis.Do("ZSCORE", c.expireName, key))
	if err != nil {
		return false
	}
	
	expireTime := time.UnixMilli(int64(score))
	return time.Now().After(expireTime)
}

