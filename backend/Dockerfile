# Build stage
FROM golang:1.26 AS builder

WORKDIR /app

# Install git and other build dependencies
RUN apt-get update && apt-get install -y git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server main.go

# Final stage
FROM debian:bookworm-slim

WORKDIR /app

# Install ca-certificates and tzdata for SSL/TLS and timezone support
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bin/server .

# Copy migrations folder if it's used at runtime by the application
COPY --from=builder /app/migrations ./migrations

# Expose port (must match PORT env)
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
