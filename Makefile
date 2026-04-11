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
	go build -o cmd/client/client ./cmd/client

run:
	go run cmd/server/main.go
