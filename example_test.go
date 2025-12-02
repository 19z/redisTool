package redisTool_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"adFastPro/backend/internal/utils/redisTool"
)

// Student 示例结构体
type Student struct {
	Name string
	Age  int
}

// TestBasicUsage 基本使用示例
func TestBasicUsage(t *testing.T) {
	// 初始化 Redis 客户端
	redis := redisTool.Builder("127.0.0.1:6379", "").
		Config(redisTool.Config{
			Prefix: "myproject:",
			NameCreator: func(config redisTool.Config, types redisTool.RedisType, name ...string) string {
				return config.Prefix + types.String() + ":" + strings.Join(name, ":")
			},
			MaxIdle:         10,
			MaxActive:       100,
			IdleTimeout:     time.Second * 300,
			MaxLifeTime:     0,
			Serializer:      redisTool.DefaultSerializer,
			SafeTypeMapName: "__SafeTypeMap__",
		}).
		Build()

	redisTool.SetDefaultConnection(redis)

	// 测试 List
	testList(redis)

	// 测试 Set
	testSet(redis)

	// 测试 Map
	testMap(redis)

	// 测试 ZSet
	testZSet(redis)

	// 测试 Queue
	testQueue(redis)

	// 测试 Cache
	testCache(redis)

	// 测试 Lock
	testLock(redis)

	// 测试辅助工具
	testHelper(redis)
}

func testList(redis *redisTool.Redis) {
	fmt.Println("=== Testing List ===")

	list := redis.NewList("listName")
	list.Push("1")
	list.Push("2")

	if value, exist := list.Pop(); exist {
		fmt.Println("Pop:", value)
	}

	// 类型化 List
	typeList := redisTool.NewTypeList[Student]("students", redis)
	typeList.Push(Student{Name: "张三", Age: 18})
	typeList.Push(Student{Name: "李四", Age: 20})

	if value, exist := typeList.Pop(); exist {
		fmt.Println("Pop Student:", value.Name, value.Age)
	}

	// SafeUpset
	oldValue, exist, _ := typeList.SafeUpset(0, Student{Name: "王五", Age: 22})
	if exist {
		fmt.Println("Old value:", oldValue.Name)
	}
}

func testSet(redis *redisTool.Redis) {
	fmt.Println("=== Testing Set ===")

	set := redis.NewSet("setName")
	set.Add("item1", "item2", "item3")
	set.Remove("item1")

	fmt.Println("Set exists item2:", set.Exists("item2"))
	fmt.Println("Set length:", set.Length())

	// 类型化 Set
	typeSet := redisTool.NewTypeSet[Student]("typeSetName", redis)
	typeSet.Add(Student{Name: "张三", Age: 18}, Student{Name: "李四", Age: 20})
	fmt.Println("TypeSet length:", typeSet.Length())

	// 迭代器
	for student := range typeSet.Iterator(10) {
		fmt.Println("Student:", student.Name, student.Age)
	}
}

func testMap(redis *redisTool.Redis) {
	fmt.Println("=== Testing Map ===")

	dict := redis.NewMap("mapName")
	dict.Set("key1", "value1")
	dict.Set("key2", "value2")

	if value, exist := dict.Get("key1"); exist {
		fmt.Println("Get key1:", value)
	}

	// 类型化 Map
	typeMap := redisTool.NewTypeMap[Student]("students", redis)
	typeMap.Set("student1", Student{Name: "张三", Age: 18})

	if value, exist := typeMap.Get("student1"); exist {
		fmt.Println("Get student1:", value.Name, value.Age)
	}

	// 数字型 Map
	numberDict := redis.NewNumberMap("numberMapName")
	numberDict.Set("score1", 100)
	numberDict.Increment("score1", 10)

	if score, exist := numberDict.Get("score1"); exist {
		fmt.Println("Score:", score)
	}
}

func testZSet(redis *redisTool.Redis) {
	fmt.Println("=== Testing ZSet ===")

	typeZset := redisTool.NewTypeZSet[Student]("zsetName", redis)
	typeZset.Add(Student{Name: "张三", Age: 18}, 95.5)
	typeZset.Add(Student{Name: "李四", Age: 20}, 88.0)
	typeZset.Add(Student{Name: "王五", Age: 22}, 92.0)

	// 按分数范围获取
	students, _ := typeZset.RangeByScore(90, 100)
	fmt.Println("Students with score 90-100:")
	for _, s := range students {
		fmt.Printf("  %s: %d\n", s.Name, s.Age)
	}

	// 使用迭代器
	fmt.Println("Iterator with score filter:")
	for item := range typeZset.IteratorFilterByScore(0, 100) {
		fmt.Printf("  %s (Score: %.1f)\n", item.Value.Name, item.Score)
	}
}

func testQueue(redis *redisTool.Redis) {
	fmt.Println("=== Testing Queue ===")

	queue := redisTool.NewQueue[Student]("queueName", redisTool.QueueConfig{
		MaxLength:   100,
		MaxWaitTime: 0, // 非阻塞
		MaxRetry:    3,
		ErrorHandler: func(value interface{}, err error, storage func(value interface{})) time.Duration {
			fmt.Println("Error handler:", err)
			if student, ok := value.(Student); ok && student.Age < 20 {
				student.Age++
				storage(student)
				return time.Second * 5
			}
			return -1 // 不重试
		},
	}, redis)

	// 添加任务
	queue.Add(Student{Name: "张三", Age: 18})
	queue.Add(Student{Name: "李四", Age: 20})

	// 添加延迟任务
	queue.AddDelayed(Student{Name: "王五", Age: 22}, time.Second*5)

	fmt.Println("Queue length:", queue.Length())
	fmt.Println("Delayed queue length:", queue.DelayedLength())

	// 处理任务
	if value, exist := queue.Take(); exist {
		fmt.Println("Take:", value.Name, value.Age)
		queue.Complete(value)
	}
}

func testCache(redis *redisTool.Redis) {
	fmt.Println("=== Testing Cache ===")

	cache := redisTool.NewCache[Student]("users", redisTool.CacheConfig{
		DefaultExpire: time.Minute * 10,
	}, redis)

	// 设置缓存
	cache.Set("student1", Student{Name: "张三", Age: 18}, time.Minute*5)

	// 获取缓存
	if student, exist := cache.Get("student1"); exist {
		fmt.Println("Get from cache:", student.Name, student.Age)
	}

	// GetOrSet
	student := cache.GetOrSet("student2", func(key string) (Student, time.Duration) {
		fmt.Println("Cache miss, creating new student")
		return Student{Name: "李四", Age: 20}, time.Minute * 5
	})
	fmt.Println("GetOrSet result:", student.Name, student.Age)

	fmt.Println("Cache length:", cache.Length())

	// 获取 TTL
	if ttl, ok := cache.GetTTL("student1"); ok {
		fmt.Println("TTL:", ttl)
	}
}

func testLock(redis *redisTool.Redis) {
	fmt.Println("=== Testing Lock ===")

	// 基本使用
	lock := redis.NewLock("lockName", redisTool.LockConfig{
		WaitTime:           time.Second * 5,
		RetryTime:          time.Second,
		MaxGetLockWaitTime: time.Second * 3,
	})

	if err := lock.Lock(); err != nil {
		fmt.Println("Lock error:", err)
		return
	}
	fmt.Println("Lock acquired")
	defer lock.Unlock()

	time.Sleep(time.Second * 2)
	fmt.Println("Lock released")

	// 使用闭包
	redis.NewLock("lockName2").LockFunc(func() {
		fmt.Println("In lock function")
		time.Sleep(time.Second)
	})

	// TryLock
	if redis.NewLock("lockName3").TryLockFunc(func() {
		fmt.Println("Got lock, executing...")
	}) {
		fmt.Println("TryLock success")
	} else {
		fmt.Println("TryLock failed")
	}
}

func testHelper(redis *redisTool.Redis) {
	fmt.Println("=== Testing Helper ===")

	// LastUseTime
	lastTime := redis.LastUseTime("mykey", true)
	fmt.Println("Last use time:", lastTime)

	time.Sleep(time.Millisecond * 100)

	lastTime = redis.LastUseTime("mykey", false)
	fmt.Println("Last use time (not updated):", lastTime)

	// AcrossMinute
	if redis.AcrossMinute("task1") {
		fmt.Println("Across minute: true")
	} else {
		fmt.Println("Across minute: false")
	}

	// 再次调用，应该返回 false（除非真的跨越了分钟）
	if redis.AcrossMinute("task1") {
		fmt.Println("Across minute again: true")
	} else {
		fmt.Println("Across minute again: false")
	}

	// AcrossTime
	if redis.AcrossTime("task2", time.Second) {
		fmt.Println("Across second: true")
	}

	time.Sleep(time.Second * 2)

	if redis.AcrossTime("task2", time.Second) {
		fmt.Println("Across second after 2s: true")
	}
}
