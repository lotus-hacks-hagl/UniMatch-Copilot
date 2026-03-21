.PHONY: all build run test clean db-up db-down swagger seed

all: build

build:
	@echo "Building backend..."
	@go build -o bin/server main.go

run:
	@echo "Running backend server..."
	@go run main.go

test:
	@echo "Running tests..."
	@go test ./... -v

clean:
	@echo "Cleaning up..."
	@rm -rf bin/

db-up:
	@echo "Starting PostgreSQL via Docker (if applicable)..."
	@docker-compose up -d db

db-down:
	@echo "Stopping PostgreSQL..."
	@docker-compose down

swagger:
	@echo "Generating Swagger Docs..."
	@swag init -pd

seed:
	@echo "Seeding initial data..."
	@go run scripts/seed.go

init-data:
	@echo "Initializing default data..."
	@go run scripts/seed.go
	@echo "Data initialization complete."
