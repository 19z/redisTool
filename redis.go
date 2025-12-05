package redisTool

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Redis Redis 客户端
type Redis struct {
	pool   *redis.Pool
	config Config
}

// 全局默认连接
var defaultConnection *Redis

// SetDefaultConnection 设置默认连接
func SetDefaultConnection(r *Redis) {
	defaultConnection = r
}

// GetDefaultConnection 获取默认连接
func GetDefaultConnection() *Redis {
	return defaultConnection
}

// Builder 构建器
type RedisBuilder struct {
	addr     string
	password string
	config   Config
}

// Builder 创建 Redis 构建器
func Builder(addr, password string) *RedisBuilder {
	return &RedisBuilder{
		addr:     addr,
		password: password,
		config:   DefaultConfig(),
	}
}

// Config 设置配置
func (b *RedisBuilder) Config(config Config) *RedisBuilder {
	// 合并配置，保留未设置的默认值
	if config.NameCreator == nil {
		config.NameCreator = b.config.NameCreator
	}
	if config.Serializer == nil {
		config.Serializer = b.config.Serializer
	}
	if config.MaxIdle == 0 {
		config.MaxIdle = b.config.MaxIdle
	}
	if config.MaxActive == 0 {
		config.MaxActive = b.config.MaxActive
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = b.config.IdleTimeout
	}
	if config.SafeTypeMapName == "" {
		config.SafeTypeMapName = b.config.SafeTypeMapName
	}

	b.config = config
	return b
}

// Build 构建 Redis 客户端
func (b *RedisBuilder) Build() *Redis {
	pool := &redis.Pool{
		MaxIdle:         b.config.MaxIdle,
		MaxActive:       b.config.MaxActive,
		IdleTimeout:     b.config.IdleTimeout,
		MaxConnLifetime: b.config.MaxLifeTime,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", b.addr)
			if err != nil {
				return nil, err
			}
			if b.password != "" {
				if _, err := c.Do("AUTH", b.password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	r := &Redis{
		pool:   pool,
		config: b.config,
	}

	// 测试 Redis 连接是否可用
	conn := r.GetConn()
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		panic(fmt.Sprintf("Redis connection test failed: %v (addr: %s)", err, b.addr))
	}

	return r
}

// GetConn 获取连接
func (r *Redis) GetConn() redis.Conn {
	return r.pool.Get()
}

// GetConnWithContext 获取带上下文的连接
func (r *Redis) GetConnWithContext(ctx context.Context) (redis.Conn, error) {
	return r.pool.GetContext(ctx)
}

// Close 关闭连接池
func (r *Redis) Close() error {
	return r.pool.Close()
}

// CreateName 创建 Redis 键名
func (r *Redis) CreateName(redisType RedisType, name ...string) string {
	return r.config.NameCreator(r.config, redisType, name...)
}

// Serialize 序列化
func (r *Redis) Serialize(v interface{}) ([]byte, error) {
	return r.config.Serializer.Serialize(v)
}

// Deserialize 反序列化
func (r *Redis) Deserialize(data []byte, v interface{}) error {
	return r.config.Serializer.Deserialize(data, v)
}

// Do 执行 Redis 命令
func (r *Redis) Do(commandName string, args ...interface{}) (interface{}, error) {
	conn := r.GetConn()
	defer conn.Close()
	return conn.Do(commandName, args...)
}

// DoWithConn 使用指定连接执行 Redis 命令
func (r *Redis) DoWithConn(conn redis.Conn, commandName string, args ...interface{}) (interface{}, error) {
	return conn.Do(commandName, args...)
}
