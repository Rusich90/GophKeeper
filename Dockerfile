# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the server
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Copy migrations
COPY --from=builder /app/internal/server/migrations ./migrations

# Install migrate tool for running migrations
RUN apk add --no-cache curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate

# Create a non-root user and group
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Change ownership of app directory to appuser
RUN chown -R appuser:appuser /app

# Expose gRPC port
EXPOSE 50051

# Switch to non-root user
USER appuser

# Run the server
CMD ["./server"]