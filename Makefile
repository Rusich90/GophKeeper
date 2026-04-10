
gen-proto:
	@echo "Generating protobuf files..."
	cd api/proto && protoc --go_out=../../pkg/pb --go_opt=paths=source_relative \
		--go-grpc_out=../../pkg/pb --go-grpc_opt=paths=source_relative \
		*.proto

build-server:
	go build -o cmd/server/server ./cmd/server

build-client:
	go build -o cmd/client/client ./cmd/client
