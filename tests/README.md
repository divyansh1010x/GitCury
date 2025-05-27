# GitCury Test Suite Documentation

This document provides an overview of the GitCury test suite, including the test structure, test types, and how to run the tests.

## Test Structure

The GitCury test suite is organized into the following directories:

```
tests/
├── config/           # Tests for the configuration package
├── core/             # Tests for the core functionality
├── embeddings/       # Tests for the embeddings package
├── git/              # Tests for the git package
├── integration_test.go # End-to-end integration tests
├── mocks/            # Mock implementations for testing
├── output/           # Tests for the output package
├── run_tests.sh      # Test runner script
├── testreport/       # Test report generator
├── testutils/        # Utilities for testing
└── utils/            # Tests for the utility package
```

## Test Types

The GitCury test suite includes the following types of tests:

### Unit Tests

Unit tests test individual functions and methods in isolation. They are organized by package and are located in the corresponding test directories.

### Integration Tests

Integration tests test the interaction between multiple components. They are located in the `integration_test.go` file.

### End-to-End Tests

End-to-end tests test the complete workflow of GitCury, from getting changed files to committing and pushing changes. They are also located in the `integration_test.go` file.

## Test Dependencies

The GitCury test suite uses the following dependencies:

- **Mocks**: Mock implementations of interfaces for testing in isolation.
- **TestUtils**: Utility functions for creating test environments, such as temporary directories and git repositories.
- **Test Report Generator**: A utility for generating comprehensive test reports.

## Running Tests

To run the tests, use the `run_tests.sh` script:

```bash
./tests/run_tests.sh
```

### Options

The test runner script supports the following options:

- `--no-coverage`: Disable code coverage collection.
- `--detailed`: Show detailed test output.
- `--report-dir DIR`: Specify a directory for test reports (default: ./test-reports).
- `--help`: Show help information.

### Test Reports

The test runner generates both JSON and HTML reports. The reports include:

- Total number of tests
- Number of passed, failed, and skipped tests
- Total duration
- Code coverage (if enabled)
- Timestamp
- GitCury version
- Go version
- Detailed test results

## Test Coverage

The GitCury test suite aims to achieve high code coverage. The current coverage status is:

| Package    | Coverage |
|------------|----------|
| config     | TBD      |
| core       | TBD      |
| embeddings | TBD      |
| git        | TBD      |
| output     | TBD      |
| utils      | TBD      |
| **Total**  | **TBD**  |

## Test Environment

The GitCury test suite is designed to be run in a clean environment. Some tests may require specific environment variables or configuration:

- `GITCURY_TEST_API_KEY`: API key for testing API integration.
- `GITCURY_SKIP_INTEGRATION`: Set to "true" to skip integration tests.
- `GITCURY_CONFIG_PATH`: Path to a test configuration file.
- `GITCURY_TEST_OUTPUT_FILE`: Path to a test output file.

## Adding New Tests

When adding new features to GitCury, please follow these guidelines for adding tests:

1. **Create unit tests** for all new functions and methods.
2. **Update integration tests** if the feature affects multiple components.
3. **Document test coverage** in this file.
4. **Run the full test suite** before submitting a pull request.

## Troubleshooting

If you encounter issues running the tests, please check the following:

- Ensure all dependencies are installed.
- Check environment variables for API keys or other configuration.
- Verify that git is installed and configured correctly.
- Check the test reports for detailed error messages.

## Future Improvements

The following improvements are planned for the GitCury test suite:

1. **Dependency Injection**: Implement proper dependency injection to make testing easier.
2. **Mocking Framework**: Improve the mocking framework to support more complex scenarios.
3. **Benchmark Tests**: Add benchmark tests to measure performance.
4. **Property-Based Testing**: Implement property-based testing for complex scenarios.
5. **Continuous Integration**: Set up CI/CD pipelines for running tests automatically.
