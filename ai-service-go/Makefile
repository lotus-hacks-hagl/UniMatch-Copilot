BIN_NAME=ai-service-go
CONTAINER_NAME=unimatch-ai

.PHONY: run

run:
	go run main.go

build:
	@echo "Building $(BIN_NAME)..."
	@mkdir -p ../bin
	@GOARCH=amd64 GOOS=linux go build -o ../bin/$(BIN_NAME) .
	@echo "Build $(BIN_NAME) done."