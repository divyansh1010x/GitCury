#!/bin/bash
# Script to set up all distribution channels for GitCury

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}GitCury Distribution Channels Setup${NC}"
echo "================================================="
echo ""

# Check if GitHub CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}GitHub CLI (gh) is not installed.${NC}"
    echo "Please install it from: https://cli.github.com/"
    exit 1
fi

# Check if logged in to GitHub
echo -e "${YELLOW}Checking GitHub authentication...${NC}"
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Not logged in to GitHub.${NC}"
    echo "Please run 'gh auth login' first."
    exit 1
fi
echo -e "${GREEN}✓ GitHub authentication verified${NC}"

# Get GitHub username
GITHUB_USERNAME=$(gh api user | jq -r '.login')
if [ -z "$GITHUB_USERNAME" ]; then
    echo -e "${RED}Failed to get GitHub username.${NC}"
    echo "Please make sure you're logged in correctly."
    exit 1
fi
echo -e "${GREEN}✓ GitHub username: $GITHUB_USERNAME${NC}"

# 1. Create Homebrew tap repository
echo ""
echo -e "${YELLOW}Setting up Homebrew tap repository...${NC}"
if gh repo view "$GITHUB_USERNAME/homebrew-gitcury" &> /dev/null; then
    echo -e "${GREEN}✓ Homebrew tap repository already exists${NC}"
else
    echo "Creating Homebrew tap repository..."
    mkdir -p /tmp/homebrew-gitcury/Formula
    cd /tmp/homebrew-gitcury
    echo "# GitCury Homebrew Tap" > README.md
    echo "" >> README.md
    echo "Homebrew tap for [GitCury](https://github.com/$GITHUB_USERNAME/GitCury) - AI-Powered Git Automation CLI tool." >> README.md
    echo "" >> README.md
    echo "## Usage" >> README.md
    echo "" >> README.md
    echo "\`\`\`bash" >> README.md
    echo "brew tap $GITHUB_USERNAME/gitcury" >> README.md
    echo "brew install gitcury" >> README.md
    echo "\`\`\`" >> README.md
    
    git init
    git add README.md
    git commit -m "Initial commit"
    
    gh repo create "$GITHUB_USERNAME/homebrew-gitcury" --public --description "Homebrew tap for GitCury" --source=. --push
    echo -e "${GREEN}✓ Homebrew tap repository created${NC}"
fi

# 2. Create Scoop bucket repository
echo ""
echo -e "${YELLOW}Setting up Scoop bucket repository...${NC}"
if gh repo view "$GITHUB_USERNAME/GitCury-Scoop-Bucket" &> /dev/null; then
    echo -e "${GREEN}✓ Scoop bucket repository already exists${NC}"
else
    echo "Creating Scoop bucket repository..."
    mkdir -p /tmp/GitCury-Scoop-Bucket
    cd /tmp/GitCury-Scoop-Bucket
    echo "# GitCury Scoop Bucket" > README.md
    echo "" >> README.md
    echo "Scoop bucket for [GitCury](https://github.com/$GITHUB_USERNAME/GitCury) - AI-Powered Git Automation CLI tool." >> README.md
    echo "" >> README.md
    echo "## Usage" >> README.md
    echo "" >> README.md
    echo "\`\`\`powershell" >> README.md
    echo "scoop bucket add gitcury https://github.com/$GITHUB_USERNAME/GitCury-Scoop-Bucket.git" >> README.md
    echo "scoop install gitcury" >> README.md
    echo "\`\`\`" >> README.md
    
    git init
    git add README.md
    git commit -m "Initial commit"
    
    gh repo create "$GITHUB_USERNAME/GitCury-Scoop-Bucket" --public --description "Scoop bucket for GitCury" --source=. --push
    echo -e "${GREEN}✓ Scoop bucket repository created${NC}"
fi

# 3. Check Docker Hub repository
echo ""
echo -e "${YELLOW}Checking Docker Hub repository...${NC}"
echo -e "${BLUE}Note: You'll need to manually create a Docker Hub repository if it doesn't exist.${NC}"
echo "1. Go to https://hub.docker.com/repositories"
echo "2. Click 'Create Repository'"
echo "3. Name it 'gitcury' and set it to public"
echo ""
echo -e "${YELLOW}Would you like to open Docker Hub now? (y/n)${NC}"
read -r open_docker
if [[ "$open_docker" == "y" || "$open_docker" == "Y" ]]; then
    echo "Opening Docker Hub in your browser..."
    if command -v xdg-open &> /dev/null; then
        xdg-open "https://hub.docker.com/repositories" &> /dev/null
    elif command -v open &> /dev/null; then
        open "https://hub.docker.com/repositories" &> /dev/null
    else
        echo -e "${YELLOW}Could not open browser automatically. Please visit:${NC}"
        echo "https://hub.docker.com/repositories"
    fi
fi

# 4. Set up GitHub Secrets
echo ""
echo -e "${YELLOW}Setting up GitHub Secrets...${NC}"
echo "You'll need to create the following secrets in your GitCury repository:"
echo "1. HOMEBREW_TAP_PAT - GitHub Personal Access Token with 'repo' scope for homebrew tap"
echo "2. SCOOP_BUCKET_PAT - GitHub Personal Access Token with 'repo' scope for scoop bucket"
echo "3. DOCKERHUB_USERNAME - Your Docker Hub username"
echo "4. DOCKERHUB_TOKEN - Your Docker Hub access token"
echo ""
echo -e "${YELLOW}Would you like to set up these secrets now? (y/n)${NC}"
read -r setup_secrets
if [[ "$setup_secrets" == "y" || "$setup_secrets" == "Y" ]]; then
    echo -e "${BLUE}Creating GitHub Personal Access Token...${NC}"
    echo "1. Go to https://github.com/settings/tokens/new"
    echo "2. Enter a note like 'GitCury Release Automation'"
    echo "3. Select repo scope"
    echo "4. Click 'Generate token'"
    echo "5. Copy the token (you'll need it twice)"
    echo ""
    echo -e "${YELLOW}Would you like to open GitHub token page now? (y/n)${NC}"
    read -r open_github
    if [[ "$open_github" == "y" || "$open_github" == "Y" ]]; then
        echo "Opening GitHub token page in your browser..."
        if command -v xdg-open &> /dev/null; then
            xdg-open "https://github.com/settings/tokens/new" &> /dev/null
        elif command -v open &> /dev/null; then
            open "https://github.com/settings/tokens/new" &> /dev/null
        else
            echo -e "${YELLOW}Could not open browser automatically. Please visit:${NC}"
            echo "https://github.com/settings/tokens/new"
        fi
    fi
    
    echo ""
    echo -e "${YELLOW}Enter the generated GitHub token:${NC}"
    read -r -s github_token
    
    if [ -n "$github_token" ]; then
        echo "Setting HOMEBREW_TAP_PAT secret..."
        gh secret set HOMEBREW_TAP_PAT -b "$github_token" -R "$GITHUB_USERNAME/GitCury"
        
        echo "Setting SCOOP_BUCKET_PAT secret..."
        gh secret set SCOOP_BUCKET_PAT -b "$github_token" -R "$GITHUB_USERNAME/GitCury"
        
        echo -e "${GREEN}✓ GitHub token secrets set${NC}"
    else
        echo -e "${RED}No token provided. Skipping GitHub token secrets.${NC}"
    fi
    
    echo ""
    echo -e "${YELLOW}Enter your Docker Hub username:${NC}"
    read -r dockerhub_username
    
    if [ -n "$dockerhub_username" ]; then
        echo "Setting DOCKERHUB_USERNAME secret..."
        gh secret set DOCKERHUB_USERNAME -b "$dockerhub_username" -R "$GITHUB_USERNAME/GitCury"
        
        echo -e "${YELLOW}Enter your Docker Hub access token:${NC}"
        echo "(Generate this at https://hub.docker.com/settings/security)"
        read -r -s dockerhub_token
        
        if [ -n "$dockerhub_token" ]; then
            echo "Setting DOCKERHUB_TOKEN secret..."
            gh secret set DOCKERHUB_TOKEN -b "$dockerhub_token" -R "$GITHUB_USERNAME/GitCury"
            echo -e "${GREEN}✓ Docker Hub secrets set${NC}"
        else
            echo -e "${RED}No Docker Hub token provided. Skipping Docker Hub token secret.${NC}"
        fi
    else
        echo -e "${RED}No Docker Hub username provided. Skipping Docker Hub secrets.${NC}"
    fi
fi

# 5. Verify CI/CD workflow files
echo ""
echo -e "${YELLOW}Verifying CI/CD workflow files...${NC}"
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo ".")
if [ -f "$REPO_ROOT/.github/workflows/release.yml" ] && \
   [ -f "$REPO_ROOT/.github/workflows/pr.yml" ] && \
   [ -f "$REPO_ROOT/.github/workflows/auto-tag.yml" ] && \
   [ -f "$REPO_ROOT/.goreleaser.yml" ]; then
    echo -e "${GREEN}✓ All required workflow files exist${NC}"
else
    echo -e "${RED}Some workflow files are missing. Please ensure you have:${NC}"
    echo "- .github/workflows/release.yml"
    echo "- .github/workflows/pr.yml"
    echo "- .github/workflows/auto-tag.yml"
    echo "- .goreleaser.yml"
fi

echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Create an initial release manually: 'git tag v0.1.0 && git push origin v0.1.0'"
echo "2. Check that the automated release workflow runs successfully"
echo "3. Verify your distribution channels after the release completes"
echo ""
