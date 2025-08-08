# Go Discord Bot Makefile

.PHONY: build run clean test docker-build docker-run migrate help

# Default target
all: build

# Build the application
build:
	go build -o discord-bot .

# Run the application
run:
	go run main.go

# Clean build artifacts
clean:
	rm -f discord-bot

# Run tests
test:
	go test -v ./...

# Build Docker image
docker-build:
	docker build -t go-discord-bot .

# Run with Docker
docker-run:
	docker run --env-file .env -p 8080:8080 go-discord-bot

# Run database migrations
migrate:
	go run main.go -migrate-only

# Show help
help:
	@echo "Go Discord Bot Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        - Build the Discord bot"
	@echo "  run          - Run the Discord bot locally"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker"
	@echo "  migrate      - Run database migrations"
	@echo "  help         - Show this help message"