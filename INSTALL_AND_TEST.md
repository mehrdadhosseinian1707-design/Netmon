# Installing Go and Testing NetMon - Step by Step Guide

**Current Time:** 2026-09-07 23:26:53 UTC  
**System:** Windows with Node.js 24.18.0, npm 11.16.0, Docker 29.7.2  
**Goal:** Get the complete NetMon system running

---

## Step 1: Install Go (5 minutes)

### Download Go

1. **Open your browser** and go to: https://go.dev/dl/

2. **Download the Windows installer:**
   - Look for: `go1.23.1.windows-amd64.msi` (or latest version)
   - File size: ~130 MB
   - Click to download

3. **Run the installer:**
   - Double-click the downloaded `.msi` file
   - Follow the installation wizard
   - Accept default installation path: `C:\Program Files\Go`
   - Click "Next" → "Install" → "Finish"

4. **Restart your terminal:**
   - Close all Git Bash / PowerShell / Command Prompt windows
   - Open a NEW terminal (important for PATH to update)

5. **Verify installation:**
   ```bash
   go version
   ```
   **Expected output:** `go version go1.23.1 windows/amd64` (or similar)

**✅ Checkpoint:** If you see the Go version, continue. If not, restart your computer and try again.

---

## Step 2: Start Docker Desktop (2 minutes)

1. **Open Docker Desktop:**
   - Press Windows key
   - Type "Docker Desktop"
   - Click to launch
   - Wait for Docker to start (whale icon in system tray should be solid)

2. **Verify Docker is running:**
   ```bash
   docker ps
   ```
   **Expected output:** Table header showing "CONTAINER ID   IMAGE   COMMAND..." (may be empty, that's OK)

**✅ Checkpoint:** If you see the table header, Docker is running.

---

## Step 3: Start PostgreSQL Database (2 minutes)

```bash
# Navigate to project
cd D:\netmon

# Start database containers
docker-compose up -d

# Wait 10 seconds for initialization
sleep 10

# Verify containers are running
docker ps
```

**Expected output:** You should see 2 containers:
- `netmon-postgres` (port 5432)
- `netmon-redis` (port 6379)

**✅ Checkpoint:** Both containers show "Up" status.

---

## Step 4: Run Database Migrations (1 minute)

```bash
# Navigate to backend
cd backend

# Install Go dependencies (first time only)
go mod download

# Run migrations
go run cmd/migrate/main.go up
```

**Expected output:**
```
Migrations completed successfully
```

**Verify tables were created:**
```bash
docker exec -it netmon-postgres psql -U netmon -d netmon -c "\dt"
```

**Expected output:** List of 15+ tables including `probes`, `targets`, `monitors`, `measurements`, `alerts`, `incidents`

**✅ Checkpoint:** Tables are listed.

---

## Step 5: Start Backend Server (1 minute)

**Open a NEW terminal window (Terminal 1) and run:**

```bash
cd D:\netmon\backend

# Start the server
go run cmd/server/main.go
```

**Expected output:**
```json
{"level":"info","msg":"starting netmon server","version":"0.2.0"}
{"level":"info","msg":"connected to database"}
{"level":"info","msg":"starting HTTP server","addr":"0.0.0.0:8080"}
```

**✅ Checkpoint:** Server is running. Keep this terminal open!

---

## Step 6: Test Backend API (2 minutes)

**Open a NEW terminal window (Terminal 2) and run:**

```bash
# Test health check
curl http://localhost:8080/health
```

**Expected:** `{"status":"healthy","time":"2026-09-07T23:XX:XX"}`

### Create a Test Target

```bash
# Create Google DNS target
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{"name":"Google DNS","target_type":"host","address":"8.8.8.8"}'
```

**Expected:** Returns JSON with `"id":"<UUID>"` - **Copy this ID!**

### Create an ICMP Monitor

**Replace `<TARGET_ID>` with the UUID from above:**

```bash
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -d '{"target_id":"<TARGET_ID>","monitor_type":"icmp","interval_seconds":60,"timeout_seconds":10,"retries":3,"config":{"count":4}}'
```

**Expected:** Returns JSON with monitor configuration

**✅ Checkpoint:** Target and monitor created. The system will start monitoring in 60 seconds!

---

## Step 7: Start Frontend Dashboard (3 minutes)

**Open a NEW terminal window (Terminal 3) and run:**

```bash
cd D:\netmon\frontend

# Install dependencies (first time only - takes 2-3 minutes)
npm install

# Start development server
npm start
```

**Expected output:**
```
Compiled successfully!

You can now view netmon-frontend in the browser.

  Local:            http://localhost:3000
```

**Your browser should automatically open to http://localhost:3000**

**✅ Checkpoint:** Browser shows NetMon dashboard.

---

## Step 8: View the Dashboard (5 minutes)

### Dashboard Page (http://localhost:3000/)

**You should see:**
- ✅ "Network Monitoring Dashboard" heading
- ✅ Statistics cards showing:
  - Total Targets: 1
  - Targets UP: 0 (will update after first check)
  - Targets DOWN: 0
  - Active Probes: 0 or 1
- ✅ Clean, modern design with blue/green/red color coding

### Targets Page

1. **Click "Targets" in navigation bar**
2. **You should see:**
   - ✅ "Google DNS" card
   - ✅ Address: 8.8.8.8
   - ✅ Status indicator (green dot)
   - ✅ Target type badge: "HOST"

### Target Detail Page

1. **Click on the "Google DNS" card**
2. **You should see:**
   - ✅ Target name and address at top
   - ✅ 6 statistics boxes (may show "N/A" initially)
   - ✅ Two empty chart areas
   - ✅ "Recent Checks" table (empty initially)

**⏱️ Wait 60 seconds for first measurement...**

### After 60 Seconds - Refresh the Page

**Now you should see:**
- ✅ Availability: ~100%
- ✅ Avg Latency: (e.g., 15.32 ms)
- ✅ Min/Max Latency values
- ✅ Latency chart with data points
- ✅ Packet loss chart (should be 0%)
- ✅ Recent checks table with one row showing:
  - Timestamp
  - Status: SUCCESS (green badge)
  - Latency value
  - Packet loss: 0.00%

**✅ Checkpoint:** You see real monitoring data!

---

## Step 9: Watch It Work! (continuous)

Every 60 seconds:
1. A new ICMP check runs
2. Data is stored in TimescaleDB
3. Dashboard updates (auto-refresh every 30 seconds)
4. Charts populate with more data points

**In Terminal 1 (backend), you'll see logs:**
```json
{"level":"debug","msg":"check completed","target":"8.8.8.8","success":true,"duration":"..."}
```

**After 5-10 minutes:**
- Charts show multiple data points
- Latency trends become visible
- You can see jitter variations
- Recent checks table fills up

---

## Step 10: Add More Targets (optional)

### Add Cloudflare DNS
```bash
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{"name":"Cloudflare DNS","target_type":"host","address":"1.1.1.1"}'
```

### Add TCP Monitor (HTTPS port check)
```bash
# Create target with port
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{"name":"Google HTTPS","target_type":"server","address":"google.com","port":443}'

# Create TCP monitor (use returned target ID)
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -d '{"target_id":"<TARGET_ID>","monitor_type":"tcp","interval_seconds":30,"timeout_seconds":10,"config":{"timeout":10}}'
```

**Refresh dashboard** - you should now see 2-3 targets!

---

## Troubleshooting

### Issue: "go: command not found" after installation
**Solution:** 
```bash
# Close all terminals
# Open NEW terminal
go version
```
If still not working, restart your computer.

### Issue: Backend shows "Permission denied" for ICMP
**Solution:** Run terminal as Administrator:
- Right-click Git Bash / PowerShell
- Select "Run as Administrator"
- Navigate to project and start backend again

### Issue: No data after 60 seconds
**Solution:**
```bash
# Check backend logs in Terminal 1
# Should see "check completed" messages

# If not, check database connection
docker ps
docker logs netmon-postgres
```

### Issue: Frontend shows "Proxy error"
**Solution:**
- Verify backend is running on port 8080
- Check Terminal 1 shows "starting HTTP server"
- Restart frontend (Ctrl+C, then `npm start`)

### Issue: Port 8080 already in use
**Solution:**
```bash
# Find what's using port 8080
netstat -ano | findstr :8080

# Kill the process (replace <PID>)
taskkill /PID <PID> /F
```

---

## Success Indicators ✅

You'll know everything is working when:

1. **Backend Terminal (Terminal 1):**
   - No errors
   - Shows "check completed" every 60 seconds
   - Logs measurement success

2. **Frontend (Browser):**
   - Dashboard loads without errors
   - Can navigate between pages
   - Targets list shows your target(s)
   - Target detail shows charts with data
   - Auto-refresh updates data

3. **Database:**
   ```bash
   docker exec -it netmon-postgres psql -U netmon -d netmon -c "SELECT COUNT(*) FROM measurements;"
   ```
   Returns count > 0 and growing

4. **Browser Console (F12):**
   - No red errors
   - API calls return 200 status
   - WebSocket connection attempts visible

---

## What You're Seeing

**This is a real network monitoring system:**
- ✅ Backend executing actual ICMP ping to 8.8.8.8
- ✅ Measuring real packet loss, latency, jitter
- ✅ Storing time-series data in TimescaleDB
- ✅ REST API serving real-time statistics
- ✅ React dashboard visualizing live data
- ✅ Auto-refresh keeping data current

**No fake data. No mocks. Real monitoring!**

---

## Next Steps After Success

1. **Explore the UI:**
   - Click around the dashboard
   - View different targets
   - Watch charts update over time

2. **Add More Targets:**
   - Your local router
   - Public DNS servers (8.8.4.4, 1.0.0.1)
   - Web servers with TCP monitors

3. **Learn the API:**
   ```bash
   # List all targets
   curl http://localhost:8080/api/v1/targets
   
   # Get target stats
   curl http://localhost:8080/api/v1/targets/<ID>/stats
   
   # Get recent measurements
   curl http://localhost:8080/api/v1/targets/<ID>/measurements
   ```

4. **Continue Development:**
   - Phase 3: DNS, HTTP, TLS monitors
   - Add UI forms for creating targets
   - Implement alert engine
   - Add authentication

---

## Stop Everything

When you're done:

```bash
# Stop frontend (Terminal 3)
Ctrl+C

# Stop backend (Terminal 1)
Ctrl+C

# Stop database
cd D:\netmon
docker-compose down
```

---

**Ready? Let's start with Step 1: Install Go!**

Visit https://go.dev/dl/ and download the Windows installer.

Let me know when you've installed Go and I'll help you continue!
