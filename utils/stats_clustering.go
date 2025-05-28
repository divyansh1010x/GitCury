package utils

import (
	"fmt"
)

// CaptureClusteringConfigFromSettings captures clustering configuration info from provided parameters
func CaptureClusteringConfigFromSettings(
	defaultMethod string,
	enabledMethods []string,
	confidenceThresholds map[string]float64,
	similarityThresholds map[string]float64,
	maxFilesForSemantic int,
	enableFallbackMethods bool,
	performanceMode string,
	maxProcessingTime int,
	enableBenchmarking bool,
	adaptiveOptimization bool,
) {
	if !IsStatsEnabled() {
		return
	}

	statsMutex.Lock()
	defer statsMutex.Unlock()

	clusteringInfo = &ClusteringMethodInfo{
		Method:                defaultMethod,
		EnabledMethods:        enabledMethods,
		ConfidenceThresholds:  confidenceThresholds,
		SimilarityThresholds:  similarityThresholds,
		MaxFilesForSemantic:   maxFilesForSemantic,
		EnableFallbackMethods: enableFallbackMethods,
		PerformanceMode:       performanceMode,
		MaxProcessingTime:     maxProcessingTime,
		EnableBenchmarking:    enableBenchmarking,
		AdaptiveOptimization:  adaptiveOptimization,
	}

	Debug("[STATS.CLUSTERING]: Captured clustering configuration - Method: " + defaultMethod +
		", Enabled Methods: " + fmt.Sprintf("%v", enabledMethods) +
		", Performance Mode: " + performanceMode)
}
