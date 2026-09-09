# NetMon - Production Network Monitoring Platform

A distributed network monitoring platform for monitoring hosts, routers, switches, servers, ISPs, and network infrastructure.

## Architecture

```
Central Server
├── API (REST + WebSocket)
├── Monitoring Scheduler
├── Alert Engine
├── Correlation Engine
├── Device Discovery
├── Database (PostgreSQL + TimescaleDB)
└── Probe Manager
    ├── Probe 1
    ├── Probe 2
    └── Probe N
```

## Features

### Phase 1 (Current)
- ICMP monitoring (IPv4/IPv6, packet loss, RTT, jitter)
- TCP port monitoring
- Distributed probe architecture
- TimescaleDB for time-series data
- REST API
- Scheduler with concurrency control

### Planned
- DNS monitoring (A, AAAA, MX, NS, TXT, SOA, PTR)
- HTTP/HTTPS monitoring with detailed timing
- TLS certificate monitoring
- Traceroute and MTR
- SNMP device monitoring
- VPN tunnel monitoring
- ISP performance tracking
- Alert engine with correlation
- React dashboard with real-time updates
- Device discovery
- ASN/BGP tracking

## Technology Stack

**Backend:**
- Go 1.21+
- PostgreSQL 14+
- TimescaleDB 2.11+
- Redis (for job queues)
- REST API + WebSocket

**Frontend:**
- React 18
- TypeScript
- Modern component architecture

**Deployment:**
- Docker
- Docker Compose

## Quick Start

```bash
# Start infrastructure
docker-compose up -d

# Run migrations
cd backend
go run cmd/migrate/main.go up

# Start server
go run cmd/server/main.go

# Start probe
go run cmd/probe/main.go
```

## Project Structure

```
netmon/
├── backend/
│   ├── cmd/
│   │   ├── server/     # Central server
│   │   ├── probe/      # Monitoring probe
│   │   └── migrate/    # Database migrations
│   ├── internal/
│   │   ├── api/        # REST API
│   │   ├── models/     # Data models
│   │   ├── monitors/   # Monitoring plugins
│   │   ├── scheduler/  # Job scheduler
│   │   ├── store/      # Database layer
│   │   └── config/     # Configuration
│   └── pkg/            # Shared packages
├── frontend/           # React dashboard
├── docker/            # Docker configurations
└── migrations/        # Database migrations
```

## Development Status

This is an active development project being built incrementally following production-grade practices.
