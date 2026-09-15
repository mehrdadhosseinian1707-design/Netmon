@echo off
echo ========================================
echo Docker Desktop COMPLETE RESET
echo WARNING: This will DELETE all Docker data!
echo ========================================
echo.
echo This will:
echo - Stop Docker Desktop completely
echo - Unregister WSL distributions
echo - Delete all Docker data (containers, images, volumes)
echo - Reset to factory defaults
echo.
echo Press Ctrl+C to cancel or
pause

echo.
echo [1/8] Stopping Docker Desktop...
taskkill /IM "Docker Desktop.exe" /F 2>nul
taskkill /IM "com.docker.backend.exe" /F 2>nul
taskkill /IM "dockerd.exe" /F 2>nul
timeout /t 3 /nobreak >nul

echo [2/8] Stopping Docker services...
net stop com.docker.service 2>nul
net stop docker 2>nul
timeout /t 2 /nobreak >nul

echo [3/8] Shutting down WSL...
wsl --shutdown
timeout /t 3 /nobreak >nul

echo [4/8] Unregistering Docker WSL distributions...
wsl --unregister docker-desktop 2>nul
wsl --unregister docker-desktop-data 2>nul
timeout /t 2 /nobreak >nul

echo [5/8] Deleting Docker data directories...
echo This may take a minute...

if exist "%LOCALAPPDATA%\Docker" (
    echo Removing %LOCALAPPDATA%\Docker
    rmdir /S /Q "%LOCALAPPDATA%\Docker" 2>nul
)

if exist "%APPDATA%\Docker" (
    echo Removing %APPDATA%\Docker
    rmdir /S /Q "%APPDATA%\Docker" 2>nul
)

if exist "%PROGRAMDATA%\Docker" (
    echo Removing %PROGRAMDATA%\Docker
    rmdir /S /Q "%PROGRAMDATA%\Docker" 2>nul
)

echo [6/8] Cleaning registry entries...
reg delete "HKCU\Software\Docker Inc." /f 2>nul

echo [7/8] Removing WSL ext4 files...
if exist "C:\Users\%USERNAME%\AppData\Local\Docker\wsl" (
    rmdir /S /Q "C:\Users\%USERNAME%\AppData\Local\Docker\wsl" 2>nul
)

echo [8/8] Starting Docker Desktop fresh...
echo This will take 30-60 seconds...
start "" "C:\Program Files\Docker\Docker\Docker Desktop.exe"
timeout /t 30 /nobreak >nul

echo.
echo ========================================
echo Reset Complete!
echo ========================================
echo.
echo Docker Desktop is starting fresh.
echo.
echo Next steps:
echo 1. Wait for Docker Desktop to fully start (check system tray)
echo 2. Accept any terms of service if prompted
echo 3. Once running, test with: docker ps
echo 4. Then run: start-all.bat
echo.
pause
