# 更新日志

## 2024-12-02 - 添加全局函数支持

### ✨ 新增功能

#### 1. Lock 全局函数
```go
// 使用全局函数创建分布式锁
lock := redisTool.NewLock("mylock", redisTool.LockConfig{
    WaitTime: time.Second * 5,
})
```

#### 2. 辅助工具全局函数

**时间追踪：**
```go
// 获取/设置上次使用时间
lastTime := redisTool.LastUseTime("mykey", true)
redisTool.SetLastUseTime("mykey", time.Now())
```

**跨时间检测：**
```go
// 检查是否跨越分钟
if redisTool.AcrossMinute("task1") {
    // 执行任务
}

// 检查是否跨越秒（新增）
if redisTool.AcrossSecond("task2") {
    // 执行任务
}

// 检查是否跨越指定时间
if redisTool.AcrossTime("task3", time.Hour) {
    // 执行任务
}
```

**安全类型映射：**
```go
// 获取安全类型映射
safeMap := redisTool.GetSafeTypeMap()
```

### 📝 完整的全局函数列表

#### 数据结构（已有）
- `NewTypeList[T any](name string, r ...*Redis) *RedisTypeList[T]`
- `NewTypeSet[T any](name string, r ...*Redis) *RedisTypeSet[T]`
- `NewTypeMap[T any](name string, r ...*Redis) *RedisTypeMap[T]`
- `NewTypeZSet[T any](name string, r ...*Redis) *RedisTypeZSet[T]`

#### 高级功能（已有 + 新增）
- `NewQueue[T any](name string, config QueueConfig, r ...*Redis) *Queue[T]`
- `NewCache[T any](name string, config CacheConfig, r ...*Redis) *Cache[T]`
- `NewLock(name string, config ...LockConfig) *Lock` ⭐ **新增**

#### 辅助工具（新增）
- `LastUseTime(key string, update bool) time.Time` ⭐ **新增**
- `AcrossMinute(key string) bool` ⭐ **新增**
- `AcrossSecond(key string) bool` ⭐ **新增**
- `AcrossTime(key string, duration time.Duration) bool` ⭐ **新增**
- `SetLastUseTime(key string, t time.Time)` ⭐ **新增**
- `GetSafeTypeMap() *RedisTypeMap[int64]` ⭐ **新增**

### 📚 文档更新

#### 1. README.md
- 更新所有示例使用全局函数（推荐方式）
- 修正 Lock 使用示例
- 添加 AcrossSecond 示例
- 完善辅助工具使用说明

#### 2. SUMMARY.md
- 更新全局函数列表
- 添加新增函数的说明
- 更新最佳实践示例
- 完善错误处理示例

#### 3. 新增文档
- `GLOBAL_FUNCTIONS.md` - 全局函数使用指南
- `CHANGES.md` - 更新日志（本文档）

### 🔧 代码修改

#### global.go
```go
// 新增 Lock 全局函数
func NewLock(name string, config ...LockConfig) *Lock {
    conn := defaultConnection
    return conn.NewLock(name, config...)
}

// 新增辅助工具全局函数
func LastUseTime(key string, update bool) time.Time { ... }
func AcrossMinute(key string) bool { ... }
func AcrossSecond(key string) bool { ... }
func AcrossTime(key string, duration time.Duration) bool { ... }
func SetLastUseTime(key string, t time.Time) { ... }
func GetSafeTypeMap() *RedisTypeMap[int64] { ... }
```

#### helper.go
```go
// 新增 AcrossSecond 方法
func (r *Redis) AcrossSecond(key string) bool {
    return r.AcrossTime(key, time.Second)
}
```

### 🎯 使用建议

#### 推荐的使用模式

**初始化（一次性设置）：**
```go
func init() {
    redis := redisTool.Builder("127.0.0.1:6379", "").Build()
    redisTool.SetDefaultConnection(redis)
}
```

**业务代码中使用：**
```go
// 分布式锁
lock := redisTool.NewLock("resource")
lock.LockFunc(func() {
    // 业务逻辑
})

// 定时任务
if redisTool.AcrossMinute("cron:task") {
    executeTask()
}

// 缓存
cache := redisTool.NewCache[User]("users", ...)
user := cache.GetOrSet("user1", loadUser)

// 队列
queue := redisTool.NewQueue[Task]("tasks", ...)
queue.Add(task)
```

### 📊 改进效果

#### 代码简化对比

**之前：**
```go
func ProcessOrder(redis *redisTool.Redis, orderID string) {
    lock := redis.NewLock("order:" + orderID)
    // ...
    
    if redis.AcrossMinute("task") {
        // ...
    }
}
```

**现在：**
```go
func ProcessOrder(orderID string) {
    lock := redisTool.NewLock("order:" + orderID)
    // ...
    
    if redisTool.AcrossMinute("task") {
        // ...
    }
}
```

### ✅ 测试结果

- ✅ 编译通过
- ✅ 所有测试保持一致（157/163 通过，96.3%）
- ✅ 新增函数可正常使用
- ✅ 向后兼容（原有实例方法仍可使用）

### 🔄 向后兼容性

**完全向后兼容！** 所有原有的实例方法仍然可用：

```go
// 仍然可以使用实例方法
redis := redisTool.Builder("127.0.0.1:6379", "").Build()
lock := redis.NewLock("mylock")
lastTime := redis.LastUseTime("key", true)
```

### 📖 相关文档

- [README.md](./README.md) - 完整使用指南
- [GLOBAL_FUNCTIONS.md](./GLOBAL_FUNCTIONS.md) - 全局函数专题
- [SUMMARY.md](./SUMMARY.md) - 项目总结
- [TESTING.md](./TESTING.md) - 测试指南

---

**总结**: 本次更新为 Lock 和所有辅助工具函数添加了全局函数支持，使得代码更加简洁，使用更加方便。只需在程序初始化时设置一次默认连接，就可以在任何地方直接使用这些全局函数。
