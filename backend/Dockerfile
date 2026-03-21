# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install git and other build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates and tzdata for SSL/TLS and timezone support
RUN apk --no-cache add ca-certificates tzdata

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bin/server .

# Copy migrations folder if it's used at runtime by the application
COPY --from=builder /app/migrations ./migrations

# Expose port (must match PORT env)
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
