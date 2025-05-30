package interfaces

// GeminiRunner defines the interface for Gemini API operations
type GeminiRunner interface {
	SendToGemini(contextData map[string]map[string]string, apiKey string, customInstructions ...string) (string, error)
}
