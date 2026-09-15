#!/bin/bash
# NetMon Complete Startup Script

echo "🚀 Starting NetMon Platform..."
echo ""

# Check if Docker is running
if ! docker ps &> /dev/null; then
    echo "❌ Docker is not running. Please start Docker Desktop first."
    exit 1
fi

# Start database containers
echo "📦 Starting database containers..."
cd "$(dirname "$0")"
docker-compose up -d

# Wait for database to be healthy
echo "⏳ Waiting for database to be ready..."
sleep 5

# Check if migrations are needed
echo "🔄 Checking database schema..."
TABLES=$(docker exec netmon-postgres psql -U netmon -d netmon -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='monitors';" 2>/dev/null | tr -d ' ')

if [ "$TABLES" = "0" ]; then
    echo "📝 Applying database migrations..."
    cat migrations/001_initial_schema.sql | docker exec -i netmon-postgres psql -U netmon -d netmon > /dev/null
    cat migrations/002_add_extended_metrics.sql | docker exec -i netmon-postgres psql -U netmon -d netmon > /dev/null 2>&1
    echo "✅ Migrations applied"
else
    echo "✅ Database schema already exists"
fi

echo ""
echo "✅ Database ready!"
echo ""
echo "To start the backend server, open a terminal and run:"
echo "  cd D:\\netmon\\backend"
echo "  go run cmd/server/main.go"
echo ""
echo "To start the frontend, open another terminal and run:"
echo "  cd D:\\netmon\\frontend"
echo "  npm install  # (only first time)"
echo "  npm start"
echo ""
echo "📊 Dashboard will be available at: http://localhost:3000"
echo "🔌 Backend API available at: http://localhost:8080"
