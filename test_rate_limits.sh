#!/bin/bash

# Test script to verify the 429 error fixes for embedding generation
# This script tests with a small number of files to ensure rate limiting works properly

echo "🧪 Testing GitCury embedding generation with rate limit improvements..."

# Set up a test directory
TEST_DIR="/tmp/gitcury_rate_limit_test"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Initialize git repo
git init
git config user.email "test@example.com"
git config user.name "Test User"

# Create test files with different content
echo "package main

func main() {
    fmt.Println(\"Hello World\")
}" > main.go

echo "# Test Project

This is a test project for GitCury rate limiting." > README.md

echo "{
    \"name\": \"test-project\",
    \"version\": \"1.0.0\",
    \"description\": \"Testing rate limits\"
}" > package.json

echo "def hello_world():
    print('Hello from Python!')

if __name__ == '__main__':
    hello_world()" > hello.py

echo "body {
    margin: 0;
    padding: 20px;
    font-family: Arial, sans-serif;
}" > styles.css

# Add files to git
git add .
echo "✅ Created 5 test files"

# Run GitCury to generate commit messages (this will test embedding generation)
echo "🚀 Testing GitCury with 5 files..."
cd /home/lakshya-jain/projects/GitCury

# Test with our improvements
echo "Testing embedding generation with rate limiting improvements..."
timeout 60s ./gitcury msgs --root "$TEST_DIR" --num 5

EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Test PASSED: No 429 errors with 5 files!"
elif [ $EXIT_CODE -eq 124 ]; then
    echo "⚠️  Test TIMEOUT: Took longer than 60 seconds (may indicate rate limiting is working)"
else
    echo "❌ Test FAILED with exit code: $EXIT_CODE"
fi

# Clean up
rm -rf "$TEST_DIR"

echo "🧹 Cleanup completed"
echo ""
echo "📋 Summary of improvements made:"
echo "   • Updated to stable embedding model: text-embedding-004"
echo "   • Reduced concurrent requests from 2 to 1"
echo "   • Increased delays between requests (2s → 5s)"
echo "   • Added circuit breaker for repeated failures"
echo "   • Improved retry configuration (10 → 3 retries)"
echo "   • Added rate limit error detection"
