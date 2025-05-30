#!/bin/bash

# Quick CI Status Check Script
# This script checks the status of GitHub Actions workflows

echo "🔍 Checking GitCury CI Status..."

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

# Check if we have any workflows that would run
echo ""
echo "🚀 Workflows that would trigger on push to $BRANCH:"

if [[ "$BRANCH" == "main" || "$BRANCH" == "master" ]]; then
    echo "  ✅ Fast Check"
    echo "  ✅ Code Quality"  
    echo "  ✅ PR Workflow (if PR exists)"
else
    echo "  ✅ Fast Check (on PR to main/master)"
    echo "  ✅ Code Quality (on PR to main/master)"
    echo "  ✅ PR Workflow (on PR to main/master)"
fi

echo ""
echo "⏰ Comprehensive Tests run:"
echo "  🌙 Nightly at 2 AM UTC"
echo "  🔧 Manually via workflow_dispatch"
echo "  🚀 On releases"

echo ""
echo "💡 To run performance test locally:"
echo "   ./performance-test.sh"

echo ""
echo "🔗 Workflow URLs:"
echo "   https://github.com/lakshyajain-0291/gitcury/actions"
