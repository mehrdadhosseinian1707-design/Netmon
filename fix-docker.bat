@echo off
echo ========================================
echo Docker Desktop VHDX Fix Script
echo ========================================
echo.
echo This will fix the "Access is denied" error on docker_data.vhdx
echo.
pause

echo [1/6] Stopping Docker Desktop...
taskkill /IM "Docker Desktop.exe" /F 2>nul
timeout /t 2 /nobreak >nul

echo [2/6] Stopping Docker services...
net stop com.docker.service 2>nul
net stop docker 2>nul
timeout /t 2 /nobreak >nul

echo [3/6] Stopping WSL...
wsl --shutdown
timeout /t 3 /nobreak >nul

echo [4/6] Checking file permissions...
set VHDX_PATH=C:\Users\%USERNAME%\AppData\Local\Docker\wsl\disk\docker_data.vhdx

if exist "%VHDX_PATH%" (
    echo Found VHDX file: %VHDX_PATH%
    echo Taking ownership...
    takeown /F "%VHDX_PATH%" /A
    icacls "%VHDX_PATH%" /grant %USERNAME%:F
) else (
    echo VHDX file not found at expected location
)

echo.
echo [5/6] Cleaning up Docker resources...
REM Clean Docker temp files
del /Q "%TEMP%\docker-*" 2>nul
rmdir /S /Q "%LOCALAPPDATA%\Docker\log" 2>nul

echo.
echo [6/6] Starting Docker Desktop...
echo Please wait 10-15 seconds for Docker to start...
start "" "C:\Program Files\Docker\Docker\Docker Desktop.exe"
timeout /t 15 /nobreak >nul

echo.
echo ========================================
echo Fix completed!
echo ========================================
echo.
echo Docker Desktop should now be starting.
echo Wait for "Docker Desktop is running" message.
echo.
echo If the error persists, run: fix-docker-reset.bat
echo.
pause
