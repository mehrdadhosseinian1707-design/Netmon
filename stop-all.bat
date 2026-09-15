@echo off
echo ========================================
echo NetMon - Stopping All Services
echo ========================================
echo.

echo [1/2] Stopping Docker containers...
docker-compose down

echo [2/2] Stopping Go and Node processes...
taskkill /FI "WindowTitle eq NetMon Backend*" /T /F 2>nul
taskkill /FI "WindowTitle eq NetMon Frontend*" /T /F 2>nul

echo.
echo All services stopped!
pause
