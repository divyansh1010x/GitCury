.PHONY: build test clean release docker-build docker-push lint check test-coverage help all

# Default target
all: build

# Build the application
build:
	@echo "Building GitCury..."
	go build -o gitcury

# Install the application
install: build
	@echo "Installing GitCury..."
	cp gitcury $(GOPATH)/bin/gitcury

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f gitcury gitcury.exe
	rm -f coverage.out coverage.html
	rm -rf dist/

# Lint the code
lint:
	@echo "Linting code..."
	golangci-lint run

# Local release simulation
check-release:
	@echo "Checking release configuration..."
	goreleaser check
	goreleaser release --snapshot --clean --skip=publish

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t lakshyajain1503/gitcury:dev .

# Run the application with Docker
docker-run:
	@echo "Running GitCury in Docker..."
	docker run -it --rm \
		-v "$(PWD):/app/data" -w "/app/data" \
		-v "$(HOME)/.gitconfig:/home/gitcuryuser/.gitconfig:ro" \
		-v "$(HOME)/.gitcury:/home/gitcuryuser/.gitcury" \
		lakshyajain1503/gitcury:dev $(CMD)

# Run the application with Docker Compose
docker-compose-up:
	@echo "Starting GitCury with Docker Compose..."
	docker-compose up

# Tag a new version
tag:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make tag VERSION=1.2.3"; exit 1; fi
	@echo "Tagging version v$(VERSION)..."
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)
	@echo "Tagged v$(VERSION)"

# Generate mockups
generate:
	@echo "Generating code..."
	go generate ./...

# Help target
help:
	@echo "GitCury Makefile Commands:"
	@echo "  make build          - Build the application"
	@echo "  make install        - Install to GOPATH/bin"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make lint           - Lint the code"
	@echo "  make check-release  - Check release configuration"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run CMD=msgs - Run with Docker"
	@echo "  make docker-compose-up - Start with Docker Compose"
	@echo "  make tag VERSION=1.2.3 - Tag a new version"
	@echo "  make generate       - Generate code"
	@echo "  make help           - Show this help"
