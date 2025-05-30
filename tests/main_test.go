package integration

import (
	"fmt"
	"os"
	"testing"
)

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	// Setup code before tests
	fmt.Println("Setting up test environment...")

	// Ensure we're in test mode
	os.Setenv("GITCURY_TEST_MODE", "true")

	// Run the tests
	exitCode := m.Run()

	// Cleanup code after tests
	fmt.Println("Cleaning up test environment...")
	os.Unsetenv("GITCURY_TEST_MODE")

	// Exit with the exit code from the tests
	os.Exit(exitCode)
}
