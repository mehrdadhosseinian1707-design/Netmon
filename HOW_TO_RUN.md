# NetMon - Quick Start Guide

## Prerequisites

Before you start, make sure you have:

1. **Docker Desktop** - [Download here](https://www.docker.com/products/docker-desktop/)
   - ⚠️ **IMPORTANT**: Docker Desktop MUST be running before starting NetMon
   - After installing, launch Docker Desktop and wait for it to fully start

2. **Go 1.24+** - [Download here](https://go.dev/dl/)

3. **Node.js 16+** - [Download here](https://nodejs.org/)

---

## Method 1: Quick Start (Recommended - Windows)

### Step 1: Start Docker Desktop
1. Open Docker Desktop application
2. Wait until you see "Docker Desktop is running" in the system tray
3. Verify by opening Command Prompt and running:
   ```bash
   docker ps
   ```
   You should see a table header (even if no containers are running yet)

### Step 2: Run Everything at Once
```bash
cd D:\netmon
start-all.bat
```

This script will:
- ✅ Check if Docker Desktop is running
- ✅ Start PostgreSQL + Redis containers
- ✅ Start the backend server (with automatic migrations)
- ✅ Start the frontend development server
- ✅ Open your browser to http://localhost:3000

**Note:** First time setup takes 2-3 minutes while npm installs dependencies.

### To Stop All Services:
```bash
stop-all.bat
```

---

## Method 2: Manual Start (Step-by-Step)

### Step 1: Start Docker Desktop
⚠️ **CRITICAL**: Open Docker Desktop application and wait for it to be fully running

**Verify Docker is running:**
```bash
docker --version
docker ps
```

### Step 2: Start Database (PostgreSQL + Redis)
```bash
cd D:\netmon
docker-compose up -d
```

**What this does:**
- Starts PostgreSQL with TimescaleDB extension
- Starts Redis cache
- Creates persistent volumes for data storage

**Verify containers are running:**
```bash
docker ps
```
You should see:
- `netmon-postgres` (port 5432)
- `netmon-redis` (port 6379)

**Check container logs if needed:**
```bash
docker logs netmon-postgres
docker logs netmon-redis
```

---

### Step 3: Start Backend Server
Open a **NEW terminal window** and run:
```bash
cd D:\netmon\backend
go run cmd/server/main.go
```

**What this does:**
- Connects to PostgreSQL database
- Runs database migrations automatically (creates all tables)
- Starts REST API server on http://localhost:8080
- Initializes monitoring scheduler
- Ready to accept probe connections

**Expected output:**
```
{"level":"info","msg":"starting netmon server","version":"0.2.0","mode":"release"}
{"level":"info","msg":"connected to database"}
{"level":"info","msg":"running database migrations"}
Applying migration 001: initial_schema
{"level":"info","msg":"migrations completed successfully"}
{"level":"info","msg":"starting HTTP server","addr":"0.0.0.0:8080"}
```

**⚠️ Keep this terminal window open!**

---

### Step 4: Start Frontend Dashboard
Open **ANOTHER NEW terminal window** and run:
```bash
cd D:\netmon\frontend

# First time only (installs dependencies - takes 2-3 minutes)
npm install

# Start development server
npm start
```

**What this does:**
- Installs React dependencies (first time only)
- Starts development server on http://localhost:3000
- Opens browser automatically
- Hot-reloads on code changes

**Expected output:**
```
Compiled successfully!

You can now view netmon-frontend in the browser.

  Local:            http://localhost:3000
  On Your Network:  http://192.168.x.x:3000
```

**Browser will open automatically to:** http://localhost:3000

---

## Verification

Once all services are running, verify:

1. **Backend API:**
   ```bash
   curl http://localhost:8080/api/health
   ```
   Should return: `{"status":"ok"}`

2. **Frontend:**
   Open http://localhost:3000 in your browser
   You should see the NetMon dashboard

3. **Database:**
   ```bash
   docker exec -it netmon-postgres psql -U netmon -d netmon -c "\dt"
   ```
   Should list tables: probes, targets, monitors, measurements, incidents, alerts

---

## Stopping Services

### Quick Method (Windows):
```bash
stop-all.bat
```

### Manual Method:

1. **Stop Frontend:** Press `Ctrl+C` in the frontend terminal

2. **Stop Backend:** Press `Ctrl+C` in the backend terminal

3. **Stop Docker containers:**
   ```bash
   docker-compose down
   ```

4. **To also remove data volumes:**
   ```bash
   docker-compose down -v
   ```
   ⚠️ Warning: This deletes all monitoring data!

---

## Troubleshooting

### Problem: Docker VHDX "Access is denied" Error

**Error:** `creating vhdx: getting VHDX metadata... Access is denied`

This is the most common Docker Desktop error on Windows. The VHDX file is locked or corrupted.

**Solution 1 - Quick Fix (Try this first):**
```bash
cd D:\netmon

# Right-click and "Run as Administrator"
fix-docker.bat
```

This will:
- Stop Docker Desktop
- Shutdown WSL
- Fix file permissions on the VHDX file
- Restart Docker Desktop

**Solution 2 - Complete Reset (if Solution 1 fails):**
```bash
cd D:\netmon

# Right-click and "Run as Administrator"  
fix-docker-reset.bat
```

⚠️ **WARNING**: This deletes ALL Docker data (containers, images, volumes)

This will:
- Completely uninstall Docker WSL distributions
- Delete all Docker data
- Reset Docker Desktop to factory defaults
- Restart fresh

**Solution 3 - Manual Fix:**
1. Close Docker Desktop completely
2. Open PowerShell as Administrator
3. Run:
   ```powershell
   wsl --shutdown
   wsl --unregister docker-desktop
   wsl --unregister docker-desktop-data
   ```
4. Delete: `C:\Users\YOUR_USERNAME\AppData\Local\Docker\wsl`
5. Start Docker Desktop again

---

### Problem: "Docker is not running" or "cannot connect to Docker daemon"

**Solution:**
1. Open Docker Desktop application
2. Wait until it says "Docker Desktop is running" (bottom-left corner)
3. Run the start script again

**Still not working?**
```bash
# Run diagnostic tool
docker-diagnose.bat

# Or restart Docker Desktop service (Windows)
# 1. Open Services (Win + R, type "services.msc")
# 2. Find "Docker Desktop Service"
# 3. Right-click → Restart
```

---

### Problem: Docker containers fail to start

**Solution:**
```bash
# Check what's wrong
docker-compose logs

# Stop and remove everything
docker-compose down -v

# Start fresh
docker-compose up -d

# Check status
docker ps
```

---

### Problem: Backend fails with "failed to connect to database"

**Solution:**
```bash
# Check if PostgreSQL is ready
docker logs netmon-postgres

# Wait a few more seconds, then try again
# On first start, PostgreSQL takes 5-10 seconds to initialize
```

---

### Problem: Port already in use (8080 or 3000)

**Solution:**
```bash
# Find what's using the port
netstat -ano | findstr :8080
netstat -ano | findstr :3000

# Kill the process (replace PID with actual process ID)
taskkill /PID <PID> /F
```

---

### Problem: Frontend "Module not found" errors

**Solution:**
```bash
cd D:\netmon\frontend

# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install
npm start
```

---

### Problem: Database tables not created

**Solution:**
The backend automatically runs migrations on startup. If tables are missing:

1. Check backend logs for migration errors
2. Manually verify database:
   ```bash
   docker exec -it netmon-postgres psql -U netmon -d netmon
   \dt
   \q
   ```
3. If tables are missing, restart the backend - migrations run on every startup

---

## Architecture Overview

```
┌─────────────────────────────────────────────────┐
│  Frontend (React + TypeScript)                  │
│  http://localhost:3000                          │
└─────────────────┬───────────────────────────────┘
                  │
                  │ REST API
                  ▼
┌─────────────────────────────────────────────────┐
│  Backend (Go)                                   │
│  http://localhost:8080                          │
│  - REST API Server (Gin)                        │
│  - Monitoring Scheduler                         │
│  - Database Migrations                          │
└─────────────────┬───────────────────────────────┘
                  │
        ┌─────────┴──────────┐
        ▼                    ▼
┌─────────────┐      ┌─────────────┐
│ PostgreSQL  │      │   Redis     │
│ + TimescaleDB      │ Cache       │
│ :5432       │      │ :6379       │
└─────────────┘      └─────────────┘
```

---

## Default Credentials

**PostgreSQL:**
- Host: localhost
- Port: 5432
- Database: netmon
- Username: netmon
- Password: netmon_dev_password

**Redis:**
- Host: localhost
- Port: 6379
- Password: (none)

---

## Next Steps

1. **Access the Dashboard:** http://localhost:3000
2. **Add Monitoring Targets:** Navigate to "Targets" → "Add Target"
3. **Configure Probes:** Set up monitoring probes in "Probes" section
4. **View Metrics:** Check Dashboard for real-time monitoring data

---

## Development Notes

- **Hot Reload:** Frontend auto-reloads on code changes
- **Backend Changes:** Restart `go run cmd/server/main.go` to apply
- **Database Changes:** Migrations run automatically on backend startup
- **Data Persistence:** Docker volumes persist data between restarts

---

## Need Help?

- Check the logs in each terminal window
- Review troubleshooting section above
- Ensure Docker Desktop is running before all else
