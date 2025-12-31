include .env
export

.PHONY: all build run test clean dev docker-build docker-up docker-down migrate sqlc lint

# Default target
all: build

# Build the application
build:
	go build -o bin/api ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Development with hot reload (requires air)
dev:
	air

# Run tests
test:
	go test -v -race ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Generate SQLC code
sqlc:
	sqlc generate

# Run linter
lint:
	golangci-lint run

# Docker commands
docker-build:
	docker build -t locolive-backend .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Database migrations
migrate-up:
	docker-compose --profile migrate run --rm migrate

migrate-down:
	docker-compose run --rm migrate -path /migrations -database "postgres://locolive:locolive@postgres:5432/locolive?sslmode=disable" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name

# Local development setup
setup:
	go mod download
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	cp .env.example .env
	@echo "Setup complete! Edit .env with your configuration."

# Initialize the database
init-db:
	docker-compose up -d postgres redis
	sleep 3
	docker-compose --profile migrate run --rm migrate

# Full dev stack
dev-stack: init-db
	docker-compose up -d

# Health check
health:
	curl -s http://localhost:8080/health | jq
