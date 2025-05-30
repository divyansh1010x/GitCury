package utils

import (
	"github.com/lakshyajain-0291/gitcury/di"
	"github.com/lakshyajain-0291/gitcury/interfaces"
)

// DefaultGeminiRunner implements the GeminiRunner interface using the real Gemini API
type DefaultGeminiRunner struct{}

// NewDefaultGeminiRunner creates a new instance of DefaultGeminiRunner
func NewDefaultGeminiRunner() interfaces.GeminiRunner {
	return &DefaultGeminiRunner{}
}

// init initializes the default Gemini runner
func init() {
	if di.GetGeminiRunner() == nil {
		di.SetGeminiRunner(NewDefaultGeminiRunner())
	}
}

// SendToGemini delegates to the real SendToGemini function
func (r *DefaultGeminiRunner) SendToGemini(contextData map[string]map[string]string, apiKey string, customInstructions ...string) (string, error) {
	return SendToGemini(contextData, apiKey, customInstructions...)
}
