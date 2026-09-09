# Quick Start - Testing Without Go

Since Go is not currently installed, here's what we can do:

## Option 1: Install Go (Recommended for Full Testing)

1. **Download Go:**
   - Visit: https://go.dev/dl/
   - Download: `go1.23.1.windows-amd64.msi` (or latest)
   - Install and restart terminal

2. **Verify Installation:**
   ```powershell
   go version
   # Should show: go version go1.23.1 windows/amd64
   ```

3. **Then proceed with full testing** (see TESTING.md)

---

## Option 2: Test Frontend Only (No Backend Needed)

You can explore the frontend code and UI structure without running the backend:

### Step 1: Install Frontend Dependencies

```powershell
cd D:\netmon\frontend
npm install
```

This will take 2-5 minutes to download all React dependencies.

### Step 2: Start Frontend in "Mock Mode"

```powershell
npm start
```

**Note:** The frontend will try to connect to the backend API. Since the backend isn't running, you'll see some API errors in the console, but you can still see:
- Dashboard page structure
- Navigation
- Component layout
- UI design
- Charts (with no data)

### Step 3: View the UI

Open browser to: http://localhost:3000

You'll see:
- ✅ Dashboard with stat boxes (will show 0s)
- ✅ Navigation bar
- ✅ Targets page (will be empty)
- ✅ Responsive design
- ❌ No real data (needs backend)

---

## Option 3: Review Code Structure

Explore the codebase to understand what was built:

### Backend Code (Go)
```powershell
cd D:\netmon\backend

# View main server
type cmd\server\main.go

# View API endpoints
type internal\api\api.go

# View ICMP monitor
type internal\monitors\icmp.go

# View database schema
type ..\migrations\001_initial_schema.sql
```

### Frontend Code (TypeScript/React)
```powershell
cd D:\netmon\frontend

# View main app
type src\App.tsx

# View dashboard
type src\pages\Dashboard.tsx

# View API service
type src\services\api.ts

# View types
type src\types\index.ts
```

---

## What You Can Test Right Now (Without Go)

### 1. Start Docker Database

```powershell
cd D:\netmon

# Start Docker Desktop first (from Start menu)
# Then run:
docker-compose up -d

# Verify containers are running
docker ps

# Should show:
# netmon-postgres (PostgreSQL + TimescaleDB)
# netmon-redis
```

### 2. Install Frontend and Preview UI

```powershell
cd frontend
npm install
npm start
```

Browser opens to http://localhost:3000

**What you'll see:**
- Complete UI structure
- Dashboard page with metrics boxes
- Navigation between pages
- Modern, clean design
- Responsive layout

**What won't work yet:**
- No data in charts (needs backend)
- API calls will fail (needs backend)
- Can't create targets (needs backend)

---

## Recommended Path Forward

### For Full Testing:

1. **Install Go** (15 minutes)
   - Download from https://go.dev/dl/
   - Run installer
   - Restart terminal

2. **Start Database** (2 minutes)
   ```powershell
   docker-compose up -d
   ```

3. **Run Migrations** (1 minute)
   ```powershell
   cd backend
   go run cmd/migrate/main.go up
   ```

4. **Start Backend** (1 minute)
   ```powershell
   go run cmd/server/main.go
   ```

5. **Start Frontend** (already done if you did Option 2)
   ```powershell
   cd ../frontend
   npm start
   ```

6. **Create Test Data** (via API)
   ```powershell
   # Create target
   curl -X POST http://localhost:8080/api/v1/targets -H "Content-Type: application/json" -d "{\"name\":\"Google DNS\",\"target_type\":\"host\",\"address\":\"8.8.8.8\"}"

   # Create monitor (use target ID from above)
   curl -X POST http://localhost:8080/api/v1/monitors -H "Content-Type: application/json" -d "{\"target_id\":\"<UUID>\",\"monitor_type\":\"icmp\",\"interval_seconds\":60,\"timeout_seconds\":10,\"config\":{\"count\":4}}"
   ```

7. **View Results** (after 60 seconds)
   - Refresh dashboard
   - See live data in charts
   - Watch measurements update

---

## Quick Demo Without Installation

If you just want to see what was built:

### File Count
```powershell
cd D:\netmon
dir /s /b *.go *.tsx *.ts *.sql | find /c /v ""
# Shows: 30+ source files
```

### Code Statistics
```powershell
# Backend (Go)
dir backend\internal\*.go /s
dir backend\cmd\*.go /s

# Frontend (React/TypeScript)
dir frontend\src\*.tsx /s
dir frontend\src\*.ts /s

# Database
type migrations\001_initial_schema.sql | find /c /v ""
# Shows: 400+ lines of SQL
```

### Documentation
```powershell
# Read documentation
type README.md
type ARCHITECTURE.md
type STATUS.md
type TESTING.md
```

---

## Summary of Current Status

**What's Installed:**
✅ Node.js 24.18.0  
✅ npm 11.16.0  
✅ Docker 29.7.2  
❌ Go (needs installation)  
⚠️ Docker Desktop (needs to be started)  

**What You Can Do Now:**
1. ✅ Install frontend and see UI structure
2. ✅ Start Docker database
3. ✅ Review all source code
4. ✅ Read documentation
5. ❌ Run backend (needs Go)
6. ❌ See real monitoring data (needs backend)

**To Get Full System Working:**
1. Install Go from https://go.dev/dl/
2. Start Docker Desktop
3. Follow TESTING.md guide

---

## Next Steps

What would you like to do?

**A)** Install Go and test the full system  
**B)** Just preview the frontend UI (npm install + npm start)  
**C)** Review code and documentation  
**D)** Something else

Let me know and I'll guide you through!
