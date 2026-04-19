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
	@echo "Building client..."
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev") && \
	BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
	echo "Version: $$VERSION" && \
	echo "Build Date: $$BUILD_DATE" && \
	go build -ldflags "-X github.com/Rusich90/GophKeeper/internal/client.Version=$$VERSION -X github.com/Rusich90/GophKeeper/internal/client.BuildDate=$$BUILD_DATE" -o keeper ./cmd/client

# Кросс-платформенная сборка клиента
build-client-mac:
	@echo "Building client for macOS..."
	@mkdir -p dist/client
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o dist/client/keeper-darwin-amd64 ./cmd/client
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -o dist/client/keeper-darwin-arm64 ./cmd/client
	@echo "macOS builds completed!"
	@ls -lh dist/client/keeper-darwin-*

build-client-linux:
	@echo "Building client for Linux..."
	@mkdir -p dist/client
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o dist/client/keeper-linux-amd64 ./cmd/client
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -o dist/client/keeper-linux-arm64 ./cmd/client
	@echo "Linux builds completed!"
	@ls -lh dist/client/keeper-linux-*

build-client-windows:
	@echo "Building client for Windows..."
	@mkdir -p dist/client
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o dist/client/keeper-windows-amd64.exe ./cmd/client
	@echo "Windows build completed!"
	@ls -lh dist/client/keeper-windows-*

# Очистка директории с бинарными файлами
clean-dist:
	@echo "Cleaning dist directory..."
	@rm -rf dist/
	@echo "Done!"

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
