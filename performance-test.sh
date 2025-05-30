#!/bin/bash

echo "⚡ Minimal CI workflow simulation"
echo "================================"

echo "📦 Testing build..."
if go build . > /dev/null 2>&1; then
    echo "✅ Build: PASS"
else
    echo "❌ Build: FAIL"
    exit 1
fi

echo "🔍 Testing vet..."  
if go vet ./... > /dev/null 2>&1; then
    echo "✅ Vet: PASS"
else
    echo "❌ Vet: FAIL"
    exit 1
fi

echo ""
echo "🎯 Workflow Summary:"
echo "   • Only essential checks (build + vet)"
echo "   • No tests (moved to comprehensive workflow)"
echo "   • No complex linting"
echo "   • Expected CI time: 30-60 seconds"
echo ""
echo "✅ Minimal workflow simulation completed!"
