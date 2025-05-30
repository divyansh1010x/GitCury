#!/bin/bash

# Ultra-Fast CI Status Check Script
echo "⚡ GitCury Ultra-Fast CI Status"
echo "==============================="

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository"
    exit 1
fi

# Get current branch
BRANCH=$(git branch --show-current)
echo "📍 Current branch: $BRANCH"

# Get latest commit
COMMIT=$(git rev-parse --short HEAD)
echo "📝 Latest commit: $COMMIT"

echo ""
echo "🚀 Active Workflows:"

if [[ "$BRANCH" == "main" || "$BRANCH" == "master" ]]; then
    echo "  ⚡ Essential Checks (30-60s) - Build + Vet"
    echo "  🔧 PR Validation - GoReleaser check (if PR exists)"
else
    echo "  ⚡ Essential Checks (30-60s) - Build + Vet (on PR)"
    echo "  🔧 PR Validation - GoReleaser check (on PR)"
fi

echo ""
echo "📅 Scheduled/Manual Workflows:"
echo "  🧪 Comprehensive Tests - Weekly on Sunday 2 AM UTC"
echo "  🔧 Manual trigger available via workflow_dispatch"
echo "  🚀 Full validation on releases"

echo ""
echo "🎯 Performance Targets:"
echo "  • Essential checks: 30-60 seconds"
echo "  • PR validation: 1-2 minutes"
echo "  • Total feedback: Under 2 minutes"

echo ""
echo "💡 Quick commands:"
echo "   ./performance-test.sh  - Test locally"
echo "   Local build: go build ."
echo "   Local vet: go vet ./..."

echo ""
echo "🔗 Monitor at: https://github.com/lakshyajain-0291/gitcury/actions"
echo ""
echo "✅ Ultra-fast CI setup active!"
