package main_test

import (
	"GitCury/tests/testutils"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigSetup(t *testing.T) {
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()

	// Print current environment
	fmt.Printf("GEMINI_API_KEY: %s\n", os.Getenv("GEMINI_API_KEY"))
	fmt.Printf("HOME: %s\n", os.Getenv("HOME"))

	// Check if config file exists
	configPath := filepath.Join(os.Getenv("HOME"), ".gitcury", "config.json")
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file exists at: %s\n", configPath)

		// Read and print config content
		content, err := os.ReadFile(configPath)
		if err == nil {
			fmt.Printf("Config content: %s\n", string(content))
		}
	} else {
		fmt.Printf("Config file does not exist at: %s\n", configPath)
	}
}
