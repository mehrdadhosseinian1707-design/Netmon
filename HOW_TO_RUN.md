# NetMon - How to Run the Application

## System Requirements ✅
- ✅ Go 1.27.1 (installed)
- ✅ Node.js 24.18.0 (installed)
- ✅ npm 11.16.0 (installed)
- ✅ Docker 29.7.2 (installed)

---

## Quick Start (3 Steps)

### Step 1: Start Database (Docker)
```bash
cd D:\netmon
docker-compose up -d
```
**What this does:** Starts PostgreSQL + TimescaleDB and Redis in Docker containers

**Verify it's running:**
```bash
docker ps
```
You should see containers: `netmon-postgres` and `netmon-redis`

---

### Step 2: Start Backend Server
```bash
cd D:\netmon\backend
go run cmd/server/main.go
```
**What this does:**
- Runs database migrations automatically
- Starts REST API on http://localhost:8080
- Starts monitoring scheduler
- Ready to accept probe connections

**You'll see:**
```
Starting NetMon Server...
Database connection established
Running migrations...
Server listening on :8080
```

**Keep this terminal open!**

---

### Step 3: Start Frontend Dashboard
Open a **NEW terminal** and run:
```bash
cd D:\netmon\frontend
npm install    # Only needed first time (takes 2-3 minutes)
npm start
```
**What this does:**
- Installs React dependencies (first time only)
- Starts development server on http://localhost:3000
- Opens browser automatically

**Browser will open to:** http://localhost:3000

---

## Optional: Start Monitoring Probe

If you want to run distributed probes (for actual monitoring):

Open a **THIRD terminal**:
```bash
cd D:\netmon\backend
go run cmd/probe/main.go
```

**What this does:**
- Connects to central server
- Executes monitoring jobs (ICMP, TCP, DNS, HTTP, etc.)
- Sends measurements back to server

---

## What You'll See

### Frontend Dashboard (http://localhost:3000)
- **Dashboard page** - Overview with metrics, live charts, status indicators
- **Targets page** - List of monitored targets (hosts, routers, servers)
- **Monitors page** - Active monitoring checks
- **Tools page** - Network diagnostic tools (ping, traceroute, DNS lookup)
- **Real-time updates** - WebSocket connection for live data

### Backend API (http://localhost:8080)
Available endpoints:
- `GET /health` - Health check
- `GET /api/v1/targets` - List targets
- `GET /api/v1/monitors` - List monitors
- `GET /api/v1/measurements` - Get measurement data
- `POST /api/v1/targets` - Create new target
- `POST /api/v1/monitors` - Create new monitor

---

## Testing the System

### 1. Create a Test Target (Google DNS)
```bash
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Google DNS\",\"target_type\":\"host\",\"address\":\"8.8.8.8\"}"
```

### 2. Create ICMP Monitor
Get the target ID from step 1, then:
```bash
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -d "{\"target_id\":\"<UUID-from-step-1>\",\"monitor_type\":\"icmp\",\"interval_seconds\":60,\"timeout_seconds\":10,\"config\":{\"count\":4}}"
```

### 3. View in Dashboard
- Open http://localhost:3000
- Click "Targets" - you'll see "Google DNS"
- Click on it to see detailed metrics
- Wait 60 seconds for first measurement

---

## Stopping the Application

### Stop Frontend
Press `Ctrl+C` in the frontend terminal

### Stop Backend
Press `Ctrl+C` in the backend terminal

### Stop Database
```bash
cd D:\netmon
docker-compose down
```

**To remove all data:**
```bash
docker-compose down -v
```

---

## Troubleshooting

### Port Already in Use
**Error:** `bind: address already in use`

**Solution:**
```bash
# Windows - Find what's using port 8080
netstat -ano | findstr :8080

# Kill the process
taskkill /PID <pid> /F
```

### Docker Not Running
**Error:** `Cannot connect to Docker daemon`

**Solution:** Start Docker Desktop from Windows Start menu

### Database Connection Failed
**Error:** `failed to connect to database`

**Solution:**
```bash
# Check if containers are running
docker ps

# View database logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

### Frontend Won't Start
**Error:** `npm ERR!` or `Module not found`

**Solution:**
```bash
cd D:\netmon\frontend
rm -rf node_modules package-lock.json
npm install
npm start
```

---

## Architecture Overview

```
┌─────────────────────────────────────────┐
│         Frontend (React)                │
│         http://localhost:3000           │
│  • Dashboard with real-time charts      │
│  • WebSocket connection                 │
└──────────────┬──────────────────────────┘
               │ HTTP/WebSocket
┌──────────────▼──────────────────────────┐
│      Backend Server (Go)                │
│      http://localhost:8080              │
│  • REST API                             │
│  • Monitoring Scheduler                 │
│  • WebSocket Server                     │
└──────────────┬──────────────────────────┘
               │
    ┌──────────┼──────────┐
    │          │          │
┌───▼────┐ ┌──▼─────┐ ┌──▼──────┐
│Postgres│ │ Redis  │ │ Probes  │
│+TimescaleDB│        │(optional)│
└────────┘ └────────┘ └─────────┘
```

---

## Features Available

### 10 Monitor Types
1. **ICMP** - Ping monitoring (packet loss, RTT, jitter)
2. **TCP** - Port connectivity checks
3. **DNS** - DNS resolution monitoring
4. **HTTP/HTTPS** - Web endpoint monitoring
5. **TLS** - Certificate expiry monitoring
6. **Traceroute** - Network path analysis
7. **MTR** - Combined ping + traceroute
8. **SNMP** - Device monitoring via SNMP
9. **VPN** - Tunnel status monitoring
10. **BGP** - BGP session monitoring

### 60+ Network Tools
- Ping, Traceroute, MTR
- DNS lookup (A, AAAA, MX, NS, TXT, SOA, PTR)
- WHOIS, GeoIP, ASN lookup
- Port scanner, SSL checker
- Bandwidth test, packet capture
- And many more...

### 21 Metrics Collected
- Packet loss, RTT (min/avg/max)
- Jitter, TCP connect time
- DNS resolution time
- HTTP timing (DNS, connect, TLS, TTFB, total)
- TLS certificate expiry days
- Hop count, path changes
- SNMP values
- And more...

---

## Next Steps

1. **Start the app** (follow steps above)
2. **Create targets** via UI or API
3. **Set up monitors** for each target
4. **View real-time data** in dashboard
5. **Explore network tools** in Tools page

---

## Support

- **Documentation:** Check ARCHITECTURE.md, STATUS.md
- **API Reference:** http://localhost:8080/api/v1/
- **Logs:** Check terminal output or `docker-compose logs`

Enjoy monitoring! 🚀
