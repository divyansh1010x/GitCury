package end_to_end

import (
	"GitCury/config"
	"GitCury/core"
	"GitCury/tests/testutils"
	"testing"
)

func TestPushOperation(t *testing.T) {
	// Set up test environment
	env, err := testutils.SetupTestEnv()
	if err != nil {
		t.Fatalf("Failed to set up test environment: %v", err)
	}
	defer env.Cleanup()

	// Configure root folders
	config.Set("root_folders", []interface{}{env.TempDir})

	// Configure mock push success
	env.GitMock.SetupMockPushResult(env.TempDir, true)

	// Run the push operation
	err = core.PushAllRoots("main")
	if err != nil {
		t.Fatalf("Push operation failed: %v", err)
	}

	// Verify the branch was pushed
	if env.GitMock.LastPushBranch != "main" {
		t.Errorf("Expected push to branch 'main', but was %s", env.GitMock.LastPushBranch)
	}

	// Verify the command calls (we should have one push command)
	pushCount := 0
	for _, call := range env.GitMock.CommandCalls {
		if len(call.Args) >= 2 && call.Args[0] == "push" {
			pushCount++
		}
	}
	if pushCount == 0 {
		t.Errorf("Expected at least one push command, but none were recorded")
	}
}

func TestPushOperationWithError(t *testing.T) {
	// Set up test environment
	env, err := testutils.SetupTestEnv()
	if err != nil {
		t.Fatalf("Failed to set up test environment: %v", err)
	}
	defer env.Cleanup()

	// Configure root folders
	config.Set("root_folders", []interface{}{env.TempDir})

	// Configure mock push to fail
	env.GitMock.SetupMockPushResult(env.TempDir, false)

	// Run the push operation
	err = core.PushAllRoots("main")

	// Verify that the function returns an error
	if err == nil {
		t.Errorf("Expected error when push fails, but got no error")
	}
}

func TestPushOperationWithInvalidConfig(t *testing.T) {
	// Set up test environment
	env, err := testutils.SetupTestEnv()
	if err != nil {
		t.Fatalf("Failed to set up test environment: %v", err)
	}
	defer env.Cleanup()

	// Set invalid root folders configuration
	config.Set("root_folders", "not-an-array")

	// Run the push operation
	err = core.PushAllRoots("main")

	// Verify that the function returns an error for invalid config
	if err == nil {
		t.Errorf("Expected error with invalid root_folders config, but got no error")
	}
}
