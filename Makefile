include .env
export

MIGRATIONS_DIR := ./internal/server/migrations

check-database-dsn:
ifndef DATABASE_DSN
	$(error DATABASE_DSN is not set. Please create .env file or set environment variable)
endif

migrate-up: check-database-dsn
	@echo "Applying migrations"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" up

migrate-down: check-database-dsn
	@echo "Rolling back last migration..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" down 1

gen-proto:
	@echo "Generating protobuf files..."
	cd api/proto && protoc --go_out=../../pkg/pb --go_opt=paths=source_relative \
		--go-grpc_out=../../pkg/pb --go-grpc_opt=paths=source_relative \
		*.proto

build-server:
	go build -o cmd/server/server ./cmd/server

build-client:
	go build -o keeper ./cmd/client

run:
	go run cmd/server/main.go

docker-up:
	@echo "Starting all services with Docker Compose..."
	docker compose up -d

docker-down:
	@echo "Stopping all services..."
	docker compose down

docker-build:
	@echo "Building Docker images..."
	docker compose build

docker-logs:
	@echo "Showing logs..."
	docker compose logs -f

docker-restart: docker-down docker-up
	@echo "Restarting all services..."

docker-migrate-up:
	@echo "Applying migrations in Docker..."
	docker compose exec -T server migrate -path /app/migrations -database "$(DATABASE_DSN)" up

docker-migrate-down:
	@echo "Rolling back last migration in Docker..."
	docker compose exec -T server migrate -path /app/migrations -database "$(DATABASE_DSN)" down 1

# Тестирование
test:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage by function:"
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "Coverage percentage:"
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}'

# Генерация моков
gen-mocks:
	@echo "Generating mocks..."
	mockery --all --dir=internal/server/storage/user
	mockery --all --dir=internal/server/storage/token
