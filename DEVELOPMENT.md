# GitCury Development Guide

This document provides guidance for developers who want to contribute to GitCury.

## Prerequisites

- Go 1.20 or higher
- Git
- [Gemini API key](https://ai.google.dev/tutorials/setup)

## Setting Up the Development Environment

1. **Clone the repository**

   ```bash
   git clone https://github.com/lakshyajain-0291/GitCury.git
   cd GitCury
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Configure your environment**

   Create a configuration file at `~/.gitcury/config.json` or copy the example:

   ```bash
   mkdir -p ~/.gitcury
   cp config.json.example ~/.gitcury/config.json
   ```

   Edit the config file to add your Gemini API key and set appropriate root folders.

4. **Build the application**

   ```bash
   go build -o gitcury
   ```

5. **Run the application**

   ```bash
   ./gitcury --help
   ```

## Development with Docker

For development using Docker:

1. **Build the Docker image**

   ```bash
   docker build -t gitcury:dev .
   ```

2. **Run with Docker Compose**

   ```bash
   export GEMINI_API_KEY=your_api_key_here
   docker-compose up
   ```

## Running Tests

Run all tests:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Run specific test packages:

```bash
go test -v ./tests/config/...
```

## Creating a New Release

GitCury uses semantic versioning and automated releases:

1. Commit your changes with conventional commit messages (feat:, fix:, etc.)
2. Push to the main branch (PRs will be automatically tested)
3. When merged to main, a new version will be automatically created and released

For manual releases (rare cases):

```bash
git tag v1.2.3
git push origin v1.2.3
```

## Contribution Guidelines

1. Create a feature branch from `main`
2. Follow Go coding standards and project conventions
3. Add tests for new functionality
4. Use conventional commit messages
5. Submit a PR against `main`
6. Ensure all CI checks pass

## Code Organization

- `cmd/`: Command line interface
- `config/`: Configuration management
- `core/`: Core application logic
- `embeddings/`: Embedding generation and processing
- `git/`: Git operations
- `utils/`: Utility functions
- `interfaces/`: Interface definitions
- `output/`: Output management
- `tests/`: Test suite

## Useful Development Commands

```bash
# Format all code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Generate mocks (if using mockgen)
go generate ./...

# Run with verbose output
./gitcury --verbose msgs

# Test specific functionality
./gitcury msgs --root /path/to/test/repo
```
