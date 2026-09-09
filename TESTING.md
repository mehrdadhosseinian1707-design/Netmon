# NetMon Testing Guide

## Quick Test Without Docker (Simplified)

If Go is not installed or you want to verify the code structure first, follow these steps:

### Step 1: Verify Project Structure

```bash
cd D:\netmon

# List backend files
dir backend\cmd\server\main.go
dir backend\cmd\migrate\main.go
dir backend\internal\api\api.go

# List frontend files
dir frontend\src\App.tsx
dir frontend\src\pages\Dashboard.tsx
dir frontend\package.json
```

### Step 2: Check Prerequisites

```powershell
# Check if Go is installed
go version
# Expected: go version go1.21.x or higher

# Check if Node.js is installed
node --version
# Expected: v18.x or higher

# Check if Docker is running
docker ps
# Should list running containers or show empty list

# Check if npm is installed
npm --version
# Expected: 9.x or higher
```

### Step 3: Install Missing Prerequisites

**If Go is not installed:**
1. Download from https://go.dev/dl/
2. Install and restart terminal
3. Verify: `go version`

**If Node.js is not installed:**
1. Download from https://nodejs.org/ (LTS version)
2. Install and restart terminal
3. Verify: `node --version`

**If Docker Desktop is not running:**
1. Download from https://docker.com/products/docker-desktop/
2. Install and start Docker Desktop
3. Verify: `docker ps`

---

## Full System Test (Requires All Prerequisites)

### Test 1: Start Database

```powershell
cd D:\netmon

# Start PostgreSQL and Redis
docker-compose up -d

# Wait 10 seconds for database to initialize
timeout /t 10

# Check containers are running
docker ps

# Should show:
# - netmon-postgres (port 5432)
# - netmon-redis (port 6379)

# View database logs
docker logs netmon-postgres

# Test database connection
docker exec -it netmon-postgres psql -U netmon -d netmon -c "SELECT version();"
```

### Test 2: Run Database Migrations

```powershell
cd D:\netmon\backend

# Run migrations
go run cmd/migrate/main.go up

# Expected output:
# "Migrations completed successfully"

# Verify tables were created
docker exec -it netmon-postgres psql -U netmon -d netmon -c "\dt"

# Should list 15+ tables including:
# - probes
# - targets
# - monitors
# - measurements
# - alerts
# - incidents
```

### Test 3: Start Backend Server

```powershell
# In Terminal 1
cd D:\netmon\backend

# Download Go dependencies (first time only)
go mod download
go mod tidy

# Start the server
go run cmd/server/main.go

# Expected output:
# {"level":"info","msg":"starting netmon server","version":"0.2.0"}
# {"level":"info","msg":"connected to database"}
# {"level":"info","msg":"starting HTTP server","addr":"0.0.0.0:8080"}

# Server is now running on http://localhost:8080
```

### Test 4: Test Backend API

```powershell
# In Terminal 2 (keep Terminal 1 running)

# Test health check
curl http://localhost:8080/health

# Expected: {"status":"healthy","time":"2026-09-08T..."}

# Create a probe
curl -X POST http://localhost:8080/api/v1/probes -H "Content-Type: application/json" -d "{\"name\":\"test-probe\",\"location\":\"local\",\"description\":\"Test probe\"}"

# Expected: Returns JSON with probe ID and API key

# List probes
curl http://localhost:8080/api/v1/probes

# Create a target
curl -X POST http://localhost:8080/api/v1/targets -H "Content-Type: application/json" -d "{\"name\":\"Google DNS\",\"target_type\":\"host\",\"address\":\"8.8.8.8\"}"

# Expected: Returns JSON with target ID
# Copy the "id" field (UUID) for next step

# List targets
curl http://localhost:8080/api/v1/targets

# Create an ICMP monitor (replace <TARGET_ID> with actual UUID)
curl -X POST http://localhost:8080/api/v1/monitors -H "Content-Type: application/json" -d "{\"target_id\":\"<TARGET_ID>\",\"monitor_type\":\"icmp\",\"interval_seconds\":60,\"timeout_seconds\":10,\"retries\":3,\"config\":{\"count\":4}}"

# Expected: Returns JSON with monitor ID

# Wait 60 seconds for first measurement

# Get target statistics (replace <TARGET_ID>)
curl http://localhost:8080/api/v1/targets/<TARGET_ID>/stats

# Expected: JSON with availability, avg_latency_ms, etc.

# Get recent measurements (replace <TARGET_ID>)
curl http://localhost:8080/api/v1/targets/<TARGET_ID>/measurements

# Expected: Array of measurement objects with latency, packet_loss, etc.
```

### Test 5: Start Frontend

```powershell
# In Terminal 3 (keep Terminal 1 and 2 running)
cd D:\netmon\frontend

# Install dependencies (first time only - may take 2-5 minutes)
npm install

# Start development server
npm start

# Expected output:
# Compiled successfully!
# webpack compiled with 0 warnings
# 
# The app is running at:
#   http://localhost:3000

# Browser should automatically open to http://localhost:3000
```

### Test 6: Verify Frontend

**In Browser at http://localhost:3000:**

1. **Dashboard Page:**
   - Should load without errors
   - Should show "Total Targets: 1"
   - Should show stats boxes
   - May show "0" for UP/DOWN until measurements arrive

2. **Click "Targets" in navbar:**
   - Should show "Google DNS" target card
   - Should show address "8.8.8.8"
   - Should show green status indicator (●)

3. **Click on "Google DNS" card:**
   - Should navigate to target detail page
   - Should show 6 statistics boxes
   - After 60+ seconds, should show data in charts
   - Should show recent checks table

4. **Wait 60 seconds, then refresh:**
   - Availability should be ~100%
   - Avg Latency should show (e.g., 15.32 ms)
   - Charts should populate with data points

5. **Check browser console (F12):**
   - Should see API calls succeeding
   - Should see WebSocket connection attempt
   - No red errors

---

## Troubleshooting

### Issue: "go: command not found"
**Solution:** Install Go from https://go.dev/dl/

### Issue: "docker: command not found"
**Solution:** Install Docker Desktop from https://docker.com

### Issue: "Database connection failed"
```powershell
# Check Docker containers
docker ps -a

# Restart PostgreSQL
docker restart netmon-postgres

# Check logs
docker logs netmon-postgres

# Verify port 5432 is not in use
netstat -ano | findstr :5432
```

### Issue: "Port 8080 already in use"
```powershell
# Find process using port 8080
netstat -ano | findstr :8080

# Kill process (replace <PID> with actual PID)
taskkill /PID <PID> /F

# Or change port in backend config
$env:NETMON_SERVER_PORT=8081
go run cmd/server/main.go
```

### Issue: "Permission denied" (ICMP)
**Windows:** Run terminal as Administrator
**Linux:** Run `sudo setcap cap_net_raw=+ep ./bin/server`

### Issue: Frontend "npm install" fails
```powershell
# Clear npm cache
npm cache clean --force

# Delete node_modules and try again
Remove-Item -Recurse -Force node_modules
npm install
```

### Issue: Frontend shows "Proxy error"
- Ensure backend is running on port 8080
- Check `package.json` has `"proxy": "http://localhost:8080"`
- Restart frontend: Ctrl+C, then `npm start`

### Issue: No measurements appearing
```powershell
# Check scheduler logs in backend terminal
# Should see: "check completed" every 60 seconds

# Query database directly
docker exec -it netmon-postgres psql -U netmon -d netmon -c "SELECT COUNT(*) FROM measurements;"

# If count is 0:
# 1. Ensure monitor is created
# 2. Wait full interval (60 seconds)
# 3. Check backend logs for errors
```

---

## Expected Results

After all tests:

✅ **Database:** PostgreSQL running with 15+ tables created  
✅ **Backend:** Server running on port 8080, API responding  
✅ **Frontend:** React app on port 3000, dashboard visible  
✅ **Monitoring:** ICMP checks running every 60 seconds  
✅ **Data Flow:** Measurements → Database → API → Frontend → Charts  
✅ **Charts:** Latency and packet loss visualization working  

---

## Quick Verification Commands

```powershell
# Check all services
docker ps                                    # PostgreSQL, Redis running
curl http://localhost:8080/health            # Backend healthy
curl http://localhost:3000                   # Frontend serving

# Check data
curl http://localhost:8080/api/v1/targets   # Targets exist
curl http://localhost:8080/api/v1/monitors  # Monitors configured

# Check database
docker exec -it netmon-postgres psql -U netmon -d netmon -c "SELECT COUNT(*) FROM measurements;"
```

---

## Stop Everything

```powershell
# Stop frontend (in Terminal 3)
Ctrl+C

# Stop backend (in Terminal 1)
Ctrl+C

# Stop database
cd D:\netmon
docker-compose down

# Or stop without removing data
docker-compose stop
```

---

## Next Steps After Testing

Once everything is working:

1. **Explore the UI:**
   - Add more targets (different IPs, ports)
   - Create TCP monitors (port 443, 80, etc.)
   - Watch data populate in real-time

2. **Customize:**
   - Change monitor intervals
   - Adjust chart colors
   - Add new target types

3. **Continue Development:**
   - Phase 3: DNS, HTTP, TLS monitors
   - Add forms for creating targets/monitors in UI
   - Implement alert engine
   - Add user authentication

---

**Ready to start? Let's begin with checking prerequisites!**
