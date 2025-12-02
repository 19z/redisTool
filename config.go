package redisTool

import (
	"strings"
	"time"
)

// Config Redis 配置
type Config struct {
	Prefix          string                                        // 项目前缀
	NameCreator     func(config Config, types RedisType, name ...string) string // 名称创建器
	MaxIdle         int                                           // 最大空闲连接数
	MaxActive       int                                           // 最大连接数
	IdleTimeout     time.Duration                                 // 空闲连接超时时间
	MaxLifeTime     time.Duration                                 // 活跃连接超时时间
	Serializer      SerializerFunc                                // 序列化器
	SafeTypeMapName string                                        // 安全类型映射名称
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Prefix:          "",
		NameCreator:     DefaultNameCreator,
		MaxIdle:         10,
		MaxActive:       100,
		IdleTimeout:     time.Second * 300,
		MaxLifeTime:     0,
		Serializer:      DefaultSerializer,
		SafeTypeMapName: "__SafeTypeMap__",
	}
}

// DefaultNameCreator 默认名称创建器
func DefaultNameCreator(config Config, types RedisType, name ...string) string {
	if len(name) == 0 {
		return config.Prefix + types.String()
	}
	return config.Prefix + types.String() + ":" + strings.Join(name, ":")
}
