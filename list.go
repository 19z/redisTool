package redisTool

import (
	"fmt"
	"reflect"

	"github.com/gomodule/redigo/redis"
)

// RedisList Redis 列表
type RedisList struct {
	redis *Redis
	name  string
}

// RedisTypeList 类型化的 Redis 列表
type RedisTypeList[T any] struct {
	list *RedisList
}

// NewList 创建列表
func (r *Redis) NewList(name string) *RedisList {
	return &RedisList{
		redis: r,
		name:  r.CreateName(RedisTypeList_, name),
	}
}


// Push 从右侧推入元素
func (l *RedisList) Push(value interface{}) error {
	data, err := l.redis.Serialize(value)
	if err != nil {
		return err
	}
	_, err = l.redis.Do("RPUSH", l.name, data)
	return err
}

// Pop 从右侧弹出元素
func (l *RedisList) Pop() (interface{}, bool) {
	data, err := redis.Bytes(l.redis.Do("RPOP", l.name))
	if err != nil || len(data) == 0 {
		return nil, false
	}

	var result interface{}
	if err := l.redis.Deserialize(data, &result); err != nil {
		return nil, false
	}
	return result, true
}

// Shift 从左侧弹出元素
func (l *RedisList) Shift() (interface{}, bool) {
	data, err := redis.Bytes(l.redis.Do("LPOP", l.name))
	if err != nil || len(data) == 0 {
		return nil, false
	}

	var result interface{}
	if err := l.redis.Deserialize(data, &result); err != nil {
		return nil, false
	}
	return result, true
}

// Unshift 从左侧推入元素
func (l *RedisList) Unshift(value interface{}) error {
	data, err := l.redis.Serialize(value)
	if err != nil {
		return err
	}
	_, err = l.redis.Do("LPUSH", l.name, data)
	return err
}

// Index 获取指定索引的元素
func (l *RedisList) Index(index int) (interface{}, bool) {
	data, err := redis.Bytes(l.redis.Do("LINDEX", l.name, index))
	if err != nil || len(data) == 0 {
		return nil, false
	}

	var result interface{}
	if err := l.redis.Deserialize(data, &result); err != nil {
		return nil, false
	}
	return result, true
}

// Length 获取列表长度
func (l *RedisList) Length() int {
	length, err := redis.Int(l.redis.Do("LLEN", l.name))
	if err != nil {
		return 0
	}
	return length
}

// Clear 清空列表
func (l *RedisList) Clear() error {
	_, err := l.redis.Do("DEL", l.name)
	return err
}

// IsEmpty 判断列表是否为空
func (l *RedisList) IsEmpty() bool {
	return l.Length() == 0
}

// IsNotEmpty 判断列表是否不为空
func (l *RedisList) IsNotEmpty() bool {
	return !l.IsEmpty()
}

// Exists 判断列表是否存在
func (l *RedisList) Exists() bool {
	exists, err := redis.Int(l.redis.Do("EXISTS", l.name))
	if err != nil {
		return false
	}
	return exists == 1
}

// Get 获取指定范围的元素
func (l *RedisList) Get(start, end int) ([]interface{}, error) {
	data, err := redis.ByteSlices(l.redis.Do("LRANGE", l.name, start, end))
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0, len(data))
	for _, d := range data {
		var item interface{}
		if err := l.redis.Deserialize(d, &item); err != nil {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// Set 设置指定索引的元素
func (l *RedisList) Set(index int, value interface{}) error {
	data, err := l.redis.Serialize(value)
	if err != nil {
		return err
	}
	_, err = l.redis.Do("LSET", l.name, index, data)
	return err
}

// DeleteIndex 删除指定索引的元素（通过设置标记并移除）
func (l *RedisList) DeleteIndex(index int) error {
	// Redis 没有直接删除索引的命令，需要先标记再删除
	marker := "__DELETED__"
	if err := l.Set(index, marker); err != nil {
		return err
	}
	_, err := l.redis.Do("LREM", l.name, 1, marker)
	return err
}

// DeleteValue 删除指定值的元素
func (l *RedisList) DeleteValue(value interface{}, count int) error {
	data, err := l.redis.Serialize(value)
	if err != nil {
		return err
	}
	_, err = l.redis.Do("LREM", l.name, count, data)
	return err
}

// DeleteRange 删除指定范围外的元素
func (l *RedisList) DeleteRange(start, end int) error {
	_, err := l.redis.Do("LTRIM", l.name, start, end)
	return err
}

// Iterator 获取迭代器
func (l *RedisList) Iterator(batchSize int) <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		index := 0
		for {
			items, err := l.Get(index, index+batchSize-1)
			if err != nil || len(items) == 0 {
				break
			}
			for _, item := range items {
				ch <- item
			}
			if len(items) < batchSize {
				break
			}
			index += batchSize
		}
	}()
	return ch
}

// SafeUpset 安全更新（多实例安全）
func (l *RedisList) SafeUpset(index int, value interface{}) (interface{}, bool, error) {
	// 使用 Lua 脚本实现原子操作
	script := `
		local old = redis.call('LINDEX', KEYS[1], ARGV[1])
		if old then
			redis.call('LSET', KEYS[1], ARGV[1], ARGV[2])
		end
		return old
	`

	data, err := l.redis.Serialize(value)
	if err != nil {
		return nil, false, err
	}

	conn := l.redis.GetConn()
	defer conn.Close()

	luaScript := redis.NewScript(1, script)
	oldData, err := redis.Bytes(luaScript.Do(conn, l.name, index, data))
	if err != nil || len(oldData) == 0 {
		return nil, false, err
	}

	var oldValue interface{}
	if err := l.redis.Deserialize(oldData, &oldValue); err != nil {
		return nil, false, err
	}
	return oldValue, true, nil
}

// === RedisTypeList 类型化方法 ===

// Push 推入元素
func (tl *RedisTypeList[T]) Push(value T) error {
	return tl.list.Push(value)
}

// Pop 弹出元素
func (tl *RedisTypeList[T]) Pop() (T, bool) {
	var zero T
	value, ok := tl.list.Pop()
	if !ok {
		return zero, false
	}

	// 类型断言
	if v, ok := value.(T); ok {
		return v, true
	}

	// 尝试通过反射转换
	result := reflect.New(reflect.TypeOf(zero)).Interface()
	data, _ := tl.list.redis.Serialize(value)
	if err := tl.list.redis.Deserialize(data, result); err != nil {
		return zero, false
	}
	return reflect.ValueOf(result).Elem().Interface().(T), true
}

// Shift 从左侧弹出元素
func (tl *RedisTypeList[T]) Shift() (T, bool) {
	var zero T
	value, ok := tl.list.Shift()
	if !ok {
		return zero, false
	}

	if v, ok := value.(T); ok {
		return v, true
	}

	result := reflect.New(reflect.TypeOf(zero)).Interface()
	data, _ := tl.list.redis.Serialize(value)
	if err := tl.list.redis.Deserialize(data, result); err != nil {
		return zero, false
	}
	return reflect.ValueOf(result).Elem().Interface().(T), true
}

// Unshift 从左侧推入元素
func (tl *RedisTypeList[T]) Unshift(value T) error {
	return tl.list.Unshift(value)
}

// Index 获取指定索引的元素
func (tl *RedisTypeList[T]) Index(index int) (T, bool) {
	var zero T
	value, ok := tl.list.Index(index)
	if !ok {
		return zero, false
	}

	if v, ok := value.(T); ok {
		return v, true
	}

	result := reflect.New(reflect.TypeOf(zero)).Interface()
	data, _ := tl.list.redis.Serialize(value)
	if err := tl.list.redis.Deserialize(data, result); err != nil {
		return zero, false
	}
	return reflect.ValueOf(result).Elem().Interface().(T), true
}

// Length 获取列表长度
func (tl *RedisTypeList[T]) Length() int {
	return tl.list.Length()
}

// Clear 清空列表
func (tl *RedisTypeList[T]) Clear() error {
	return tl.list.Clear()
}

// IsEmpty 判断列表是否为空
func (tl *RedisTypeList[T]) IsEmpty() bool {
	return tl.list.IsEmpty()
}

// IsNotEmpty 判断列表是否不为空
func (tl *RedisTypeList[T]) IsNotEmpty() bool {
	return tl.list.IsNotEmpty()
}

// Exists 判断列表是否存在
func (tl *RedisTypeList[T]) Exists() bool {
	return tl.list.Exists()
}

// Get 获取指定范围的元素
func (tl *RedisTypeList[T]) Get(start, end int) ([]T, error) {
	values, err := tl.list.Get(start, end)
	if err != nil {
		return nil, err
	}

	var zero T
	result := make([]T, 0, len(values))
	for _, value := range values {
		if v, ok := value.(T); ok {
			result = append(result, v)
		} else {
			item := reflect.New(reflect.TypeOf(zero)).Interface()
			data, _ := tl.list.redis.Serialize(value)
			if err := tl.list.redis.Deserialize(data, item); err == nil {
				result = append(result, reflect.ValueOf(item).Elem().Interface().(T))
			}
		}
	}
	return result, nil
}

// Set 设置指定索引的元素
func (tl *RedisTypeList[T]) Set(index int, value T) error {
	return tl.list.Set(index, value)
}

// DeleteIndex 删除指定索引的元素
func (tl *RedisTypeList[T]) DeleteIndex(index int) error {
	return tl.list.DeleteIndex(index)
}

// DeleteValue 删除指定值的元素
func (tl *RedisTypeList[T]) DeleteValue(value T, count int) error {
	return tl.list.DeleteValue(value, count)
}

// DeleteRange 删除指定范围外的元素
func (tl *RedisTypeList[T]) DeleteRange(start, end int) error {
	return tl.list.DeleteRange(start, end)
}

// Iterator 获取迭代器
func (tl *RedisTypeList[T]) Iterator(batchSize int) <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for value := range tl.list.Iterator(batchSize) {
			if v, ok := value.(T); ok {
				ch <- v
			} else {
				var zero T
				item := reflect.New(reflect.TypeOf(zero)).Interface()
				data, _ := tl.list.redis.Serialize(value)
				if err := tl.list.redis.Deserialize(data, item); err == nil {
					ch <- reflect.ValueOf(item).Elem().Interface().(T)
				}
			}
		}
	}()
	return ch
}

// SafeUpset 安全更新
func (tl *RedisTypeList[T]) SafeUpset(index int, value T) (T, bool, error) {
	var zero T
	oldValue, ok, err := tl.list.SafeUpset(index, value)
	if !ok || err != nil {
		return zero, false, err
	}

	if v, ok := oldValue.(T); ok {
		return v, true, nil
	}

	result := reflect.New(reflect.TypeOf(zero)).Interface()
	data, _ := tl.list.redis.Serialize(oldValue)
	if err := tl.list.redis.Deserialize(data, result); err != nil {
		return zero, false, fmt.Errorf("deserialize error: %w", err)
	}
	return reflect.ValueOf(result).Elem().Interface().(T), true, nil
}
