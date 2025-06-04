package utils

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type PromptRequest struct {
	Question string
	Default  bool
	RespChan chan bool
}

// ResourceManager monitors and manages system resources for the application
type ResourceManager struct {
	maxMemoryPercent float64
	maxCPUPercent    float64
	checkInterval    time.Duration

	// State
	running      bool
	stopChan     chan struct{}
	resourceLock sync.RWMutex

	// Metrics
	memStats         runtime.MemStats
	lastMemoryUsage  uint64
	lastCPUUsage     float64 //nolint:unused // Will be used in future CPU monitoring
	lastChecked      time.Time
	resourceWarnings int
}

func StartPromptCoordinator(promptChan <-chan PromptRequest) {
	for req := range promptChan {
		StopCreativeLoader() // Optional: Stop loader/spinner if active
		fmt.Println()

		Info("🔔 Question from a folder:")
		Info("👉 " + req.Question)

		resp := ConfirmAction(req.Question, req.Default)
		req.RespChan <- resp

		StartCreativeLoader("Resuming processing...", BrailleAnimation)
	}
}

// NewResourceManager creates a resource manager with default settings
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		maxMemoryPercent: 80.0, // Default to 80% max memory usage
		maxCPUPercent:    90.0, // Default to 90% max CPU usage
		checkInterval:    5 * time.Second,
		stopChan:         make(chan struct{}),
		resourceWarnings: 0,
		lastChecked:      time.Now(),
	}
}

// SetMaxMemoryPercent sets the maximum allowed memory usage as a percentage
func (rm *ResourceManager) SetMaxMemoryPercent(percent float64) {
	if percent <= 0 || percent > 100 {
		Warning("Invalid memory percent value, using default")
		percent = 80.0
	}
	rm.resourceLock.Lock()
	defer rm.resourceLock.Unlock()
	rm.maxMemoryPercent = percent
}

// SetMaxCPUPercent sets the maximum allowed CPU usage as a percentage
func (rm *ResourceManager) SetMaxCPUPercent(percent float64) {
	if percent <= 0 || percent > 100 {
		Warning("Invalid CPU percent value, using default")
		percent = 90.0
	}
	rm.resourceLock.Lock()
	defer rm.resourceLock.Unlock()
	rm.maxCPUPercent = percent
}

// SetCheckInterval sets how often resources are checked
func (rm *ResourceManager) SetCheckInterval(interval time.Duration) {
	if interval < time.Second {
		Warning("Check interval too small, using minimum of 1 second")
		interval = time.Second
	}
	rm.resourceLock.Lock()
	defer rm.resourceLock.Unlock()
	rm.checkInterval = interval
}

// Start begins monitoring system resources
func (rm *ResourceManager) Start() {
	rm.resourceLock.Lock()
	if rm.running {
		rm.resourceLock.Unlock()
		return
	}
	rm.running = true
	rm.resourceLock.Unlock()

	Debug("[RESOURCE]: Resource monitoring started")

	go func() {
		ticker := time.NewTicker(rm.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rm.checkResources()
			case <-rm.stopChan:
				Debug("[RESOURCE]: Resource monitoring stopped")
				return
			}
		}
	}()
}

// Stop ends resource monitoring
func (rm *ResourceManager) Stop() {
	rm.resourceLock.Lock()
	defer rm.resourceLock.Unlock()

	if !rm.running {
		return
	}

	rm.running = false
	rm.stopChan <- struct{}{}
}

// IsRunning returns whether resource monitoring is active
func (rm *ResourceManager) IsRunning() bool {
	rm.resourceLock.RLock()
	defer rm.resourceLock.RUnlock()
	return rm.running
}

// GetResourceUsage returns the current resource usage metrics
func (rm *ResourceManager) GetResourceUsage() map[string]interface{} {
	rm.resourceLock.RLock()
	defer rm.resourceLock.RUnlock()

	// Force an update of the metrics
	runtime.ReadMemStats(&rm.memStats)

	return map[string]interface{}{
		"memoryUsageMB":    rm.memStats.Alloc / 1024 / 1024,
		"totalMemoryMB":    rm.memStats.Sys / 1024 / 1024,
		"numGoroutines":    runtime.NumGoroutine(),
		"resourceWarnings": rm.resourceWarnings,
		"lastChecked":      rm.lastChecked,
		"maxMemoryPercent": rm.maxMemoryPercent,
		"maxCPUPercent":    rm.maxCPUPercent,
	}
}

// checkResources examines current resource usage and takes action if thresholds are exceeded
func (rm *ResourceManager) checkResources() {
	rm.resourceLock.Lock()
	defer rm.resourceLock.Unlock()

	rm.lastChecked = time.Now()

	// Get current memory stats
	runtime.ReadMemStats(&rm.memStats)

	// Check memory usage
	totalMemory := rm.memStats.Sys
	usedMemory := rm.memStats.Alloc
	memoryPercent := float64(usedMemory) / float64(totalMemory) * 100

	// Check CPU usage (simplified since Go doesn't provide direct CPU usage)
	numGoroutines := runtime.NumGoroutine()

	// Store current values
	rm.lastMemoryUsage = usedMemory

	// Log resource usage in debug mode
	Debug(fmt.Sprintf("[RESOURCE]: Memory: %.2f%% (%d MB / %d MB), Goroutines: %d",
		memoryPercent, usedMemory/1024/1024, totalMemory/1024/1024, numGoroutines))

	// Check if we're exceeding thresholds
	if memoryPercent > rm.maxMemoryPercent {
		rm.resourceWarnings++
		Warning(fmt.Sprintf("[RESOURCE.WARNING]: Memory usage is high: %.2f%% (threshold: %.2f%%)",
			memoryPercent, rm.maxMemoryPercent))

		// Take action - force garbage collection if memory usage is critical
		if memoryPercent > rm.maxMemoryPercent+10 {
			Info("[RESOURCE]: Forcing garbage collection due to high memory usage")
			runtime.GC()
		}
	}

	// Check if number of goroutines is unusually high (simplistic approach)
	if numGoroutines > 1000 {
		rm.resourceWarnings++
		Warning(fmt.Sprintf("[RESOURCE.WARNING]: High number of goroutines: %d", numGoroutines))
	}
}

// GetRecommendedWorkerCount returns the recommended number of worker goroutines
// based on current system load and available CPU cores
func (rm *ResourceManager) GetRecommendedWorkerCount(defaultWorkers int) int {
	rm.resourceLock.RLock()
	defer rm.resourceLock.RUnlock()

	// Get number of CPU cores
	numCPU := runtime.NumCPU()

	// Calculate recommended worker count
	var recommendedWorkers int

	// If we're under resource pressure, reduce worker count
	if rm.resourceWarnings > 5 {
		// Under significant resource pressure
		recommendedWorkers = max(1, numCPU/4)
	} else if rm.resourceWarnings > 0 {
		// Under mild resource pressure
		recommendedWorkers = max(1, numCPU/2)
	} else {
		// No resource pressure, use 75% of available cores
		recommendedWorkers = max(1, (numCPU*3)/4)
	}

	// If defaultWorkers is specified and smaller than our calculation, use that instead
	if defaultWorkers > 0 && defaultWorkers < recommendedWorkers {
		recommendedWorkers = defaultWorkers
	}

	return recommendedWorkers
}

// Default instance
var defaultResourceManager *ResourceManager
var rmOnce sync.Once

// GetResourceManager returns the default resource manager instance
func GetResourceManager() *ResourceManager {
	rmOnce.Do(func() {
		defaultResourceManager = NewResourceManager()
	})
	return defaultResourceManager
}

// Helper function for Go < 1.21
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
