@echo off
echo ========================================
echo NetMon - Network Monitoring Platform
echo Starting All Services
echo ========================================
echo.

REM Check if Docker Desktop is running
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker Desktop is not running!
    echo.
    echo Please start Docker Desktop first, then run this script again.
    echo.
    pause
    exit /b 1
)

echo [1/4] Starting Docker containers (PostgreSQL + Redis)...
docker-compose up -d
if %errorlevel% neq 0 (
    echo [ERROR] Failed to start Docker containers
    pause
    exit /b 1
)

echo.
echo [2/4] Waiting for database to be ready...
timeout /t 5 /nobreak >nul

echo [3/4] Starting backend server...
start "NetMon Backend" cmd /k "cd backend && go run cmd/server/main.go"

echo.
echo [4/4] Starting frontend (this may take a moment on first run)...
timeout /t 3 /nobreak >nul
start "NetMon Frontend" cmd /k "cd frontend && npm start"

echo.
echo ========================================
echo All services started successfully!
echo ========================================
echo.
echo Backend:  http://localhost:8080
echo Frontend: http://localhost:3000
echo.
echo Press any key to view logs or close this window...
pause >nul
