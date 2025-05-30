package utils

import (
	"GitCury/api"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"google.golang.org/grpc/status"
)

var (
	maxRetries int
	retryDelay int
)

func init() {
	// Initialize with default values
	maxRetries = 3
	retryDelay = 5
}

func SetTimeoutVar(retries, delay int) {
	if retries <= 0 {
		Warning("[GEMINI]: Invalid maxRetries value: " + fmt.Sprintf("%d", retries) + ", using default (3)")
		retries = 3
	}

	if delay <= 0 {
		Warning("[GEMINI]: Invalid retryDelay value: " + fmt.Sprintf("%d", delay) + ", using default (5)")
		delay = 5
	}

	maxRetries = retries
	retryDelay = delay

	// Also update the API config
	api.SetRetryConfig(retries, delay)

	Debug(fmt.Sprintf("[GEMINI]: Updated retry settings: maxRetries=%d, retryDelay=%d", maxRetries, retryDelay))
}

func SendToGemini(contextData map[string]map[string]string, apiKey string, customInstructions ...string) (string, error) {
	// Validate API key with helpful guidance
	if apiKey == "" {
		Error("[GEMINI]: ❌ API key is empty")
		Error("🔑 GEMINI_API_KEY is required but not set!")
		Error("💡 To fix this, run one of these commands:")
		Error("   • gitcury config set --key GEMINI_API_KEY --value YOUR_API_KEY_HERE")
		Error("   • export GEMINI_API_KEY=your_api_key_here")
		Error("📖 Get your API key from: https://aistudio.google.com/app/apikey")
		return "", NewAPIError("GEMINI_API_KEY is not set", nil, map[string]interface{}{
			"suggestion": "Set GEMINI_API_KEY using the commands shown above",
			"docs_url":   "https://aistudio.google.com/app/apikey",
		})
	}

	// Get retry configuration from API package
	maxRetries, retryDelay = api.GetRetryConfig()

	// Ensure we have sensible defaults
	if maxRetries <= 0 {
		maxRetries = 3
	}

	if retryDelay <= 0 {
		retryDelay = 5
	}

	Debug(fmt.Sprintf("[GEMINI]: 🔑 Using API key (length: %d)", len(apiKey)))
	Debug(fmt.Sprintf("[GEMINI]: ⚙️ Retry config: maxRetries=%d, retryDelay=%d", maxRetries, retryDelay))

	ctx := context.Background()
	Debug("[GEMINI]: 🔐 Initializing Gemini client...")
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		Error("[GEMINI]: 🚨 Failed to initialize Gemini client: " + err.Error())
		return "", NewAPIError("Failed to initialize Gemini client", err, map[string]interface{}{
			"api_key_length": len(apiKey),
		})
	}
	defer client.Close()
	Debug("[GEMINI]: ✅ Gemini client initialized successfully")

	Debug("[GEMINI]: ⚙️ Configuring Gemini model...")
	model := client.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.5)
	model.SetMaxOutputTokens(100)
	model.ResponseMIMEType = "application/json"
	Debug("[GEMINI]: ⚙️ Model configuration: temperature=0.5, max_tokens=100, mime=application/json")

	// Configure safety settings to be more permissive for code content
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
	}

	// Build the system instruction with default guidelines
	baseInstruction := `
	Generate and return only a commit message as JSON with the key "message".
	Follow these guidelines for the commit message:
	• Capitalize the first word, omit final punctuation. If using conventional commits, use lowercase for the commit type.
	• Use imperative mood in the subject line.
	• Include a commit type (e.g. fix, update, refactor, bump).
	• Limit the first line to ≤ 50 characters, subsequent lines ≤ 72.
	• Be concise and direct; avoid filler words.
	• Do not include newline characters (\n) or similar formatting.

	The commit type can include the following:
	feat – a new feature
	fix – a bug fix
	chore – non-source changes
	refactor – refactored code
	docs – documentation updates
	style – formatting changes
	test – tests
	perf – performance improvements
	ci – continuous integration
	build – build system changes
	revert – revert a previous commit
	`

	// Check for user-provided custom instructions
	if len(customInstructions) > 0 && customInstructions[0] != "" {
		userInstructions := customInstructions[0]
		Debug("[GEMINI]: 📝 Found custom commit instructions")

		// Sanitize the instructions to prevent misuse (warnings handled inside function)
		sanitized := SanitizeUserInstructions(userInstructions)

		// Add user instructions at the beginning of the base instruction
		baseInstruction = `
	Generate and return only a commit message as JSON with the key "message".
	
	CUSTOM INSTRUCTIONS FROM USER:
	` + sanitized + `
	 
	Additionally, follow these guidelines strictly for the commit message:
	• Limit the first line to ≤ 50 characters, subsequent lines ≤ 72.
	• Be concise and direct; avoid filler words.
	• Do not include newline characters (\n) or similar formatting.
	`
	}

	// Save the instruction for testing
	lastSystemInstruction = baseInstruction

	model.SystemInstruction = genai.NewUserContent(genai.Text(baseInstruction))

	var promptBuilder strings.Builder
	promptBuilder.WriteString("Summarize the following file changes:\n\n")
	for file, data := range contextData {
		promptBuilder.WriteString(fmt.Sprintf("File: %s\nType: %s\nDiff:\n%s\n\n", file, data["type"], data["diff"]))
	}
	prompt := promptBuilder.String()

	// Debug logging for API request
	Debug(fmt.Sprintf("[GEMINI]: 📤 Making API request with prompt length: %d characters", len(prompt)))
	Debug(fmt.Sprintf("[GEMINI]: 📤 Context data files: %d", len(contextData)))

	Debug("[GEMINI]: Max retries set to: " + fmt.Sprintf("%d", maxRetries))

	var resp *genai.GenerateContentResponse
	var lastErr error

	// Ensure maxRetries is at least 1
	if maxRetries < 1 {
		Debug("[GEMINI]: maxRetries was less than 1, setting to default (3)")
		maxRetries = 3
	}

	for retries := 0; retries < maxRetries; retries++ {
		Debug(fmt.Sprintf("[GEMINI]: 🔄 Attempt %d/%d - Calling GenerateContent", retries+1, maxRetries))

		resp, err = model.GenerateContent(ctx, genai.Text(prompt))
		lastErr = err

		Debug(fmt.Sprintf("[GEMINI]: 📥 API call completed. Error: %v, Response nil: %v",
			err != nil, resp == nil))

		// Successful response
		if err == nil && resp != nil && len(resp.Candidates) > 0 &&
			resp.Candidates[0] != nil && resp.Candidates[0].Content != nil &&
			len(resp.Candidates[0].Content.Parts) > 0 {
			Debug("[GEMINI]: ✅ Received valid response")
			break
		}

		// Handle errors
		if err != nil {
			Debug(fmt.Sprintf("[GEMINI]: ⚠️ Error received: %v", err))
			if strings.Contains(err.Error(), "googleapi: Error 429: You exceeded your current quota") ||
				status.Code(err) == 14 { // Retry on specific 429 error or UNAVAILABLE
				Warning("[GEMINI]: ⚠️ Quota exceeded or service unavailable. Retrying in " +
					fmt.Sprintf("%d", retryDelay) + " seconds... (attempt " +
					fmt.Sprintf("%d/%d", retries+1, maxRetries) + ")")
			} else {
				Debug(fmt.Sprintf("[GEMINI]: ⚠️ API error: %v. Retrying...", err))
			}
		} else if resp == nil {
			Debug("[GEMINI]: ⚠️ Received nil response. Retrying...")
		} else if len(resp.Candidates) == 0 {
			Debug("[GEMINI]: ⚠️ No candidates in response. Retrying...")
		} else {
			Debug("[GEMINI]: ⚠️ Invalid response structure. Retrying...")
		}

		// Don't sleep on the last retry
		if retries < maxRetries-1 {
			Debug(fmt.Sprintf("[GEMINI]: 😴 Sleeping for %d seconds before next retry...", retryDelay))
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}
	}

	// If we still have an error after all retries, return it
	if lastErr != nil {
		return "", NewAPIError("Failed to generate content from Gemini after "+
			fmt.Sprintf("%d", maxRetries)+" attempts", lastErr, map[string]interface{}{
			"retries_attempted": maxRetries,
		})
	}

	// Comprehensive response validation
	if resp == nil {
		Error("[GEMINI]: ❌ Received nil response from Gemini API.")
		return "", NewAPIError("Received nil response from Gemini API", nil, map[string]interface{}{
			"context_files": len(contextData),
		})
	}

	if len(resp.Candidates) == 0 {
		Error("[GEMINI]: ❌ No candidates returned by Gemini.")
		return "", NewAPIError("No candidates returned by Gemini API", nil, map[string]interface{}{
			"context_files":   len(contextData),
			"prompt_feedback": fmt.Sprintf("%+v", resp.PromptFeedback),
		})
	}

	candidate := resp.Candidates[0]
	if candidate == nil {
		Error("[GEMINI]: ❌ First candidate is nil.")
		return "", NewAPIError("First candidate is nil", nil, map[string]interface{}{
			"total_candidates": len(resp.Candidates),
		})
	}

	if candidate.Content == nil {
		Error("[GEMINI]: ❌ Candidate content is nil.")
		return "", NewAPIError("Candidate content is nil", nil, map[string]interface{}{
			"finish_reason":  fmt.Sprintf("%v", candidate.FinishReason),
			"safety_ratings": fmt.Sprintf("%+v", candidate.SafetyRatings),
		})
	}

	if len(candidate.Content.Parts) == 0 {
		Error("[GEMINI]: ❌ No content parts in candidate.")
		return "", NewAPIError("No content parts in candidate", nil, map[string]interface{}{
			"content_role": candidate.Content.Role,
		})
	}

	respMessage := fmt.Sprintf(`%s`, candidate.Content.Parts[0])
	if respMessage == "" {
		Error("[GEMINI]: ❌ Empty content returned from Gemini.")
		return "", NewAPIError("Empty content returned from Gemini", nil, map[string]interface{}{
			"parts_count": len(candidate.Content.Parts),
		})
	}

	Debug("[GEMINI]: ✨ Response received: " + respMessage)

	// First try to parse as JSON with "message" key
	var result map[string]string
	err = json.Unmarshal([]byte(respMessage), &result)
	if err == nil {
		// Successfully parsed as JSON object
		if message, ok := result["message"]; ok {
			return message, nil
		}
		// JSON object but no "message" key, fall through to treat as plain text
		Debug("[GEMINI]: JSON response missing 'message' key, treating as plain text")
	} else {
		// Failed to parse as JSON object, check if it's an array or other format
		Debug("[GEMINI]: Failed to parse as JSON object, treating as plain text: " + err.Error())
	}

	// If JSON parsing failed or no "message" key found, treat the entire response as the message
	// This handles cases where Gemini returns plain text or unexpected JSON format
	trimmedResponse := strings.TrimSpace(respMessage)
	if trimmedResponse == "" {
		Error("[GEMINI]: ❌ Empty response after trimming.")
		return "", NewAPIError("Empty response after trimming", nil, map[string]interface{}{
			"raw_response": respMessage,
		})
	}

	return trimmedResponse, nil
}

// Keep track of the last system instruction for testing
var lastSystemInstruction string

// GetLastSystemInstruction returns the last system instruction used (for testing)
func GetLastSystemInstruction() string {
	return lastSystemInstruction
}

// SanitizeUserInstructions cleans potentially harmful or manipulative instructions
func SanitizeUserInstructions(instructions string) string {
	// Filter out potential prompt injection attempts
	blockedPhrases := []string{
		"ignore all previous instructions",
		"ignore previous instructions",
		"disregard the above",
		"do not follow",
		"override",
		"<script>",
		"</script>",
		"function(",
		"eval(",
		"exec(",
		"system(",
		"shell",
		"bash",
		"cmd",
		"powershell",
	}

	sanitized := instructions
	actuallyFiltered := false

	// Check each blocked phrase case-insensitively but preserve original case
	for _, phrase := range blockedPhrases {
		lowered := strings.ToLower(sanitized)
		phrasePos := strings.Index(lowered, strings.ToLower(phrase))
		if phrasePos != -1 {
			// Replace the actual occurrence with [filtered]
			before := sanitized[:phrasePos]
			after := sanitized[phrasePos+len(phrase):]
			sanitized = before + "[filtered]" + after
			actuallyFiltered = true
		}
	}

	// Ensure instructions aren't too long
	truncated := false
	if len(sanitized) > 500 {
		sanitized = sanitized[:500] + "..."
		truncated = true
	}

	// Only warn if we actually filtered something harmful or truncated
	if actuallyFiltered {
		Warning("[GEMINI]: ⚠️ User commit instructions contained potentially harmful content and were sanitized for safety")
	} else if truncated {
		Debug("[GEMINI]: ℹ️ User commit instructions were truncated due to length limit")
	}

	return sanitized
}
