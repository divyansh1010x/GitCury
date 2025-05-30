# GitCury Test Suite

This directory contains comprehensive tests for the GitCury project.

## Directory Structure

- `end_to_end/`: End-to-end tests that verify complete workflows
- `mock/`: Mock implementations of external dependencies
- `testutils/`: Utilities for setting up test environments

## Running Tests

### Running All Tests

```bash
go test -v ./tests/...
```

### Running Tests with Coverage

```bash
# Using the coverage script (recommended)
./tests/run_coverage.sh

# Or manually
go test -v -race -coverprofile=coverage.out -covermode=atomic ./tests/...
go tool cover -html=coverage.out -o coverage.html
```

### Running Specific Test Files

```bash
go test -v ./tests/end_to_end/message_generation_test.go
```

### Running Specific Test Functions

```bash
go test -v ./tests/end_to_end -run TestMessageGeneration
```

## Test Design

### Mock Implementation

All external dependencies (Git commands, API calls) are mocked to ensure:

1. Tests don't make actual Git changes or API calls
2. Tests are fast and reliable
3. Tests can simulate error conditions easily

### Test Environment

Each test creates an isolated test environment with:

1. Temporary directory for test files
2. Mock Git implementation
3. Mock Gemini API
4. Clean configuration

### Coverage Goals

The test suite aims for high code coverage, especially in core functionality:

- Message generation
- Commit operations
- Push operations
- Configuration management
- Error handling and recovery
- End-to-end workflows

## Adding New Tests

When adding new features to GitCury, please add corresponding tests:

1. For new commands, add end-to-end tests in `end_to_end/`
2. For new utility functions, add unit tests
3. Update mocks in `mock/` as needed for new dependencies

## Running Tests with Different Go Versions

```bash
# With Go 1.19
go1.19 test -v ./tests/...

# With Go 1.20
go1.20 test -v ./tests/...
```
