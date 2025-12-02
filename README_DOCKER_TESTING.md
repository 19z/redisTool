# Docker 真实 Redis 测试指南

## 🚀 快速开始

### 自动化测试（推荐）

```bash
# Windows
.\test_with_docker.bat

# Linux/Mac
chmod +x test_with_docker.sh
./test_with_docker.sh
```

脚本会自动：
1. ✅ 检查 Docker 是否运行
2. ✅ 启动 Redis/KVRocks 容器
3. ✅ 等待服务就绪
4. ✅ 运行完整测试
5. ✅ 清理容器

### 手动测试

#### 1. 启动 Redis 容器

```bash
docker-compose up -d
```

#### 2. 运行测试

```bash
# Windows
set USE_REAL_REDIS=1
set REDIS_ADDR=localhost:16379
go test -v ./...

# Linux/Mac
export USE_REAL_REDIS=1
export REDIS_ADDR=localhost:16379
go test -v ./...
```

#### 3. 清理

```bash
docker-compose down
```

## 📦 可用服务

### Redis 7 (端口 16379)
```bash
# 连接测试
docker exec redistool-test-redis redis-cli ping

# 手动测试
redis-cli -p 16379
```

### KVRocks (端口 16380)
```bash
# 连接测试
docker exec redistool-test-kvrocks redis-cli -p 6666 ping

# 手动测试
redis-cli -p 16380
```

## 🎯 测试模式对比

| 特性 | miniredis | 真实 Redis | 说明 |
|------|-----------|-----------|------|
| 启动速度 | ⚡ 极快 | 🐢 较慢 | miniredis 是内存模拟 |
| TTL/过期 | ❌ 不支持 | ✅ 完整支持 | miniredis 限制 |
| 延迟队列 | ❌ 部分支持 | ✅ 完整支持 | 依赖 TTL 功能 |
| 时间快进 | ✅ 支持 | ❌ 不支持 | miniredis 特性 |
| 测试通过率 | 96.3% (157/163) | 100% (163/163) | 6个测试在 miniredis 失败 |
| 适用场景 | 开发/CI | 集成测试 | 根据需求选择 |

## 📊 测试结果预期

### miniredis 模式
```
总测试: 163
通过: 157 (96.3%)
失败: 6 (TTL相关)
```

### 真实 Redis 模式
```
总测试: 163
通过: 163 (100%)
失败: 0
```

## ⚠️ miniredis 限制的测试

以下测试在 miniredis 模式下会失败，在真实 Redis 中正常：

1. **TestCache_ClearExpired** - Cache 自动过期清理
2. **TestCache_Expiration** - Cache 键自动过期
3. **TestAcrossMinute** - 跨分钟时间检测
4. **TestAcrossTime** - 跨时间间隔检测
5. **TestQueue_AddDelayed** - 队列延迟任务处理
6. **TestQueue_Complete** - 队列任务完成流程

这些测试依赖 Redis 的 TTL/EXPIREAT 自动过期机制，miniredis 不完全支持。

## 🔧 故障排查

### Docker 未运行
```
[ERROR] Docker is not running
```
**解决**: 启动 Docker Desktop

### 端口被占用
```
Error: Port 16379 is already in use
```
**解决**: 
```bash
# 查看占用端口的进程
netstat -ano | findstr 16379  # Windows
lsof -i :16379               # Linux/Mac

# 停止旧容器
docker-compose down
```

### Redis 连接失败
```
[ERROR] Redis is not responding
```
**解决**:
```bash
# 查看容器状态
docker ps

# 查看日志
docker logs redistool-test-redis

# 重启容器
docker-compose restart
```

### 测试数据残留
```bash
# 清空 Redis
docker exec redistool-test-redis redis-cli FLUSHALL

# 或重建容器
docker-compose down
docker-compose up -d
```

## 📝 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `USE_REAL_REDIS` | 启用真实 Redis | `0` (使用 miniredis) |
| `REDIS_ADDR` | Redis 服务器地址 | `localhost:16379` |

## 🔍 调试技巧

### 查看测试详情
```bash
go test -v -run TestCache_Expiration ./...
```

### 只运行特定测试包
```bash
go test -v ./cache_test.go
```

### 启用 race 检测
```bash
go test -race ./...
```

### 生成覆盖率报告
```bash
export USE_REAL_REDIS=1
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🎬 CI/CD 集成

### GitHub Actions 示例

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test-real-redis:
    runs-on: ubuntu-latest
    
    services:
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests with real Redis
        env:
          USE_REAL_REDIS: 1
          REDIS_ADDR: localhost:6379
        run: |
          cd backend/internal/utils/redisTool
          go test -v -timeout 30s ./...
      
      - name: Generate coverage
        env:
          USE_REAL_REDIS: 1
          REDIS_ADDR: localhost:6379
        run: |
          cd backend/internal/utils/redisTool
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
```

## 💡 最佳实践

### 开发阶段
- ✅ 使用 miniredis（快速反馈）
- ✅ 使用 `run_tests.bat/sh`
- ✅ 忽略 TTL 相关测试失败

### 集成测试
- ✅ 使用真实 Redis
- ✅ 使用 `test_with_docker.bat/sh`
- ✅ 确保所有测试通过

### CI/CD 流程
- ✅ 快速测试用 miniredis
- ✅ 完整测试用真实 Redis
- ✅ 定期运行两种模式

## 📚 相关文档

- [TESTING.md](./TESTING.md) - 完整测试指南
- [README.md](./README.md) - 项目使用文档
- [docker-compose.yml](./docker-compose.yml) - Docker 配置

## 🤝 贡献

发现问题或有改进建议？欢迎：
1. 提交 Issue
2. 创建 Pull Request
3. 更新文档

---

**提示**: 首次运行 Docker 测试需要下载镜像，可能需要几分钟。后续运行会很快。
