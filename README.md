# Redis Tool

一个功能丰富的 Redis 工具库，提供了对 Redis 各种数据结构的高级封装和实用功能。

## 特性

- **Builder 模式** - 灵活的配置和初始化
- **类型安全** - 支持泛型，提供类型化的数据结构
- **序列化** - 灵活的序列化机制，支持自定义序列化器
- **数据结构** - List、Set、Map、ZSet 等常用数据结构
- **高级功能** - Queue（队列）、Cache（缓存）、Lock（分布式锁）
- **辅助工具** - 时间追踪、跨时间段检测等实用功能

## 安装

```bash
go get github.com/19z/redisTool
```

依赖包会自动安装：
- `github.com/gomodule/redigo` - Redis 客户端
- `github.com/google/uuid` - UUID 生成器
- `github.com/alicebob/miniredis/v2` - 测试用 Redis 模拟器

## 使用示例

### 1. 初始化 Redis 客户端

```go
import "github.com/19z/redisTool"

func main() {
    redis := redisTool.Builder("127.0.0.1:6379", "password").
        Config(redisTool.Config{
            Prefix: "myproject:",
            MaxIdle: 10,
            MaxActive: 100,
            IdleTimeout: time.Second * 300,
        }).
        Build()
    
    // 设置为全局默认连接
    redisTool.SetDefaultConnection(redis)
}
```

### 2. 使用 List

```go
// 创建列表
list := redis.NewList("mylist")
list.Push("item1")
list.Push("item2")

// 弹出元素
if value, ok := list.Pop(); ok {
    fmt.Println(value)
}

// 使用类型化列表
type Student struct {
    Name string
    Age  int
}

// 方式一：使用全局函数（需要先设置默认连接）
redisTool.SetDefaultConnection(redis)
typeList := redisTool.NewTypeList[Student]("students")
typeList.Push(Student{Name: "李四", Age: 20})

if student, ok := typeList.Pop(); ok {
    fmt.Println(student.Name, student.Age)
}

// 方式二：使用全局函数并指定连接
typeList2 := redisTool.NewTypeList[Student]("students2", redis)
typeList2.Push(Student{Name: "王五", Age: 22})
```

### 3. 使用 Set

```go
set := redis.NewSet("myset")
set.Add("item1", "item2", "item3")
set.Remove("item1")

exists := set.Exists("item2")
fmt.Println("exists:", exists)

// 类型化 Set
typeSet := redisTool.NewTypeSet[Student]("students")
typeSet.Add(Student{Name: "张三", Age: 18})
```

### 4. 使用 Map

```go
dict := redis.NewMap("mymap")
dict.Set("key1", "value1")

if value, ok := dict.Get("key1"); ok {
    fmt.Println(value)
}

// 类型化 Map
typeDict := redisTool.NewTypeMap[Student]("students")
typeDict.Set("student1", Student{Name: "张三", Age: 18})

// 数字型 Map
numberMap := redis.NewNumberMap("scores")
numberMap.Set("user1", 100)
numberMap.Increment("user1", 10)
```

### 5. 使用 ZSet

```go
zset := redisTool.NewTypeZSet[Student]("students")
zset.Add(Student{Name: "张三", Age: 18}, 95.5)
zset.Add(Student{Name: "李四", Age: 20}, 88.0)

// 按分数范围获取
students, _ := zset.RangeByScore(90, 100)
for _, s := range students {
    fmt.Println(s.Name, s.Age)
}

// 使用迭代器
for item := range zset.IteratorFilterByScore(0, 100) {
    fmt.Println(item.Value.Name, item.Score)
}
```

### 6. 使用 Queue

```go
// 使用全局函数
queue := redisTool.NewQueue[Student]("tasks", redisTool.QueueConfig{
    MaxLength: 100,
    MaxWaitTime: time.Second * 5,
    MaxRetry: 3,
    ErrorHandler: func(value interface{}, err error, storage func(value interface{})) time.Duration {
        fmt.Println("Error:", err)
        return time.Second * 5 // 5 秒后重试
    },
})

// 添加任务
queue.Add(Student{Name: "张三", Age: 18})

// 添加延迟任务
queue.AddDelayed(Student{Name: "李四", Age: 20}, time.Second * 10)

// 处理任务
if value, ok := queue.Take(); ok {
    fmt.Println(value.Name)
    queue.Complete(value)
}

// 启动工作线程
queue.StartWorkers(5, func(value Student) error {
    fmt.Println("Processing:", value.Name)
    return nil
})
```

### 7. 使用 Cache

```go
// 使用全局函数
cache := redisTool.NewCache[Student]("students", redisTool.CacheConfig{
    DefaultExpire: time.Minute * 10,
})

// 设置缓存
cache.Set("student1", Student{Name: "张三", Age: 18}, time.Minute*5)

// 获取缓存
if student, ok := cache.Get("student1"); ok {
    fmt.Println(student.Name)
}

// 获取或设置
student := cache.GetOrSet("student2", func(key string) (Student, time.Duration) {
    return Student{Name: "李四", Age: 20}, time.Minute * 5
})

// 注意：缓存会在 Set 操作时自动清理过期数据
// 使用概率性机制，在横跨分钟时触发清理，无需手动调用
```

### 8. 使用分布式锁

```go
// 使用全局函数（推荐）
lock := redisTool.NewLock("mylock", redisTool.LockConfig{
    WaitTime: time.Second * 5,
    RetryTime: time.Second,
    MaxGetLockWaitTime: time.Second * 30,
})

if err := lock.Lock(); err != nil {
    fmt.Println("获取锁失败:", err)
    return
}
defer lock.Unlock()

// 执行业务逻辑
fmt.Println("执行业务逻辑")

// 使用闭包简化
redisTool.NewLock("mylock").LockFunc(func() {
    fmt.Println("执行业务逻辑")
})

// 尝试获取锁
if redisTool.NewLock("mylock").TryLockFunc(func() {
    fmt.Println("获得锁，执行业务逻辑")
}) {
    fmt.Println("成功")
} else {
    fmt.Println("未获得锁")
}
```

### 9. 使用辅助工具

```go
// 使用全局函数（推荐）
// 获取上次使用时间
lastTime := redisTool.LastUseTime("mykey", true)
fmt.Println("上次使用时间:", lastTime)

// 检查是否跨越分钟
if redisTool.AcrossMinute("task1") {
    fmt.Println("跨越了分钟，执行任务")
}

// 检查是否跨越秒
if redisTool.AcrossSecond("task2") {
    fmt.Println("跨越了秒，执行任务")
}

// 检查是否跨越指定时间
if redisTool.AcrossTime("task3", time.Hour) {
    fmt.Println("跨越了小时，执行任务")
}

// 设置上次使用时间
redisTool.SetLastUseTime("mykey", time.Now())

// 获取安全类型映射
safeMap := redisTool.GetSafeTypeMap()
safeMap.Set("key1", time.Now().UnixMilli())
```

## 序列化

默认序列化器会自动处理：
- 实现了 `Serializer` 接口的类型
- 基本类型（string, int, float, bool 等）
- 其他类型使用 gob 编码

自定义序列化：

```go
type Student struct {
    Name string
    Age  int
}

func (s Student) Serialize() ([]byte, error) {
    return json.Marshal(s)
}

func (s *Student) Deserialize(data []byte) error {
    return json.Unmarshal(data, s)
}
```

## 配置选项

### Config

- `Prefix` - Redis 键前缀
- `NameCreator` - 自定义键名生成函数
- `MaxIdle` - 最大空闲连接数
- `MaxActive` - 最大连接数
- `IdleTimeout` - 空闲连接超时时间
- `MaxLifeTime` - 连接最大生存时间
- `Serializer` - 序列化器
- `SafeTypeMapName` - 安全类型映射名称

### QueueConfig

- `MaxLength` - 队列最大长度
- `MaxWaitTime` - 阻塞等待时间
- `MaxRetry` - 最大重试次数
- `ErrorHandler` - 错误处理函数

### CacheConfig

- `DefaultExpire` - 默认过期时间

### LockConfig

- `WaitTime` - 锁的持有时间
- `RetryTime` - 重试间隔
- `MaxGetLockWaitTime` - 获取锁的最长等待时间

## 注意事项

1. 使用前需要初始化 Redis 连接
2. 建议使用 `SetDefaultConnection` 设置全局连接
3. 大量数据操作建议使用 Iterator 避免阻塞
4. 分布式锁使用后记得释放，建议使用 defer
5. 缓存会自动清理过期数据，也可以手动调用 `ClearExpired`

## 许可证

MIT License
