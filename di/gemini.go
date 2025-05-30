package di

import "github.com/lakshyajain-0291/gitcury/interfaces"

// GeminiRunnerInstance allows dependency injection for testing
var GeminiRunnerInstance interfaces.GeminiRunner

// SetGeminiRunner allows injecting a custom GeminiRunner (used in tests)
func SetGeminiRunner(runner interfaces.GeminiRunner) {
	GeminiRunnerInstance = runner
}

// GetGeminiRunner returns the current Gemini runner instance
func GetGeminiRunner() interfaces.GeminiRunner {
	return GeminiRunnerInstance
}
