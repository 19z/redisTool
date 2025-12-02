package redisTool

import (
	"strings"
	"testing"
	"time"
)

func TestBuilder(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	if tr.Redis == nil {
		t.Fatal("Builder() returned nil")
	}

	if tr.Redis.pool == nil {
		t.Fatal("Builder() pool is nil")
	}

	if tr.Redis.config.MaxIdle != 10 {
		t.Errorf("Builder() MaxIdle = %v, want 10", tr.Redis.config.MaxIdle)
	}
}

func TestBuilder_WithConfig(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	redis := Builder(tr.MiniRedis.Addr(), "").
		Config(Config{
			Prefix:      "test:",
			MaxIdle:     20,
			MaxActive:   200,
			IdleTimeout: time.Minute,
		}).
		Build()
	defer redis.Close()

	if redis.config.Prefix != "test:" {
		t.Errorf("Config Prefix = %v, want test:", redis.config.Prefix)
	}

	if redis.config.MaxIdle != 20 {
		t.Errorf("Config MaxIdle = %v, want 20", redis.config.MaxIdle)
	}
}

func TestBuilder_CustomNameCreator(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	customCreator := func(config Config, types RedisType, name ...string) string {
		return "custom:" + strings.Join(name, "-")
	}

	redis := Builder(tr.MiniRedis.Addr(), "").
		Config(Config{
			NameCreator: customCreator,
		}).
		Build()
	defer redis.Close()

	name := redis.CreateName(RedisTypeList_, "test", "name")
	if name != "custom:test-name" {
		t.Errorf("CreateName() = %v, want custom:test-name", name)
	}
}

func TestRedis_CreateName(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	tests := []struct {
		name      string
		prefix    string
		redisType RedisType
		names     []string
		want      string
	}{
		{
			name:      "list with prefix",
			prefix:    "app:",
			redisType: RedisTypeList_,
			names:     []string{"users"},
			want:      "app:list:users",
		},
		{
			name:      "set without prefix",
			prefix:    "",
			redisType: RedisTypeSet_,
			names:     []string{"tags"},
			want:      "set:tags",
		},
		{
			name:      "hash with multiple names",
			prefix:    "app:",
			redisType: RedisTypeHash_,
			names:     []string{"user", "profile"},
			want:      "app:hash:user:profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis := Builder(tr.MiniRedis.Addr(), "").
				Config(Config{
					Prefix: tt.prefix,
				}).
				Build()
			defer redis.Close()

			got := redis.CreateName(tt.redisType, tt.names...)
			if got != tt.want {
				t.Errorf("CreateName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedis_Do(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	// Test SET and GET
	_, err := tr.Redis.Do("SET", "testkey", "testvalue")
	if err != nil {
		t.Fatalf("Do(SET) error = %v", err)
	}

	value, err := tr.Redis.Do("GET", "testkey")
	if err != nil {
		t.Fatalf("Do(GET) error = %v", err)
	}

	if string(value.([]byte)) != "testvalue" {
		t.Errorf("Do(GET) = %v, want testvalue", value)
	}
}

func TestRedis_SerializeDeserialize(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	original := TestStruct{Name: "Alice", Age: 30}

	data, err := tr.Redis.Serialize(original)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	var result TestStruct
	err = tr.Redis.Deserialize(data, &result)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if result.Name != original.Name || result.Age != original.Age {
		t.Errorf("Deserialize() = %v, want %v", result, original)
	}
}

func TestSetDefaultConnection(t *testing.T) {
	tr := NewTestRedis(t)
	defer tr.Close()

	SetDefaultConnection(tr.Redis)

	got := GetDefaultConnection()
	if got != tr.Redis {
		t.Errorf("GetDefaultConnection() != SetDefaultConnection()")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxIdle != 10 {
		t.Errorf("DefaultConfig MaxIdle = %v, want 10", config.MaxIdle)
	}

	if config.MaxActive != 100 {
		t.Errorf("DefaultConfig MaxActive = %v, want 100", config.MaxActive)
	}

	if config.IdleTimeout != time.Second*300 {
		t.Errorf("DefaultConfig IdleTimeout = %v, want 300s", config.IdleTimeout)
	}

	if config.SafeTypeMapName != "__SafeTypeMap__" {
		t.Errorf("DefaultConfig SafeTypeMapName = %v, want __SafeTypeMap__", config.SafeTypeMapName)
	}

	if config.NameCreator == nil {
		t.Error("DefaultConfig NameCreator is nil")
	}

	if config.Serializer == nil {
		t.Error("DefaultConfig Serializer is nil")
	}
}

func TestDefaultNameCreator(t *testing.T) {
	config := Config{Prefix: "app:"}

	tests := []struct {
		name      string
		redisType RedisType
		names     []string
		want      string
	}{
		{
			name:      "no names",
			redisType: RedisTypeList_,
			names:     []string{},
			want:      "app:list",
		},
		{
			name:      "single name",
			redisType: RedisTypeSet_,
			names:     []string{"users"},
			want:      "app:set:users",
		},
		{
			name:      "multiple names",
			redisType: RedisTypeHash_,
			names:     []string{"user", "profile", "data"},
			want:      "app:hash:user:profile:data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultNameCreator(config, tt.redisType, tt.names...)
			if got != tt.want {
				t.Errorf("DefaultNameCreator() = %v, want %v", got, tt.want)
			}
		})
	}
}
