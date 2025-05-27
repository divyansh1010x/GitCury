#!/bin/bash
# GitCury Test Runner
# This script runs all tests and generates a comprehensive report

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
REPORT_DIR="./test-reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="${REPORT_DIR}/gitcury_test_report_${TIMESTAMP}.json"
COVERAGE_ENABLED=true
DETAILED_OUTPUT=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-coverage)
      COVERAGE_ENABLED=false
      shift
      ;;
    --detailed)
      DETAILED_OUTPUT=true
      shift
      ;;
    --report-dir)
      REPORT_DIR="$2"
      REPORT_FILE="${REPORT_DIR}/gitcury_test_report_${TIMESTAMP}.json"
      shift 2
      ;;
    --help)
      echo "GitCury Test Runner"
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  --no-coverage     Disable code coverage collection"
      echo "  --detailed        Show detailed test output"
      echo "  --report-dir DIR  Directory for test reports (default: ./test-reports)"
      echo "  --help            Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information."
      exit 1
      ;;
  esac
done

# Create report directory
mkdir -p "$REPORT_DIR"

echo -e "${GREEN}Starting GitCury test suite...${NC}"
echo "Report will be saved to: $REPORT_FILE"

# Build test report generator
echo -e "${YELLOW}Building test report generator...${NC}"
go build -o testreport_runner ./tests/testreport/testreport.go

# Build test runner
echo -e "${YELLOW}Building test runner...${NC}"
cat > runner.go << 'EOF'
package main

import (
	"GitCury/tests/testreport"
	"fmt"
	"os"
)

func main() {
	reportPath := os.Args[1]
	coverageEnabled := os.Args[2] == "true"
	
	report, err := testreport.GenerateTestReport(reportPath, coverageEnabled)
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		os.Exit(1)
	}
	
	testreport.PrintReportSummary(report)
	
	// Exit with error if any tests failed
	if report.FailedTests > 0 {
		os.Exit(1)
	}
}
EOF

go build -o test_runner runner.go
rm runner.go

# Run tests with detailed output if requested
if [ "$DETAILED_OUTPUT" = true ]; then
  echo -e "${YELLOW}Running tests with detailed output...${NC}"
  go test -v ./...
fi

# Run test runner
echo -e "${YELLOW}Running tests and generating report...${NC}"
./test_runner "$REPORT_FILE" "$COVERAGE_ENABLED"
TEST_EXIT_CODE=$?

# Clean up
rm test_runner testreport_runner

# Generate HTML report
echo -e "${YELLOW}Generating HTML report...${NC}"
cat > "$REPORT_DIR/report.html" << EOF
<!DOCTYPE html>
<html>
<head>
  <title>GitCury Test Report</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
    h1 { color: #333; }
    .summary { background-color: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
    .pass { color: green; }
    .fail { color: red; }
    .skip { color: orange; }
    table { width: 100%; border-collapse: collapse; }
    th, td { padding: 8px; text-align: left; border-bottom: 1px solid #ddd; }
    th { background-color: #f2f2f2; }
    tr:hover { background-color: #f5f5f5; }
    .test-result { margin-bottom: 10px; padding: 10px; border-radius: 5px; }
    .test-pass { background-color: #dff0d8; }
    .test-fail { background-color: #f2dede; }
    .test-skip { background-color: #fcf8e3; }
    pre { background-color: #f8f8f8; padding: 10px; border-radius: 5px; overflow-x: auto; }
  </style>
  <script>
    function loadReport() {
      fetch('$(basename "$REPORT_FILE")')
        .then(response => response.json())
        .then(data => {
          document.getElementById('totalTests').textContent = data.totalTests;
          document.getElementById('passedTests').textContent = data.passedTests;
          document.getElementById('failedTests').textContent = data.failedTests;
          document.getElementById('skippedTests').textContent = data.skippedTests;
          document.getElementById('duration').textContent = (data.totalDuration / 1000000000).toFixed(2) + ' seconds';
          document.getElementById('coverage').textContent = data.coverage ? data.coverage.toFixed(2) + '%' : 'N/A';
          document.getElementById('timestamp').textContent = new Date(data.timestamp).toLocaleString();
          document.getElementById('gitCuryVersion').textContent = data.gitCuryVersion;
          document.getElementById('goVersion').textContent = data.goVersion;
          
          const resultsList = document.getElementById('results');
          resultsList.innerHTML = '';
          
          data.results.forEach(result => {
            const resultClass = result.success ? 'test-pass' : (result.skipMsg ? 'test-skip' : 'test-fail');
            const resultStatus = result.success ? 'PASS' : (result.skipMsg ? 'SKIP' : 'FAIL');
            const statusClass = result.success ? 'pass' : (result.skipMsg ? 'skip' : 'fail');
            
            const resultDiv = document.createElement('div');
            resultDiv.className = 'test-result ' + resultClass;
            
            resultDiv.innerHTML = \`
              <h3>\${result.package}.\${result.name}</h3>
              <p>Status: <span class="\${statusClass}">\${resultStatus}</span></p>
              <p>Duration: \${(result.duration / 1000000).toFixed(2)} ms</p>
              \${result.errorMsg ? '<pre>' + result.errorMsg + '</pre>' : ''}
              \${result.skipMsg ? '<pre>' + result.skipMsg + '</pre>' : ''}
            \`;
            
            resultsList.appendChild(resultDiv);
          });
        })
        .catch(error => {
          console.error('Error loading report:', error);
          document.getElementById('results').innerHTML = '<p>Error loading report: ' + error.message + '</p>';
        });
    }
  </script>
</head>
<body onload="loadReport()">
  <h1>GitCury Test Report</h1>
  
  <div class="summary">
    <h2>Summary</h2>
    <table>
      <tr><td>Total Tests:</td><td id="totalTests">-</td></tr>
      <tr><td>Passed Tests:</td><td id="passedTests" class="pass">-</td></tr>
      <tr><td>Failed Tests:</td><td id="failedTests" class="fail">-</td></tr>
      <tr><td>Skipped Tests:</td><td id="skippedTests" class="skip">-</td></tr>
      <tr><td>Total Duration:</td><td id="duration">-</td></tr>
      <tr><td>Code Coverage:</td><td id="coverage">-</td></tr>
      <tr><td>Timestamp:</td><td id="timestamp">-</td></tr>
      <tr><td>GitCury Version:</td><td id="gitCuryVersion">-</td></tr>
      <tr><td>Go Version:</td><td id="goVersion">-</td></tr>
    </table>
  </div>
  
  <h2>Test Results</h2>
  <div id="results">
    <p>Loading test results...</p>
  </div>
</body>
</html>
EOF

echo -e "${GREEN}Test execution completed!${NC}"
echo "HTML report generated at: $REPORT_DIR/report.html"
echo "JSON report available at: $REPORT_FILE"

# Exit with the test exit code
exit $TEST_EXIT_CODE
