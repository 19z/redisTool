package redisTool

import "time"

// RedisType Redis 数据类型枚举
type RedisType int

const (
	RedisTypeString RedisType = iota
	RedisTypeList_
	RedisTypeSet_
	RedisTypeZSet_
	RedisTypeHash_
	RedisTypeQueue_
	RedisTypeCache_
	RedisTypeLock_
	RedisTypeSafeTypeMap_
)

// String 返回 RedisType 的字符串表示
func (rt RedisType) String() string {
	switch rt {
	case RedisTypeString:
		return "string"
	case RedisTypeList_:
		return "list"
	case RedisTypeSet_:
		return "set"
	case RedisTypeZSet_:
		return "zset"
	case RedisTypeHash_:
		return "hash"
	case RedisTypeQueue_:
		return "queue"
	case RedisTypeCache_:
		return "cache"
	case RedisTypeLock_:
		return "lock"
	case RedisTypeSafeTypeMap_:
		return "safetypemap"
	default:
		return "unknown"
	}
}

// Serializer 序列化接口
type Serializer interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

// SerializerFunc 序列化函数接口
type SerializerFunc interface {
	Serialize(v interface{}) ([]byte, error)
	Deserialize(data []byte, v interface{}) error
}

// ZSetItem ZSet 项
type ZSetItem[T any] struct {
	Value T
	Score float64
}

// QueueConfig 队列配置
type QueueConfig struct {
	MaxLength    int                                                                           // 队列最大长度，0 表示不限制
	MaxWaitTime  time.Duration                                                                 // 队列阻塞等待时间，0 表示不阻塞
	MaxRetry     int                                                                           // 队列重试次数，0 表示不重试
	ErrorHandler func(value interface{}, err error, storage func(value interface{})) time.Duration // 错误处理器
}

// CacheConfig 缓存配置
type CacheConfig struct {
	DefaultExpire time.Duration // 默认过期时间，0 表示不过期
}

// LockConfig 锁配置
type LockConfig struct {
	WaitTime           time.Duration // 最长的锁时间，获得锁后，如果在这个时间内没有释放锁，视为出错，自动释放锁
	RetryTime          time.Duration // 尝试获取锁的间隔时间
	MaxGetLockWaitTime time.Duration // 获取锁最长等待时间
}
