package utils

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	statsEnabled      = false
	statsMutex        sync.RWMutex
	commandStartTime  time.Time
	operationProgress = make(map[string]ProgressInfo)
	totalOperations   int
	completedOps      int
	clusteringInfo    *ClusteringMethodInfo // New field for clustering info
)

type ProgressInfo struct {
	Name      string
	StartTime time.Time
	EndTime   *time.Time
	Duration  time.Duration
	Progress  float64
	Status    string
}

type ClusteringMethodInfo struct {
	Method                string             `json:"method"`
	EnabledMethods        []string           `json:"enabledMethods"`
	ConfidenceThresholds  map[string]float64 `json:"confidenceThresholds"`
	SimilarityThresholds  map[string]float64 `json:"similarityThresholds"`
	MaxFilesForSemantic   int                `json:"maxFilesForSemantic"`
	EnableFallbackMethods bool               `json:"enableFallbackMethods"`
	PerformanceMode       string             `json:"performanceMode"`
	MaxProcessingTime     int                `json:"maxProcessingTime"`
	EnableBenchmarking    bool               `json:"enableBenchmarking"`
	AdaptiveOptimization  bool               `json:"adaptiveOptimization"`
}

type CommandStats struct {
	Command        string
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	Operations     []ProgressInfo
	TotalOps       int
	CompletedOps   int
	SuccessRate    float64
	MemoryUsed     string
	CPUTime        string
	ClusteringInfo *ClusteringMethodInfo // New field for clustering info
}

// EnableStats turns on statistics tracking
func EnableStats() {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	statsEnabled = true
	commandStartTime = time.Now()
	operationProgress = make(map[string]ProgressInfo)
	totalOperations = 0
	completedOps = 0
	Info("📊 Statistics tracking enabled")
}

// IsStatsEnabled returns whether stats tracking is currently enabled
func IsStatsEnabled() bool {
	statsMutex.RLock()
	defer statsMutex.RUnlock()
	return statsEnabled
}

// StartOperation tracks the start of an operation
func StartOperation(name string) {
	if !IsStatsEnabled() {
		return
	}

	statsMutex.Lock()
	defer statsMutex.Unlock()

	operationProgress[name] = ProgressInfo{
		Name:      name,
		StartTime: time.Now(),
		Progress:  0.0,
		Status:    "running",
	}
	totalOperations++
}

// UpdateProgress updates the progress of an operation
func UpdateProgress(name string, progress float64, status string) {
	if !IsStatsEnabled() {
		return
	}

	statsMutex.Lock()
	defer statsMutex.Unlock()

	if info, exists := operationProgress[name]; exists {
		info.Progress = progress
		info.Status = status
		operationProgress[name] = info
	}
}

// CompleteOperation marks an operation as completed
func CompleteOperation(name string) {
	if !IsStatsEnabled() {
		return
	}

	statsMutex.Lock()
	defer statsMutex.Unlock()

	if info, exists := operationProgress[name]; exists {
		now := time.Now()
		info.EndTime = &now
		info.Duration = now.Sub(info.StartTime)
		info.Progress = 100.0 // Always set progress to 100% on completion
		info.Status = "completed"
		operationProgress[name] = info
		completedOps++
	}
}

// FailOperation marks an operation as failed
func FailOperation(name string, reason string) {
	if !IsStatsEnabled() {
		return
	}

	statsMutex.Lock()
	defer statsMutex.Unlock()

	if info, exists := operationProgress[name]; exists {
		now := time.Now()
		info.EndTime = &now
		info.Duration = now.Sub(info.StartTime)
		info.Status = fmt.Sprintf("failed: %s", reason)
		operationProgress[name] = info
		completedOps++
	}
}

// PrintStats displays comprehensive command statistics
func PrintStats() {
	if !IsStatsEnabled() {
		return
	}

	statsMutex.RLock()
	defer statsMutex.RUnlock()

	endTime := time.Now()
	totalDuration := endTime.Sub(commandStartTime)

	// Calculate success rate
	successCount := 0
	for _, op := range operationProgress {
		if op.Status == "completed" {
			successCount++
		}
	}

	successRate := 0.0
	if totalOperations > 0 {
		successRate = (float64(successCount) / float64(totalOperations)) * 100
	}

	// Get memory usage
	memUsage := GetMemoryUsage()

	fmt.Printf("\n%s%s╔══════════════════════════════════════════════════════════════════════════════╗%s\n", Green, BlackBg, Reset)
	fmt.Printf("%s%s║                            📊 COMMAND STATISTICS                             ║%s\n", Green, BlackBg, Reset)
	fmt.Printf("%s%s╚══════════════════════════════════════════════════════════════════════════════╝%s\n", Green, BlackBg, Reset)

	fmt.Printf("\n%s⏱️  Total Execution Time:%s %v\n", Cyan, Reset, totalDuration.Round(time.Millisecond))
	fmt.Printf("%s📈 Operations Completed:%s %d/%d (%.1f%% success rate)\n", Cyan, Reset, successCount, totalOperations, successRate)
	fmt.Printf("%s💾 Memory Usage:%s %s\n", Cyan, Reset, memUsage)
	fmt.Printf("%s⚡ Start Time:%s %s\n", Cyan, Reset, commandStartTime.Format("15:04:05"))
	fmt.Printf("%s🏁 End Time:%s %s\n", Cyan, Reset, endTime.Format("15:04:05"))

	// Display clustering configuration if available
	if clusteringInfo != nil {
		fmt.Printf("\n%s%s🧠 CLUSTERING CONFIGURATION:%s\n", Yellow, Bold, Reset)
		fmt.Printf("%s🎯 Primary Method:%s %s\n", Cyan, Reset, clusteringInfo.Method)

		if len(clusteringInfo.EnabledMethods) > 0 {
			fmt.Printf("%s🔧 Enabled Methods:%s %v\n", Cyan, Reset, clusteringInfo.EnabledMethods)
		}

		fmt.Printf("%s🚀 Performance Mode:%s %s\n", Cyan, Reset, clusteringInfo.PerformanceMode)
		fmt.Printf("%s⏰ Max Processing Time:%s %ds\n", Cyan, Reset, clusteringInfo.MaxProcessingTime)
		fmt.Printf("%s📁 Max Files for Semantic:%s %d\n", Cyan, Reset, clusteringInfo.MaxFilesForSemantic)
		fmt.Printf("%s🔄 Fallback Methods:%s %t\n", Cyan, Reset, clusteringInfo.EnableFallbackMethods)
		fmt.Printf("%s📊 Benchmarking:%s %t\n", Cyan, Reset, clusteringInfo.EnableBenchmarking)
		fmt.Printf("%s🧠 Adaptive Optimization:%s %t\n", Cyan, Reset, clusteringInfo.AdaptiveOptimization)

		// Display confidence thresholds
		if len(clusteringInfo.ConfidenceThresholds) > 0 {
			fmt.Printf("\n%s%s🎚️ Confidence Thresholds:%s\n", Yellow, Bold, Reset)
			for method, threshold := range clusteringInfo.ConfidenceThresholds {
				fmt.Printf("   %s• %s:%s %.2f\n", Cyan, method, Reset, threshold)
			}
		}

		// Display similarity thresholds
		if len(clusteringInfo.SimilarityThresholds) > 0 {
			fmt.Printf("\n%s%s🔍 Similarity Thresholds:%s\n", Yellow, Bold, Reset)
			for method, threshold := range clusteringInfo.SimilarityThresholds {
				fmt.Printf("   %s• %s:%s %.2f\n", Cyan, method, Reset, threshold)
			}
		}
	}

	if len(operationProgress) > 0 {
		fmt.Printf("\n%s%s📋 Operation Details:%s\n", Yellow, Bold, Reset)
		for name, info := range operationProgress {
			status := info.Status
			statusColor := Green
			if info.Status != "completed" && info.Status != "running" {
				statusColor = Red
			} else if info.Status == "running" {
				statusColor = Yellow
			}

			duration := "ongoing"
			if info.EndTime != nil {
				duration = info.Duration.Round(time.Millisecond).String()
			}

			fmt.Printf("   %s• %s:%s %s%s%s (Duration: %s, Progress: %.1f%%)\n",
				Cyan, name, Reset, statusColor, status, Reset, duration, info.Progress)
		}
	}

	fmt.Printf("\n%s%s════════════════════════════════════════════════════════════════════════════════%s\n", Green, BlackBg, Reset)
}

// GetMemoryUsage returns current memory usage as a formatted string
func GetMemoryUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert bytes to megabytes for easier reading
	memMB := float64(m.Alloc) / 1024 / 1024
	return fmt.Sprintf("%.1f MB", memMB)
}

// AddStatsPostRunToCommand adds a PostRun function to display stats for a command
func AddStatsPostRunToCommand(cmd *cobra.Command) {
	// Save existing PostRun if there is one
	existingPostRun := cmd.PostRun

	cmd.PostRun = func(cmd *cobra.Command, args []string) {
		// Call the existing PostRun if there is one
		if existingPostRun != nil {
			existingPostRun(cmd, args)
		}

		// Mark command operation as complete if it was started
		if IsStatsEnabled() {
			commandName := "Command:" + cmd.Name()
			if _, exists := operationProgress[commandName]; exists {
				MarkOperationComplete(commandName)
			}

			// Print stats
			PrintStats()
		}
	}
}

// UpdateOperationProgress is a helper function to update operation progress
// and ensure consistent progress tracking in the stats system
func UpdateOperationProgress(name string, progress float64) {
	if !IsStatsEnabled() {
		return
	}

	// Update the progress
	UpdateProgress(name, progress, "running")
}

// MarkOperationComplete is a helper function to mark an operation as complete
// with proper progress handling
func MarkOperationComplete(name string) {
	if !IsStatsEnabled() {
		return
	}

	// Ensure progress is set to 100% before completion
	UpdateProgress(name, 100.0, "running")
	CompleteOperation(name)
}

// SetClusteringInfo stores clustering configuration information for stats display
func SetClusteringInfo(method string, enabledMethods []string, confidenceThresholds, similarityThresholds map[string]float64,
	maxFilesForSemantic int, enableFallbackMethods bool, performanceMode string, maxProcessingTime int,
	enableBenchmarking, adaptiveOptimization bool) {
	if !IsStatsEnabled() {
		return
	}

	statsMutex.Lock()
	defer statsMutex.Unlock()

	clusteringInfo = &ClusteringMethodInfo{
		Method:                method,
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
}

// GetClusteringInfo returns the stored clustering configuration
func GetClusteringInfo() *ClusteringMethodInfo {
	statsMutex.RLock()
	defer statsMutex.RUnlock()
	return clusteringInfo
}

// CaptureClusteringConfig captures clustering configuration from the config package
// This function should be called when clustering operations begin
func CaptureClusteringConfig() {
	if !IsStatsEnabled() {
		return
	}

	// We'll need to import the config package to get clustering configuration
	// For now, we'll create a basic version that can be enhanced
	statsMutex.Lock()
	defer statsMutex.Unlock()

	// Default clustering info - this will be enhanced when we integrate with config
	clusteringInfo = &ClusteringMethodInfo{
		Method:         "auto",
		EnabledMethods: []string{"directory", "pattern", "cached", "semantic"},
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
		MaxFilesForSemantic:   10,
		EnableFallbackMethods: true,
		PerformanceMode:       "balanced",
		MaxProcessingTime:     60,
		EnableBenchmarking:    false,
		AdaptiveOptimization:  true,
	}
}
