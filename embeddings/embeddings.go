package embeddings

import (
	"GitCury/config"
	"GitCury/utils"
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

// Circuit breaker for rate limiting
var (
	lastFailureTime         time.Time
	consecutiveFailures     int
	circuitBreakerThreshold = 3
	circuitBreakerTimeout   = 5 * time.Minute
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
				"nextRetryAfter":      time.Until(lastFailureTime.Add(circuitBreakerTimeout)).String(),
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
		MaxRetries:   3,                // Reduced from 10
		InitialDelay: 5 * time.Second,  // Reduced from 30
		MaxDelay:     30 * time.Second, // Reduced from 120
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

	rng := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // Non-cryptographic use, data shuffling
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

// K-means++ initialization for better centroid selection
func kMeansPlusPlus(data [][]float32, k int) [][]float32 {
	n := len(data)
	dim := len(data[0])
	centroids := make([][]float32, k)
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // Non-cryptographic use, clustering initialization

	// Choose first centroid randomly
	centroids[0] = make([]float32, dim)
	copy(centroids[0], data[rng.Intn(n)])

	// Choose remaining centroids based on distance probability
	for i := 1; i < k; i++ {
		distances := make([]float64, n)
		totalDistance := 0.0

		// Calculate minimum distance to existing centroids for each point
		for j := 0; j < n; j++ {
			minDist := math.MaxFloat64
			for c := 0; c < i; c++ {
				dist := euclideanDistance(data[j], centroids[c])
				if dist < minDist {
					minDist = dist
				}
			}
			distances[j] = minDist * minDist // Square the distance for probability
			totalDistance += distances[j]
		}

		// Choose next centroid with probability proportional to squared distance
		if totalDistance > 0 {
			target := rng.Float64() * totalDistance
			cumulative := 0.0

			for j := 0; j < n; j++ {
				cumulative += distances[j]
				if cumulative >= target {
					centroids[i] = make([]float32, dim)
					copy(centroids[i], data[j])
					break
				}
			}
		} else {
			// Fallback to random selection if all distances are zero
			centroids[i] = make([]float32, dim)
			copy(centroids[i], data[rng.Intn(n)])
		}
	}

	return centroids
}

// Euclidean distance helper function
func euclideanDistance(a, b []float32) float64 {
	var sum float64
	for i := 0; i < len(a); i++ {
		diff := float64(a[i] - b[i])
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// Enhanced K-means with K-means++ initialization and convergence detection
func KMeansOptimized(data [][]float32, k int, maxIter int) ([]int, error) {
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

	// Use K-means++ initialization for better starting centroids
	centroids := kMeansPlusPlus(data, k)

	convergenceThreshold := 0.001
	var prevCentroids [][]float32

	for iter := 0; iter < maxIter; iter++ {
		// Assign points to closest centroids
		for i := 0; i < n; i++ {
			labels[i] = closestCentroid(data[i], centroids)
		}

		// Save previous centroids for convergence check
		prevCentroids = make([][]float32, k)
		for i := 0; i < k; i++ {
			prevCentroids[i] = make([]float32, dim)
			copy(prevCentroids[i], centroids[i])
		}

		// Update centroids
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

		// Handle empty clusters and normalize
		rng := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // Non-cryptographic use, clustering initialization
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

		// Check for convergence
		if iter > 0 {
			converged := true
			for i := 0; i < k; i++ {
				if euclideanDistance(centroids[i], prevCentroids[i]) > convergenceThreshold {
					converged = false
					break
				}
			}
			if converged {
				utils.Debug(fmt.Sprintf("[EMBEDDINGS]: K-means converged after %d iterations", iter+1))
				break
			}
		}
	}

	return labels, nil
}

// Hierarchical clustering using agglomerative approach
func HierarchicalClustering(data [][]float32, maxClusters int, threshold float32) ([]int, error) {
	if len(data) == 0 {
		return nil, utils.NewValidationError(
			"Cannot perform hierarchical clustering on empty data",
			nil,
			map[string]interface{}{
				"dataPoints": 0,
			},
		)
	}

	n := len(data)

	// Initialize each point as its own cluster
	clusters := make([][]int, n)
	labels := make([]int, n)
	for i := 0; i < n; i++ {
		clusters[i] = []int{i}
		labels[i] = i
	}

	// Calculate initial distance matrix
	distances := make([][]float32, n)
	for i := 0; i < n; i++ {
		distances[i] = make([]float32, n)
		for j := i + 1; j < n; j++ {
			dist := float32(euclideanDistance(data[i], data[j]))
			distances[i][j] = dist
			distances[j][i] = dist
		}
	}

	numClusters := n
	clusterID := n

	// Merge clusters until we reach the desired number or threshold
	for numClusters > maxClusters {
		// Find the two closest clusters
		minDist := float32(math.MaxFloat32)
		var mergeI, mergeJ int = -1, -1

		for i := 0; i < len(clusters); i++ {
			if len(clusters[i]) == 0 {
				continue
			}
			for j := i + 1; j < len(clusters); j++ {
				if len(clusters[j]) == 0 {
					continue
				}

				// Calculate minimum distance between clusters (single linkage)
				clusterDist := float32(math.MaxFloat32)
				for _, pi := range clusters[i] {
					for _, pj := range clusters[j] {
						if distances[pi][pj] < clusterDist {
							clusterDist = distances[pi][pj]
						}
					}
				}

				if clusterDist < minDist && clusterDist <= threshold {
					minDist = clusterDist
					mergeI, mergeJ = i, j
				}
			}
		}

		// If no clusters are close enough to merge, break
		if mergeI == -1 || mergeJ == -1 {
			break
		}

		// Merge clusters
		clusters[mergeI] = append(clusters[mergeI], clusters[mergeJ]...)
		clusters[mergeJ] = nil // Mark as empty

		// Update labels
		for _, pointIdx := range clusters[mergeI] {
			labels[pointIdx] = clusterID
		}

		numClusters--
		clusterID++
	}

	// Renumber labels to be consecutive
	labelMap := make(map[int]int)
	newLabel := 0
	finalLabels := make([]int, n)

	for i := 0; i < n; i++ {
		if _, exists := labelMap[labels[i]]; !exists {
			labelMap[labels[i]] = newLabel
			newLabel++
		}
		finalLabels[i] = labelMap[labels[i]]
	}

	utils.Debug(fmt.Sprintf("[EMBEDDINGS]: Hierarchical clustering: %d points -> %d clusters", n, newLabel))
	return finalLabels, nil
}

// Adaptive clustering that chooses between K-means and hierarchical based on data characteristics
func AdaptiveClustering(data [][]float32, targetClusters int, maxIter int) ([]int, error) {
	if len(data) < 10 {
		// For small datasets, use hierarchical clustering
		return HierarchicalClustering(data, targetClusters, 0.5)
	}

	// Calculate data variance to determine clustering approach
	variance := calculateDataVariance(data)

	// If variance is low (data is tightly clustered), use hierarchical
	// If variance is high (data is spread out), use K-means
	if variance < 0.1 {
		utils.Debug("[EMBEDDINGS]: Using hierarchical clustering for low-variance data")
		return HierarchicalClustering(data, targetClusters, 0.3)
	} else {
		utils.Debug("[EMBEDDINGS]: Using optimized K-means for high-variance data")
		return KMeansOptimized(data, targetClusters, maxIter)
	}
}

// Calculate data variance helper function
func calculateDataVariance(data [][]float32) float64 {
	if len(data) == 0 {
		return 0
	}

	n := len(data)
	dim := len(data[0])

	// Calculate mean
	mean := make([]float64, dim)
	for i := 0; i < n; i++ {
		for j := 0; j < dim; j++ {
			mean[j] += float64(data[i][j])
		}
	}
	for j := 0; j < dim; j++ {
		mean[j] /= float64(n)
	}

	// Calculate variance
	var totalVariance float64
	for i := 0; i < n; i++ {
		for j := 0; j < dim; j++ {
			diff := float64(data[i][j]) - mean[j]
			totalVariance += diff * diff
		}
	}

	return totalVariance / float64(n*dim)
}

// ClusteringMetrics holds performance metrics for clustering operations
type ClusteringMetrics struct {
	Algorithm      string        `json:"algorithm"`
	DataPoints     int           `json:"dataPoints"`
	TargetClusters int           `json:"targetClusters"`
	ActualClusters int           `json:"actualClusters"`
	Duration       time.Duration `json:"duration"`
	Iterations     int           `json:"iterations,omitempty"`
	Silhouette     float64       `json:"silhouetteScore,omitempty"`
	Inertia        float64       `json:"inertia,omitempty"`
}

// Global metrics collection
var clusteringMetrics []ClusteringMetrics

// CalculateSilhouetteScore calculates the silhouette score for clustering quality assessment
func CalculateSilhouetteScore(data [][]float32, labels []int) float64 {
	n := len(data)
	if n < 2 {
		return 0
	}

	silhouetteScores := make([]float64, n)

	for i := 0; i < n; i++ {
		// Calculate a(i) - average distance to points in same cluster
		sameClusterDist := 0.0
		sameClusterCount := 0

		// Calculate b(i) - minimum average distance to points in other clusters
		otherClusterDists := make(map[int][]float64)

		for j := 0; j < n; j++ {
			if i == j {
				continue
			}

			dist := euclideanDistance(data[i], data[j])

			if labels[i] == labels[j] {
				sameClusterDist += dist
				sameClusterCount++
			} else {
				if _, exists := otherClusterDists[labels[j]]; !exists {
					otherClusterDists[labels[j]] = []float64{}
				}
				otherClusterDists[labels[j]] = append(otherClusterDists[labels[j]], dist)
			}
		}

		a := 0.0
		if sameClusterCount > 0 {
			a = sameClusterDist / float64(sameClusterCount)
		}

		b := math.MaxFloat64
		for _, dists := range otherClusterDists {
			avgDist := 0.0
			for _, d := range dists {
				avgDist += d
			}
			avgDist /= float64(len(dists))
			if avgDist < b {
				b = avgDist
			}
		}

		if b == math.MaxFloat64 {
			silhouetteScores[i] = 0
		} else {
			silhouetteScores[i] = (b - a) / math.Max(a, b)
		}
	}

	// Calculate average silhouette score
	avgSilhouette := 0.0
	for _, score := range silhouetteScores {
		avgSilhouette += score
	}
	return avgSilhouette / float64(n)
}

// CalculateInertia calculates the within-cluster sum of squares
func CalculateInertia(data [][]float32, labels []int) float64 {
	// Calculate centroids
	clusterMap := make(map[int][]int)
	for i, label := range labels {
		clusterMap[label] = append(clusterMap[label], i)
	}

	centroids := make(map[int][]float32)
	for label, indices := range clusterMap {
		dim := len(data[0])
		centroid := make([]float32, dim)

		for _, idx := range indices {
			for j := 0; j < dim; j++ {
				centroid[j] += data[idx][j]
			}
		}

		for j := 0; j < dim; j++ {
			centroid[j] /= float32(len(indices))
		}

		centroids[label] = centroid
	}

	// Calculate inertia
	var inertia float64
	for i, point := range data {
		centroid := centroids[labels[i]]
		dist := euclideanDistance(point, centroid)
		inertia += dist * dist
	}

	return inertia
}

// KMeansWithMetrics performs K-means clustering with performance monitoring
func KMeansWithMetrics(data [][]float32, k int, maxIter int) ([]int, ClusteringMetrics, error) {
	startTime := time.Now()
	labels, err := KMeansOptimized(data, k, maxIter)
	duration := time.Since(startTime)

	metrics := ClusteringMetrics{
		Algorithm:      "KMeans++",
		DataPoints:     len(data),
		TargetClusters: k,
		Duration:       duration,
		Iterations:     maxIter, // We could track actual iterations in KMeansOptimized
	}

	if err == nil && len(labels) > 0 {
		// Count actual clusters
		clusterSet := make(map[int]bool)
		for _, label := range labels {
			clusterSet[label] = true
		}
		metrics.ActualClusters = len(clusterSet)

		// Calculate quality metrics
		metrics.Silhouette = CalculateSilhouetteScore(data, labels)
		metrics.Inertia = CalculateInertia(data, labels)
	}

	// Store metrics
	clusteringMetrics = append(clusteringMetrics, metrics)

	utils.Debug(fmt.Sprintf("[EMBEDDINGS]: K-means metrics - Duration: %v, Silhouette: %.3f, Inertia: %.3f",
		duration, metrics.Silhouette, metrics.Inertia))

	return labels, metrics, err
}

// HierarchicalWithMetrics performs hierarchical clustering with performance monitoring
func HierarchicalWithMetrics(data [][]float32, maxClusters int, threshold float32) ([]int, ClusteringMetrics, error) {
	startTime := time.Now()
	labels, err := HierarchicalClustering(data, maxClusters, threshold)
	duration := time.Since(startTime)

	metrics := ClusteringMetrics{
		Algorithm:      "Hierarchical",
		DataPoints:     len(data),
		TargetClusters: maxClusters,
		Duration:       duration,
	}

	if err == nil && len(labels) > 0 {
		// Count actual clusters
		clusterSet := make(map[int]bool)
		for _, label := range labels {
			clusterSet[label] = true
		}
		metrics.ActualClusters = len(clusterSet)

		// Calculate quality metrics
		metrics.Silhouette = CalculateSilhouetteScore(data, labels)
		metrics.Inertia = CalculateInertia(data, labels)
	}

	// Store metrics
	clusteringMetrics = append(clusteringMetrics, metrics)

	utils.Debug(fmt.Sprintf("[EMBEDDINGS]: Hierarchical metrics - Duration: %v, Silhouette: %.3f, Clusters: %d",
		duration, metrics.Silhouette, metrics.ActualClusters))

	return labels, metrics, err
}

// GetClusteringMetrics returns all collected clustering metrics
func GetClusteringMetrics() []ClusteringMetrics {
	return clusteringMetrics
}

// ResetMetrics clears all collected metrics
func ResetMetrics() {
	clusteringMetrics = []ClusteringMetrics{}
}
