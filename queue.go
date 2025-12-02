package redisTool

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Queue 队列
type Queue[T any] struct {
	redis         *Redis
	name          string
	delayedName   string
	processingName string
	retryName     string
	config        QueueConfig
}


// Add 添加任务到队列
func (q *Queue[T]) Add(value T) error {
	// 检查队列长度
	if q.config.MaxLength > 0 {
		length, err := redis.Int(q.redis.Do("LLEN", q.name))
		if err != nil {
			return err
		}
		if length >= q.config.MaxLength {
			return fmt.Errorf("queue is full, max length: %d", q.config.MaxLength)
		}
	}
	
	data, err := q.redis.Serialize(value)
	if err != nil {
		return err
	}
	
	_, err = q.redis.Do("RPUSH", q.name, data)
	return err
}

// AddDelayed 添加延迟任务
func (q *Queue[T]) AddDelayed(value T, delay time.Duration) error {
	data, err := q.redis.Serialize(value)
	if err != nil {
		return err
	}
	
	score := float64(time.Now().Add(delay).UnixMilli())
	_, err = q.redis.Do("ZADD", q.delayedName, score, data)
	return err
}

// Take 获取任务
func (q *Queue[T]) Take() (T, bool) {
	var zero T
	
	// 首先处理延迟任务
	q.processDelayedTasks()
	
	var data []byte
	var err error
	
	if q.config.MaxWaitTime > 0 {
		// 阻塞获取
		values, err := redis.ByteSlices(q.redis.Do("BLPOP", q.name, int(q.config.MaxWaitTime.Seconds())))
		if err != nil || len(values) < 2 {
			return zero, false
		}
		data = values[1]
	} else {
		// 非阻塞获取
		data, err = redis.Bytes(q.redis.Do("LPOP", q.name))
		if err != nil || len(data) == 0 {
			return zero, false
		}
	}
	
	var value T
	result := reflect.New(reflect.TypeOf(zero)).Interface()
	if err := q.redis.Deserialize(data, result); err != nil {
		return zero, false
	}
	value = reflect.ValueOf(result).Elem().Interface().(T)
	
	// 将任务移到处理中队列
	if q.config.MaxRetry > 0 {
		q.redis.Do("ZADD", q.processingName, float64(time.Now().UnixMilli()), data)
	}
	
	return value, true
}

// Complete 完成任务
func (q *Queue[T]) Complete(value T) error {
	if q.config.MaxRetry == 0 {
		return nil
	}
	
	data, err := q.redis.Serialize(value)
	if err != nil {
		return err
	}
	
	_, err = q.redis.Do("ZREM", q.processingName, data)
	return err
}

// Fail 任务失败
func (q *Queue[T]) Fail(value T, err error) error {
	if q.config.ErrorHandler == nil {
		return q.Complete(value)
	}
	
	// 调用错误处理器
	retryDelay := q.config.ErrorHandler(value, err, func(v interface{}) {
		if updatedValue, ok := v.(T); ok {
			value = updatedValue
		}
	})
	
	if retryDelay < 0 {
		// 不重试，直接完成
		return q.Complete(value)
	}
	
	// 重新加入延迟队列
	if retryDelay > 0 {
		return q.AddDelayed(value, retryDelay)
	}
	
	// 立即重试
	return q.Add(value)
}

// Length 获取队列长度
func (q *Queue[T]) Length() int {
	length, err := redis.Int(q.redis.Do("LLEN", q.name))
	if err != nil {
		return 0
	}
	return length
}

// DelayedLength 获取延迟队列长度
func (q *Queue[T]) DelayedLength() int {
	length, err := redis.Int(q.redis.Do("ZCARD", q.delayedName))
	if err != nil {
		return 0
	}
	return length
}

// ProcessingLength 获取处理中队列长度
func (q *Queue[T]) ProcessingLength() int {
	length, err := redis.Int(q.redis.Do("ZCARD", q.processingName))
	if err != nil {
		return 0
	}
	return length
}

// Clear 清空队列
func (q *Queue[T]) Clear() error {
	conn := q.redis.GetConn()
	defer conn.Close()
	
	_, err := conn.Do("DEL", q.name, q.delayedName, q.processingName, q.retryName)
	return err
}

// processDelayedTasks 处理延迟任务
func (q *Queue[T]) processDelayedTasks() {
	now := float64(time.Now().UnixMilli())
	
	// 使用 Lua 脚本原子性地移动任务
	script := `
		local items = redis.call('ZRANGEBYSCORE', KEYS[1], 0, ARGV[1])
		for i, item in ipairs(items) do
			redis.call('ZREM', KEYS[1], item)
			redis.call('RPUSH', KEYS[2], item)
		end
		return #items
	`
	
	conn := q.redis.GetConn()
	defer conn.Close()
	
	luaScript := redis.NewScript(2, script)
	luaScript.Do(conn, q.delayedName, q.name, now)
}

// StartWorker 启动工作线程
func (q *Queue[T]) StartWorker(handler func(value T) error) {
	go func() {
		for {
			value, ok := q.Take()
			if !ok {
				time.Sleep(time.Second)
				continue
			}
			
			if err := handler(value); err != nil {
				q.Fail(value, err)
			} else {
				q.Complete(value)
			}
		}
	}()
}

// StartWorkers 启动多个工作线程
func (q *Queue[T]) StartWorkers(count int, handler func(value T) error) {
	for i := 0; i < count; i++ {
		q.StartWorker(handler)
	}
}
