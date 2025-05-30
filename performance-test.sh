#!/bin/bash

echo "🚀 Testing GitCury build and test performance..."

echo "📦 Building project..."
time go build -v ./...

echo ""
echo "🧪 Running fast tests (no race detection)..."
time go test -v ./...

echo ""
echo "🏃 Running tests with race detection..."
time go test -v -race ./...

echo ""
echo "📊 Running benchmark tests..."
time go test -bench=. -benchmem ./tests/

echo ""
echo "🔍 Running quick lint checks..."
time golangci-lint run --fast --timeout=1m --disable-all --enable=errcheck,gosimple,govet,ineffassign,staticcheck,typecheck,unused

echo ""
echo "✅ Performance test completed!"
