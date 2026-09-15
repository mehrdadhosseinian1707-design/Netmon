# Docker Troubleshooting Guide for NetMon

## Common Docker Desktop Errors on Windows

### Error 1: VHDX "Access is denied"
```
creating vhdx: getting VHDX metadata for virtual disk 
C:\Users\xxx\AppData\Local\Docker\wsl\disk\docker_data.vhdx: 
Access is denied.
```

**Cause:** The VHDX file is locked by another process or has incorrect permissions.

**Fix:**
```bash
# Run as Administrator
fix-docker.bat
```

If that doesn't work:
```bash
# Complete reset (deletes all Docker data)
# Run as Administrator
fix-docker-reset.bat
```

---

### Error 2: WSL Integration Issues
```
WSL 2 installation is incomplete
The WSL 2 Linux kernel is now installed using a separate MSI update package
```

**Fix:**
1. Download WSL2 kernel update: https://aka.ms/wsl2kernel
2. Install it
3. Restart Docker Desktop

---

### Error 3: Port Already in Use
```
Error starting userland proxy: listen tcp 0.0.0.0:5432: 
bind: address already in use
```

**Fix:**
```bash
# Find what's using the port
netstat -ano | findstr :5432

# Kill the process (replace PID)
taskkill /PID <PID> /F

# Or use different ports in docker-compose.yml
```

---

### Error 4: Docker Service Won't Start
```
Docker Desktop service is not running
```

**Fix:**
```bash
# Open Services (Win + R, type "services.msc")
# Find "Docker Desktop Service"
# Right-click → Start

# Or from PowerShell as Administrator:
net start com.docker.service
```

---

## Diagnostic Tools

### Quick Health Check
```bash
docker-diagnose.bat
```

This checks:
- Docker installation
- Docker processes
- WSL status
- VHDX file status and permissions
- Docker API connectivity

### Manual Diagnostics
```bash
# Check Docker version
docker --version

# Check if daemon is running
docker info

# Check WSL
wsl --list --verbose

# Should show:
# - docker-desktop         Running
# - docker-desktop-data    Running

# Check Docker processes
tasklist | findstr docker
```

---

## Complete Clean Reinstall (Last Resort)

If nothing else works:

### 1. Uninstall Docker Desktop
```bash
# From Windows Settings → Apps → Docker Desktop → Uninstall
```

### 2. Clean up WSL
```bash
wsl --list
wsl --unregister docker-desktop
wsl --unregister docker-desktop-data
```

### 3. Delete Docker directories
```bash
rmdir /S /Q "%LOCALAPPDATA%\Docker"
rmdir /S /Q "%APPDATA%\Docker"
rmdir /S /Q "%PROGRAMDATA%\Docker"
```

### 4. Reinstall Docker Desktop
Download from: https://www.docker.com/products/docker-desktop/

---

## Preventive Maintenance

### Keep Docker Healthy

1. **Regular cleanup:**
```bash
# Remove unused containers, images, volumes
docker system prune -a --volumes
```

2. **Don't force quit Docker Desktop**
   - Always close gracefully
   - Wait for "Docker Desktop stopped" message

3. **Keep WSL updated:**
```bash
wsl --update
```

4. **Allocate enough resources:**
   - Docker Desktop Settings → Resources
   - Minimum: 4GB RAM, 2 CPUs

---

## Alternative: Run Without Docker

If Docker continues to fail, you can run PostgreSQL and Redis directly on Windows:

### Install PostgreSQL
1. Download from: https://www.postgresql.org/download/windows/
2. Install with TimescaleDB extension
3. Create database:
```sql
CREATE DATABASE netmon;
CREATE USER netmon WITH PASSWORD 'netmon_dev_password';
GRANT ALL PRIVILEGES ON DATABASE netmon TO netmon;
```

### Install Redis
1. Download from: https://github.com/tporadowski/redis/releases
2. Extract and run: `redis-server.exe`

### Update backend/config.yaml
```yaml
database:
  host: localhost
  port: 5432
  user: netmon
  password: netmon_dev_password
  database: netmon
  sslmode: disable

redis:
  host: localhost
  port: 6379
```

Then start normally:
```bash
cd D:\netmon\backend
go run cmd/server/main.go
```

---

## Getting Help

If you're still stuck:

1. **Check Docker Desktop logs:**
   - Click Docker icon in system tray
   - Troubleshoot → View logs

2. **Check Windows Event Viewer:**
   - Win + R → eventvwr.msc
   - Windows Logs → Application
   - Filter for "Docker" events

3. **Run full diagnostic:**
```bash
docker-diagnose.bat > docker-diagnostic.txt
```
   Share the diagnostic.txt file when asking for help
