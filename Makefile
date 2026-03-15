.PHONY: setup run build test clean

# Default target
all: run

# Install dependencies and setup environment
setup:
	@echo "=> Setting up Fube-Go environment..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "=> Created .env file. Please update it with your local credentials."; fi
	@go mod download
	@go install github.com/air-verse/air@latest
	@echo "=> Setup complete! Run 'make dev' to start the server with hot-reload."

# Run the application with hot-reload (requires Air)
dev:
	@echo "=> Starting Fube-Go with hot-reload..."
	@air

# Run the application normally
run:
	@echo "=> Starting Fube-Go..."
	@go run main.go

# Build the binary
build:
	@echo "=> Building Fube-Go binary..."
	@go build -o bin/fube-go main.go
	@echo "=> Build complete: bin/fube-go"

# Run tests
test:
	@echo "=> Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "=> Cleaning up..."
	@rm -rf bin/
	@rm -rf tmp/
