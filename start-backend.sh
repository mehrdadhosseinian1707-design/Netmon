#!/bin/bash
# NetMon Backend Startup Script

cd "$(dirname "$0")/backend"

echo "Starting NetMon Backend Server..."
echo "API will be available at: http://localhost:8080"
echo ""

go run cmd/server/main.go
