package config

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// ClusteringConfig holds all clustering-related configuration
type ClusteringConfig struct {
	DefaultMethod                 string             `json:"defaultMethod"`
	EnableFallbackMethods         bool               `json:"enableFallbackMethods"`
	MaxFilesForSemanticClustering int                `json:"maxFilesForSemanticClustering"`
	ConfidenceThresholds          map[string]float64 `json:"confidenceThresholds"`
	SimilarityThresholds          map[string]float64 `json:"similarityThresholds"`
	Methods                       ClusteringMethods  `json:"methods"`
	Performance                   PerformanceConfig  `json:"performance"`
}

// ClusteringMethods holds method-specific configurations
type ClusteringMethods struct {
	Directory DirectoryConfig `json:"directory"`
	Pattern   PatternConfig   `json:"pattern"`
	Cached    CachedConfig    `json:"cached"`
	Semantic  SemanticConfig  `json:"semantic"`
}

// DirectoryConfig holds directory-based clustering settings
type DirectoryConfig struct {
	Enabled bool    `json:"enabled"`
	Weight  float64 `json:"weight"`
}

// PatternConfig holds pattern-based clustering settings
type PatternConfig struct {
	Enabled bool    `json:"enabled"`
	Weight  float64 `json:"weight"`
}

// CachedConfig holds cached embedding clustering settings
type CachedConfig struct {
	Enabled          bool    `json:"enabled"`
	Weight           float64 `json:"weight"`
	MinCacheHitRatio float64 `json:"minCacheHitRatio"`
	MaxCacheAge      int     `json:"maxCacheAge"` // hours
}

// SemanticConfig holds semantic clustering settings
type SemanticConfig struct {
	Enabled                 bool    `json:"enabled"`
	Weight                  float64 `json:"weight"`
	RateLimitDelay          int     `json:"rateLimitDelay"` // milliseconds
	MaxConcurrentEmbeddings int     `json:"maxConcurrentEmbeddings"`
	EmbeddingTimeout        int     `json:"embeddingTimeout"` // seconds
}

// PerformanceConfig holds performance-related settings
type PerformanceConfig struct {
	PreferSpeed          bool `json:"preferSpeed"`
	MaxProcessingTime    int  `json:"maxProcessingTime"` // seconds
	EnableBenchmarking   bool `json:"enableBenchmarking"`
	AdaptiveOptimization bool `json:"adaptiveOptimization"`
}

// ClusteringMethod represents available clustering methods
type ClusteringMethod string

const (
	DirectoryMethod ClusteringMethod = "directory"
	PatternMethod   ClusteringMethod = "pattern"
	CachedMethod    ClusteringMethod = "cached"
	SemanticMethod  ClusteringMethod = "semantic"
	AutoMethod      ClusteringMethod = "auto" // Uses the smart multi-layered approach
)

// GetClusteringConfig retrieves the clustering configuration
func GetClusteringConfig() *ClusteringConfig {
	mu.RLock()
	defer mu.RUnlock()

	clusteringSettings, exists := settings["clustering"]
	if !exists {
		log.Println("[Config]: Clustering configuration not found, using defaults")
		return getDefaultClusteringConfig()
	}

	clusteringMap, ok := clusteringSettings.(map[string]interface{})
	if !ok {
		log.Println("[Config]: Invalid clustering configuration f ormat, using defaults")
		return getDefaultClusteringConfig()
	}

	config := &ClusteringConfig{}

	// Parse basic settings
	if defaultMethod, ok := clusteringMap["defaultMethod"].(string); ok {
		config.DefaultMethod = defaultMethod
	} else {
		config.DefaultMethod = "directory"
	}

	if enableFallback, ok := clusteringMap["enableFallbackMethods"].(bool); ok {
		config.EnableFallbackMethods = enableFallback
	} else {
		config.EnableFallbackMethods = true
	}

	if maxFiles, ok := clusteringMap["maxFilesForSemanticClustering"].(float64); ok {
		config.MaxFilesForSemanticClustering = int(maxFiles)
	} else if maxFiles, ok := clusteringMap["maxFilesForSemanticClustering"].(int); ok {
		config.MaxFilesForSemanticClustering = maxFiles
	} else {
		config.MaxFilesForSemanticClustering = 10
	}

	// Parse confidence thresholds
	config.ConfidenceThresholds = parseFloatMap(clusteringMap, "confidenceThresholds", map[string]float64{
		"directory": 0.8,
		"pattern":   0.7,
		"cached":    0.6,
		"semantic":  0.5,
	})

	// Parse similarity thresholds
	config.SimilarityThresholds = parseFloatMap(clusteringMap, "similarityThresholds", map[string]float64{
		"directory": 0.7,
		"pattern":   0.6,
		"cached":    0.5,
		"semantic":  0.4,
	})

	// Parse methods configuration
	if methodsMap, ok := clusteringMap["methods"].(map[string]interface{}); ok {
		config.Methods = parseMethodsConfig(methodsMap)
	} else {
		config.Methods = getDefaultMethodsConfig()
	}

	// Parse performance configuration
	if perfMap, ok := clusteringMap["performance"].(map[string]interface{}); ok {
		config.Performance = parsePerformanceConfig(perfMap)
	} else {
		config.Performance = getDefaultPerformanceConfig()
	}

	return config
}

// SetClusteringConfig updates the clustering configuration
func SetClusteringConfig(config *ClusteringConfig) error {
	// Use global config mutex to prevent deadlocks
	// The global config system will handle the locking
	clusteringMap := map[string]interface{}{
		"defaultMethod":                 config.DefaultMethod,
		"enableFallbackMethods":         config.EnableFallbackMethods,
		"maxFilesForSemanticClustering": config.MaxFilesForSemanticClustering,
		"confidenceThresholds":          config.ConfidenceThresholds,
		"similarityThresholds":          config.SimilarityThresholds,
		"methods": map[string]interface{}{
			"directory": map[string]interface{}{
				"enabled": config.Methods.Directory.Enabled,
				"weight":  config.Methods.Directory.Weight,
			},
			"pattern": map[string]interface{}{
				"enabled": config.Methods.Pattern.Enabled,
				"weight":  config.Methods.Pattern.Weight,
			},
			"cached": map[string]interface{}{
				"enabled":          config.Methods.Cached.Enabled,
				"weight":           config.Methods.Cached.Weight,
				"minCacheHitRatio": config.Methods.Cached.MinCacheHitRatio,
				"maxCacheAge":      config.Methods.Cached.MaxCacheAge,
			},
			"semantic": map[string]interface{}{
				"enabled":                 config.Methods.Semantic.Enabled,
				"weight":                  config.Methods.Semantic.Weight,
				"rateLimitDelay":          config.Methods.Semantic.RateLimitDelay,
				"maxConcurrentEmbeddings": config.Methods.Semantic.MaxConcurrentEmbeddings,
				"embeddingTimeout":        config.Methods.Semantic.EmbeddingTimeout,
			},
		},
		"performance": map[string]interface{}{
			"preferSpeed":          config.Performance.PreferSpeed,
			"maxProcessingTime":    config.Performance.MaxProcessingTime,
			"enableBenchmarking":   config.Performance.EnableBenchmarking,
			"adaptiveOptimization": config.Performance.AdaptiveOptimization,
		},
	}

	// Use the global config system to set and save
	Set("clustering", clusteringMap)

	log.Println("[Config]: Clustering configuration updated successfully")
	return nil
}

// SetClusteringConfigByKey updates a specific clustering configuration value
func SetClusteringConfigByKey(key, value string) error {
	// Get current config (this will acquire and release a read lock)
	config := GetClusteringConfig()

	// Make a copy to avoid modifying the original while it might still be read-locked
	configCopy := *config

	switch key {
	// Global settings
	case "similarity_threshold":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.SimilarityThresholds["directory"] = floatVal
			configCopy.SimilarityThresholds["pattern"] = floatVal
			// configCopy.SimilarityThresholds["cached"] = floatVal
			// configCopy.SimilarityThresholds["semantic"] = floatVal
		} else {
			return fmt.Errorf("invalid float value for similarity_threshold: %s", value)
		}
	case "max_processing_time":
		if intVal, err := parseInt(value); err == nil {
			configCopy.Performance.MaxProcessingTime = intVal
		} else {
			return fmt.Errorf("invalid integer value for max_processing_time: %s", value)
		}
	case "adaptive_optimization":
		if boolVal, err := parseBool(value); err == nil {
			configCopy.Performance.AdaptiveOptimization = boolVal
		} else {
			return fmt.Errorf("invalid boolean value for adaptive_optimization: %s", value)
		}
	case "performance_mode":
		switch value {
		case "speed":
			configCopy.Performance.PreferSpeed = true
			configCopy.Performance.MaxProcessingTime = 30
		case "quality":
			configCopy.Performance.PreferSpeed = false
			configCopy.Performance.MaxProcessingTime = 120
		case "balanced":
			configCopy.Performance.PreferSpeed = true
			configCopy.Performance.MaxProcessingTime = 60
		default:
			return fmt.Errorf("invalid performance mode: %s (valid: speed, quality, balanced)", value)
		}

	// Directory method settings
	case "directory_enabled":
		if boolVal, err := parseBool(value); err == nil {
			configCopy.Methods.Directory.Enabled = boolVal
		} else {
			return fmt.Errorf("invalid boolean value for directory_enabled: %s", value)
		}
	case "directory_weight":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.Methods.Directory.Weight = floatVal
		} else {
			return fmt.Errorf("invalid float value for directory_weight: %s", value)
		}
	case "directory_confidence_threshold":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.ConfidenceThresholds["directory"] = floatVal
		} else {
			return fmt.Errorf("invalid float value for directory_confidence_threshold: %s", value)
		}
	case "directory_similarity_threshold":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.SimilarityThresholds["directory"] = floatVal
		} else {
			return fmt.Errorf("invalid float value for directory_similarity_threshold: %s", value)
		}

	// Pattern method settings
	case "pattern_enabled":
		if boolVal, err := parseBool(value); err == nil {
			configCopy.Methods.Pattern.Enabled = boolVal
		} else {
			return fmt.Errorf("invalid boolean value for pattern_enabled: %s", value)
		}
	case "pattern_weight":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.Methods.Pattern.Weight = floatVal
		} else {
			return fmt.Errorf("invalid float value for pattern_weight: %s", value)
		}
	case "pattern_confidence_threshold":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.ConfidenceThresholds["pattern"] = floatVal
		} else {
			return fmt.Errorf("invalid float value for pattern_confidence_threshold: %s", value)
		}
	case "pattern_similarity_threshold":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.SimilarityThresholds["pattern"] = floatVal
		} else {
			return fmt.Errorf("invalid float value for pattern_similarity_threshold: %s", value)
		}

	// Cached method settings
	case "cached_enabled":
		if boolVal, err := parseBool(value); err == nil {
			configCopy.Methods.Cached.Enabled = boolVal
		} else {
			return fmt.Errorf("invalid boolean value for cached_enabled: %s", value)
		}
	case "cached_weight":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.Methods.Cached.Weight = floatVal
		} else {
			return fmt.Errorf("invalid float value for cached_weight: %s", value)
		}
	case "cached_confidence_threshold":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.ConfidenceThresholds["cached"] = floatVal
		} else {
			return fmt.Errorf("invalid float value for cached_confidence_threshold: %s", value)
		}
	case "cached_similarity_threshold":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.SimilarityThresholds["cached"] = floatVal
		} else {
			return fmt.Errorf("invalid float value for cached_similarity_threshold: %s", value)
		}
	case "cached_delay_ms":
		if _, err := parseInt(value); err == nil {
			// Note: This could be added to CachedConfig if needed
			log.Println("cached_delay_ms setting not yet implemented in configuration structure")
		} else {
			return fmt.Errorf("invalid integer value for cached_delay_ms: %s", value)
		}

	// Semantic method settings
	case "semantic_enabled":
		if boolVal, err := parseBool(value); err == nil {
			configCopy.Methods.Semantic.Enabled = boolVal
		} else {
			return fmt.Errorf("invalid boolean value for semantic_enabled: %s", value)
		}
	case "semantic_weight":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.Methods.Semantic.Weight = floatVal
		} else {
			return fmt.Errorf("invalid float value for semantic_weight: %s", value)
		}
	case "semantic_confidence_threshold":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.ConfidenceThresholds["semantic"] = floatVal
		} else {
			return fmt.Errorf("invalid float value for semantic_confidence_threshold: %s", value)
		}
	case "semantic_similarity_threshold":
		if floatVal, err := parseFloat(value); err == nil {
			configCopy.SimilarityThresholds["semantic"] = floatVal
		} else {
			return fmt.Errorf("invalid float value for semantic_similarity_threshold: %s", value)
		}
	case "semantic_rate_limit_delay":
		if intVal, err := parseInt(value); err == nil {
			configCopy.Methods.Semantic.RateLimitDelay = intVal
		} else {
			return fmt.Errorf("invalid integer value for semantic_rate_limit_delay: %s", value)
		}

	default:
		return fmt.Errorf("unknown clustering configuration key: %s", key)
	}

	// Now save the modified config (this will acquire and release a write lock)
	return SetClusteringConfig(&configCopy)
}

// ApplyClusteringPreset applies a predefined clustering configuration preset
func ApplyClusteringPreset(presetName string) error {
	var config *ClusteringConfig

	switch presetName {
	case "speed":
		config = GetSpeedOptimizedConfig()
	case "balanced":
		config = GetBalancedConfig()
	case "quality":
		config = GetQualityOptimizedConfig()
	default:
		return fmt.Errorf("unknown preset: %s (valid presets: speed, balanced, quality)", presetName)
	}

	return SetClusteringConfig(config)
}

// Helper functions for parsing configuration

func parseFloatMap(source map[string]interface{}, key string, defaultMap map[string]float64) map[string]float64 {
	if thresholds, ok := source[key].(map[string]interface{}); ok {
		result := make(map[string]float64)
		for k, v := range thresholds {
			if floatVal, ok := v.(float64); ok {
				result[k] = floatVal
			} else if intVal, ok := v.(int); ok {
				result[k] = float64(intVal)
			} else {
				// Use default if conversion fails
				if defaultVal, exists := defaultMap[k]; exists {
					result[k] = defaultVal
				}
			}
		}
		return result
	}
	return defaultMap
}

func parseMethodsConfig(methodsMap map[string]interface{}) ClusteringMethods {
	methods := ClusteringMethods{}

	// Parse directory config
	if dirMap, ok := methodsMap["directory"].(map[string]interface{}); ok {
		methods.Directory = DirectoryConfig{
			Enabled: getBoolOrDefault(dirMap, "enabled", true),
			Weight:  getFloatOrDefault(dirMap, "weight", 1.0),
		}
	} else {
		methods.Directory = DirectoryConfig{Enabled: true, Weight: 1.0}
	}

	// Parse pattern config
	if patternMap, ok := methodsMap["pattern"].(map[string]interface{}); ok {
		methods.Pattern = PatternConfig{
			Enabled: getBoolOrDefault(patternMap, "enabled", true),
			Weight:  getFloatOrDefault(patternMap, "weight", 0.8),
		}
	} else {
		methods.Pattern = PatternConfig{Enabled: true, Weight: 0.8}
	}

	// Parse cached config
	if cachedMap, ok := methodsMap["cached"].(map[string]interface{}); ok {
		methods.Cached = CachedConfig{
			Enabled:          getBoolOrDefault(cachedMap, "enabled", true),
			Weight:           getFloatOrDefault(cachedMap, "weight", 0.6),
			MinCacheHitRatio: getFloatOrDefault(cachedMap, "minCacheHitRatio", 0.4),
			MaxCacheAge:      getIntOrDefault(cachedMap, "maxCacheAge", 24),
		}
	} else {
		methods.Cached = CachedConfig{
			Enabled: true, Weight: 0.6, MinCacheHitRatio: 0.4, MaxCacheAge: 24,
		}
	}

	// Parse semantic config
	if semanticMap, ok := methodsMap["semantic"].(map[string]interface{}); ok {
		methods.Semantic = SemanticConfig{
			Enabled:                 getBoolOrDefault(semanticMap, "enabled", true),
			Weight:                  getFloatOrDefault(semanticMap, "weight", 0.4),
			RateLimitDelay:          getIntOrDefault(semanticMap, "rateLimitDelay", 2000),
			MaxConcurrentEmbeddings: getIntOrDefault(semanticMap, "maxConcurrentEmbeddings", 1),
			EmbeddingTimeout:        getIntOrDefault(semanticMap, "embeddingTimeout", 30),
		}
	} else {
		methods.Semantic = SemanticConfig{
			Enabled: true, Weight: 0.4, RateLimitDelay: 2000,
			MaxConcurrentEmbeddings: 1, EmbeddingTimeout: 30,
		}
	}

	return methods
}

func parsePerformanceConfig(perfMap map[string]interface{}) PerformanceConfig {
	return PerformanceConfig{
		PreferSpeed:          getBoolOrDefault(perfMap, "preferSpeed", true),
		MaxProcessingTime:    getIntOrDefault(perfMap, "maxProcessingTime", 60),
		EnableBenchmarking:   getBoolOrDefault(perfMap, "enableBenchmarking", false),
		AdaptiveOptimization: getBoolOrDefault(perfMap, "adaptiveOptimization", true),
	}
}

// Helper functions for type conversion with defaults

func getBoolOrDefault(m map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return defaultVal
}

func getFloatOrDefault(m map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	} else if val, ok := m[key].(int); ok {
		return float64(val)
	}
	return defaultVal
}

func getIntOrDefault(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	} else if val, ok := m[key].(int); ok {
		return val
	}
	return defaultVal
}

// Default configuration builders

func getDefaultClusteringConfig() *ClusteringConfig {
	return &ClusteringConfig{
		DefaultMethod:                 "directory",
		EnableFallbackMethods:         true,
		MaxFilesForSemanticClustering: 10,
		ConfidenceThresholds: map[string]float64{
			"directory": 0.8,
			"pattern":   0.7,
			"cached":    0.6,
			"semantic":  0.5,
		},
		SimilarityThresholds: map[string]float64{
			"directory": 0.7,
			"pattern":   0.6,
			"cached":    0.5,
			"semantic":  0.4,
		},
		Methods:     getDefaultMethodsConfig(),
		Performance: getDefaultPerformanceConfig(),
	}
}

func getDefaultMethodsConfig() ClusteringMethods {
	return ClusteringMethods{
		Directory: DirectoryConfig{Enabled: true, Weight: 1.0},
		Pattern:   PatternConfig{Enabled: true, Weight: 0.8},
		Cached: CachedConfig{
			Enabled: true, Weight: 0.6, MinCacheHitRatio: 0.4, MaxCacheAge: 24,
		},
		Semantic: SemanticConfig{
			Enabled: true, Weight: 0.4, RateLimitDelay: 2000,
			MaxConcurrentEmbeddings: 1, EmbeddingTimeout: 30,
		},
	}
}

func getDefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		PreferSpeed:          true,
		MaxProcessingTime:    60,
		EnableBenchmarking:   false,
		AdaptiveOptimization: true,
	}
}

// Configuration presets for different use cases

// GetSpeedOptimizedConfig returns configuration optimized for speed
func GetSpeedOptimizedConfig() *ClusteringConfig {
	config := getDefaultClusteringConfig()
	config.DefaultMethod = "directory"
	config.EnableFallbackMethods = false    // Only use the fastest method
	config.Methods.Semantic.Enabled = false // Disable slow semantic clustering
	config.Methods.Cached.Enabled = false   // Disable caching overhead
	config.Performance.PreferSpeed = true
	config.Performance.MaxProcessingTime = 30 // Shorter timeout
	return config
}

// GetQualityOptimizedConfig returns configuration optimized for clustering quality
func GetQualityOptimizedConfig() *ClusteringConfig {
	config := getDefaultClusteringConfig()
	config.DefaultMethod = "semantic"
	config.EnableFallbackMethods = true
	config.MaxFilesForSemanticClustering = 20 // Allow more files for semantic analysis

	// Lower thresholds for more permissive clustering
	config.ConfidenceThresholds["directory"] = 0.9
	config.ConfidenceThresholds["pattern"] = 0.8
	config.ConfidenceThresholds["cached"] = 0.7
	config.ConfidenceThresholds["semantic"] = 0.6

	config.Methods.Semantic.Enabled = true
	config.Methods.Semantic.Weight = 1.0
	config.Methods.Semantic.RateLimitDelay = 1000 // Faster API calls if possible
	config.Performance.PreferSpeed = false
	config.Performance.MaxProcessingTime = 120 // Longer timeout for quality
	config.Performance.EnableBenchmarking = true
	return config
}

// GetBalancedConfig returns a balanced configuration for speed and quality
func GetBalancedConfig() *ClusteringConfig {
	config := getDefaultClusteringConfig()
	config.DefaultMethod = "auto" // Use smart multi-layered approach
	config.EnableFallbackMethods = true
	config.MaxFilesForSemanticClustering = 10
	config.Performance.PreferSpeed = true
	config.Performance.AdaptiveOptimization = true
	return config
}

// Helper functions for parsing string values
func parseFloat(s string) (float64, error) {
	// Try to parse as float
	return strconv.ParseFloat(s, 64)
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "yes", "1", "on", "enabled":
		return true, nil
	case "false", "no", "0", "off", "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean: %s", s)
	}
}

// IsMethodEnabled checks if a specific clustering method is enabled
func IsMethodEnabled(method ClusteringMethod) bool {
	config := GetClusteringConfig()
	switch method {
	case DirectoryMethod:
		return config.Methods.Directory.Enabled
	case PatternMethod:
		return config.Methods.Pattern.Enabled
	case CachedMethod:
		return config.Methods.Cached.Enabled
	case SemanticMethod:
		return config.Methods.Semantic.Enabled
	default:
		return false
	}
}

// GetConfidenceThreshold returns the confidence threshold for a specific method
func GetConfidenceThreshold(method ClusteringMethod) float64 {
	config := GetClusteringConfig()
	if threshold, exists := config.ConfidenceThresholds[string(method)]; exists {
		return threshold
	}
	return 0.5 // Default threshold
}

// GetSimilarityThreshold returns the similarity threshold for a specific method
func GetSimilarityThreshold(method ClusteringMethod) float64 {
	config := GetClusteringConfig()
	if threshold, exists := config.SimilarityThresholds[string(method)]; exists {
		return threshold
	}
	return 0.7 // Default threshold
}
