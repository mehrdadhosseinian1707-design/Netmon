# NetMon Setup Guide

## Prerequisites

### Required Software
- **Go 1.21+** - [Download](https://go.dev/dl/)
- **PostgreSQL 14+** - [Download](https://www.postgresql.org/download/)
- **TimescaleDB 2.11+** - [Install Guide](https://docs.timescale.com/install/latest/)
- **Redis** (optional) - [Download](https://redis.io/download/)
- **Docker Desktop** (for containerized deployment) - [Download](https://www.docker.com/products/docker-desktop/)

### System Requirements
- **Windows**: Administrator privileges required for ICMP/ping operations
- **Linux**: Raw socket privileges or `setcap cap_net_raw=+ep` on binary
- **macOS**: Sudo may be required for ICMP operations

## Installation

### 1. Clone/Download the Project

```bash
cd D:\netmon
```

### 2. Install Go Dependencies

```bash
cd backend
go mod download
go mod tidy
```

### 3. Database Setup

#### Option A: Using Docker (Recommended)

```bash
# Start PostgreSQL + TimescaleDB
docker-compose up -d

# Wait for database to be ready
timeout /t 10

# Run migrations
go run cmd/migrate/main.go up
```

#### Option B: Manual PostgreSQL Setup

1. Install PostgreSQL 14+
2. Install TimescaleDB extension
3. Create database:

```sql
CREATE DATABASE netmon;
CREATE USER netmon WITH PASSWORD 'netmon_dev_password';
GRANT ALL PRIVILEGES ON DATABASE netmon TO netmon;
```

4. Enable extensions:

```sql
\c netmon
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

5. Run migrations:

```bash
# Set connection string
$env:DATABASE_URL="host=localhost port=5432 user=netmon password=netmon_dev_password dbname=netmon sslmode=disable"

# Run migration
cd backend
go run cmd/migrate/main.go up
```

### 4. Configuration

The application uses environment variables or configuration files:

**Environment Variables:**

```bash
# Server
NETMON_SERVER_HOST=0.0.0.0
NETMON_SERVER_PORT=8080

# Database
NETMON_DATABASE_HOST=localhost
NETMON_DATABASE_PORT=5432
NETMON_DATABASE_USER=netmon
NETMON_DATABASE_PASSWORD=netmon_dev_password
NETMON_DATABASE_DATABASE=netmon

# Redis (optional)
NETMON_REDIS_HOST=localhost
NETMON_REDIS_PORT=6379

# Probe
NETMON_PROBE_ID=<probe-uuid>
NETMON_PROBE_NAME="Main Probe"
NETMON_PROBE_LOCATION="Local"
NETMON_PROBE_APIKEY=<api-key>
```

## Running the System

### Method 1: Using Make (Recommended)

```bash
# Start infrastructure
make docker-up

# Run migrations
make migrate-up

# Start server (in one terminal)
make run-server

# Start probe (in another terminal)
make run-probe
```

### Method 2: Manual

```bash
# Terminal 1: Start server
cd backend
go run cmd/server/main.go

# Terminal 2: Start probe
cd backend
go run cmd/probe/main.go
```

### Method 3: Build and Run Binaries

```bash
# Build
make build

# Run server
./bin/server

# Run probe
./bin/probe
```

## Testing the System

### 1. Run Unit Tests

```bash
cd backend
go test -v ./internal/monitors/...
```

### 2. Test API Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Create a probe
curl -X POST http://localhost:8080/api/v1/probes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "probe-01",
    "location": "datacenter-1",
    "description": "Primary monitoring probe"
  }'

# Create a target
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Google DNS",
    "target_type": "host",
    "address": "8.8.8.8"
  }'

# List targets
curl http://localhost:8080/api/v1/targets

# List probes
curl http://localhost:8080/api/v1/probes
```

### 3. Create Monitors

```bash
# ICMP monitor
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -d '{
    "target_id": "<target-uuid>",
    "monitor_type": "icmp",
    "interval_seconds": 60,
    "timeout_seconds": 10,
    "retries": 3,
    "config": {
      "count": 4,
      "packet_size": 64
    }
  }'

# TCP monitor
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -d '{
    "target_id": "<target-uuid>",
    "monitor_type": "tcp",
    "interval_seconds": 30,
    "timeout_seconds": 10,
    "config": {
      "timeout": 10
    }
  }'
```

### 4. View Measurements

```bash
# Get target statistics
curl http://localhost:8080/api/v1/targets/<target-uuid>/stats

# Get recent measurements
curl http://localhost:8080/api/v1/targets/<target-uuid>/measurements
```

## Quick Test Without Database

Run standalone monitor tests:

```bash
cd backend
go test -v ./internal/monitors/monitor_test.go -run TestICMPMonitor
go test -v ./internal/monitors/monitor_test.go -run TestTCPMonitor
```

## Troubleshooting

### ICMP "Permission Denied"

**Windows:**
- Run terminal as Administrator
- Or set capabilities: This is automatically handled by the ping library

**Linux:**
```bash
sudo setcap cap_net_raw=+ep ./bin/server
sudo setcap cap_net_raw=+ep ./bin/probe
```

### Database Connection Failed

Check:
1. PostgreSQL is running: `psql -U netmon -d netmon`
2. TimescaleDB extension: `SELECT * FROM pg_extension WHERE extname='timescaledb';`
3. Connection string is correct
4. Firewall allows port 5432

### Docker Issues

```bash
# Check Docker is running
docker ps

# View logs
docker-compose logs postgres
docker-compose logs redis

# Restart containers
docker-compose restart

# Reset everything
docker-compose down -v
docker-compose up -d
```

### Port Already in Use

```bash
# Windows - Find process on port 8080
netstat -ano | findstr :8080

# Kill process
taskkill /PID <pid> /F
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                   Central Server                    │
├─────────────────────────────────────────────────────┤
│  • REST API (port 8080)                            │
│  • Monitoring Scheduler                             │
│  • PostgreSQL + TimescaleDB                         │
│  • Redis (job queue)                                │
└─────────────────────────────────────────────────────┘
                         │
            ┌────────────┴────────────┐
            │                         │
    ┌───────▼──────┐          ┌──────▼───────┐
    │   Probe 1    │          │   Probe 2    │
    │              │          │              │
    │ • ICMP       │          │ • ICMP       │
    │ • TCP        │          │ • TCP        │
    │ • DNS        │          │ • DNS        │
    └──────────────┘          └──────────────┘
```

## Next Steps

After Phase 1 is running:

1. **Phase 2**: React dashboard with real-time WebSocket updates
2. **Phase 3**: DNS, HTTP/HTTPS, TLS, Traceroute, MTR monitors
3. **Phase 4**: SNMP, device discovery, vendor integrations
4. **Phase 5**: Multi-probe architecture, ISP analysis, ASN tracking
5. **Phase 6**: BGP monitoring, advanced correlation, SLA reporting

## Support

- Check logs: `docker-compose logs -f`
- Review database: `psql -U netmon -d netmon`
- API documentation: http://localhost:8080/api/v1/
