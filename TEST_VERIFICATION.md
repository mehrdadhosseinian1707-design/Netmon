# Phase 1 & 2 Test Verification Checklist

**Date:** 2026-09-07  
**Time:** 23:35 UTC  
**Status:** Ready for Testing

---

## Prerequisites Checklist

Before starting Phase 3, verify these tests pass:

### ✅ Test 1: Go Installation
```bash
go version
```
**Expected:** `go version go1.23.x windows/amd64`

### ✅ Test 2: Docker Running
```bash
docker ps
```
**Expected:** Shows container list (may be empty)

### ✅ Test 3: Database Started
```bash
cd D:\netmon
docker-compose up -d
docker ps
```
**Expected:** Shows `netmon-postgres` and `netmon-redis` containers

### ✅ Test 4: Migrations Completed
```bash
cd D:\netmon\backend
go run cmd/migrate/main.go up
```
**Expected:** "Migrations completed successfully"

### ✅ Test 5: Backend Running
```bash
cd D:\netmon\backend
go run cmd/server/main.go
```
**Expected:** 
```json
{"level":"info","msg":"starting netmon server","version":"0.2.0"}
{"level":"info","msg":"connected to database"}
{"level":"info","msg":"starting HTTP server","addr":"0.0.0.0:8080"}
```

### ✅ Test 6: API Health Check
```bash
curl http://localhost:8080/health
```
**Expected:** `{"status":"healthy","time":"2026-09-07T23:35:..."}`

### ✅ Test 7: Create Test Target
```bash
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{"name":"Google DNS","target_type":"host","address":"8.8.8.8"}'
```
**Expected:** JSON with target ID (copy this UUID)

### ✅ Test 8: Create Monitor
```bash
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -d '{"target_id":"<UUID>","monitor_type":"icmp","interval_seconds":60,"timeout_seconds":10,"retries":3,"config":{"count":4}}'
```
**Expected:** JSON with monitor configuration

### ✅ Test 9: Frontend Running
```bash
cd D:\netmon\frontend
npm install
npm start
```
**Expected:** Browser opens to http://localhost:3000

### ✅ Test 10: Dashboard Loads
**Open:** http://localhost:3000
**Expected:** Dashboard page with statistics boxes

### ✅ Test 11: Targets List Shows Data
**Navigate to:** http://localhost:3000/targets
**Expected:** "Google DNS" target card visible

### ✅ Test 12: Target Detail Loads
**Click:** "Google DNS" card
**Expected:** Detail page with 6 stat boxes and 2 chart areas

### ✅ Test 13: First Measurement (Wait 60 seconds)
**After 60 seconds, refresh target detail page**
**Expected:** 
- Availability: 100%
- Avg Latency: ~10-30 ms
- Charts show data points
- Recent checks table has 1 row

### ✅ Test 14: Data Flow Verification
```bash
docker exec -it netmon-postgres psql -U netmon -d netmon -c "SELECT COUNT(*) FROM measurements;"
```
**Expected:** count > 0

### ✅ Test 15: Backend Logs Show Activity
**Check Terminal 1 (backend):**
**Expected:** See logs like:
```json
{"level":"debug","msg":"check completed","target":"8.8.8.8","success":true}
```

---

## Test Results Summary

**If all 15 tests pass:**
✅ Backend is working correctly  
✅ Database is storing measurements  
✅ Frontend is displaying data  
✅ Real monitoring is happening  
✅ **READY FOR PHASE 3**

**If any test fails:**
- Check `INSTALL_AND_TEST.md` for troubleshooting
- Verify all prerequisites are installed
- Check Docker containers are running
- Review backend logs for errors

---

## Quick Validation (30 seconds)

```bash
# All in one check
go version && \
docker ps && \
curl -s http://localhost:8080/health && \
curl -s http://localhost:3000 > /dev/null && \
echo "✅ All services running!"
```

---

## Proceeding to Phase 3

Once tests pass, Phase 3 will add:
- DNS monitoring (A, AAAA, MX, NS, TXT, SOA, PTR records)
- HTTP/HTTPS monitoring with detailed timing
- TLS certificate monitoring and expiration alerts
- Traceroute with ASN tracking
- MTR (continuous path monitoring)

**Estimated Phase 3 Development Time:** 3-4 hours  
**Additional Code:** ~2,000 lines  
**New Components:** 5 monitors, UI pages, API endpoints

---

**Date:** 2026-09-07 23:35 UTC  
**Status:** Awaiting test verification to proceed to Phase 3
