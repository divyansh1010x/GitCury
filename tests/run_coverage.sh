#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Fancy Header
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                                                        ║${NC}"
echo -e "${CYAN}║  ${PURPLE}GitCury Test Coverage Report 🧪${CYAN}                      ║${NC}"
echo -e "${CYAN}║                                                        ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Run tests with coverage
echo -e "${BLUE}Running tests with coverage...${NC}"
go test -v -race -coverprofile=coverage.out -covermode=atomic ./tests/... -timeout 30s

# Check if tests passed
if [ $? -ne 0 ]; then
  echo -e "${RED}Tests failed! 😥${NC}"
  exit 1
fi

echo -e "${GREEN}All tests passed successfully! 🎉${NC}"
echo ""

# Generate coverage report
go tool cover -func=coverage.out -o coverage.txt

# Extract total coverage
COVERAGE=$(tail -1 coverage.txt | awk '{print $3}')

echo -e "${YELLOW}═════════════════════════════════════════════════${NC}"
echo -e "${GREEN}📊 Code Coverage Summary${NC}"
echo -e "${YELLOW}═════════════════════════════════════════════════${NC}"

# Convert coverage to a number for comparison
COVERAGE_NUM=$(echo $COVERAGE | tr -d '%')

# Print coverage with color based on percentage
if (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
  echo -e "${GREEN}Total Coverage: $COVERAGE${NC} 🚀"
elif (( $(echo "$COVERAGE_NUM >= 60" | bc -l) )); then
  echo -e "${YELLOW}Total Coverage: $COVERAGE${NC} 👍"
else
  echo -e "${RED}Total Coverage: $COVERAGE${NC} ⚠️"
fi

echo ""
echo -e "${BLUE}Detailed Package Coverage:${NC}"
echo -e "${YELLOW}═════════════════════════════════════════════════${NC}"

# Print package coverage in a more readable format
while read line; do
  PKG=$(echo $line | awk '{print $1}')
  FUNC=$(echo $line | awk '{print $2}')
  COV=$(echo $line | awk '{print $3}')
  
  # Skip the total line at the end
  if [[ "$FUNC" == "total:" ]]; then
    continue
  fi
  
  # Convert coverage to a number for comparison
  COV_NUM=$(echo $COV | tr -d '%')
  
  # Print with color based on percentage
  if (( $(echo "$COV_NUM >= 80" | bc -l) )); then
    echo -e "${GREEN}$PKG${NC}\t${CYAN}$FUNC${NC}\t${GREEN}$COV${NC}"
  elif (( $(echo "$COV_NUM >= 60" | bc -l) )); then
    echo -e "${GREEN}$PKG${NC}\t${CYAN}$FUNC${NC}\t${YELLOW}$COV${NC}"
  else
    echo -e "${GREEN}$PKG${NC}\t${CYAN}$FUNC${NC}\t${RED}$COV${NC}"
  fi
done < coverage.txt

echo ""
echo -e "${PURPLE}Generating HTML coverage report...${NC}"
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}HTML coverage report generated at coverage.html${NC}"

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                                                        ║${NC}"
echo -e "${CYAN}║  ${GREEN}Testing complete! View coverage.html for details${CYAN}      ║${NC}"
echo -e "${CYAN}║                                                        ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
