package redisTool

import (
	"reflect"

	"github.com/gomodule/redigo/redis"
)

// RedisZSet Redis 有序集合
type RedisZSet struct {
	redis *Redis
	name  string
}

// RedisTypeZSet 类型化的 Redis 有序集合
type RedisTypeZSet[T any] struct {
	zset *RedisZSet
}

// NewZSet 创建有序集合
func (r *Redis) NewZSet(name string) *RedisZSet {
	return &RedisZSet{
		redis: r,
		name:  r.CreateName(RedisTypeZSet_, name),
	}
}

// Add 添加元素
func (z *RedisZSet) Add(value interface{}, score float64) error {
	data, err := z.redis.Serialize(value)
	if err != nil {
		return err
	}
	_, err = z.redis.Do("ZADD", z.name, score, data)
	return err
}

// Remove 移除元素
func (z *RedisZSet) Remove(values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	args := make([]interface{}, 0, len(values)+1)
	args = append(args, z.name)

	for _, value := range values {
		data, err := z.redis.Serialize(value)
		if err != nil {
			return err
		}
		args = append(args, data)
	}

	_, err := z.redis.Do("ZREM", args...)
	return err
}

// Score 获取元素的分数
func (z *RedisZSet) Score(value interface{}) (float64, bool) {
	data, err := z.redis.Serialize(value)
	if err != nil {
		return 0, false
	}

	score, err := redis.Float64(z.redis.Do("ZSCORE", z.name, data))
	if err != nil {
		return 0, false
	}
	return score, true
}

// IncrementScore 增加元素的分数
func (z *RedisZSet) IncrementScore(value interface{}, delta float64) (float64, error) {
	data, err := z.redis.Serialize(value)
	if err != nil {
		return 0, err
	}

	score, err := redis.Float64(z.redis.Do("ZINCRBY", z.name, delta, data))
	if err != nil {
		return 0, err
	}
	return score, nil
}

// Length 获取有序集合大小
func (z *RedisZSet) Length() int {
	length, err := redis.Int(z.redis.Do("ZCARD", z.name))
	if err != nil {
		return 0
	}
	return length
}

// IsEmpty 判断有序集合是否为空
func (z *RedisZSet) IsEmpty() bool {
	return z.Length() == 0
}

// IsNotEmpty 判断有序集合是否不为空
func (z *RedisZSet) IsNotEmpty() bool {
	return !z.IsEmpty()
}

// Clear 清空有序集合
func (z *RedisZSet) Clear() error {
	_, err := z.redis.Do("DEL", z.name)
	return err
}

// RangeByScore 按分数范围获取元素
func (z *RedisZSet) RangeByScore(min, max float64, withScores bool) ([]interface{}, error) {
	var reply interface{}
	var err error

	if withScores {
		reply, err = z.redis.Do("ZRANGEBYSCORE", z.name, min, max, "WITHSCORES")
	} else {
		reply, err = z.redis.Do("ZRANGEBYSCORE", z.name, min, max)
	}

	if err != nil {
		return nil, err
	}

	data, err := redis.ByteSlices(reply, nil)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0)
	if withScores {
		for i := 0; i < len(data); i += 2 {
			var value interface{}
			if err := z.redis.Deserialize(data[i], &value); err == nil {
				result = append(result, value)
			}
		}
	} else {
		for _, d := range data {
			var value interface{}
			if err := z.redis.Deserialize(d, &value); err == nil {
				result = append(result, value)
			}
		}
	}

	return result, nil
}

// RangeByRank 按排名范围获取元素
func (z *RedisZSet) RangeByRank(start, stop int) ([]interface{}, error) {
	data, err := redis.ByteSlices(z.redis.Do("ZRANGE", z.name, start, stop))
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0, len(data))
	for _, d := range data {
		var value interface{}
		if err := z.redis.Deserialize(d, &value); err == nil {
			result = append(result, value)
		}
	}
	return result, nil
}

// RemoveRangeByScore 按分数范围删除元素
func (z *RedisZSet) RemoveRangeByScore(min, max float64) error {
	_, err := z.redis.Do("ZREMRANGEBYSCORE", z.name, min, max)
	return err
}

// Iterator 获取迭代器（按分数过滤）
func (z *RedisZSet) Iterator(batchSize int) <-chan struct {
	Value interface{}
	Score float64
} {
	ch := make(chan struct {
		Value interface{}
		Score float64
	})
	go func() {
		defer close(ch)

		conn := z.redis.GetConn()
		defer conn.Close()

		cursor := 0
		for {
			values, err := redis.Values(conn.Do("ZSCAN", z.name, cursor, "COUNT", batchSize))
			if err != nil {
				break
			}

			if len(values) != 2 {
				break
			}

			cursor, _ = redis.Int(values[0], nil)
			items, _ := redis.ByteSlices(values[1], nil)

			for i := 0; i < len(items); i += 2 {
				var value interface{}
				if err := z.redis.Deserialize(items[i], &value); err == nil {
					score, _ := redis.Float64(items[i+1], nil)
					ch <- struct {
						Value interface{}
						Score float64
					}{Value: value, Score: score}
				}
			}

			if cursor == 0 {
				break
			}
		}
	}()
	return ch
}

// === RedisTypeZSet 类型化方法 ===

// Add 添加元素
func (tz *RedisTypeZSet[T]) Add(value T, score float64) error {
	return tz.zset.Add(value, score)
}

// Remove 移除元素
func (tz *RedisTypeZSet[T]) Remove(values ...T) error {
	if len(values) == 0 {
		return nil
	}

	interfaces := make([]interface{}, len(values))
	for i, v := range values {
		interfaces[i] = v
	}
	return tz.zset.Remove(interfaces...)
}

// Score 获取元素的分数
func (tz *RedisTypeZSet[T]) Score(value T) (v float64, exist bool) {
	return tz.zset.Score(value)
}

// IncrementScore 增加元素的分数
func (tz *RedisTypeZSet[T]) IncrementScore(value T, delta float64) (float64, error) {
	return tz.zset.IncrementScore(value, delta)
}

// Length 获取有序集合大小
func (tz *RedisTypeZSet[T]) Length() int {
	return tz.zset.Length()
}

// IsEmpty 判断有序集合是否为空
func (tz *RedisTypeZSet[T]) IsEmpty() bool {
	return tz.zset.IsEmpty()
}

// IsNotEmpty 判断有序集合是否不为空
func (tz *RedisTypeZSet[T]) IsNotEmpty() bool {
	return tz.zset.IsNotEmpty()
}

// Clear 清空有序集合
func (tz *RedisTypeZSet[T]) Clear() error {
	return tz.zset.Clear()
}

// RangeByScore 按分数范围获取元素
func (tz *RedisTypeZSet[T]) RangeByScore(min, max float64) ([]T, error) {
	values, err := tz.zset.RangeByScore(min, max, false)
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
			data, _ := tz.zset.redis.Serialize(value)
			if err := tz.zset.redis.Deserialize(data, item); err == nil {
				result = append(result, reflect.ValueOf(item).Elem().Interface().(T))
			}
		}
	}
	return result, nil
}

// RangeByRank 按排名范围获取元素
func (tz *RedisTypeZSet[T]) RangeByRank(start, stop int) ([]T, error) {
	values, err := tz.zset.RangeByRank(start, stop)
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
			data, _ := tz.zset.redis.Serialize(value)
			if err := tz.zset.redis.Deserialize(data, item); err == nil {
				result = append(result, reflect.ValueOf(item).Elem().Interface().(T))
			}
		}
	}
	return result, nil
}

// RemoveRangeByScore 按分数范围删除元素
func (tz *RedisTypeZSet[T]) RemoveRangeByScore(min, max float64) error {
	return tz.zset.RemoveRangeByScore(min, max)
}

// Iterator 获取迭代器
func (tz *RedisTypeZSet[T]) Iterator(batchSize int) <-chan ZSetItem[T] {
	ch := make(chan ZSetItem[T])
	go func() {
		defer close(ch)

		var zero T
		for item := range tz.zset.Iterator(batchSize) {
			if v, ok := item.Value.(T); ok {
				ch <- ZSetItem[T]{Value: v, Score: item.Score}
			} else {
				result := reflect.New(reflect.TypeOf(zero)).Interface()
				data, _ := tz.zset.redis.Serialize(item.Value)
				if err := tz.zset.redis.Deserialize(data, result); err == nil {
					ch <- ZSetItem[T]{
						Value: reflect.ValueOf(result).Elem().Interface().(T),
						Score: item.Score,
					}
				}
			}
		}
	}()
	return ch
}

// IteratorFilterByScore 按分数过滤的迭代器
func (tz *RedisTypeZSet[T]) IteratorFilterByScore(min, max float64) <-chan ZSetItem[T] {
	ch := make(chan ZSetItem[T])
	go func() {
		defer close(ch)

		conn := tz.zset.redis.GetConn()
		defer conn.Close()

		data, err := redis.ByteSlices(conn.Do("ZRANGEBYSCORE", tz.zset.name, min, max, "WITHSCORES"))
		if err != nil {
			return
		}

		var zero T
		for i := 0; i < len(data); i += 2 {
			var value interface{}
			if err := tz.zset.redis.Deserialize(data[i], &value); err == nil {
				score, _ := redis.Float64(data[i+1], nil)

				if v, ok := value.(T); ok {
					ch <- ZSetItem[T]{Value: v, Score: score}
				} else {
					item := reflect.New(reflect.TypeOf(zero)).Interface()
					itemData, _ := tz.zset.redis.Serialize(value)
					if err := tz.zset.redis.Deserialize(itemData, item); err == nil {
						ch <- ZSetItem[T]{
							Value: reflect.ValueOf(item).Elem().Interface().(T),
							Score: score,
						}
					}
				}
			}
		}
	}()
	return ch
}
