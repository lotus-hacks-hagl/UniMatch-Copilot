# Build stage
FROM --platform=linux/amd64 golang:1.26 AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/ai-service-go ./main.go

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/ai-service-go /app/ai-service-go
COPY --from=builder /app/docs /app/docs

EXPOSE 8895

CMD ["./ai-service-go"]