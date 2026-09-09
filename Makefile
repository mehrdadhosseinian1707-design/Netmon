.PHONY: help build test docker-up docker-down migrate-up run-server run-probe

help:
	@echo "NetMon - Network Monitoring Platform"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build all binaries"
	@echo "  test          - Run tests"
	@echo "  docker-up     - Start Docker containers"
	@echo "  docker-down   - Stop Docker containers"
	@echo "  migrate-up    - Run database migrations"
	@echo "  run-server    - Run the server"
	@echo "  run-probe     - Run a probe"
	@echo "  clean         - Clean build artifacts"

build:
	cd backend && go build -o ../bin/server ./cmd/server
	cd backend && go build -o ../bin/probe ./cmd/probe
	cd backend && go build -o ../bin/migrate ./cmd/migrate

test:
	cd backend && go test -v ./...

docker-up:
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 5

docker-down:
	docker-compose down

migrate-up: docker-up
	cd backend && go run cmd/migrate/main.go up

run-server: migrate-up
	cd backend && go run cmd/server/main.go

run-probe:
	cd backend && go run cmd/probe/main.go

clean:
	rm -rf bin/
	docker-compose down -v
