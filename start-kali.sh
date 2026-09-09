#!/bin/bash
# NetMon startup script for Kali Linux

set -e

echo "🚀 Starting NetMon System..."

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo "❌ Docker not installed"; exit 1; }
command -v go >/dev/null 2>&1 || { echo "❌ Go not installed"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "❌ Node.js not installed"; exit 1; }

# Start Docker containers
echo "📦 Starting Docker containers..."
docker-compose up -d

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for database..."
sleep 10

# Check if migrations have been run
echo "🗄️ Checking database migrations..."
cd backend
if ! go run cmd/migrate/main.go up 2>/dev/null; then
    echo "✅ Migrations already applied or completed"
fi

# Start backend in background
echo "🔧 Starting backend server..."
go run cmd/server/main.go > /tmp/netmon-backend.log 2>&1 &
BACKEND_PID=$!
echo "Backend PID: $BACKEND_PID"

# Wait for backend to start
sleep 5

# Check backend health
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ Backend is running"
else
    echo "❌ Backend failed to start. Check /tmp/netmon-backend.log"
    exit 1
fi

# Start frontend
echo "🎨 Starting frontend..."
cd ../frontend
PORT=3001 npm start &
FRONTEND_PID=$!
echo "Frontend PID: $FRONTEND_PID"

echo ""
echo "✅ NetMon is running!"
echo "📊 Dashboard: http://localhost:3001"
echo "🔌 API: http://localhost:8080"
echo ""
echo "To stop:"
echo "  kill $BACKEND_PID $FRONTEND_PID"
echo "  docker-compose down"
