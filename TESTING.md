# 测试文档

## 测试覆盖率目标

本项目的单元测试覆盖率目标为 **≥ 90%**。

## 运行测试

### Windows

```bash
# 运行所有测试
.\run_tests.bat

# 或手动运行
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linux/macOS

```bash
# 运行所有测试
chmod +x run_tests.sh
./run_tests.sh

# 或手动运行
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 测试文件说明

### 核心测试文件

- **test_helper.go** - 测试辅助工具，提供 miniredis 模拟 Redis 服务器
- **serializer_test.go** - 序列化器测试
- **redis_test.go** - Redis 客户端和 Builder 测试
- **list_test.go** - List 和 TypeList 功能测试
- **set_test.go** - Set 和 TypeSet 功能测试
- **map_test.go** - Map、TypeMap 和 NumberMap 功能测试
- **zset_test.go** - ZSet 和 TypeZSet 功能测试
- **queue_test.go** - Queue 队列功能测试
- **cache_test.go** - Cache 缓存功能测试
- **lock_test.go** - Lock 分布式锁功能测试
- **helper_test.go** - 辅助工具函数测试
- **global_test.go** - 全局泛型函数测试

## 测试覆盖的场景

### 1. 正常功能测试
- 基本的 CRUD 操作
- 数据结构的各种方法
- 类型转换和序列化

### 2. 边界条件测试
- 空列表/集合/映射操作
- 索引越界
- 不存在的键

### 3. 异常情况测试
- 无效的输入
- 序列化/反序列化错误
- 连接失败

### 4. 并发测试
- 分布式锁的并发访问
- 多个工作线程处理队列
- SafeUpset 原子操作

### 5. 过期和清理测试
- Cache 过期机制
- 自动清理过期数据
- Lock 自动释放

## 使用 miniredis

测试使用 [miniredis](https://github.com/alicebob/miniredis) 模拟 Redis 服务器，无需真实的 Redis 实例。

### 优点
- **快速** - 不需要启动外部 Redis 服务
- **隔离** - 每个测试独立运行
- **时间控制** - 可以快进时间测试过期功能

### 示例

```go
func TestExample(t *testing.T) {
    tr := NewTestRedis(t)
    defer tr.Close()

    // 使用 tr.Redis 进行测试
    list := tr.Redis.NewList("test")
    list.Push("value")

    // 快进时间
    tr.FastForward(60) // 快进 60 秒

    // 清空数据
    tr.FlushAll()
}
```

## 全局函数使用说明

由于 Go 语言的限制，方法不能有类型参数，因此提供了两种使用方式：

### 方式一：使用 Redis 实例的方法

```go
redis := redisTool.Builder("127.0.0.1:6379", "").Build()
typeList := redis.NewTypeList[Student]("students")
```

### 方式二：使用全局函数

```go
// 设置默认连接
redisTool.SetDefaultConnection(redis)

// 使用默认连接
typeList := redisTool.NewTypeList[Student]("students")

// 或指定连接
typeList := redisTool.NewTypeList[Student]("students", redis)
```

## 测试最佳实践

1. **独立性** - 每个测试应该独立运行，不依赖其他测试
2. **清理** - 使用 `defer tr.Close()` 确保资源被释放
3. **命名** - 测试函数名应清晰描述测试内容
4. **断言** - 使用明确的错误消息，便于定位问题
5. **覆盖率** - 确保测试覆盖正常流程和异常流程

## 持续集成

建议在 CI/CD 流程中运行测试：

```yaml
# .github/workflows/test.yml 示例
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out
```

## 报告问题

如果发现测试失败或覆盖率不足，请：
1. 查看错误信息
2. 运行单个测试：`go test -v -run TestName`
3. 添加调试输出
4. 提交 issue 或 PR
