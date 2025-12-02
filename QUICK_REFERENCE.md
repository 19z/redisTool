# Redis Tool - 快速参考

## 🚀 初始化

```go
import "github.com/19z/redisTool"

// 创建并设置默认连接（只需一次）
redis := redisTool.Builder("127.0.0.1:6379", "password").
    Config(redisTool.Config{
        Prefix: "myapp:",
        MaxIdle: 10,
        MaxActive: 100,
    }).
    Build()

redisTool.SetDefaultConnection(redis)
```

## 📦 数据结构

### List（列表）
```go
// 类型化列表
list := redisTool.NewTypeList[Student]("students")
list.Push(student)                    // 添加
if value, ok := list.Pop(); ok {}    // 弹出
length := list.Length()               // 长度
list.Clear()                          // 清空
```

### Set（集合）
```go
// 类型化集合
set := redisTool.NewTypeSet[Student]("students")
set.Add(student)                      // 添加
set.Remove(student)                   // 移除
exists := set.Exists(student)         // 检查
items, _ := set.ToArray()            // 转数组
```

### Map（哈希表）
```go
// 类型化哈希表
dict := redisTool.NewTypeMap[Student]("students")
dict.Set("key", student)              // 设置
if value, ok := dict.Get("key") {}   // 获取
dict.Delete("key")                    // 删除
all, _ := dict.ToArray()             // 获取所有
```

### ZSet（有序集合）
```go
// 类型化有序集合
zset := redisTool.NewTypeZSet[Student]("students")
zset.Add(student, 95.5)               // 添加（带分数）
score, ok := zset.Score(student)      // 获取分数
items, _ := zset.RangeByScore(90, 100) // 按分数范围获取
```

## 🎯 高级功能

### Queue（队列）
```go
queue := redisTool.NewQueue[Task]("tasks", redisTool.QueueConfig{
    MaxLength: 100,
    MaxRetry: 3,
})

queue.Add(task)                       // 添加任务
queue.AddDelayed(task, time.Second*10) // 延迟任务
if value, ok := queue.Take(); ok {    // 获取任务
    queue.Complete(value)             // 完成任务
}
```

### Cache（缓存）
```go
cache := redisTool.NewCache[User]("users", redisTool.CacheConfig{
    DefaultExpire: time.Minute * 10,
})

cache.Set("key", user, time.Minute*5) // 设置（带过期时间）
if user, ok := cache.Get("key") {}   // 获取
cache.Delete("key")                   // 删除

// GetOrSet 模式
user := cache.GetOrSet("key", func(key string) (User, time.Duration) {
    return loadUser(key), time.Minute * 5
})
```

### Lock（分布式锁）
```go
// 基本使用
lock := redisTool.NewLock("resource")
if err := lock.Lock(); err != nil {
    return
}
defer lock.Unlock()

// 使用闭包（推荐）
redisTool.NewLock("resource").LockFunc(func() {
    // 自动处理锁
})

// 尝试获取锁
if redisTool.NewLock("resource").TryLockFunc(func() {
    // 业务逻辑
}) {
    fmt.Println("成功")
}
```

## 🛠 辅助工具

### 时间追踪
```go
// 获取上次使用时间
lastTime := redisTool.LastUseTime("key", true)

// 设置上次使用时间
redisTool.SetLastUseTime("key", time.Now())
```

### 跨时间检测
```go
// 每分钟执行一次
if redisTool.AcrossMinute("task:minute") {
    doMinuteTask()
}

// 每秒执行一次
if redisTool.AcrossSecond("task:second") {
    doSecondTask()
}

// 自定义时间间隔
if redisTool.AcrossTime("task:hour", time.Hour) {
    doHourlyTask()
}
```

### 安全类型映射
```go
safeMap := redisTool.GetSafeTypeMap()
safeMap.Set("counter", timestamp)
```

## 📋 常用模式

### 缓存模式
```go
func GetUser(id string) (*User, error) {
    cache := redisTool.NewCache[User]("users", ...)
    
    user := cache.GetOrSet(id, func(key string) (User, time.Duration) {
        user := loadFromDB(key)
        return user, time.Minute * 5
    })
    
    return &user, nil
}
```

### 分布式锁模式
```go
func UpdateResource(id string) error {
    lock := redisTool.NewLock("resource:" + id)
    
    return lock.LockFunc(func() {
        // 业务逻辑
        return updateDB(id)
    })
}
```

### 定时任务模式
```go
func CronJob() {
    // 每分钟任务
    if redisTool.AcrossMinute("cron:minute") {
        doMinuteTask()
    }
    
    // 每小时任务
    if redisTool.AcrossTime("cron:hour", time.Hour) {
        doHourlyTask()
    }
}
```

### 队列工作模式
```go
func StartWorker() {
    queue := redisTool.NewQueue[Task]("tasks", ...)
    
    queue.StartWorkers(5, func(task Task) error {
        return processTask(task)
    })
}
```

## ⚙️ 配置选项

### Config（全局配置）
```go
redisTool.Config{
    Prefix:          "myapp:",        // 键前缀
    MaxIdle:         10,               // 最大空闲连接
    MaxActive:       100,              // 最大连接数
    IdleTimeout:     time.Second * 300, // 空闲超时
    MaxLifeTime:     time.Hour,        // 连接最大生命周期
    SafeTypeMapName: "safemap",        // 安全映射名称
}
```

### QueueConfig（队列配置）
```go
redisTool.QueueConfig{
    MaxLength:   100,                  // 最大长度
    MaxWaitTime: time.Second * 5,      // 等待时间
    MaxRetry:    3,                    // 最大重试
    ErrorHandler: func(value interface{}, err error, storage func(interface{})) time.Duration {
        return time.Second * 5         // 返回重试延迟，-1 表示不重试
    },
}
```

### CacheConfig（缓存配置）
```go
redisTool.CacheConfig{
    DefaultExpire: time.Minute * 10,   // 默认过期时间，0 表示不过期
}
```

### LockConfig（锁配置）
```go
redisTool.LockConfig{
    WaitTime:           time.Second * 5,   // 锁持有时间
    RetryTime:          time.Second,       // 重试间隔
    MaxGetLockWaitTime: time.Second * 30,  // 最长等待时间
}
```

## 📊 性能优化

### 使用迭代器处理大数据
```go
// List 迭代器
for item := range list.Iterator(100) {
    process(item)
}

// Set 迭代器
for item := range set.Iterator(100) {
    process(item)
}

// ZSet 迭代器（带分数）
for item := range zset.Iterator(100) {
    fmt.Println(item.Value, item.Score)
}
```

### 批量操作
```go
// 批量添加
set.Add(item1, item2, item3)

// 批量删除
dict.Delete("key1", "key2", "key3")
```

### 原子操作
```go
// SafeUpset 原子更新
oldValue, exist, err := dict.SafeUpset("key", newValue)
```

## 🔗 完整文档

- [README.md](./README.md) - 完整功能说明
- [GLOBAL_FUNCTIONS.md](./GLOBAL_FUNCTIONS.md) - 全局函数详解
- [SUMMARY.md](./SUMMARY.md) - 项目总结
- [TESTING.md](./TESTING.md) - 测试指南
- [CHANGES.md](./CHANGES.md) - 更新日志

## 💡 提示

1. **初始化**: 使用前必须调用 `SetDefaultConnection` 设置默认连接
2. **错误处理**: 大部分操作返回 `(value, bool)` 或 `error`，记得检查
3. **并发安全**: 所有操作都是并发安全的
4. **连接池**: 自动管理连接池，无需手动管理连接
5. **过期清理**: Cache 支持自动清理过期数据
