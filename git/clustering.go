package git

import (
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

// SmartClusterFiles groups files using multi-layered approach
// When targetClusters is 0 or negative, it uses threshold-based clustering without limits
func SmartClusterFiles(changedFiles []string, rootFolder string, targetClusters int) ([][]string, error) {
	if len(changedFiles) == 0 {
		return [][]string{}, nil
	}

	if len(changedFiles) == 1 {
		return [][]string{changedFiles}, nil
	}

	// Determine if we should use threshold-based clustering (no cluster limit)
	useThresholdClustering := targetClusters <= 0

	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Starting multi-layered clustering for %d files, threshold-based: %v",
		len(changedFiles), useThresholdClustering))

	// Layer 1: Directory-based clustering
	dirClusters, dirConfidence := directoryBasedClustering(changedFiles, rootFolder, targetClusters)
	if dirConfidence >= 0.8 && (!useThresholdClustering || validateClustersByThreshold(dirClusters, rootFolder, 0.7)) {
		utils.Debug("[GIT.CLUSTER]: Directory-based clustering successful with high confidence")
		return dirClusters, nil
	}

	// Layer 2: Pattern-based clustering
	patternClusters, patternConfidence := patternBasedClustering(changedFiles, targetClusters)
	if patternConfidence >= 0.7 && (!useThresholdClustering || validateClustersByThreshold(patternClusters, rootFolder, 0.6)) {
		utils.Debug("[GIT.CLUSTER]: Pattern-based clustering successful with good confidence")
		return patternClusters, nil
	}

	// Layer 3: Cached embedding clustering
	cachedClusters, cachedConfidence, cacheHitRatio := cachedEmbeddingClustering(changedFiles, rootFolder, targetClusters)
	if cachedConfidence >= 0.6 && cacheHitRatio >= 0.4 && (!useThresholdClustering || validateClustersByThreshold(cachedClusters, rootFolder, 0.5)) {
		utils.Debug("[GIT.CLUSTER]: Cached embedding clustering successful")
		return cachedClusters, nil
	}

	// Layer 4: Smart sampling for large file sets
	if len(changedFiles) > 10 {
		return smartSamplingClustering(changedFiles, rootFolder, targetClusters, useThresholdClustering)
	}

	// Layer 5: Full semantic clustering (fallback)
	return fullSemanticClustering(changedFiles, rootFolder, targetClusters, useThresholdClustering)
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

// cachedEmbeddingClustering uses cached embeddings when available
func cachedEmbeddingClustering(files []string, rootFolder string, targetClusters int) ([][]string, float64, float64) {
	cache := loadEmbeddingCache(rootFolder)
	fileEmbeddings := make(map[string][]float32)
	cacheHits := 0

	// Check cache for existing embeddings
	for _, file := range files {
		if cached, exists := getCachedEmbedding(file, cache); exists {
			fileEmbeddings[file] = cached.Embedding
			cacheHits++
		}
	}

	cacheHitRatio := float64(cacheHits) / float64(len(files))

	// If we don't have enough cached embeddings, return early
	if cacheHitRatio < 0.3 {
		return createSingleFileClusters(files), 0.0, cacheHitRatio
	}

	// Generate embeddings for missing files (limit to prevent API overload)
	maxNewEmbeddings := 5
	newEmbeddings := 0

	for _, file := range files {
		if _, exists := fileEmbeddings[file]; !exists && newEmbeddings < maxNewEmbeddings {
			diff, err := embeddings.GetFileDiff(file)
			if err != nil {
				utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Could not get diff for file: %s - %v", file, err))
				continue
			}

			// Skip very large diffs to prevent API overload
			if len(diff) > 15000 {
				diff = diff[:15000] + "... [truncated]"
			}

			embedding, err := embeddings.GenerateEmbedding(diff)
			if err != nil {
				utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: Could not generate embedding for file: %s - %v", file, err))
				continue
			}

			fileEmbeddings[file] = embedding
			saveEmbeddingToCache(file, embedding, diff, cache)
			newEmbeddings++
		}
	}

	// Perform clustering using available embeddings
	if len(fileEmbeddings) < 2 {
		return createSingleFileClusters(files), 0.0, cacheHitRatio
	}

	clusters := performEmbeddingBasedClustering(fileEmbeddings, targetClusters)
	confidence := calculateEmbeddingClusterConfidence(clusters, fileEmbeddings)

	// Save updated cache
	saveEmbeddingCache(cache)

	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Cached embedding clustering: %d files -> %d clusters, cache hit ratio: %.2f, confidence: %.2f",
		len(files), len(clusters), cacheHitRatio, confidence))

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

	fileEmbeddings := make(map[string][]float32)

	// Generate embeddings for all files with rate limiting
	for i, file := range files {
		// Rate limiting: add delay between requests
		if i > 0 {
			time.Sleep(2 * time.Second)
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

	// Perform clustering
	var clusters [][]string
	if useThresholdClustering {
		clusters = performThresholdBasedClustering(fileEmbeddings, 0.6) // 0.6 similarity threshold
	} else {
		clusters = performEmbeddingBasedClustering(fileEmbeddings, targetClusters)
	}

	utils.Debug(fmt.Sprintf("[GIT.CLUSTER]: Full semantic clustering: %d files -> %d clusters",
		len(files), len(clusters)))

	return clusters, nil
}

// performThresholdBasedClustering clusters files based on similarity threshold instead of target count
func performThresholdBasedClustering(fileEmbeddings map[string][]float32, threshold float64) [][]string {
	files := make([]string, 0, len(fileEmbeddings))
	embeddings := make([][]float32, 0, len(fileEmbeddings))

	for file, embedding := range fileEmbeddings {
		files = append(files, file)
		embeddings = append(embeddings, embedding)
	}

	// Calculate similarity matrix
	similarities := make([][]float64, len(files))
	for i := range similarities {
		similarities[i] = make([]float64, len(files))
		for j := range similarities[i] {
			if i == j {
				similarities[i][j] = 1.0
			} else {
				similarities[i][j] = cosineSimilarity(embeddings[i], embeddings[j])
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

	// Perform K-means clustering
	labels, err := embeddings.KMeans(vectors, targetClusters, 20)
	if err != nil {
		utils.Warning(fmt.Sprintf("[GIT.CLUSTER]: K-means clustering failed: %v", err))
		return createSingleFileClusters(files)
	}

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

func getCachedEmbedding(filePath string, cache *EmbeddingCache) (*FileCacheEntry, bool) {
	entry, exists := cache.Embeddings[filePath]
	if !exists {
		return nil, false
	}

	// Check if file has been modified since cache entry
	if currentHash := getFileContentHash(filePath); currentHash != entry.ContentHash {
		delete(cache.Embeddings, filePath)
		return nil, false
	}

	return &entry, true
}

func saveEmbeddingToCache(filePath string, embedding []float32, content string, cache *EmbeddingCache) {
	entry := FileCacheEntry{
		FilePath:    filePath,
		Embedding:   embedding,
		ContentHash: getFileContentHash(filePath),
		LastUpdated: time.Now(),
	}

	cache.Embeddings[filePath] = entry
	cache.LastUpdated = time.Now()
}

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
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	hasher := sha256.Sum256(data)
	return hex.EncodeToString(hasher[:])
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

	// Sort by size (smallest first for merging)
	sort.Slice(clusters, func(i, j int) bool {
		return len(clusters[i]) < len(clusters[j])
	})

	// Merge smallest clusters until we reach target count
	for len(clusters) > targetCount {
		// Merge the two smallest clusters
		smallest := clusters[0]
		secondSmallest := clusters[1]
		merged := append(smallest, secondSmallest...)

		// Remove the two smallest and add the merged one
		clusters = clusters[2:]
		clusters = append(clusters, merged)

		// Re-sort
		sort.Slice(clusters, func(i, j int) bool {
			return len(clusters[i]) < len(clusters[j])
		})
	}

	return clusters
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Additional helper functions for confidence calculation
func calculateDirectoryGroupingConfidence(files []string, clusters [][]string, rootFolder string) float64 {
	// Calculate how well files are grouped by directory
	totalFiles := len(files)
	wellGroupedFiles := 0

	for _, cluster := range clusters {
		if len(cluster) <= 1 {
			wellGroupedFiles++
			continue
		}

		// Check if files in cluster are from same or related directories
		dirs := make(map[string]int)
		for _, file := range cluster {
			relPath, _ := filepath.Rel(rootFolder, file)
			dir := filepath.Dir(relPath)
			dirs[dir]++
		}

		// If most files are from the same directory, consider them well grouped
		maxDirCount := 0
		for _, count := range dirs {
			if count > maxDirCount {
				maxDirCount = count
			}
		}

		if float64(maxDirCount)/float64(len(cluster)) >= 0.7 {
			wellGroupedFiles += len(cluster)
		}
	}

	return float64(wellGroupedFiles) / float64(totalFiles)
}

func calculatePatternConfidence(files []string, clusters [][]string) float64 {
	totalFiles := len(files)
	wellGroupedFiles := 0

	for _, cluster := range clusters {
		if len(cluster) <= 1 {
			wellGroupedFiles++
			continue
		}

		// Check extension consistency
		exts := make(map[string]int)
		for _, file := range cluster {
			ext := strings.ToLower(filepath.Ext(file))
			exts[ext]++
		}

		maxExtCount := 0
		for _, count := range exts {
			if count > maxExtCount {
				maxExtCount = count
			}
		}

		// If most files have the same extension, consider them well grouped
		if float64(maxExtCount)/float64(len(cluster)) >= 0.6 {
			wellGroupedFiles += len(cluster)
		}
	}

	return float64(wellGroupedFiles) / float64(totalFiles)
}

func calculateEmbeddingClusterConfidence(clusters [][]string, embeddings map[string][]float32) float64 {
	if len(clusters) == 0 {
		return 0
	}

	totalSimilarity := 0.0
	totalPairs := 0

	for _, cluster := range clusters {
		if len(cluster) <= 1 {
			continue
		}

		// Calculate average intra-cluster similarity
		clusterSimilarity := 0.0
		pairs := 0

		for i := 0; i < len(cluster); i++ {
			for j := i + 1; j < len(cluster); j++ {
				if emb1, exists1 := embeddings[cluster[i]]; exists1 {
					if emb2, exists2 := embeddings[cluster[j]]; exists2 {
						similarity := cosineSimilarity(emb1, emb2)
						clusterSimilarity += similarity
						pairs++
					}
				}
			}
		}

		if pairs > 0 {
			totalSimilarity += clusterSimilarity / float64(pairs)
			totalPairs++
		}
	}

	if totalPairs == 0 {
		return 0.5 // Neutral confidence for single-file clusters
	}

	return totalSimilarity / float64(totalPairs)
}

// Additional utility functions for smart sampling
func selectRepresentativeFiles(files []string, sampleSize int) []string {
	if len(files) <= sampleSize {
		return files
	}

	// Group files by extension and directory to ensure diverse sampling
	extGroups := make(map[string][]string)
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		if ext == "" {
			ext = "no-extension"
		}
		extGroups[ext] = append(extGroups[ext], file)
	}

	representatives := make([]string, 0, sampleSize)

	// Select at least one file from each extension group
	for _, group := range extGroups {
		if len(representatives) >= sampleSize {
			break
		}

		// Select representative from this group (prefer files with meaningful names)
		var selected string
		for _, file := range group {
			basename := strings.ToLower(filepath.Base(file))
			if strings.Contains(basename, "main") || strings.Contains(basename, "index") ||
				strings.Contains(basename, "core") || strings.Contains(basename, "app") {
				selected = file
				break
			}
		}
		if selected == "" {
			selected = group[0] // Fallback to first file
		}

		representatives = append(representatives, selected)
	}

	// Fill remaining slots with largest files (likely to be more significant)
	remaining := sampleSize - len(representatives)
	if remaining > 0 {
		unselected := make([]string, 0)
		selectedMap := make(map[string]bool)
		for _, repr := range representatives {
			selectedMap[repr] = true
		}

		for _, file := range files {
			if !selectedMap[file] {
				unselected = append(unselected, file)
			}
		}

		// Sort by file size (approximate by name length as a simple heuristic)
		sort.Slice(unselected, func(i, j int) bool {
			return len(unselected[i]) > len(unselected[j])
		})

		for i := 0; i < remaining && i < len(unselected); i++ {
			representatives = append(representatives, unselected[i])
		}
	}

	return representatives
}

func assignFilesToClusters(allFiles []string, representativeClusters [][]string, rootFolder string) [][]string {
	if len(representativeClusters) == 0 {
		return createSingleFileClusters(allFiles)
	}

	// Create map of representatives to their cluster index
	reprToCluster := make(map[string]int)
	for clusterIdx, cluster := range representativeClusters {
		for _, repr := range cluster {
			reprToCluster[repr] = clusterIdx
		}
	}

	// Initialize final clusters with representatives
	finalClusters := make([][]string, len(representativeClusters))
	copy(finalClusters, representativeClusters)

	// Assign remaining files to clusters based on similarity
	for _, file := range allFiles {
		if _, isRepr := reprToCluster[file]; isRepr {
			continue // Skip representatives as they're already assigned
		}

		bestCluster := 0
		bestSimilarity := -1.0

		// Find the most similar cluster
		for clusterIdx, cluster := range representativeClusters {
			similarity := calculateFileSimilarityToCluster(file, cluster, rootFolder)
			if similarity > bestSimilarity {
				bestSimilarity = similarity
				bestCluster = clusterIdx
			}
		}

		finalClusters[bestCluster] = append(finalClusters[bestCluster], file)
	}

	return finalClusters
}

func calculateFileSimilarityToCluster(file string, cluster []string, rootFolder string) float64 {
	if len(cluster) == 0 {
		return 0
	}

	// Calculate similarity based on directory and extension
	fileExt := strings.ToLower(filepath.Ext(file))
	fileDir := filepath.Dir(file)

	totalSimilarity := 0.0
	for _, clusterFile := range cluster {
		clusterExt := strings.ToLower(filepath.Ext(clusterFile))
		clusterDir := filepath.Dir(clusterFile)

		similarity := 0.0

		// Extension similarity
		if fileExt == clusterExt {
			similarity += 0.5
		}

		// Directory similarity
		if fileDir == clusterDir {
			similarity += 0.5
		} else if strings.Contains(fileDir, clusterDir) || strings.Contains(clusterDir, fileDir) {
			similarity += 0.3
		}

		totalSimilarity += similarity
	}

	return totalSimilarity / float64(len(cluster))
}

func filterClustersByThreshold(clusters [][]string, rootFolder string, threshold float64) [][]string {
	filtered := make([][]string, 0)

	for _, cluster := range clusters {
		if len(cluster) <= 1 {
			filtered = append(filtered, cluster)
			continue
		}

		avgSimilarity := calculateClusterSimilarity(cluster, rootFolder)
		if avgSimilarity >= threshold {
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

func fallbackToPatterClustering(files []string, targetClusters int) [][]string {
	utils.Debug("[GIT.CLUSTER]: Falling back to pattern-based clustering")
	clusters, _ := patternBasedClustering(files, targetClusters)
	return clusters
}
