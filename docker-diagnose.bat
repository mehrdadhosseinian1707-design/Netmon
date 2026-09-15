@echo off
echo ========================================
echo Docker Diagnostic Tool
echo ========================================
echo.

echo [System Information]
echo Windows Version:
ver
echo.

echo [Docker Installation]
where docker 2>nul
if %errorlevel% equ 0 (
    echo Docker CLI found
    docker --version 2>nul
) else (
    echo ERROR: Docker CLI not found in PATH
)
echo.

echo [Docker Desktop Process]
tasklist | findstr /I "docker" 2>nul
if %errorlevel% equ 0 (
    echo Docker processes are running
) else (
    echo WARNING: No Docker processes found
)
echo.

echo [Docker Service Status]
sc query com.docker.service 2>nul
echo.

echo [WSL Status]
wsl --status 2>nul
echo.

echo [WSL Distributions]
wsl --list --verbose 2>nul
echo.

echo [VHDX File Status]
set VHDX_PATH=C:\Users\%USERNAME%\AppData\Local\Docker\wsl\disk\docker_data.vhdx
echo Checking: %VHDX_PATH%
if exist "%VHDX_PATH%" (
    echo File exists
    dir "%VHDX_PATH%" 2>nul
    echo.
    echo File permissions:
    icacls "%VHDX_PATH%" 2>nul
) else (
    echo File NOT found
)
echo.

echo [Docker API Test]
docker info 2>nul
if %errorlevel% equ 0 (
    echo SUCCESS: Docker is responding
) else (
    echo ERROR: Cannot connect to Docker daemon
)
echo.

echo [Port Status]
netstat -ano | findstr ":2375" >nul 2>&1
if %errorlevel% equ 0 (
    echo Docker port 2375 is in use
) else (
    echo Docker port 2375 is free
)
echo.

echo ========================================
echo Diagnostic Complete
echo ========================================
echo.
echo If you see errors above:
echo - VHDX Access Denied: Run fix-docker.bat as Administrator
echo - WSL issues: Run: wsl --shutdown, then restart Docker
echo - Complete failure: Run fix-docker-reset.bat as Administrator
echo.
pause
