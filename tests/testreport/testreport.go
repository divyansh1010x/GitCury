// Package testreport provides utilities for generating test reports
package testreport

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestResult represents the result of a single test
type TestResult struct {
	Name      string        `json:"name"`
	Package   string        `json:"package"`
	Success   bool          `json:"success"`
	Duration  time.Duration `json:"duration"`
	ErrorMsg  string        `json:"errorMsg,omitempty"`
	SkipMsg   string        `json:"skipMsg,omitempty"`
	TimeStamp time.Time     `json:"timestamp"`
}

// TestReport represents a complete test report
type TestReport struct {
	TotalTests      int           `json:"totalTests"`
	PassedTests     int           `json:"passedTests"`
	FailedTests     int           `json:"failedTests"`
	SkippedTests    int           `json:"skippedTests"`
	TotalDuration   time.Duration `json:"totalDuration"`
	Coverage        float64       `json:"coverage,omitempty"`
	Results         []TestResult  `json:"results"`
	TimeStamp       time.Time     `json:"timestamp"`
	GitCuryVersion  string        `json:"gitCuryVersion"`
	GoVersion       string        `json:"goVersion"`
	OperatingSystem string        `json:"operatingSystem"`
}

// GenerateTestReport runs all tests and generates a comprehensive report
func GenerateTestReport(outputPath string, coverageEnabled bool) (*TestReport, error) {
	startTime := time.Now()

	// Initialize report
	report := &TestReport{
		Results:   make([]TestResult, 0),
		TimeStamp: startTime,
	}

	// Get system information
	report.GoVersion = getGoVersion()
	report.OperatingSystem = getOS()
	report.GitCuryVersion = getGitCuryVersion()

	// Build test command
	args := []string{"test", "./..."}
	if coverageEnabled {
		args = append(args, "-cover")
	}
	args = append(args, "-v", "-json")

	// Run tests with JSON output
	cmd := exec.Command("go", args...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Don't return error here, we want to generate a report even if tests fail
		fmt.Printf("Warning: tests completed with error: %v\n", err)
	}

	// Parse test output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Try to parse as JSON
		var testEvent map[string]interface{}
		if err := json.Unmarshal([]byte(line), &testEvent); err != nil {
			continue
		}

		// Check if this is a test result
		if eventType, ok := testEvent["Action"].(string); ok && eventType == "run" {
			// Process test result
			testName, _ := testEvent["Test"].(string)
			pkgName, _ := testEvent["Package"].(string)

			// Find corresponding pass/fail/skip event
			result := findTestResult(lines, testName, pkgName)
			if result != nil {
				report.Results = append(report.Results, *result)
			}
		}
	}

	// Calculate statistics
	for _, result := range report.Results {
		report.TotalTests++
		if result.Success {
			report.PassedTests++
		} else if result.SkipMsg != "" {
			report.SkippedTests++
		} else {
			report.FailedTests++
		}
		report.TotalDuration += result.Duration
	}

	// Get coverage if enabled
	if coverageEnabled {
		report.Coverage = extractCoverage(string(output))
	}

	// Write report to file if path provided
	if outputPath != "" {
		if err := writeReport(report, outputPath); err != nil {
			return report, fmt.Errorf("failed to write report: %w", err)
		}
	}

	return report, nil
}

// Helper to find test result from output
func findTestResult(lines []string, testName, pkgName string) *TestResult {
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var testEvent map[string]interface{}
		if err := json.Unmarshal([]byte(line), &testEvent); err != nil {
			continue
		}

		eventTestName, _ := testEvent["Test"].(string)
		eventPkgName, _ := testEvent["Package"].(string)
		eventType, _ := testEvent["Action"].(string)

		if eventTestName == testName && eventPkgName == pkgName &&
			(eventType == "pass" || eventType == "fail" || eventType == "skip") {

			// Create test result
			result := &TestResult{
				Name:      testName,
				Package:   pkgName,
				Success:   eventType == "pass",
				TimeStamp: time.Now(),
			}

			// Get duration
			if duration, ok := testEvent["Elapsed"].(float64); ok {
				result.Duration = time.Duration(duration * float64(time.Second))
			}

			// Get error message for failed tests
			if eventType == "fail" {
				if output, ok := testEvent["Output"].([]interface{}); ok && len(output) > 0 {
					var errorMsg strings.Builder
					for _, line := range output {
						errorMsg.WriteString(line.(string))
						errorMsg.WriteString("\n")
					}
					result.ErrorMsg = errorMsg.String()
				}
			}

			// Get skip message
			if eventType == "skip" {
				if output, ok := testEvent["Output"].([]interface{}); ok && len(output) > 0 {
					var skipMsg strings.Builder
					for _, line := range output {
						skipMsg.WriteString(line.(string))
						skipMsg.WriteString("\n")
					}
					result.SkipMsg = skipMsg.String()
				}
			}

			return result
		}
	}

	return nil
}

// Helper to extract coverage percentage
func extractCoverage(output string) float64 {
	// Look for coverage line in output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			// Parse coverage percentage
			parts := strings.Split(line, "coverage:")
			if len(parts) > 1 {
				coveragePart := strings.TrimSpace(parts[1])
				coveragePct := strings.TrimSuffix(coveragePart, "%")
				if coverage, err := parseFloat(coveragePct); err == nil {
					return coverage
				}
			}
		}
	}

	return 0.0
}

// Helper to parse float
func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// Helper to write report to file
func writeReport(report *TestReport, path string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0644)
}

// Helper to get Go version
func getGoVersion() string {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// Helper to get OS information
func getOS() string {
	return fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH"))
}

// Helper to get GitCury version
func getGitCuryVersion() string {
	// This would ideally use the GitCury version from a version package
	return "development" // Replace with actual version
}

// PrintReportSummary prints a summary of the test report
func PrintReportSummary(report *TestReport) {
	fmt.Println("=== GitCury Test Report Summary ===")
	fmt.Printf("Total Tests: %d\n", report.TotalTests)
	fmt.Printf("Passed: %d\n", report.PassedTests)
	fmt.Printf("Failed: %d\n", report.FailedTests)
	fmt.Printf("Skipped: %d\n", report.SkippedTests)
	fmt.Printf("Total Duration: %v\n", report.TotalDuration)

	if report.Coverage > 0 {
		fmt.Printf("Code Coverage: %.2f%%\n", report.Coverage)
	}

	fmt.Println("===============================")

	// Print failed tests if any
	if report.FailedTests > 0 {
		fmt.Println("Failed Tests:")
		for _, result := range report.Results {
			if !result.Success && result.SkipMsg == "" {
				fmt.Printf("- %s.%s\n", result.Package, result.Name)
			}
		}
	}
}
