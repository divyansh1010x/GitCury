package embeddings

import (
	"os"
	"GitCury/config"
	"errors"
	"math"
	"fmt"
	"context"
	"math/rand"
	"time"
	
	"google.golang.org/genai"

)

func GenerateEmbedding(text string) ([]float32, error) {
	var err error

	apiKeyInterface := config.Get("GEMINI_API_KEY")
	apiKey, ok := apiKeyInterface.(string)
	if !ok || apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Google API key not found")
		}
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %v", err)
	}

	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	result, err := client.Models.EmbedContent(ctx,
		"gemini-embedding-exp-03-07",
		contents,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting embeddings: %v", err)
	}

    embedding := result.Embeddings[0]
    
    flatEmbeddings := embedding.Values  

    if flatEmbeddings == nil {
        return nil, fmt.Errorf("embedding vector is nil")
    }

    return flatEmbeddings, nil
}

func KMeans(data [][]float32, k int, maxIter int) ([]int, error) {
	if k <= 0 || len(data) == 0 {
		return nil, errors.New("invalid parameters for KMeans")
	}
	if len(data) < k {
		return nil, errors.New("number of clusters cannot exceed data points")
	}

	n := len(data)
	dim := len(data[0])
	labels := make([]int, n)
	centroids := make([][]float32, k)

	rand.Seed(time.Now().UnixNano())
	perm := rand.Perm(n)
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
				copy(newCentroids[i], data[rand.Intn(n)])
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