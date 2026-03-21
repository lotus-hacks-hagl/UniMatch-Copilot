FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o ai-service-go .

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/ai-service-go /app/ai-service-go
EXPOSE 9000
CMD ["/app/ai-service-go"]
