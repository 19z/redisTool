#!/bin/bash

set -e

echo "========================================"
echo "Redis Tool - Docker Integration Tests"
echo "========================================"
echo ""

cd "$(dirname "$0")"

# 检查 Docker 是否运行
if ! docker info >/dev/null 2>&1; then
    echo "[ERROR] Docker is not running. Please start Docker."
    exit 1
fi

# 选择测试环境
echo "Select test environment:"
echo "1. Redis 7 (recommended)"
echo "2. KVRocks (Redis-compatible)"
echo "3. Both"
echo ""
read -p "Enter choice (1-3): " choice

case $choice in
    1) TEST_ENV="redis" ;;
    2) TEST_ENV="kvrocks" ;;
    3) TEST_ENV="both" ;;
    *)
        echo "Invalid choice!"
        exit 1
        ;;
esac

echo ""
echo "Starting Docker containers..."
docker-compose up -d

echo ""
echo "Waiting for Redis to be ready..."
sleep 3

# 测试连接
if ! docker exec redistool-test-redis redis-cli ping >/dev/null 2>&1; then
    echo "[ERROR] Redis is not responding"
    docker-compose down
    exit 1
fi

echo "[OK] Redis is ready!"
echo ""

TEST_RESULT=0

# 运行测试函数
test_redis() {
    echo "========================================"
    echo "Testing with Redis 7"
    echo "========================================"
    export USE_REAL_REDIS=1
    export REDIS_ADDR=localhost:16379
    go test -v -timeout 30s ./... || return 1
    return 0
}

test_kvrocks() {
    echo "========================================"
    echo "Testing with KVRocks"
    echo "========================================"
    export USE_REAL_REDIS=1
    export REDIS_ADDR=localhost:16380
    go test -v -timeout 30s ./... || return 1
    return 0
}

# 根据选择运行测试
case $TEST_ENV in
    redis)
        test_redis || TEST_RESULT=1
        ;;
    kvrocks)
        test_kvrocks || TEST_RESULT=1
        ;;
    both)
        test_redis || TEST_RESULT=1
        echo ""
        test_kvrocks || TEST_RESULT=1
        ;;
esac

echo ""
echo "========================================"
echo "Cleaning up..."
echo "========================================"
docker-compose down

echo ""
if [ $TEST_RESULT -eq 0 ]; then
    echo "[SUCCESS] All tests passed!"
    echo ""
else
    echo "[FAILED] Some tests failed!"
    echo ""
fi

exit $TEST_RESULT
