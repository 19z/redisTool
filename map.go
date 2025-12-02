package redisTool

import (
	"fmt"
	"reflect"

	"github.com/gomodule/redigo/redis"
)

// RedisMap Redis 哈希表
type RedisMap struct {
	redis *Redis
	name  string
}

// RedisTypeMap 类型化的 Redis 哈希表
type RedisTypeMap[T any] struct {
	rmap *RedisMap
}

// RedisNumberMap 数字型 Redis 哈希表
type RedisNumberMap struct {
	rmap *RedisMap
}

// NewMap 创建哈希表
func (r *Redis) NewMap(name string) *RedisMap {
	return &RedisMap{
		redis: r,
		name:  r.CreateName(RedisTypeHash_, name),
	}
}

// NewNumberMap 创建数字型哈希表
func (r *Redis) NewNumberMap(name string) *RedisNumberMap {
	return &RedisNumberMap{
		rmap: r.NewMap(name),
	}
}

// Set 设置键值
func (m *RedisMap) Set(key string, value interface{}) error {
	data, err := m.redis.Serialize(value)
	if err != nil {
		return err
	}
	_, err = m.redis.Do("HSET", m.name, key, data)
	return err
}

// Get 获取值
func (m *RedisMap) Get(key string) (interface{}, bool) {
	data, err := redis.Bytes(m.redis.Do("HGET", m.name, key))
	if err != nil || len(data) == 0 {
		return nil, false
	}
	
	var result interface{}
	if err := m.redis.Deserialize(data, &result); err != nil {
		return nil, false
	}
	return result, true
}

// Delete 删除键
func (m *RedisMap) Delete(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, m.name)
	for _, key := range keys {
		args = append(args, key)
	}
	
	_, err := m.redis.Do("HDEL", args...)
	return err
}

// Exists 判断键是否存在
func (m *RedisMap) Exists(key string) bool {
	exists, err := redis.Int(m.redis.Do("HEXISTS", m.name, key))
	if err != nil {
		return false
	}
	return exists == 1
}

// Length 获取哈希表大小
func (m *RedisMap) Length() int {
	length, err := redis.Int(m.redis.Do("HLEN", m.name))
	if err != nil {
		return 0
	}
	return length
}

// IsEmpty 判断哈希表是否为空
func (m *RedisMap) IsEmpty() bool {
	return m.Length() == 0
}

// IsNotEmpty 判断哈希表是否不为空
func (m *RedisMap) IsNotEmpty() bool {
	return !m.IsEmpty()
}

// Clear 清空哈希表
func (m *RedisMap) Clear() error {
	_, err := m.redis.Do("DEL", m.name)
	return err
}

// ToArray 获取所有键值对
func (m *RedisMap) ToArray() (map[string]interface{}, error) {
	data, err := redis.ByteSlices(m.redis.Do("HGETALL", m.name))
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]interface{})
	for i := 0; i < len(data); i += 2 {
		key := string(data[i])
		var value interface{}
		if err := m.redis.Deserialize(data[i+1], &value); err != nil {
			continue
		}
		result[key] = value
	}
	return result, nil
}

// Keys 获取所有键
func (m *RedisMap) Keys() ([]string, error) {
	keys, err := redis.Strings(m.redis.Do("HKEYS", m.name))
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Iterator 获取迭代器
func (m *RedisMap) Iterator(batchSize int) <-chan struct {
	Key   string
	Value interface{}
} {
	ch := make(chan struct {
		Key   string
		Value interface{}
	})
	go func() {
		defer close(ch)
		
		conn := m.redis.GetConn()
		defer conn.Close()
		
		cursor := 0
		for {
			values, err := redis.Values(conn.Do("HSCAN", m.name, cursor, "COUNT", batchSize))
			if err != nil {
				break
			}
			
			if len(values) != 2 {
				break
			}
			
			cursor, _ = redis.Int(values[0], nil)
			items, _ := redis.ByteSlices(values[1], nil)
			
			for i := 0; i < len(items); i += 2 {
				key := string(items[i])
				var value interface{}
				if err := m.redis.Deserialize(items[i+1], &value); err == nil {
					ch <- struct {
						Key   string
						Value interface{}
					}{Key: key, Value: value}
				}
			}
			
			if cursor == 0 {
				break
			}
		}
	}()
	return ch
}

// === RedisTypeMap 类型化方法 ===

// Set 设置键值
func (tm *RedisTypeMap[T]) Set(key string, value T) error {
	return tm.rmap.Set(key, value)
}

// Get 获取值
func (tm *RedisTypeMap[T]) Get(key string) (T, bool) {
	var zero T
	value, ok := tm.rmap.Get(key)
	if !ok {
		return zero, false
	}
	
	if v, ok := value.(T); ok {
		return v, true
	}
	
	result := reflect.New(reflect.TypeOf(zero)).Interface()
	data, _ := tm.rmap.redis.Serialize(value)
	if err := tm.rmap.redis.Deserialize(data, result); err != nil {
		return zero, false
	}
	return reflect.ValueOf(result).Elem().Interface().(T), true
}

// Delete 删除键
func (tm *RedisTypeMap[T]) Delete(keys ...string) error {
	return tm.rmap.Delete(keys...)
}

// Exists 判断键是否存在
func (tm *RedisTypeMap[T]) Exists(key string) bool {
	return tm.rmap.Exists(key)
}

// Length 获取哈希表大小
func (tm *RedisTypeMap[T]) Length() int {
	return tm.rmap.Length()
}

// IsEmpty 判断哈希表是否为空
func (tm *RedisTypeMap[T]) IsEmpty() bool {
	return tm.rmap.IsEmpty()
}

// IsNotEmpty 判断哈希表是否不为空
func (tm *RedisTypeMap[T]) IsNotEmpty() bool {
	return tm.rmap.IsNotEmpty()
}

// Clear 清空哈希表
func (tm *RedisTypeMap[T]) Clear() error {
	return tm.rmap.Clear()
}

// ToArray 获取所有键值对
func (tm *RedisTypeMap[T]) ToArray() (map[string]T, error) {
	values, err := tm.rmap.ToArray()
	if err != nil {
		return nil, err
	}
	
	var zero T
	result := make(map[string]T)
	for key, value := range values {
		if v, ok := value.(T); ok {
			result[key] = v
		} else {
			item := reflect.New(reflect.TypeOf(zero)).Interface()
			data, _ := tm.rmap.redis.Serialize(value)
			if err := tm.rmap.redis.Deserialize(data, item); err == nil {
				result[key] = reflect.ValueOf(item).Elem().Interface().(T)
			}
		}
	}
	return result, nil
}

// Keys 获取所有键
func (tm *RedisTypeMap[T]) Keys() ([]string, error) {
	return tm.rmap.Keys()
}

// Iterator 获取迭代器
func (tm *RedisTypeMap[T]) Iterator(batchSize int) <-chan struct {
	Key   string
	Value T
} {
	ch := make(chan struct {
		Key   string
		Value T
	})
	go func() {
		defer close(ch)
		
		var zero T
		for item := range tm.rmap.Iterator(batchSize) {
			if v, ok := item.Value.(T); ok {
				ch <- struct {
					Key   string
					Value T
				}{Key: item.Key, Value: v}
			} else {
				result := reflect.New(reflect.TypeOf(zero)).Interface()
				data, _ := tm.rmap.redis.Serialize(item.Value)
				if err := tm.rmap.redis.Deserialize(data, result); err == nil {
					ch <- struct {
						Key   string
						Value T
					}{Key: item.Key, Value: reflect.ValueOf(result).Elem().Interface().(T)}
				}
			}
		}
	}()
	return ch
}

// === RedisNumberMap 数字型方法 ===

// Set 设置数值
func (nm *RedisNumberMap) Set(key string, value float64) error {
	_, err := nm.rmap.redis.Do("HSET", nm.rmap.name, key, value)
	return err
}

// Get 获取数值
func (nm *RedisNumberMap) Get(key string) (float64, bool) {
	value, err := redis.Float64(nm.rmap.redis.Do("HGET", nm.rmap.name, key))
	if err != nil {
		return 0, false
	}
	return value, true
}

// Increment 增加数值
func (nm *RedisNumberMap) Increment(key string, delta float64) (float64, error) {
	value, err := redis.Float64(nm.rmap.redis.Do("HINCRBYFLOAT", nm.rmap.name, key, delta))
	if err != nil {
		return 0, err
	}
	return value, nil
}

// Decrement 减少数值
func (nm *RedisNumberMap) Decrement(key string, delta float64) (float64, error) {
	return nm.Increment(key, -delta)
}

// Delete 删除键
func (nm *RedisNumberMap) Delete(keys ...string) error {
	return nm.rmap.Delete(keys...)
}

// Exists 判断键是否存在
func (nm *RedisNumberMap) Exists(key string) bool {
	return nm.rmap.Exists(key)
}

// Length 获取哈希表大小
func (nm *RedisNumberMap) Length() int {
	return nm.rmap.Length()
}

// IsEmpty 判断哈希表是否为空
func (nm *RedisNumberMap) IsEmpty() bool {
	return nm.rmap.IsEmpty()
}

// IsNotEmpty 判断哈希表是否不为空
func (nm *RedisNumberMap) IsNotEmpty() bool {
	return nm.rmap.IsNotEmpty()
}

// Clear 清空哈希表
func (nm *RedisNumberMap) Clear() error {
	return nm.rmap.Clear()
}

// ToArray 获取所有键值对
func (nm *RedisNumberMap) ToArray() (map[string]float64, error) {
	data, err := redis.ByteSlices(nm.rmap.redis.Do("HGETALL", nm.rmap.name))
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]float64)
	for i := 0; i < len(data); i += 2 {
		key := string(data[i])
		value, err := redis.Float64(data[i+1], nil)
		if err != nil {
			continue
		}
		result[key] = value
	}
	return result, nil
}

// Keys 获取所有键
func (nm *RedisNumberMap) Keys() ([]string, error) {
	return nm.rmap.Keys()
}

// SafeUpset 安全更新（多实例安全）
func (tm *RedisTypeMap[T]) SafeUpset(key string, value T) (T, bool, error) {
	var zero T
	
	// 使用 Lua 脚本实现原子操作
	script := `
		local old = redis.call('HGET', KEYS[1], ARGV[1])
		redis.call('HSET', KEYS[1], ARGV[1], ARGV[2])
		return old
	`
	
	data, err := tm.rmap.redis.Serialize(value)
	if err != nil {
		return zero, false, err
	}
	
	conn := tm.rmap.redis.GetConn()
	defer conn.Close()
	
	luaScript := redis.NewScript(1, script)
	oldData, err := redis.Bytes(luaScript.Do(conn, tm.rmap.name, key, data))
	if err != nil || len(oldData) == 0 {
		return zero, false, nil
	}
	
	result := reflect.New(reflect.TypeOf(zero)).Interface()
	if err := tm.rmap.redis.Deserialize(oldData, result); err != nil {
		return zero, false, fmt.Errorf("deserialize error: %w", err)
	}
	return reflect.ValueOf(result).Elem().Interface().(T), true, nil
}
