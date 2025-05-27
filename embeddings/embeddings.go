package embeddings

import (
	"GitCury/config"
	"GitCury/utils"
	"context"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

// Circuit breaker for rate limiting
var (
	lastFailureTime      time.Time
	consecutiveFailures   int
	circuitBreakerThreshold = 3
	circuitBreakerTimeout = 5 * time.Minute
)

// checkCircuitBreaker returns true if we should skip the request due to circuit breaker
func checkCircuitBreaker() bool {
	if consecutiveFailures >= circuitBreakerThreshold {
		if time.Since(lastFailureTime) < circuitBreakerTimeout {
			utils.Warning("[EMBEDDINGS]: Circuit breaker active - skipping request to prevent further rate limiting")
			return true
		}
		// Reset circuit breaker after timeout
		consecutiveFailures = 0
	}
	return false
}

func GenerateEmbedding(text string) ([]float32, error) {
	// Check circuit breaker first
	if checkCircuitBreaker() {
		return nil, utils.NewAPIError(
			"Circuit breaker active due to repeated rate limit errors",
			nil,
			map[string]interface{}{
				"consecutiveFailures": consecutiveFailures,
				"nextRetryAfter":     time.Until(lastFailureTime.Add(circuitBreakerTimeout)).String(),
			},
		)
	}

	// Get the API key from config or environment
	apiKeyInterface := config.Get("GEMINI_API_KEY")
	apiKey, ok := apiKeyInterface.(string)
	if !ok || apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, utils.NewConfigError(
				"Google API key not found",
				nil,
				map[string]interface{}{
					"configKey": "GEMINI_API_KEY",
					"envVar":    "GEMINI_API_KEY",
				},
			)
		}
	}

	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	var embedding *genai.ContentEmbedding

	// Initialize the client outside the retry loop for better performance
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, utils.NewAPIError(
			"Error creating Gemini client",
			err,
			map[string]interface{}{
				"apiProvider": "Google Gemini",
			},
		)
	}

	// Prepare the content once for all retries
	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	// Define the operation to retry
	embedOperation := func() error {
		result, err := client.Models.EmbedContent(ctx,
			"text-embedding-004",
			contents,
			nil,
		)
		if err != nil {
			// Check for rate limit errors and update circuit breaker
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
				consecutiveFailures++
				lastFailureTime = time.Now()
				utils.Warning("[EMBEDDINGS]: Rate limit detected, updating circuit breaker")
			}
			return utils.NewAPIError(
				"Error getting embeddings from Gemini API",
				err,
				map[string]interface{}{
					"modelName":  "text-embedding-004",
					"textLength": len(text),
				},
			)
		}

		// Reset circuit breaker on success
		consecutiveFailures = 0

		if len(result.Embeddings) == 0 || result.Embeddings[0] == nil {
			return utils.NewAPIError(
				"Received empty embedding response from API",
				nil,
				map[string]interface{}{
					"modelName": "text-embedding-004",
				},
			)
		}

		embedding = result.Embeddings[0]
		return nil
	}

	// Use the retry mechanism with a more conservative configuration for rate limit issues
	retryConfig := utils.RetryConfig{
		MaxRetries:   3,                      // Reduced from 10
		InitialDelay: 5 * time.Second,       // Reduced from 30
		MaxDelay:     30 * time.Second,      // Reduced from 120
		Factor:       2.0,
	}

	err = utils.WithRetry(ctx, "GetEmbeddings", retryConfig, embedOperation)
	if err != nil {
		return nil, err // Already wrapped with context by WithRetry
	}

	flatEmbeddings := embedding.Values

	if len(flatEmbeddings) == 0 {
		return nil, utils.NewAPIError(
			"Received empty embedding vector from API",
			nil,
			map[string]interface{}{
				"modelName":      "text-embedding-004",
				"responseStatus": "empty vector",
			},
		)
	}

	return flatEmbeddings, nil
}

func KMeans(data [][]float32, k int, maxIter int) ([]int, error) {
	if k <= 0 || len(data) == 0 {
		return nil, utils.NewValidationError(
			"Invalid parameters for KMeans clustering",
			nil,
			map[string]interface{}{
				"k":             k,
				"dataPoints":    len(data),
				"maxIterations": maxIter,
			},
		)
	}
	if len(data) < k {
		return nil, utils.NewValidationError(
			"Number of clusters cannot exceed data points",
			nil,
			map[string]interface{}{
				"k":          k,
				"dataPoints": len(data),
			},
		)
	}

	n := len(data)
	dim := len(data[0])
	labels := make([]int, n)
	centroids := make([][]float32, k)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	perm := rng.Perm(n)
	for i := 0; i < k; i++ {
		centroids[i] = make([]float32, dim)
		copy(centroids[i], data[perm[i]])
	}

	for iter := 0; iter < maxIter; iter++ {
		for i := 0; i < n; i++ {
			labels[i] = closestCentroid(data[i], centroids)
		}

		newCentroids := make([][]float32, k)
		counts := make([]int, k)

		for i := 0; i < k; i++ {
			newCentroids[i] = make([]float32, dim)
		}

		for i := 0; i < n; i++ {
			label := labels[i]
			counts[label]++
			for j := 0; j < dim; j++ {
				newCentroids[label][j] += data[i][j]
			}
		}

		for i := 0; i < k; i++ {
			if counts[i] == 0 {
				newCentroids[i] = make([]float32, dim)
				copy(newCentroids[i], data[rng.Intn(n)])
			} else {
				for j := 0; j < dim; j++ {
					newCentroids[i][j] /= float32(counts[i])
				}
			}
		}

		centroids = newCentroids
	}

	return labels, nil
}

func closestCentroid(point []float32, centroids [][]float32) int {
	minDist := float64(math.MaxFloat64)
	minIndex := 0

	for i, c := range centroids {
		dist := float64(0.0)
		for j := range point {
			d := float64(point[j] - c[j])
			dist += d * d
		}
		if dist < minDist {
			minDist = dist
			minIndex = i
		}
	}
	return minIndex
}

// GetFileDiff gets the diff for a file using git commands
func GetFileDiff(filePath string) (string, error) {
	// This is a simplified version for testing
	// In real implementation, this would use git commands
	return "mock diff for " + filePath, nil
}

// GenerateCommitMessage generates a commit message for the given file path
func GenerateCommitMessage(filePath string) (string, error) {
	// This is a simplified version for testing
	// In real implementation, this would use AI to generate commit messages
	return "feat: update " + filePath, nil
}

// CosineSimilarity calculates the cosine similarity between two vectors
func CosineSimilarity(vec1, vec2 []float32) float32 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := range vec1 {
		dotProduct += float64(vec1[i] * vec2[i])
		normA += float64(vec1[i] * vec1[i])
		normB += float64(vec2[i] * vec2[i])
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}
