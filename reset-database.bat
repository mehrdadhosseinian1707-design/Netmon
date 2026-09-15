@echo off
echo ========================================
echo NetMon Database Reset
echo WARNING: This will DELETE all data!
echo ========================================
echo.
echo This will:
echo - Drop and recreate the netmon database
echo - Clear all monitoring data
echo - Reset migrations
echo.
echo Press Ctrl+C to cancel or
pause

echo.
echo [1/3] Dropping existing database...
docker exec netmon-postgres psql -U netmon -d postgres -c "DROP DATABASE IF EXISTS netmon;"

echo [2/3] Creating fresh database...
docker exec netmon-postgres psql -U netmon -d postgres -c "CREATE DATABASE netmon;"

echo [3/3] Installing extensions...
docker exec netmon-postgres psql -U netmon -d netmon -c "CREATE EXTENSION IF NOT EXISTS timescaledb;"
docker exec netmon-postgres psql -U netmon -d netmon -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"

echo.
echo ========================================
echo Database reset complete!
echo ========================================
echo.
echo Next steps:
echo 1. Start the backend: cd backend ^&^& go run cmd/server/main.go
echo 2. Migrations will run automatically
echo.
pause
