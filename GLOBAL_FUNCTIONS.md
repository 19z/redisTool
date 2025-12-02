# 全局函数使用指南

## 📖 概述

为了简化 Redis 工具库的使用，我们提供了一套完整的全局函数。只需要设置一次默认连接，就可以在任何地方使用这些全局函数。

## 🚀 快速开始

### 1. 设置默认连接

```go
package main

import (
    "github.com/19z/redisTool"
    "time"
)

func main() {
    // 创建 Redis 客户端
    redis := redisTool.Builder("127.0.0.1:6379", "").
        Config(redisTool.Config{
            Prefix: "myapp:",
            MaxIdle: 10,
            MaxActive: 100,
        }).
        Build()
    
    // 设置为全局默认连接（只需要设置一次）
    redisTool.SetDefaultConnection(redis)
    
    // 现在就可以在任何地方使用全局函数了！
}
```

### 2. 使用全局函数

设置默认连接后，您可以在项目的任何地方直接使用全局函数：

```go
package service

import "github.com/19z/redisTool"

func ProcessUser(userID string) {
    // 使用分布式锁
    lock := redisTool.NewLock("user:" + userID)
    if err := lock.Lock(); err != nil {
        return
    }
    defer lock.Unlock()
    
    // 检查是否跨越分钟
    if redisTool.AcrossMinute("task:daily") {
        // 执行每分钟任务
    }
    
    // 使用类型化数据结构
    cache := redisTool.NewCache[User]("users", redisTool.CacheConfig{
        DefaultExpire: time.Minute * 10,
    })
    
    // ... 业务逻辑
}
```

## 📚 完整的全局函数列表

### 数据结构

#### 类型化列表
```go
typeList := redisTool.NewTypeList[Student]("students")
typeList.Push(Student{Name: "张三", Age: 18})
```

#### 类型化集合
```go
typeSet := redisTool.NewTypeSet[Student]("students")
typeSet.Add(Student{Name: "李四", Age: 20})
```

#### 类型化哈希表
```go
typeMap := redisTool.NewTypeMap[Student]("students")
typeMap.Set("stu1", Student{Name: "王五", Age: 22})
```

#### 类型化有序集合
```go
typeZSet := redisTool.NewTypeZSet[Student]("students")
typeZSet.Add(Student{Name: "赵六", Age: 19}, 95.5)
```

### 高级功能

#### 队列
```go
queue := redisTool.NewQueue[Task]("tasks", redisTool.QueueConfig{
    MaxLength: 100,
    MaxWaitTime: time.Second * 5,
    MaxRetry: 3,
})
queue.Add(Task{ID: 1, Name: "Task1"})
```

#### 缓存
```go
cache := redisTool.NewCache[User]("users", redisTool.CacheConfig{
    DefaultExpire: time.Minute * 10,
})
cache.Set("user1", User{Name: "张三"}, time.Minute*5)
```

#### 分布式锁
```go
lock := redisTool.NewLock("mylock")
lock.LockFunc(func() {
    // 自动处理锁的获取和释放
    // 执行业务逻辑
})
```

### 辅助工具

#### 时间追踪
```go
// 获取上次使用时间
lastTime := redisTool.LastUseTime("mykey", true)
fmt.Println("上次使用:", lastTime)

// 设置上次使用时间
redisTool.SetLastUseTime("mykey", time.Now())
```

#### 跨时间检测
```go
// 检查是否跨越分钟
if redisTool.AcrossMinute("task:minute") {
    fmt.Println("执行每分钟任务")
}

// 检查是否跨越秒
if redisTool.AcrossSecond("task:second") {
    fmt.Println("执行每秒任务")
}

// 检查是否跨越指定时间
if redisTool.AcrossTime("task:hour", time.Hour) {
    fmt.Println("执行每小时任务")
}
```

#### 安全类型映射
```go
safeMap := redisTool.GetSafeTypeMap()
safeMap.Set("counter", time.Now().UnixMilli())
```

## 🔄 切换连接

如果需要使用不同的 Redis 连接，可以在调用时指定：

```go
// 创建另一个 Redis 连接
redis2 := redisTool.Builder("127.0.0.1:6380", "").Build()

// 使用特定连接
typeList := redisTool.NewTypeList[Student]("students", redis2)
```

## ✨ 优势

### 1. 简化代码
**之前：**
```go
redis := getRedisConnection()
typeList := redis.NewTypeList[Student]("students")
```

**现在：**
```go
typeList := redisTool.NewTypeList[Student]("students")
```

### 2. 全局可用
在项目的任何地方都可以使用，无需传递 Redis 实例：

```go
// service/user_service.go
func GetUser(id string) {
    cache := redisTool.NewCache[User]("users", ...)
    // ...
}

// service/order_service.go
func ProcessOrder(id string) {
    lock := redisTool.NewLock("order:" + id)
    // ...
}

// handler/api_handler.go
func HandleRequest() {
    if redisTool.AcrossMinute("rate-limit") {
        // ...
    }
}
```

### 3. 灵活切换
需要时仍然可以指定特定的连接：

```go
// 使用默认连接
list1 := redisTool.NewTypeList[int]("list1")

// 使用特定连接
list2 := redisTool.NewTypeList[int]("list2", specificRedis)
```

## 🎯 最佳实践

### 1. 在程序初始化时设置默认连接

```go
func init() {
    redis := redisTool.Builder(config.RedisAddr, config.RedisPassword).
        Config(redisTool.Config{
            Prefix: config.AppName + ":",
            MaxIdle: 10,
            MaxActive: 100,
        }).
        Build()
    
    redisTool.SetDefaultConnection(redis)
}
```

### 2. 分布式锁使用闭包

```go
// 推荐：自动处理锁的获取和释放
redisTool.NewLock("resource").LockFunc(func() {
    // 业务逻辑
})

// 或使用 TryLock
if redisTool.NewLock("resource").TryLockFunc(func() {
    // 业务逻辑
}) {
    fmt.Println("执行成功")
} else {
    fmt.Println("未获得锁")
}
```

### 3. 跨时间检测用于定时任务

```go
func cronJob() {
    // 每分钟执行一次
    if redisTool.AcrossMinute("cron:minute") {
        doMinuteTask()
    }
    
    // 每小时执行一次
    if redisTool.AcrossTime("cron:hour", time.Hour) {
        doHourlyTask()
    }
    
    // 每天执行一次
    if redisTool.AcrossTime("cron:daily", 24*time.Hour) {
        doDailyTask()
    }
}
```

### 4. 缓存模式

```go
func GetUser(id string) (*User, error) {
    cache := redisTool.NewCache[User]("users", redisTool.CacheConfig{
        DefaultExpire: time.Minute * 10,
    })
    
    // 使用 GetOrSet 模式
    user := cache.GetOrSet(id, func(key string) (User, time.Duration) {
        // 从数据库加载
        user := loadUserFromDB(key)
        return user, time.Minute * 5
    })
    
    return &user, nil
}
```

## 📊 对比表

| 特性 | 实例方法 | 全局函数 |
|------|---------|---------|
| 需要传递 Redis 实例 | ✅ 是 | ❌ 否 |
| 代码简洁性 | 一般 | ✅ 优秀 |
| 支持多连接 | ✅ 是 | ✅ 是（可选参数）|
| 适用场景 | 需要明确控制连接时 | 大部分业务场景 |

## 🎓 完整示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/19z/redisTool"
)

type User struct {
    ID   string
    Name string
    Age  int
}

func main() {
    // 1. 初始化（程序启动时只需一次）
    redis := redisTool.Builder("127.0.0.1:6379", "").
        Config(redisTool.Config{
            Prefix: "myapp:",
        }).
        Build()
    redisTool.SetDefaultConnection(redis)
    
    // 2. 使用缓存
    cache := redisTool.NewCache[User]("users", redisTool.CacheConfig{
        DefaultExpire: time.Minute * 10,
    })
    cache.Set("user1", User{ID: "1", Name: "张三", Age: 18}, time.Minute*5)
    
    // 3. 使用分布式锁
    redisTool.NewLock("user:update").LockFunc(func() {
        fmt.Println("更新用户数据...")
    })
    
    // 4. 使用队列
    queue := redisTool.NewQueue[User]("tasks", redisTool.QueueConfig{
        MaxLength: 100,
    })
    queue.Add(User{ID: "2", Name: "李四", Age: 20})
    
    // 5. 定时任务检测
    if redisTool.AcrossMinute("task:sync") {
        fmt.Println("执行同步任务...")
    }
    
    // 6. 使用类型化数据结构
    userList := redisTool.NewTypeList[User]("online-users")
    userList.Push(User{ID: "3", Name: "王五", Age: 22})
    
    // 7. 使用有序集合
    ranking := redisTool.NewTypeZSet[User]("ranking")
    ranking.Add(User{ID: "4", Name: "赵六", Age: 19}, 1000.0)
    
    fmt.Println("所有操作完成！")
}
```

## 🔗 相关文档

- [README.md](./README.md) - 完整功能介绍
- [SUMMARY.md](./SUMMARY.md) - 项目总结
- [TESTING.md](./TESTING.md) - 测试指南

---

**提示**: 全局函数使用默认连接，确保在使用前调用 `SetDefaultConnection` 设置默认连接。
