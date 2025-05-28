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
)

type ProgressInfo struct {
	Name      string
	StartTime time.Time
	EndTime   *time.Time
	Duration  time.Duration
	Progress  float64
	Status    string
}

type CommandStats struct {
	Command         string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Operations      []ProgressInfo
	TotalOps        int
	CompletedOps    int
	SuccessRate     float64
	MemoryUsed      string
	CPUTime         string
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
		info.Progress = 100.0  // Always set progress to 100% on completion
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
