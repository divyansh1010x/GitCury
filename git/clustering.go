package git

import (
	"GitCury/config"
	"GitCury/embeddings"
	"GitCury/utils"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// FileCluster represents a group of related files
type FileCluster struct {
	Files       []string  `json:"files"`
	Similarity  float64   `json:"similarity"`
	ClusterType string    `json:"clusterType"`
	Created     time.Time `json:"created"`
}

// EmbeddingCache stores file embeddings to avoid regenerating them
type EmbeddingCache struct {
	RootFolder  string                    `json:"rootFolder"`
	Embeddings  map[string]FileCacheEntry `json:"embeddings"`
	LastUpdated time.Time                 `json:"lastUpdated"`
}

// FileCacheEntry stores file embedding with metadata
type FileCacheEntry struct {
	FilePath    string    `json:"filePath"`
	Embedding   []float32 `json:"embedding"`
	ContentHash string    `json:"contentHash"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// Enhanced caching structures for better performance
type CacheStats struct {
	TotalRequests  int       `json:"totalRequests"`
	CacheHits      int       `json:"cacheHits"`
	CacheMisses    int       `json:"cacheMisses"`
	HitRatio       float64   `json:"hitRatio"`
	LastCleanup    time.Time `json:"lastCleanup"`
	CacheSizeBytes int64     `json:"cacheSizeBytes"`
	EvictedEntries int       `json:"evictedEntries"`
}

type CacheConfig struct {
	MaxSize         int           `json:"maxSize"`         // Maximum number of entries
	MaxSizeBytes    int64         `json:"maxSizeBytes"`    // Maximum cache size in bytes
	TTL             time.Duration `json:"ttl"`             // Time to live for entries
	CleanupInterval time.Duration `json:"cleanupInterval"` // How often to clean expired entries
}

type LRUNode struct {
	Key   string
	Entry *FileCacheEntry
	Prev  *LRUNode
	Next  *LRUNode
}

type LRUCache struct {
	capacity int
	size     int
	cache    map[string]*LRUNode
	head     *LRUNode
	tail     *LRUNode
}

// Enhanced EmbeddingCache with LRU and statistics
type EnhancedEmbeddingCache struct {
	*EmbeddingCache
	Stats      CacheStats           `json:"stats"`
	Config     CacheConfig          `json:"config"`
	LRU        *LRUCache            `json:"-"` // Don't serialize LRU structure
	LastAccess map[string]time.Time `json:"lastAccess"`
}

// SimilarityMethod defines the method used for calculating similarity
type SimilarityMethod int

const (
	CosineSimilarity SimilarityMethod = iota
	JaccardSimilarity
	ManhattanSimilarity
	WeightedSemanticSimilarity
	HybridSimilarity
)

// String returns the string representation of SimilarityMethod
func (sm SimilarityMethod) String() string {
	switch sm {
	case CosineSimilarity:
		return "cosine"
	case JaccardSimilarity:
		return "jaccard"
	case ManhattanSimilarity:
		return "manhattan"
	case WeightedSemanticSimilarity:
		return "weighted_semantic"
	case HybridSimilarity:
		return "hybrid"
	default:
		return "unknown"
	}
}

// SmartClusterFiles groups files using multi-layered approach with configurable methods
// When targetClusters is 0 or negative, it uses threshold-based clustering without limits
func SmartClusterFiles(changedFiles []string, rootFolder string, targetClusters int) ([][]string, error) {
	if len(changedFiles) == 0 {
		return [][]string{}, nil
	}

	if len(changedFiles) == 1 {
		return [][]string{changedFiles}, nil
	}

	// Get clustering configuration
	clusteringConfig := config.GetClusteringConfig()

	// Determine if we should use threshold-based clustering (no cluster limit)
	useThresholdClustering := targetClusters <= 0

	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Starting configurable clustering for %d files, method: %s, threshold-based: %v",
		len(changedFiles), clusteringConfig.DefaultMethod, useThresholdClustering))

	// Enable benchmarking if configured
	if clusteringConfig.Performance.EnableBenchmarking {
		EnableBenchmarking()
		defer DisableBenchmarking()
	}

	// Check if specific method is requested (not auto)
	if clusteringConfig.DefaultMethod != "auto" && !clusteringConfig.EnableFallbackMethods {
		return executeSpecificMethod(changedFiles, rootFolder, targetClusters, clusteringConfig.DefaultMethod, useThresholdClustering)
	}

	// Use multi-layered approach (auto method or with fallbacks enabled)

	// Layer 1: Directory-based clustering
	if config.IsMethodEnabled(config.DirectoryMethod) {
		dirClusters, dirConfidence := directoryBasedClustering(changedFiles, rootFolder, targetClusters)
		dirThreshold := config.GetConfidenceThreshold(config.DirectoryMethod)
		dirSimilarity := config.GetSimilarityThreshold(config.DirectoryMethod)

		if dirConfidence >= dirThreshold && (!useThresholdClustering || validateClustersByThreshold(dirClusters, rootFolder, dirSimilarity)) {
			utils.Debug("[GIT.CLUSTER]: Directory-based clustering successful with configured thresholds")
			return dirClusters, nil
		}
	}

	// Layer 2: Pattern-based clustering
	if config.IsMethodEnabled(config.PatternMethod) {
		patternClusters, patternConfidence := patternBasedClustering(changedFiles, targetClusters)
		patternThreshold := config.GetConfidenceThreshold(config.PatternMethod)
		patternSimilarity := config.GetSimilarityThreshold(config.PatternMethod)

		if patternConfidence >= patternThreshold && (!useThresholdClustering || validateClustersByThreshold(patternClusters, rootFolder, patternSimilarity)) {
			utils.Debug("[GIT.CLUSTER]: Pattern-based clustering successful with configured thresholds")
			return patternClusters, nil
		}
	}

	// Layer 3: Cached embedding clustering
	if config.IsMethodEnabled(config.CachedMethod) {
		cachedClusters, cachedConfidence, cacheHitRatio := cachedEmbeddingClustering(changedFiles, rootFolder, targetClusters)
		cachedThreshold := config.GetConfidenceThreshold(config.CachedMethod)
		cachedSimilarity := config.GetSimilarityThreshold(config.CachedMethod)
		minCacheHitRatio := clusteringConfig.Methods.Cached.MinCacheHitRatio

		if cachedConfidence >= cachedThreshold && cacheHitRatio >= minCacheHitRatio && (!useThresholdClustering || validateClustersByThreshold(cachedClusters, rootFolder, cachedSimilarity)) {
			utils.Debug("[GIT.CLUSTER]: Cached embedding clustering successful with configured parameters")
			return cachedClusters, nil
		}
	}

	// Layer 4: Smart sampling for large file sets
	if config.IsMethodEnabled(config.SemanticMethod) && len(changedFiles) > clusteringConfig.MaxFilesForSemanticClustering {
		return smartSamplingClustering(changedFiles, rootFolder, targetClusters, useThresholdClustering)
	}

	// Layer 5: Full semantic clustering (fallback)
	if config.IsMethodEnabled(config.SemanticMethod) {
		return fullSemanticClustering(changedFiles, rootFolder, targetClusters, useThresholdClustering)
	}

	// If all methods are disabled, fall back to single file clusters
	utils.Warning("[GIT.CLUSTER]: All clustering methods disabled, using single file clusters")
	return createSingleFileClusters(changedFiles), nil
}

// validateClustersByThreshold checks if clusters meet similarity threshold requirements
func validateClustersByThreshold(clusters [][]string, rootFolder string, threshold float64) bool {
	for _, cluster := range clusters {
		if len(cluster) <= 1 {
			continue // Single file clusters are always valid
		}

		// Calculate average similarity within cluster
		avgSimilarity := calculateClusterSimilarity(cluster, rootFolder)
		if avgSimilarity < threshold {
			utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Cluster failed threshold validation: %.2f < %.2f", avgSimilarity, threshold))
			return false
		}
	}
	return true
}

// calculateClusterSimilarity calculates average similarity within a cluster
func calculateClusterSimilarity(files []string, rootFolder string) float64 {
	if len(files) <= 1 {
		return 1.0
	}

	// Use file extension and directory similarity as proxy
	extSimilarity := calculateExtensionSimilarity(files)
	dirSimilarity := calculateDirectorySimilarity(files, rootFolder)

	return (extSimilarity + dirSimilarity) / 2.0
}

// calculateExtensionSimilarity calculates similarity based on file extensions
func calculateExtensionSimilarity(files []string) float64 {
	if len(files) <= 1 {
		return 1.0
	}

	extMap := make(map[string]int)
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		extMap[ext]++
	}

	// Find the most common extension
	maxCount := 0
	for _, count := range extMap {
		if count > maxCount {
			maxCount = count
		}
	}

	return float64(maxCount) / float64(len(files))
}

// calculateDirectorySimilarity calculates similarity based on directory structure
func calculateDirectorySimilarity(files []string, rootFolder string) float64 {
	if len(files) <= 1 {
		return 1.0
	}

	dirs := make(map[string]int)
	for _, file := range files {
		relPath, _ := filepath.Rel(rootFolder, file)
		dir := filepath.Dir(relPath)
		dirs[dir]++
	}

	// Find the most common directory
	maxCount := 0
	for _, count := range dirs {
		if count > maxCount {
			maxCount = count
		}
	}

	return float64(maxCount) / float64(len(files))
}

// directoryBasedClustering groups files by directory structure
func directoryBasedClustering(files []string, rootFolder string, targetClusters int) ([][]string, float64) {
	dirGroups := make(map[string][]string)

	for _, file := range files {
		relPath, err := filepath.Rel(rootFolder, file)
		if err != nil {
			relPath = file
		}

		dir := filepath.Dir(relPath)
		// Group by immediate parent directory
		dirGroups[dir] = append(dirGroups[dir], file)
	}

	clusters := make([][]string, 0, len(dirGroups))
	for _, group := range dirGroups {
		clusters = append(clusters, group)
	}

	// Sort clusters by size (largest first)
	sort.Slice(clusters, func(i, j int) bool {
		return len(clusters[i]) > len(clusters[j])
	})

	// If we have a target cluster count and it's less than what we have, merge smallest clusters
	if targetClusters > 0 && len(clusters) > targetClusters {
		clusters = mergeClusters(clusters, targetClusters)
	}

	// Calculate confidence based on how well files are grouped by directory
	confidence := calculateDirectoryGroupingConfidence(files, clusters, rootFolder)

	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Directory-based clustering: %d files -> %d clusters, confidence: %.2f",
		len(files), len(clusters), confidence))

	return clusters, confidence
}

// patternBasedClustering groups files by patterns and relationships
func patternBasedClustering(files []string, targetClusters int) ([][]string, float64) {
	// Group by file extensions
	extGroups := make(map[string][]string)
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		if ext == "" {
			ext = "no-extension"
		}
		extGroups[ext] = append(extGroups[ext], file)
	}

	// Find test-implementation relationships
	testImplRelations := findTestImplementationRelations(files)

	// Create clusters based on patterns
	clusters := make([][]string, 0)
	processed := make(map[string]bool)

	// First, handle test-implementation pairs
	for testFile, implFile := range testImplRelations {
		if !processed[testFile] && !processed[implFile] {
			cluster := []string{testFile, implFile}
			clusters = append(clusters, cluster)
			processed[testFile] = true
			processed[implFile] = true
		}
	}

	// Then group remaining files by extension
	for _, group := range extGroups {
		unprocessed := make([]string, 0)
		for _, file := range group {
			if !processed[file] {
				unprocessed = append(unprocessed, file)
			}
		}

		if len(unprocessed) > 0 {
			ext := ""
			if len(group) > 0 {
				ext = strings.ToLower(filepath.Ext(group[0]))
			}
			if len(unprocessed) <= 3 || ext == ".md" || ext == ".txt" {
				// Keep small groups together, and documentation files together
				clusters = append(clusters, unprocessed)
			} else {
				// Split large groups into smaller clusters
				for len(unprocessed) > 0 {
					chunkSize := int(math.Min(3, float64(len(unprocessed))))
					chunk := unprocessed[:chunkSize]
					unprocessed = unprocessed[chunkSize:]
					clusters = append(clusters, chunk)
				}
			}
		}
	}

	// Merge clusters if we have too many for the target
	if targetClusters > 0 && len(clusters) > targetClusters {
		clusters = mergeClusters(clusters, targetClusters)
	}

	confidence := calculatePatternConfidence(files, clusters)

	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Pattern-based clustering: %d files -> %d clusters, confidence: %.2f",
		len(files), len(clusters), confidence))

	return clusters, confidence
}

// findTestImplementationRelations finds test files and their corresponding implementation files
func findTestImplementationRelations(files []string) map[string]string {
	relations := make(map[string]string)

	testFiles := make([]string, 0)
	implFiles := make([]string, 0)

	testPatterns := []*regexp.Regexp{
		regexp.MustCompile(`.*_test\.(go|js|ts|py|java|cpp|c)$`),
		regexp.MustCompile(`.*/test/.*\.(go|js|ts|py|java|cpp|c)$`),
		regexp.MustCompile(`.*/tests/.*\.(go|js|ts|py|java|cpp|c)$`),
		regexp.MustCompile(`.*\.test\.(js|ts)$`),
		regexp.MustCompile(`.*\.spec\.(js|ts)$`),
	}

	for _, file := range files {
		isTest := false
		for _, pattern := range testPatterns {
			if pattern.MatchString(strings.ToLower(file)) {
				testFiles = append(testFiles, file)
				isTest = true
				break
			}
		}
		if !isTest {
			implFiles = append(implFiles, file)
		}
	}

	// Try to match test files with implementation files
	for _, testFile := range testFiles {
		baseName := filepath.Base(testFile)
		testDir := filepath.Dir(testFile)

		// Remove test suffixes to find the implementation file
		implName := strings.ReplaceAll(baseName, "_test.", ".")
		implName = strings.ReplaceAll(implName, ".test.", ".")
		implName = strings.ReplaceAll(implName, ".spec.", ".")

		// Look for corresponding implementation file
		for _, implFile := range implFiles {
			implBaseName := filepath.Base(implFile)
			implDir := filepath.Dir(implFile)

			// Check if names match and directories are related
			if implBaseName == implName || strings.Contains(testDir, implDir) || strings.Contains(implDir, testDir) {
				relations[testFile] = implFile
				break
			}
		}
	}

	return relations
}

// cachedEmbeddingClustering uses enhanced cached embeddings when available
func cachedEmbeddingClustering(files []string, rootFolder string, targetClusters int) ([][]string, float64, float64) {
	enhancedCache := NewEnhancedEmbeddingCache(rootFolder)
	fileEmbeddings := make(map[string][]float32)
	cacheHits := 0

	// Check enhanced cache for existing embeddings
	for _, file := range files {
		if cached, exists := enhancedCache.GetEmbedding(file); exists {
			fileEmbeddings[file] = cached.Embedding
			cacheHits++
		}
	}

	cacheHitRatio := float64(cacheHits) / float64(len(files))
	stats := enhancedCache.GetStats()

	utils.Debug(fmt.Sprintf("[CACHE]: Enhanced cache stats - Hit ratio: %.3f, Total requests: %d, Cache size: %d bytes",
		stats.HitRatio, stats.TotalRequests, stats.CacheSizeBytes))

	// If we don't have enough cached embeddings, return early
	if cacheHitRatio < 0.3 {
		enhancedCache.Save()
		return createSingleFileClusters(files), 0.0, cacheHitRatio
	}

	// Generate embeddings for missing files with intelligent batching and rate limiting
	clusteringConfig := config.GetClusteringConfig()
	maxNewEmbeddings := calculateOptimalBatchSize(len(files), cacheHitRatio)
	newEmbeddings := 0
	recentMisses := len(files) - cacheHits
	rateLimitDelay := time.Duration(clusteringConfig.Methods.Semantic.RateLimitDelay) * time.Millisecond

	for _, file := range files {
		if _, exists := fileEmbeddings[file]; !exists && newEmbeddings < maxNewEmbeddings {
			diff, err := embeddings.GetFileDiff(file)
			if err != nil {
				utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Could not get diff for file: %s - %v", file, err))
				continue
			}

			// Use optimized diff size based on file type
			optimalSize := getOptimalDiffSize(file)
			if len(diff) > optimalSize {
				diff = diff[:optimalSize] + "... [truncated]"
			}

			embedding, err := embeddings.GenerateEmbedding(diff)
			if err != nil {
				utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Could not generate embedding for file: %s - %v", file, err))
				continue
			}

			fileEmbeddings[file] = embedding
			enhancedCache.PutEmbedding(file, embedding)
			newEmbeddings++

			// Use configurable adaptive delay
			delay := time.Duration(recentMisses*100) * time.Millisecond
			if delay < rateLimitDelay {
				delay = rateLimitDelay
			}
			if delay > 0 {
				time.Sleep(delay)
			}
		}
	}

	// Perform clustering using available embeddings
	if len(fileEmbeddings) < 2 {
		enhancedCache.Save()
		return createSingleFileClusters(files), 0.0, cacheHitRatio
	}

	clusters := performEmbeddingBasedClustering(fileEmbeddings, targetClusters)
	confidence := calculateEmbeddingClusterConfidence(clusters, fileEmbeddings)

	// Save enhanced cache with automatic cleanup
	enhancedCache.Save()

	finalStats := enhancedCache.GetStats()
	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Enhanced cached embedding clustering: %d files -> %d clusters, cache hit ratio: %.2f, confidence: %.2f, final cache hit ratio: %.3f",
		len(files), len(clusters), cacheHitRatio, confidence, finalStats.HitRatio))

	return clusters, confidence, cacheHitRatio
}

// smartSamplingClustering handles large file sets by sampling representative files
func smartSamplingClustering(files []string, rootFolder string, targetClusters int, useThresholdClustering bool) ([][]string, error) {
	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Using smart sampling for %d files", len(files)))

	// Sample representative files (max 8 to prevent API overload)
	sampleSize := int(math.Min(8, float64(len(files))/2))
	representatives := selectRepresentativeFiles(files, sampleSize)

	// Cluster representatives using embeddings
	reprClusters, err := fullSemanticClustering(representatives, rootFolder, -1, true) // Always use threshold for sampling
	if err != nil {
		return fallbackToPatterClustering(files, targetClusters), nil
	}

	// Assign remaining files to clusters based on similarity
	finalClusters := assignFilesToClusters(files, reprClusters, rootFolder)

	// Validate clusters meet threshold requirements if using threshold clustering
	if useThresholdClustering {
		finalClusters = filterClustersByThreshold(finalClusters, rootFolder, 0.4)
	}

	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Smart sampling clustering: %d files -> %d clusters",
		len(files), len(finalClusters)))

	return finalClusters, nil
}

// fullSemanticClustering performs complete semantic analysis
func fullSemanticClustering(files []string, rootFolder string, targetClusters int, useThresholdClustering bool) ([][]string, error) {
	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Performing full semantic clustering for %d files", len(files)))

	// Get clustering configuration for rate limiting
	clusteringConfig := config.GetClusteringConfig()
	rateLimitDelay := time.Duration(clusteringConfig.Methods.Semantic.RateLimitDelay) * time.Millisecond

	fileEmbeddings := make(map[string][]float32)

	// Generate embeddings for all files with configurable rate limiting
	for i, file := range files {
		// Rate limiting: add configurable delay between requests
		if i > 0 {
			time.Sleep(rateLimitDelay)
		}

		diff, err := GetFileDiff(file, rootFolder)
		if err != nil {
			utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Could not get diff for file: %s - %v", file, err))
			continue
		}

		// Limit diff size
		if len(diff) > 10000 {
			diff = diff[:10000] + "... [truncated]"
		}

		embedding, err := embeddings.GenerateEmbedding(diff)
		if err != nil {
			utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Could not generate embedding for file: %s - %v", file, err))
			continue
		}

		fileEmbeddings[file] = embedding
	}

	if len(fileEmbeddings) < 2 {
		return createSingleFileClusters(files), nil
	}

	// Perform clustering using configured similarity threshold
	var clusters [][]string
	if useThresholdClustering {
		semanticThreshold := config.GetSimilarityThreshold(config.SemanticMethod)
		clusters = performThresholdBasedClustering(fileEmbeddings, semanticThreshold)
	} else {
		clusters = performEmbeddingBasedClustering(fileEmbeddings, targetClusters)
	}

	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Full semantic clustering: %d files -> %d clusters",
		len(files), len(clusters)))

	return clusters, nil
}

// executeSpecificMethod runs a single specific clustering method
func executeSpecificMethod(files []string, rootFolder string, targetClusters int, methodName string, useThresholdClustering bool) ([][]string, error) {
	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Executing specific method: %s", methodName))

	switch methodName {
	case "directory":
		if !config.IsMethodEnabled(config.DirectoryMethod) {
			return nil, fmt.Errorf("directory clustering method is disabled")
		}
		clusters, _ := directoryBasedClustering(files, rootFolder, targetClusters)
		return clusters, nil

	case "pattern":
		if !config.IsMethodEnabled(config.PatternMethod) {
			return nil, fmt.Errorf("pattern clustering method is disabled")
		}
		clusters, _ := patternBasedClustering(files, targetClusters)
		return clusters, nil

	case "cached":
		if !config.IsMethodEnabled(config.CachedMethod) {
			return nil, fmt.Errorf("cached clustering method is disabled")
		}
		clusters, _, _ := cachedEmbeddingClustering(files, rootFolder, targetClusters)
		return clusters, nil

	case "semantic":
		if !config.IsMethodEnabled(config.SemanticMethod) {
			return nil, fmt.Errorf("semantic clustering method is disabled")
		}
		return fullSemanticClustering(files, rootFolder, targetClusters, useThresholdClustering)

	default:
		return nil, fmt.Errorf("unknown clustering method: %s", methodName)
	}
}

// performThresholdBasedClustering clusters files based on similarity threshold instead of target count
func performThresholdBasedClustering(fileEmbeddings map[string][]float32, threshold float64) [][]string {
	files := make([]string, 0, len(fileEmbeddings))
	embeddingVectors := make([][]float32, 0, len(fileEmbeddings))

	for file, embedding := range fileEmbeddings {
		files = append(files, file)
		embeddingVectors = append(embeddingVectors, embedding)
	}

	if len(files) <= 1 {
		return createSingleFileClusters(files)
	}

	// Use hierarchical clustering with threshold for better quality
	labels, metrics, err := embeddings.HierarchicalWithMetrics(embeddingVectors, len(files), float32(threshold))
	if err != nil {
		utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Hierarchical clustering failed: %v", err))
		// Fallback to original similarity-based approach
		return performOriginalThresholdClustering(fileEmbeddings, threshold)
	}

	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Threshold clustering metrics - Silhouette: %.3f, Clusters: %d",
		metrics.Silhouette, metrics.ActualClusters))

	// Group files by cluster labels
	clusterMap := make(map[int][]string)
	for i, label := range labels {
		clusterMap[label] = append(clusterMap[label], files[i])
	}

	clusters := make([][]string, 0, len(clusterMap))
	for _, cluster := range clusterMap {
		clusters = append(clusters, cluster)
	}

	return clusters
}

// performOriginalThresholdClustering is the fallback implementation with enhanced similarity
func performOriginalThresholdClustering(fileEmbeddings map[string][]float32, threshold float64) [][]string {
	files := make([]string, 0, len(fileEmbeddings))
	embeddings := make([][]float32, 0, len(fileEmbeddings))

	for file, embedding := range fileEmbeddings {
		files = append(files, file)
		embeddings = append(embeddings, embedding)
	}

	// Calculate similarity matrix using enhanced similarity methods
	similarities := make([][]float64, len(files))
	config := DefaultSimilarityConfig()

	for i := range similarities {
		similarities[i] = make([]float64, len(files))
		for j := range similarities[i] {
			if i == j {
				similarities[i][j] = 1.0
			} else {
				// Use hybrid similarity for better clustering quality
				similarities[i][j] = hybridSimilarity(embeddings[i], embeddings[j], config)
			}
		}
	}

	// Group files that have similarity above threshold
	clusters := make([][]string, 0)
	assigned := make([]bool, len(files))

	for i := 0; i < len(files); i++ {
		if assigned[i] {
			continue
		}

		cluster := []string{files[i]}
		assigned[i] = true

		// Find all files similar to this one
		for j := i + 1; j < len(files); j++ {
			if !assigned[j] && similarities[i][j] >= threshold {
				cluster = append(cluster, files[j])
				assigned[j] = true
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}

// performEmbeddingBasedClustering performs K-means clustering on embeddings
func performEmbeddingBasedClustering(fileEmbeddings map[string][]float32, targetClusters int) [][]string {
	if targetClusters <= 0 {
		targetClusters = int(math.Max(1, math.Min(float64(len(fileEmbeddings))/2, 5)))
	}

	files := make([]string, 0, len(fileEmbeddings))
	vectors := make([][]float32, 0, len(fileEmbeddings))

	for file, embedding := range fileEmbeddings {
		files = append(files, file)
		vectors = append(vectors, embedding)
	}

	if len(vectors) <= targetClusters {
		return createSingleFileClusters(files)
	}

	// Use adaptive clustering for better performance
	labels, metrics, err := embeddings.KMeansWithMetrics(vectors, targetClusters, 20)
	if err != nil {
		utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: K-means clustering failed: %v", err))
		return createSingleFileClusters(files)
	}

	// Log clustering performance metrics
	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Clustering completed - Algorithm: %s, Silhouette: %.3f, Duration: %v",
		metrics.Algorithm, metrics.Silhouette, metrics.Duration))

	// Group files by cluster labels
	clusterMap := make(map[int][]string)
	for i, label := range labels {
		clusterMap[label] = append(clusterMap[label], files[i])
	}

	clusters := make([][]string, 0, len(clusterMap))
	for _, cluster := range clusterMap {
		clusters = append(clusters, cluster)
	}

	return clusters
}

// LRU Cache implementation for embedding management
func NewLRUCache(capacity int) *LRUCache {
	head := &LRUNode{}
	tail := &LRUNode{}
	head.Next = tail
	tail.Prev = head

	return &LRUCache{
		capacity: capacity,
		size:     0,
		cache:    make(map[string]*LRUNode),
		head:     head,
		tail:     tail,
	}
}

func (lru *LRUCache) addToHead(node *LRUNode) {
	node.Prev = lru.head
	node.Next = lru.head.Next
	lru.head.Next.Prev = node
	lru.head.Next = node
}

func (lru *LRUCache) removeNode(node *LRUNode) {
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
}

func (lru *LRUCache) moveToHead(node *LRUNode) {
	lru.removeNode(node)
	lru.addToHead(node)
}

func (lru *LRUCache) removeTail() *LRUNode {
	lastNode := lru.tail.Prev
	lru.removeNode(lastNode)
	return lastNode
}

func (lru *LRUCache) Get(key string) (*FileCacheEntry, bool) {
	if node, exists := lru.cache[key]; exists {
		lru.moveToHead(node)
		return node.Entry, true
	}
	return nil, false
}

func (lru *LRUCache) Put(key string, entry *FileCacheEntry) *FileCacheEntry {
	if node, exists := lru.cache[key]; exists {
		node.Entry = entry
		lru.moveToHead(node)
		return nil
	}

	newNode := &LRUNode{Key: key, Entry: entry}

	if lru.size >= lru.capacity {
		tail := lru.removeTail()
		delete(lru.cache, tail.Key)
		lru.size--
		lru.addToHead(newNode)
		lru.cache[key] = newNode
		lru.size++
		return tail.Entry // Return evicted entry
	}

	lru.addToHead(newNode)
	lru.cache[key] = newNode
	lru.size++
	return nil
}

// Enhanced cache management functions
func NewEnhancedEmbeddingCache(rootFolder string) *EnhancedEmbeddingCache {
	baseCache := loadEmbeddingCache(rootFolder)

	config := CacheConfig{
		MaxSize:         1000,             // Maximum 1000 entries
		MaxSizeBytes:    50 * 1024 * 1024, // 50MB max cache size
		TTL:             24 * time.Hour,   // 24 hour TTL
		CleanupInterval: 1 * time.Hour,    // Cleanup every hour
	}

	enhanced := &EnhancedEmbeddingCache{
		EmbeddingCache: baseCache,
		Config:         config,
		LRU:            NewLRUCache(config.MaxSize),
		LastAccess:     make(map[string]time.Time),
		Stats: CacheStats{
			LastCleanup: time.Now(),
		},
	}

	// Initialize LRU with existing cache entries
	for key, entry := range baseCache.Embeddings {
		enhanced.LRU.Put(key, &entry)
		enhanced.LastAccess[key] = entry.LastUpdated
	}

	return enhanced
}

func (ec *EnhancedEmbeddingCache) GetEmbedding(filePath string) (*FileCacheEntry, bool) {
	ec.Stats.TotalRequests++

	// Check if entry exists and is valid
	if entry, exists := ec.LRU.Get(filePath); exists {
		// Check TTL
		if time.Since(ec.LastAccess[filePath]) < ec.Config.TTL {
			// Validate file hasn't changed
			if currentHash := getFileContentHash(filePath); currentHash == entry.ContentHash {
				ec.Stats.CacheHits++
				ec.LastAccess[filePath] = time.Now()
				ec.updateHitRatio()
				return entry, true
			}
		}
		// Remove expired or invalid entry
		delete(ec.Embeddings, filePath)
		delete(ec.LastAccess, filePath)
	}

	ec.Stats.CacheMisses++
	ec.updateHitRatio()
	return nil, false
}

func (ec *EnhancedEmbeddingCache) PutEmbedding(filePath string, embedding []float32) {
	entry := FileCacheEntry{
		FilePath:    filePath,
		Embedding:   embedding,
		ContentHash: getFileContentHash(filePath),
		LastUpdated: time.Now(),
	}

	// Check if we need to evict due to size constraints
	evicted := ec.LRU.Put(filePath, &entry)
	if evicted != nil {
		ec.Stats.EvictedEntries++
		delete(ec.Embeddings, evicted.FilePath)
		delete(ec.LastAccess, evicted.FilePath)
	}

	ec.Embeddings[filePath] = entry
	ec.LastAccess[filePath] = time.Now()
	ec.LastUpdated = time.Now()

	// Check cache size constraints
	ec.enforceMemoryLimits()
}

func (ec *EnhancedEmbeddingCache) updateHitRatio() {
	if ec.Stats.TotalRequests > 0 {
		ec.Stats.HitRatio = float64(ec.Stats.CacheHits) / float64(ec.Stats.TotalRequests)
	}
}

func (ec *EnhancedEmbeddingCache) enforceMemoryLimits() {
	// Calculate current cache size in bytes
	totalSize := int64(0)
	for _, entry := range ec.Embeddings {
		totalSize += int64(len(entry.Embedding)*4) + int64(len(entry.FilePath)) + int64(len(entry.ContentHash)) + 100 // overhead
	}
	ec.Stats.CacheSizeBytes = totalSize

	// Evict entries if over memory limit
	if totalSize > ec.Config.MaxSizeBytes {
		utils.Debug(fmt.Sprintf("[CACHE]: Memory limit exceeded (%d bytes), performing cleanup", totalSize))
		ec.performMemoryCleanup()
	}
}

func (ec *EnhancedEmbeddingCache) performMemoryCleanup() {
	// Sort entries by last access time (oldest first)
	type accessTime struct {
		path string
		time time.Time
	}

	var sortedEntries []accessTime
	for path, lastAccess := range ec.LastAccess {
		sortedEntries = append(sortedEntries, accessTime{path, lastAccess})
	}

	sort.Slice(sortedEntries, func(i, j int) bool {
		return sortedEntries[i].time.Before(sortedEntries[j].time)
	})

	// Remove oldest 25% of entries
	removeCount := len(sortedEntries) / 4
	if removeCount < 1 {
		removeCount = 1
	}

	for i := 0; i < removeCount && i < len(sortedEntries); i++ {
		path := sortedEntries[i].path
		delete(ec.Embeddings, path)
		delete(ec.LastAccess, path)
		ec.Stats.EvictedEntries++
	}

	utils.Debug(fmt.Sprintf("[CACHE]: Cleaned up %d old cache entries", removeCount))
}

func (ec *EnhancedEmbeddingCache) performScheduledCleanup() {
	if time.Since(ec.Stats.LastCleanup) < ec.Config.CleanupInterval {
		return
	}

	// Remove expired entries
	expiredCount := 0
	for path, lastAccess := range ec.LastAccess {
		if time.Since(lastAccess) > ec.Config.TTL {
			delete(ec.Embeddings, path)
			delete(ec.LastAccess, path)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		utils.Debug(fmt.Sprintf("[CACHE]: Removed %d expired cache entries", expiredCount))
	}

	ec.Stats.LastCleanup = time.Now()
	ec.enforceMemoryLimits()
}

func (ec *EnhancedEmbeddingCache) Save() {
	ec.performScheduledCleanup()
	saveEmbeddingCache(ec.EmbeddingCache)
}

func (ec *EnhancedEmbeddingCache) GetStats() CacheStats {
	return ec.Stats
}

// Core cache management functions
func saveEmbeddingCache(cache *EmbeddingCache) {
	cacheFile := getCacheFilePath(cache.RootFolder)

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0755); err != nil {
		utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Failed to create cache directory: %v", err))
		return
	}

	data, err := json.Marshal(cache)
	if err != nil {
		utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Failed to marshal cache: %v", err))
		return
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Failed to save cache file: %v", err))
	}
}

func getCacheFilePath(rootFolder string) string {
	homeDir, _ := os.UserHomeDir()
	hasher := sha256.Sum256([]byte(rootFolder))
	folderHash := hex.EncodeToString(hasher[:])[:12]
	return filepath.Join(homeDir, ".gitcury", fmt.Sprintf("embedding_cache_%s.json", folderHash))
}

func getFileContentHash(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	hasher := sha256.Sum256(content)
	return hex.EncodeToString(hasher[:])
}

// Helper functions for embedding cache management
func loadEmbeddingCache(rootFolder string) *EmbeddingCache {
	cacheFile := getCacheFilePath(rootFolder)

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return &EmbeddingCache{
			RootFolder:  rootFolder,
			Embeddings:  make(map[string]FileCacheEntry),
			LastUpdated: time.Now(),
		}
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Failed to read cache file: %v", err))
		return &EmbeddingCache{
			RootFolder:  rootFolder,
			Embeddings:  make(map[string]FileCacheEntry),
			LastUpdated: time.Now(),
		}
	}

	var cache EmbeddingCache
	if err := json.Unmarshal(data, &cache); err != nil {
		utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Failed to parse cache file: %v", err))
		return &EmbeddingCache{
			RootFolder:  rootFolder,
			Embeddings:  make(map[string]FileCacheEntry),
			LastUpdated: time.Now(),
		}
	}

	return &cache
}

func getCachedEmbedding(filePath string, cache *EmbeddingCache) (*FileCacheEntry, bool) { //nolint:unused // Future caching optimization
	entry, exists := cache.Embeddings[filePath]
	if !exists {
		return nil, false
	}

	// Check if file has been modified since caching
	currentHash := getFileContentHash(filePath)
	if currentHash != entry.ContentHash {
		return nil, false
	}

	return &entry, true
}

func saveEmbeddingToCache(filePath string, embedding []float32, content string, cache *EmbeddingCache) { //nolint:unused // Future caching optimization
	entry := FileCacheEntry{
		FilePath:    filePath,
		Embedding:   embedding,
		ContentHash: getFileContentHash(filePath),
		LastUpdated: time.Now(),
	}

	cache.Embeddings[filePath] = entry
	cache.LastUpdated = time.Now()
}

// Helper functions for enhanced caching optimization
func calculateOptimalBatchSize(totalFiles int, hitRatio float64) int {
	baseSize := 5

	// Increase batch size if cache is performing well
	if hitRatio > 0.8 {
		baseSize = 8
	} else if hitRatio > 0.6 {
		baseSize = 7
	} else if hitRatio < 0.3 {
		baseSize = 3
	}

	// Scale based on total files
	if totalFiles > 20 {
		baseSize = int(math.Min(float64(baseSize), float64(totalFiles)/4))
	}

	return baseSize
}

func getOptimalDiffSize(filePath string) int {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Different optimal sizes for different file types
	switch ext {
	case ".go", ".py", ".js", ".ts":
		return 12000 // Larger for code files
	case ".md", ".txt", ".rst":
		return 8000 // Medium for documentation
	case ".json", ".yaml", ".yml", ".xml":
		return 6000 // Smaller for config files
	default:
		return 10000 // Default size
	}
}

func calculateAdaptiveDelay(recentMisses int) time.Duration { //nolint:unused // Future rate limiting optimization
	baseDelay := 1000 * time.Millisecond

	// Increase delay based on cache miss rate
	if recentMisses > 15 {
		return baseDelay * 5 // 5 seconds for high miss rate
	} else if recentMisses > 10 {
		return baseDelay * 3 // 3 seconds for medium miss rate
	} else if recentMisses > 5 {
		return baseDelay * 2 // 2 seconds for low miss rate
	}

	return baseDelay // 1 second default
}

// Utility functions
func createSingleFileClusters(files []string) [][]string {
	clusters := make([][]string, len(files))
	for i, file := range files {
		clusters[i] = []string{file}
	}
	return clusters
}

func mergeClusters(clusters [][]string, targetCount int) [][]string {
	if len(clusters) <= targetCount {
		return clusters
	}

	// Merge smallest clusters first
	for len(clusters) > targetCount {
		// Find two smallest clusters
		minSize1, minSize2 := math.MaxInt32, math.MaxInt32
		idx1, idx2 := -1, -1

		for i, cluster := range clusters {
			size := len(cluster)
			if size < minSize1 {
				minSize2 = minSize1
				idx2 = idx1
				minSize1 = size
				idx1 = i
			} else if size < minSize2 {
				minSize2 = size
				idx2 = i
			}
		}

		if idx1 != -1 && idx2 != -1 {
			// Merge cluster at idx2 into cluster at idx1
			clusters[idx1] = append(clusters[idx1], clusters[idx2]...)
			// Remove cluster at idx2
			clusters = append(clusters[:idx2], clusters[idx2+1:]...)
		} else {
			break
		}
	}

	return clusters
}

// AdvancedSimilarityConfig configures advanced similarity calculations
type AdvancedSimilarityConfig struct {
	SemanticWeight   float64 `json:"semanticWeight"`
	StructuralWeight float64 `json:"structuralWeight"`
	LexicalWeight    float64 `json:"lexicalWeight"`
	ThresholdValue   float32 `json:"thresholdValue"`
	UseNormalization bool    `json:"useNormalization"`
}

// DefaultSimilarityConfig returns optimized default configuration
func DefaultSimilarityConfig() AdvancedSimilarityConfig {
	return AdvancedSimilarityConfig{
		SemanticWeight:   0.6,
		StructuralWeight: 0.3,
		LexicalWeight:    0.1,
		ThresholdValue:   0.5,
		UseNormalization: true,
	}
}

// Advanced similarity calculation methods
func advancedCosineSimilarity(vec1, vec2 []float32, normalize bool) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	// Normalize vectors if requested
	if normalize {
		vec1 = normalizeVector(vec1)
		vec2 = normalizeVector(vec2)
	}

	// Calculate dot product and magnitudes
	dotProduct := float64(0)
	magnitude1 := float64(0)
	magnitude2 := float64(0)

	for i := 0; i < len(vec1); i++ {
		dotProduct += float64(vec1[i]) * float64(vec2[i])
		magnitude1 += float64(vec1[i]) * float64(vec1[i])
		magnitude2 += float64(vec2[i]) * float64(vec2[i])
	}

	magnitude1 = math.Sqrt(magnitude1)
	magnitude2 = math.Sqrt(magnitude2)

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0
	}

	similarity := dotProduct / (magnitude1 * magnitude2)
	return math.Max(0, math.Min(1, similarity)) // Clamp to [0, 1]
}

func jaccardSimilarity(vec1, vec2 []float32, threshold float32) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	// Convert to binary features using adaptive threshold
	if threshold <= 0 {
		threshold = calculateAdaptiveThreshold(vec1, vec2)
	}

	intersection := 0
	union := 0

	for i := 0; i < len(vec1); i++ {
		v1 := vec1[i] >= threshold
		v2 := vec2[i] >= threshold

		if v1 && v2 {
			intersection++
		}
		if v1 || v2 {
			union++
		}
	}

	if union == 0 {
		return 1.0 // Both vectors are zero
	}

	return float64(intersection) / float64(union)
}

func manhattanSimilarity(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	distance := float64(0)
	for i := 0; i < len(vec1); i++ {
		distance += math.Abs(float64(vec1[i]) - float64(vec2[i]))
	}

	// Convert distance to similarity using exponential decay
	maxDistance := float64(len(vec1)) * 2.0 // Rough max distance estimate
	similarity := math.Exp(-distance / maxDistance)

	return similarity
}

func hybridSimilarity(vec1, vec2 []float32, config AdvancedSimilarityConfig) float64 {
	// Calculate multiple similarity measures
	cosine := advancedCosineSimilarity(vec1, vec2, true)
	jaccard := jaccardSimilarity(vec1, vec2, 0)
	manhattan := manhattanSimilarity(vec1, vec2)

	// Ensemble voting with confidence weighting
	similarities := []float64{cosine, jaccard, manhattan}
	weights := []float64{0.5, 0.3, 0.2} // Default weights

	// Calculate confidence for each method
	confidences := []float64{
		calculateSimilarityConfidence(cosine),
		calculateSimilarityConfidence(jaccard),
		calculateSimilarityConfidence(manhattan),
	}

	// Weighted average with confidence adjustment
	weightedSum := 0.0
	totalWeight := 0.0

	for i, sim := range similarities {
		adjustedWeight := weights[i] * confidences[i]
		weightedSum += sim * adjustedWeight
		totalWeight += adjustedWeight
	}

	if totalWeight == 0 {
		return cosine // Fallback to cosine
	}

	return weightedSum / totalWeight
}

func calculateWeightedSemanticSimilarity(vec1, vec2 []float32, config AdvancedSimilarityConfig) float64 {
	// Combine semantic similarity with structural similarity
	semanticSim := advancedCosineSimilarity(vec1, vec2, true)
	structuralSim := jaccardSimilarity(vec1, vec2, 0)
	lexicalSim := manhattanSimilarity(vec1, vec2)

	// Apply configured weights
	totalWeight := config.SemanticWeight + config.StructuralWeight + config.LexicalWeight
	if totalWeight == 0 {
		totalWeight = 1.0
	}

	weighted := (semanticSim*config.SemanticWeight +
		structuralSim*config.StructuralWeight +
		lexicalSim*config.LexicalWeight) / totalWeight

	return weighted
}

// Helper functions for advanced similarity calculations
func normalizeVector(vec []float32) []float32 {
	normalized := make([]float32, len(vec))
	magnitude := float32(0)

	for _, v := range vec {
		magnitude += v * v
	}
	magnitude = float32(math.Sqrt(float64(magnitude)))

	if magnitude == 0 {
		return normalized
	}

	for i, v := range vec {
		normalized[i] = v / magnitude
	}

	return normalized
}

func calculateAdaptiveThreshold(vec1, vec2 []float32) float32 {
	// Calculate threshold as mean of both vectors
	sum := float32(0)
	count := 0

	for i := 0; i < len(vec1); i++ {
		sum += vec1[i] + vec2[i]
		count += 2
	}

	if count == 0 {
		return 0.5
	}

	return sum / float32(count)
}

func calculateSimilarityConfidence(similarity float64) float64 {
	// Higher confidence for extreme values (close to 0 or 1)
	return 1.0 - 4.0*similarity*(1.0-similarity)
}

func cosineSimilarity(vec1, vec2 []float32) float64 { //nolint:unused // Future similarity calculations
	return advancedCosineSimilarity(vec1, vec2, false)
}

// Helper functions for confidence calculations
func calculateDirectoryGroupingConfidence(files []string, clusters [][]string, rootFolder string) float64 {
	if len(clusters) == 0 || len(files) == 0 {
		return 0.0
	}

	// Calculate how well files are grouped by directory structure
	totalScore := 0.0
	for _, cluster := range clusters {
		if len(cluster) <= 1 {
			totalScore += 1.0 // Single file clusters are always perfect
			continue
		}

		// Calculate directory coherence within cluster
		dirSimilarity := calculateDirectorySimilarity(cluster, rootFolder)
		totalScore += dirSimilarity
	}

	return totalScore / float64(len(clusters))
}

func calculatePatternConfidence(files []string, clusters [][]string) float64 {
	if len(clusters) == 0 || len(files) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, cluster := range clusters {
		if len(cluster) <= 1 {
			totalScore += 1.0
			continue
		}

		// Calculate extension similarity within cluster
		extSimilarity := calculateExtensionSimilarity(cluster)

		// Calculate test-impl relationship bonus
		testImplRelations := findTestImplementationRelations(cluster)
		relationBonus := float64(len(testImplRelations)) / float64(len(cluster))

		clusterScore := (extSimilarity + relationBonus) / 2.0
		totalScore += clusterScore
	}

	return totalScore / float64(len(clusters))
}

func calculateEmbeddingClusterConfidence(clusters [][]string, fileEmbeddings map[string][]float32) float64 {
	if len(clusters) == 0 || len(fileEmbeddings) == 0 {
		return 0.0
	}

	totalScore := 0.0
	validClusters := 0
	config := DefaultSimilarityConfig()

	for _, cluster := range clusters {
		if len(cluster) <= 1 {
			totalScore += 1.0
			validClusters++
			continue
		}

		// Calculate average pairwise similarity within cluster using enhanced methods
		clusterEmbeddings := make([][]float32, 0)
		for _, file := range cluster {
			if embedding, exists := fileEmbeddings[file]; exists {
				clusterEmbeddings = append(clusterEmbeddings, embedding)
			}
		}

		if len(clusterEmbeddings) < 2 {
			continue
		}

		avgSimilarity := 0.0
		pairs := 0
		for i := 0; i < len(clusterEmbeddings); i++ {
			for j := i + 1; j < len(clusterEmbeddings); j++ {
				// Use weighted semantic similarity for better cluster quality assessment
				similarity := calculateWeightedSemanticSimilarity(clusterEmbeddings[i], clusterEmbeddings[j], config)
				avgSimilarity += similarity
				pairs++
			}
		}

		if pairs > 0 {
			avgSimilarity /= float64(pairs)
			totalScore += avgSimilarity
			validClusters++
		}
	}

	if validClusters > 0 {
		return totalScore / float64(validClusters)
	}
	return 0.0
}

func selectRepresentativeFiles(files []string, sampleSize int) []string {
	if len(files) <= sampleSize {
		return files
	}

	// Group files by extension and directory to ensure diverse sampling
	extGroups := make(map[string][]string)
	dirGroups := make(map[string][]string)

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		if ext == "" {
			ext = "no-ext"
		}
		extGroups[ext] = append(extGroups[ext], file)

		dir := filepath.Dir(file)
		dirGroups[dir] = append(dirGroups[dir], file)
	}

	representatives := make([]string, 0, sampleSize)
	used := make(map[string]bool)

	// First, pick one representative from each extension group
	for _, group := range extGroups {
		if len(representatives) >= sampleSize {
			break
		}
		if len(group) > 0 && !used[group[0]] {
			representatives = append(representatives, group[0])
			used[group[0]] = true
		}
	}

	// Fill remaining slots from directory groups
	for _, group := range dirGroups {
		if len(representatives) >= sampleSize {
			break
		}
		for _, file := range group {
			if len(representatives) >= sampleSize {
				break
			}
			if !used[file] {
				representatives = append(representatives, file)
				used[file] = true
			}
		}
	}

	// Fill any remaining slots randomly
	for _, file := range files {
		if len(representatives) >= sampleSize {
			break
		}
		if !used[file] {
			representatives = append(representatives, file)
			used[file] = true
		}
	}

	return representatives
}

func fallbackToPatterClustering(files []string, targetClusters int) [][]string {
	clusters, _ := patternBasedClustering(files, targetClusters)
	return clusters
}

func assignFilesToClusters(files []string, reprClusters [][]string, rootFolder string) [][]string {
	if len(reprClusters) == 0 {
		return createSingleFileClusters(files)
	}

	// Create map of representative files to their cluster index
	reprToCluster := make(map[string]int)
	for clusterIdx, cluster := range reprClusters {
		for _, file := range cluster {
			reprToCluster[file] = clusterIdx
		}
	}

	// Initialize final clusters
	finalClusters := make([][]string, len(reprClusters))
	for i, cluster := range reprClusters {
		finalClusters[i] = make([]string, len(cluster))
		copy(finalClusters[i], cluster)
	}

	// Assign non-representative files to closest cluster
	for _, file := range files {
		if _, isRepr := reprToCluster[file]; isRepr {
			continue // Already assigned
		}

		// Find best cluster based on directory and extension similarity
		bestCluster := 0
		bestScore := -1.0

		for i, cluster := range finalClusters {
			if len(cluster) == 0 {
				continue
			}

			// Calculate similarity to cluster centroid (using first file as proxy)
			dirSim := calculateDirectorySimilarity([]string{file, cluster[0]}, rootFolder)
			extSim := calculateExtensionSimilarity([]string{file, cluster[0]})
			score := (dirSim + extSim) / 2.0

			if score > bestScore {
				bestScore = score
				bestCluster = i
			}
		}

		finalClusters[bestCluster] = append(finalClusters[bestCluster], file)
	}

	return finalClusters
}

func filterClustersByThreshold(clusters [][]string, rootFolder string, threshold float64) [][]string {
	filtered := make([][]string, 0)

	for _, cluster := range clusters {
		if len(cluster) <= 1 {
			filtered = append(filtered, cluster)
			continue
		}

		// Check if cluster meets similarity threshold
		similarity := calculateClusterSimilarity(cluster, rootFolder)
		if similarity >= threshold {
			filtered = append(filtered, cluster)
		} else {
			// Split cluster into individual files if it doesn't meet threshold
			for _, file := range cluster {
				filtered = append(filtered, []string{file})
			}
		}
	}

	return filtered
}

// Performance benchmarking and metrics collection for clustering optimization
type ClusteringBenchmark struct {
	TestName         string           `json:"testName"`
	FileCount        int              `json:"fileCount"`
	TargetClusters   int              `json:"targetClusters"`
	ActualClusters   int              `json:"actualClusters"`
	Method           string           `json:"method"`
	SimilarityMethod SimilarityMethod `json:"similarityMethod"`
	ExecutionTime    time.Duration    `json:"executionTime"`
	CacheHitRatio    float64          `json:"cacheHitRatio"`
	ConfidenceScore  float64          `json:"confidenceScore"`
	SilhouetteScore  float64          `json:"silhouetteScore"`
	MemoryUsage      int64            `json:"memoryUsage"`
	ApiCalls         int              `json:"apiCalls"`
	Timestamp        time.Time        `json:"timestamp"`
}

// BenchmarkResult aggregates multiple benchmark runs
type BenchmarkResult struct {
	Benchmarks    []ClusteringBenchmark `json:"benchmarks"`
	AverageTime   time.Duration         `json:"averageTime"`
	AverageConf   float64               `json:"averageConfidence"`
	TotalApiCalls int                   `json:"totalApiCalls"`
	BestMethod    string                `json:"bestMethod"`
	Summary       string                `json:"summary"`
}

// Global benchmark collector
var clusteringBenchmarksEnabled = false
var collectedBenchmarks []ClusteringBenchmark

// EnableBenchmarking enables collection of clustering performance metrics
func EnableBenchmarking() {
	clusteringBenchmarksEnabled = true
	collectedBenchmarks = []ClusteringBenchmark{}
	utils.Debug("[BENCHMARK]: Clustering performance benchmarking enabled")
}

// DisableBenchmarking disables benchmark collection
func DisableBenchmarking() {
	clusteringBenchmarksEnabled = false
	utils.Debug("[BENCHMARK]: Clustering performance benchmarking disabled")
}

// recordBenchmark records a clustering benchmark if enabled
func recordBenchmark(benchmark ClusteringBenchmark) {
	if !clusteringBenchmarksEnabled {
		return
	}

	benchmark.Timestamp = time.Now()
	collectedBenchmarks = append(collectedBenchmarks, benchmark)

	utils.Debug(fmt.Sprintf("[BENCHMARK]: Recorded %s - %d files -> %d clusters in %v",
		benchmark.TestName, benchmark.FileCount, benchmark.ActualClusters, benchmark.ExecutionTime))
}

// GetBenchmarkResults returns all collected benchmarks with analysis
func GetBenchmarkResults() BenchmarkResult {
	if len(collectedBenchmarks) == 0 {
		return BenchmarkResult{
			Summary: "No benchmarks collected",
		}
	}

	totalTime := time.Duration(0)
	totalConf := 0.0
	totalApiCalls := 0
	methodCounts := make(map[string]int)
	methodPerformance := make(map[string]float64)

	for _, benchmark := range collectedBenchmarks {
		totalTime += benchmark.ExecutionTime
		totalConf += benchmark.ConfidenceScore
		totalApiCalls += benchmark.ApiCalls

		methodCounts[benchmark.Method]++
		methodPerformance[benchmark.Method] += benchmark.ConfidenceScore
	}

	avgTime := totalTime / time.Duration(len(collectedBenchmarks))
	avgConf := totalConf / float64(len(collectedBenchmarks))

	// Find best performing method
	bestMethod := ""
	bestScore := 0.0
	for method, totalScore := range methodPerformance {
		avgScore := totalScore / float64(methodCounts[method])
		if avgScore > bestScore {
			bestScore = avgScore
			bestMethod = method
		}
	}

	summary := fmt.Sprintf("Analyzed %d benchmarks. Best method: %s (%.3f avg confidence). Avg time: %v",
		len(collectedBenchmarks), bestMethod, bestScore, avgTime)

	return BenchmarkResult{
		Benchmarks:    collectedBenchmarks,
		AverageTime:   avgTime,
		AverageConf:   avgConf,
		TotalApiCalls: totalApiCalls,
		BestMethod:    bestMethod,
		Summary:       summary,
	}
}

// ResetBenchmarks clears all collected benchmarks
func ResetBenchmarks() {
	collectedBenchmarks = []ClusteringBenchmark{}
	utils.Debug("[BENCHMARK]: Cleared all benchmark data")
}

// saveBenchmarkResults saves benchmark results to disk
func saveBenchmarkResults(results BenchmarkResult, filepath string) error { //nolint:unused // Future benchmarking features
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// RunClusteringPerformanceTest runs comprehensive performance tests
func RunClusteringPerformanceTest(files []string, rootFolder string) BenchmarkResult {
	if len(files) == 0 {
		return BenchmarkResult{Summary: "No files provided for testing"}
	}

	EnableBenchmarking()
	defer DisableBenchmarking()

	utils.Debug(fmt.Sprintf("[BENCHMARK]: Starting performance test with %d files", len(files)))

	// Test different clustering methods
	testConfigurations := []struct {
		name           string
		targetClusters int
		method         string
	}{
		{"threshold-clustering", 0, "threshold"},
		{"small-clusters", 3, "target"},
		{"medium-clusters", 5, "target"},
		{"large-clusters", 8, "target"},
	}

	for _, config := range testConfigurations {
		startTime := time.Now()

		clusters, err := SmartClusterFiles(files, rootFolder, config.targetClusters)
		if err != nil {
			utils.Warning(fmt.Sprintf("[BENCHMARK]: Test %s failed: %v", config.name, err))
			continue
		}

		executionTime := time.Since(startTime)

		benchmark := ClusteringBenchmark{
			TestName:        config.name,
			FileCount:       len(files),
			TargetClusters:  config.targetClusters,
			ActualClusters:  len(clusters),
			Method:          config.method,
			ExecutionTime:   executionTime,
			ConfidenceScore: calculateOverallClusteringConfidence(clusters, files, rootFolder),
		}

		recordBenchmark(benchmark)
	}

	return GetBenchmarkResults()
}

// calculateOverallClusteringConfidence calculates overall confidence for cluster quality
func calculateOverallClusteringConfidence(clusters [][]string, allFiles []string, rootFolder string) float64 {
	if len(clusters) == 0 || len(allFiles) == 0 {
		return 0.0
	}

	// Calculate multiple confidence metrics
	dirConfidence := calculateDirectoryGroupingConfidence(allFiles, clusters, rootFolder)
	patternConfidence := calculatePatternConfidence(allFiles, clusters)

	// Size distribution confidence (prefer balanced cluster sizes)
	sizeVariance := calculateClusterSizeVariance(clusters)
	sizeConfidence := math.Max(0, 1.0-sizeVariance)

	// Coverage confidence (all files should be clustered)
	totalClustered := 0
	for _, cluster := range clusters {
		totalClustered += len(cluster)
	}
	coverageConfidence := float64(totalClustered) / float64(len(allFiles))

	// Weighted average of all confidence measures
	weights := []float64{0.3, 0.3, 0.2, 0.2}
	scores := []float64{dirConfidence, patternConfidence, sizeConfidence, coverageConfidence}

	weightedSum := 0.0
	for i, score := range scores {
		weightedSum += score * weights[i]
	}

	return weightedSum
}

// calculateClusterSizeVariance calculates variance in cluster sizes
func calculateClusterSizeVariance(clusters [][]string) float64 {
	if len(clusters) == 0 {
		return 0
	}

	// Calculate mean cluster size
	totalFiles := 0
	for _, cluster := range clusters {
		totalFiles += len(cluster)
	}
	mean := float64(totalFiles) / float64(len(clusters))

	// Calculate variance
	variance := 0.0
	for _, cluster := range clusters {
		diff := float64(len(cluster)) - mean
		variance += diff * diff
	}
	variance /= float64(len(clusters))

	// Normalize variance by mean to get coefficient of variation
	if mean > 0 {
		return math.Sqrt(variance) / mean
	}
	return 0
}
