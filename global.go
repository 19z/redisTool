package redisTool

import "time"

// 全局泛型函数，用于创建类型化的 Redis 数据结构
// 这些函数可以使用默认连接或提供的连接

// NewTypeList 创建类型化列表（全局函数）
func NewTypeList[T any](name string, r ...*Redis) *RedisTypeList[T] {
	var conn *Redis
	if len(r) == 0 || r[0] == nil {
		conn = defaultConnection
	} else {
		conn = r[0]
	}
	return &RedisTypeList[T]{
		list: conn.NewList(name),
	}
}

// NewTypeSet 创建类型化集合（全局函数）
func NewTypeSet[T any](name string, r ...*Redis) *RedisTypeSet[T] {
	var conn *Redis
	if len(r) == 0 || r[0] == nil {
		conn = defaultConnection
	} else {
		conn = r[0]
	}
	return &RedisTypeSet[T]{
		set: conn.NewSet(name),
	}
}

// NewTypeMap 创建类型化哈希表（全局函数）
func NewTypeMap[T any](name string, r ...*Redis) *RedisTypeMap[T] {
	var conn *Redis
	if len(r) == 0 || r[0] == nil {
		conn = defaultConnection
	} else {
		conn = r[0]
	}
	return &RedisTypeMap[T]{
		rmap: conn.NewMap(name),
	}
}

// NewTypeZSet 创建类型化有序集合（全局函数）
func NewTypeZSet[T any](name string, r ...*Redis) *RedisTypeZSet[T] {
	var conn *Redis
	if len(r) == 0 || r[0] == nil {
		conn = defaultConnection
	} else {
		conn = r[0]
	}
	return &RedisTypeZSet[T]{
		zset: conn.NewZSet(name),
	}
}

// NewQueue 创建队列（全局函数）
func NewQueue[T any](name string, config QueueConfig, r ...*Redis) *Queue[T] {
	var conn *Redis
	if len(r) == 0 || r[0] == nil {
		conn = defaultConnection
	} else {
		conn = r[0]
	}
	baseName := conn.CreateName(RedisTypeQueue_, name)
	return &Queue[T]{
		redis:          conn,
		name:           baseName,
		delayedName:    baseName + ":delayed",
		processingName: baseName + ":processing",
		retryName:      baseName + ":retry",
		config:         config,
	}
}

// NewCache 创建缓存（全局函数）
func NewCache[T any](name string, config CacheConfig, r ...*Redis) *Cache[T] {
	var conn *Redis
	if len(r) == 0 || r[0] == nil {
		conn = defaultConnection
	} else {
		conn = r[0]
	}
	baseName := conn.CreateName(RedisTypeCache_, name)
	return &Cache[T]{
		redis:      conn,
		dataName:   baseName + ":data",
		expireName: baseName + ":expire",
		config:     config,
	}
}

// NewLock 创建分布式锁（全局函数）
func NewLock(name string, config ...LockConfig) *Lock {
	conn := defaultConnection
	return conn.NewLock(name, config...)
}

// LastUseTime 获取上次使用时间（全局函数）
func LastUseTime(key string, update bool) time.Time {
	conn := defaultConnection
	return conn.LastUseTime(key, update)
}

// AcrossMinute 是否跨越了分钟（全局函数）
func AcrossMinute(key string) bool {
	conn := defaultConnection
	return conn.AcrossMinute(key)
}

// AcrossSecond 是否跨越了秒（全局函数）
func AcrossSecond(key string) bool {
	conn := defaultConnection
	return conn.AcrossSecond(key)
}

// AcrossTime 是否跨越了指定的时间间隔（全局函数）
func AcrossTime(key string, duration time.Duration) bool {
	conn := defaultConnection
	return conn.AcrossTime(key, duration)
}

// SetLastUseTime 设置上次使用时间（全局函数）
func SetLastUseTime(key string, t time.Time) {
	conn := defaultConnection
	conn.SetLastUseTime(key, t)
}

// GetSafeTypeMap 获取安全类型映射（全局函数）
func GetSafeTypeMap() *RedisTypeMap[int64] {
	conn := defaultConnection
	return conn.GetSafeTypeMap()
}
