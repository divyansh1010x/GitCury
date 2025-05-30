#!/bin/bash

# GitCury Manual Release Script
# This script allows you to manually create and release a specific version using GoReleaser
# Usage: ./manual-release.sh [version] [--dry-run] [--force]
# Example: ./manual-release.sh v1.2.3
# Example: ./manual-release.sh v1.2.3 --dry-run (test without actually releasing)

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DRY_RUN=false
FORCE=false
VERSION=""

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to show usage
show_usage() {
    echo "GitCury Manual Release Script"
    echo ""
    echo "Usage: $0 [version] [options]"
    echo ""
    echo "Arguments:"
    echo "  version     Version to release (e.g., v1.2.3, v2.0.0-beta.1)"
    echo ""
    echo "Options:"
    echo "  --dry-run   Test the release process without actually releasing"
    echo "  --force     Force release even if tag already exists"
    echo "  --help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 v1.2.3                    # Release version 1.2.3"
    echo "  $0 v1.2.3 --dry-run          # Test release of version 1.2.3"
    echo "  $0 v2.0.0-beta.1 --force     # Force release of beta version"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --help|-h)
            show_usage
            exit 0
            ;;
        v*)
            VERSION="$1"
            shift
            ;;
        *)
            if [[ -z "$VERSION" ]]; then
                VERSION="v$1"
            else
                print_error "Unknown option: $1"
                show_usage
                exit 1
            fi
            shift
            ;;
    esac
done

# Validate version format
if [[ -z "$VERSION" ]]; then
    print_error "Version is required!"
    show_usage
    exit 1
fi

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
    print_error "Invalid version format: $VERSION"
    print_info "Version should follow semantic versioning (e.g., v1.2.3, v2.0.0-beta.1)"
    exit 1
fi

print_info "Starting GitCury release process for version: $VERSION"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "Not in a git repository!"
    exit 1
fi

# Check if working directory is clean
if [[ -n $(git status --porcelain) ]]; then
    print_warning "Working directory is not clean!"
    git status --short
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Aborting release"
        exit 1
    fi
fi

# Check if tag already exists (both locally and remotely)
LOCAL_TAG_EXISTS=false
REMOTE_TAG_EXISTS=false

if git tag -l | grep -q "^$VERSION$"; then
    LOCAL_TAG_EXISTS=true
fi

# Check if tag exists on remote
if git ls-remote --tags origin | grep -q "refs/tags/$VERSION$"; then
    REMOTE_TAG_EXISTS=true
fi

if [[ "$LOCAL_TAG_EXISTS" == true ]] || [[ "$REMOTE_TAG_EXISTS" == true ]]; then
    if [[ "$FORCE" == false ]]; then
        if [[ "$LOCAL_TAG_EXISTS" == true ]] && [[ "$REMOTE_TAG_EXISTS" == true ]]; then
            print_error "Tag $VERSION already exists locally and remotely!"
        elif [[ "$LOCAL_TAG_EXISTS" == true ]]; then
            print_error "Tag $VERSION already exists locally!"
        else
            print_error "Tag $VERSION already exists on remote!"
        fi
        print_info "Use --force to overwrite existing tag"
        exit 1
    else
        if [[ "$LOCAL_TAG_EXISTS" == true ]] && [[ "$REMOTE_TAG_EXISTS" == true ]]; then
            print_warning "Tag $VERSION exists locally and remotely, will be overwritten"
        elif [[ "$LOCAL_TAG_EXISTS" == true ]]; then
            print_warning "Tag $VERSION exists locally, will be overwritten"
        else
            print_warning "Tag $VERSION exists on remote, will be overwritten"
        fi
        
        if [[ "$DRY_RUN" == false ]]; then
            # Delete local tag if it exists
            if [[ "$LOCAL_TAG_EXISTS" == true ]]; then
                git tag -d "$VERSION" || true
            fi
            # Delete remote tag if it exists
            if [[ "$REMOTE_TAG_EXISTS" == true ]]; then
                git push origin ":refs/tags/$VERSION" || true
            fi
        fi
    fi
fi

# Ensure we're on main/master branch
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "master" ]]; then
    print_warning "Current branch is '$CURRENT_BRANCH', not 'main' or 'master'"
    read -p "Continue from this branch? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Please switch to main/master branch and try again"
        exit 1
    fi
fi

# Pull latest changes and fetch tags
print_info "Pulling latest changes and fetching tags..."
if [[ "$DRY_RUN" == false ]]; then
    git pull origin "$CURRENT_BRANCH"
    git fetch --tags
else
    print_info "DRY RUN: Would pull latest changes and fetch tags"
fi

# Check if GoReleaser is installed
if ! command -v goreleaser &> /dev/null; then
    print_error "GoReleaser is not installed!"
    print_info "Install it with: go install github.com/goreleaser/goreleaser@latest"
    exit 1
fi

# Validate GoReleaser configuration
print_info "Validating GoReleaser configuration..."
if ! goreleaser check; then
    print_error "GoReleaser configuration is invalid!"
    exit 1
fi

print_success "GoReleaser configuration is valid"

# Check required environment variables for release
if [[ "$DRY_RUN" == false ]]; then
    if [[ -z "$GITHUB_TOKEN" ]]; then
        print_warning "GITHUB_TOKEN not set. Release will fail without it."
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Set GITHUB_TOKEN environment variable and try again"
            exit 1
        fi
    fi
fi

# Show what will be released
print_info "Release summary:"
print_info "  Version: $VERSION"
print_info "  Branch: $CURRENT_BRANCH"
print_info "  Commit: $(git rev-parse --short HEAD)"
print_info "  Dry run: $DRY_RUN"

if [[ "$DRY_RUN" == false ]]; then
    echo
    print_warning "This will create a tag and release $VERSION"
    read -p "Are you sure you want to proceed? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Release cancelled"
        exit 0
    fi
fi

# Create and push tag
if [[ "$DRY_RUN" == false ]]; then
    print_info "Creating tag $VERSION..."
    if ! git tag -a "$VERSION" -m "Release $VERSION"; then
        print_error "Failed to create tag $VERSION"
        exit 1
    fi
    
    print_info "Pushing tag to origin..."
    if ! git push origin "$VERSION"; then
        print_error "Failed to push tag $VERSION to origin"
        print_warning "Cleaning up local tag..."
        git tag -d "$VERSION" || true
        print_info "If the tag exists remotely, you can use --force to overwrite it"
        exit 1
    fi
    
    print_success "Tag $VERSION created and pushed"
else
    print_info "DRY RUN: Would create and push tag $VERSION"
fi

# Run GoReleaser
print_info "Running GoReleaser..."

if [[ "$DRY_RUN" == true ]]; then
    # Use GoReleaser's snapshot mode for dry run
    print_info "Running GoReleaser in snapshot mode (dry run)..."
    if goreleaser release --snapshot --skip=publish --clean; then
        print_success "GoReleaser dry run completed successfully!"
        print_info "Check the 'dist' directory for generated artifacts"
    else
        print_error "GoReleaser dry run failed!"
        exit 1
    fi
else
    # Actual release
    print_info "Running GoReleaser for real release..."
    if goreleaser release --clean; then
        print_success "Release $VERSION completed successfully! 🎉"
        print_info "Check GitHub releases: https://github.com/lakshyajain-0291/gitcury/releases"
    else
        print_error "GoReleaser release failed!"
        print_warning "You may need to manually delete the tag if it was created:"
        print_warning "  git tag -d $VERSION"
        print_warning "  git push origin :refs/tags/$VERSION"
        exit 1
    fi
fi

print_success "Release process completed!"

if [[ "$DRY_RUN" == false ]]; then
    echo
    print_info "Post-release checklist:"
    print_info "  ✓ Tag $VERSION created and pushed"
    print_info "  ✓ GitHub release created"
    print_info "  ✓ Binaries uploaded"
    print_info "  → Check release page: https://github.com/lakshyajain-0291/gitcury/releases/tag/$VERSION"
    print_info "  → Verify binary installation: go install github.com/lakshyajain-0291/gitcury@$VERSION"
fi
