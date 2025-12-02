package redisTool

import (
	"reflect"

	"github.com/gomodule/redigo/redis"
)

// RedisSet Redis 集合
type RedisSet struct {
	redis *Redis
	name  string
}

// RedisTypeSet 类型化的 Redis 集合
type RedisTypeSet[T any] struct {
	set *RedisSet
}

// NewSet 创建集合
func (r *Redis) NewSet(name string) *RedisSet {
	return &RedisSet{
		redis: r,
		name:  r.CreateName(RedisTypeSet_, name),
	}
}


// Add 添加元素
func (s *RedisSet) Add(values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}
	
	args := make([]interface{}, 0, len(values)+1)
	args = append(args, s.name)
	
	for _, value := range values {
		data, err := s.redis.Serialize(value)
		if err != nil {
			return err
		}
		args = append(args, data)
	}
	
	_, err := s.redis.Do("SADD", args...)
	return err
}

// Remove 移除元素
func (s *RedisSet) Remove(values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}
	
	args := make([]interface{}, 0, len(values)+1)
	args = append(args, s.name)
	
	for _, value := range values {
		data, err := s.redis.Serialize(value)
		if err != nil {
			return err
		}
		args = append(args, data)
	}
	
	_, err := s.redis.Do("SREM", args...)
	return err
}

// Length 获取集合大小
func (s *RedisSet) Length() int {
	length, err := redis.Int(s.redis.Do("SCARD", s.name))
	if err != nil {
		return 0
	}
	return length
}

// IsEmpty 判断集合是否为空
func (s *RedisSet) IsEmpty() bool {
	return s.Length() == 0
}

// IsNotEmpty 判断集合是否不为空
func (s *RedisSet) IsNotEmpty() bool {
	return !s.IsEmpty()
}

// Exists 判断元素是否存在
func (s *RedisSet) Exists(value interface{}) bool {
	data, err := s.redis.Serialize(value)
	if err != nil {
		return false
	}
	
	exists, err := redis.Int(s.redis.Do("SISMEMBER", s.name, data))
	if err != nil {
		return false
	}
	return exists == 1
}

// Clear 清空集合
func (s *RedisSet) Clear() error {
	_, err := s.redis.Do("DEL", s.name)
	return err
}

// ToArray 获取所有元素
func (s *RedisSet) ToArray() ([]interface{}, error) {
	data, err := redis.ByteSlices(s.redis.Do("SMEMBERS", s.name))
	if err != nil {
		return nil, err
	}
	
	result := make([]interface{}, 0, len(data))
	for _, d := range data {
		var item interface{}
		if err := s.redis.Deserialize(d, &item); err != nil {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// Iterator 获取迭代器
func (s *RedisSet) Iterator(batchSize int) <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		
		conn := s.redis.GetConn()
		defer conn.Close()
		
		cursor := 0
		for {
			values, err := redis.Values(conn.Do("SSCAN", s.name, cursor, "COUNT", batchSize))
			if err != nil {
				break
			}
			
			if len(values) != 2 {
				break
			}
			
			cursor, _ = redis.Int(values[0], nil)
			items, _ := redis.ByteSlices(values[1], nil)
			
			for _, data := range items {
				var item interface{}
				if err := s.redis.Deserialize(data, &item); err == nil {
					ch <- item
				}
			}
			
			if cursor == 0 {
				break
			}
		}
	}()
	return ch
}

// === RedisTypeSet 类型化方法 ===

// Add 添加元素
func (ts *RedisTypeSet[T]) Add(values ...T) error {
	if len(values) == 0 {
		return nil
	}
	
	interfaces := make([]interface{}, len(values))
	for i, v := range values {
		interfaces[i] = v
	}
	return ts.set.Add(interfaces...)
}

// Remove 移除元素
func (ts *RedisTypeSet[T]) Remove(values ...T) error {
	if len(values) == 0 {
		return nil
	}
	
	interfaces := make([]interface{}, len(values))
	for i, v := range values {
		interfaces[i] = v
	}
	return ts.set.Remove(interfaces...)
}

// Length 获取集合大小
func (ts *RedisTypeSet[T]) Length() int {
	return ts.set.Length()
}

// IsEmpty 判断集合是否为空
func (ts *RedisTypeSet[T]) IsEmpty() bool {
	return ts.set.IsEmpty()
}

// IsNotEmpty 判断集合是否不为空
func (ts *RedisTypeSet[T]) IsNotEmpty() bool {
	return ts.set.IsNotEmpty()
}

// Exists 判断元素是否存在
func (ts *RedisTypeSet[T]) Exists(value T) bool {
	return ts.set.Exists(value)
}

// Clear 清空集合
func (ts *RedisTypeSet[T]) Clear() error {
	return ts.set.Clear()
}

// ToArray 获取所有元素
func (ts *RedisTypeSet[T]) ToArray() ([]T, error) {
	values, err := ts.set.ToArray()
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
			data, _ := ts.set.redis.Serialize(value)
			if err := ts.set.redis.Deserialize(data, item); err == nil {
				result = append(result, reflect.ValueOf(item).Elem().Interface().(T))
			}
		}
	}
	return result, nil
}

// Iterator 获取迭代器
func (ts *RedisTypeSet[T]) Iterator(batchSize int) <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for value := range ts.set.Iterator(batchSize) {
			if v, ok := value.(T); ok {
				ch <- v
			} else {
				var zero T
				item := reflect.New(reflect.TypeOf(zero)).Interface()
				data, _ := ts.set.redis.Serialize(value)
				if err := ts.set.redis.Deserialize(data, item); err == nil {
					ch <- reflect.ValueOf(item).Elem().Interface().(T)
				}
			}
		}
	}()
	return ch
}
