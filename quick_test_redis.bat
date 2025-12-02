@echo off
echo Testing with Real Redis...
echo.

REM 确保 Redis 正在运行
docker ps | findstr "redistool-test-redis" >nul
if errorlevel 1 (
    echo Starting Redis container...
    docker-compose up -d redis
    timeout /t 3 /nobreak >nul
)

REM 测试连接
docker exec redistool-test-redis redis-cli ping >nul 2>&1
if errorlevel 1 (
    echo ERROR: Redis is not responding
    exit /b 1
)

echo Redis is ready!
echo.

set USE_REAL_REDIS=1
set REDIS_ADDR=localhost:16379

C:\Development\.golang\gopath\pkg\mod\golang.org\toolchain@v0.0.1-go1.25.0.windows-amd64\bin\go.exe test -v -timeout 60s
