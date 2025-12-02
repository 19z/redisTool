@echo off
setlocal enabledelayedexpansion

echo ========================================
echo Redis Tool - Docker Integration Tests
echo ========================================
echo.

cd /d "%~dp0"

REM 检查 Docker 是否运行
docker info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not running. Please start Docker Desktop.
    pause
    exit /b 1
)

REM 选择测试环境
echo Select test environment:
echo 1. Redis 7 (recommended)
echo 2. KVRocks (Redis-compatible)
echo 3. Both
echo.
set /p choice="Enter choice (1-3): "

if "%choice%"=="1" set TEST_ENV=redis
if "%choice%"=="2" set TEST_ENV=kvrocks
if "%choice%"=="3" set TEST_ENV=both

if not defined TEST_ENV (
    echo Invalid choice!
    pause
    exit /b 1
)

echo.
echo Starting Docker containers...
docker-compose up -d
if errorlevel 1 (
    echo [ERROR] Failed to start Docker containers
    pause
    exit /b 1
)

echo.
echo Waiting for Redis to be ready...
timeout /t 3 /nobreak >nul

REM 测试连接
docker exec redistool-test-redis redis-cli ping >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Redis is not responding
    docker-compose down
    pause
    exit /b 1
)

echo [OK] Redis is ready!
echo.

REM 运行测试
if "%TEST_ENV%"=="redis" goto test_redis
if "%TEST_ENV%"=="kvrocks" goto test_kvrocks
if "%TEST_ENV%"=="both" goto test_both

:test_redis
echo ========================================
echo Testing with Redis 7
echo ========================================
set USE_REAL_REDIS=1
set REDIS_ADDR=localhost:16379
go test -v -timeout 30s ./...
set TEST_RESULT=%ERRORLEVEL%
goto cleanup

:test_kvrocks
echo ========================================
echo Testing with KVRocks
echo ========================================
set USE_REAL_REDIS=1
set REDIS_ADDR=localhost:16380
go test -v -timeout 30s ./...
set TEST_RESULT=%ERRORLEVEL%
goto cleanup

:test_both
echo ========================================
echo Testing with Redis 7
echo ========================================
set USE_REAL_REDIS=1
set REDIS_ADDR=localhost:16379
go test -v -timeout 30s ./...
set REDIS_RESULT=%ERRORLEVEL%

echo.
echo ========================================
echo Testing with KVRocks
echo ========================================
set REDIS_ADDR=localhost:16380
go test -v -timeout 30s ./...
set KVROCKS_RESULT=%ERRORLEVEL%

if %REDIS_RESULT% NEQ 0 set TEST_RESULT=1
if %KVROCKS_RESULT% NEQ 0 set TEST_RESULT=1
if not defined TEST_RESULT set TEST_RESULT=0

:cleanup
echo.
echo ========================================
echo Cleaning up...
echo ========================================
docker-compose down

echo.
if %TEST_RESULT% EQU 0 (
    echo [SUCCESS] All tests passed!
    echo.
) else (
    echo [FAILED] Some tests failed!
    echo.
)

pause
exit /b %TEST_RESULT%
