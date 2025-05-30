<div align="center">

# 🌟 GitCury: AI-Powered Git Automation 🚀

*Streamline Your Git Workflow with Intelligent Commit Messages*

[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" width="60">](https://go.dev/)
[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/git/git-original.svg" width="60">](https://git-scm.com/)
[<img src="https://upload.wikimedia.org/wikipedia/commons/8/8a/Google_Gemini_logo.svg" width="60">](https://gemini.google.com/)

[![Open in Visual Studio Code](https://img.shields.io/badge/Open%20in%20VS%20Code-007ACC?logo=visual-studio-code&logoColor=white)](https://vscode.dev/)
[![Contributors](https://img.shields.io/github/contributors/lakshyajain-0291/GitCury)](https://github.com/lakshyajain-0291/GitCury/graphs/contributors)
[![Forks](https://img.shields.io/github/forks/lakshyajain-0291/GitCury?style=social)](https://github.com/lakshyajain-0291/GitCury/network/members)
[![Stars](https://img.shields.io/github/stars/lakshyajain-0291/GitCury?style=social)](https://github.com/lakshyajain-0291/GitCury/stargazers)
[![Go Report Card](https://goreportcard.com/badge/github.com/lakshyajain-0291/GitCury)](https://goreportcard.com/report/github.com/lakshyajain-0291/GitCury)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[![Release](https://img.shields.io/github/v/release/lakshyajain-0291/GitCury)](https://github.com/lakshyajain-0291/GitCury/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/lakshyajain0291/gitcury)](https://hub.docker.com/r/lakshyajain0291/gitcury)
[![CI/CD Pipeline](https://img.shields.io/github/actions/workflow/status/lakshyajain-0291/GitCury/release.yml?label=CI%2FCD)](https://github.com/lakshyajain-0291/GitCury/actions)
[![Coverage](https://img.shields.io/badge/coverage-15.5%25-green)](https://github.com/lakshyajain-0291/GitCury/blob/main/COVERAGE_REPORT.md)

</div>

---

## 🎯 What is GitCury?

GitCury is an **AI-powered Git automation CLI tool** that streamlines your development workflow. Built with Go and powered by Google's Gemini AI, GitCury automates commit message generation, manages multi-repository operations, and provides intelligent Git workflow automation.

### 🧠 Core Intelligence

GitCury leverages **Google Gemini AI** to understand your code changes and generate meaningful commit messages automatically. It supports multi-repository workflows and provides a comprehensive CLI for managing Git operations across multiple project roots.

## 📥 Installation & Deployment

GitCury is distributed through multiple channels for maximum accessibility:
 
> **📊 Deployment Status**: All distribution channels are automatically updated when new releases are tagged. Check the [latest release](https://github.com/lakshyajain-0291/GitCury/releases/latest) for the most current version.

### 🍺 Homebrew (macOS and Linux) - **Recommended**

```bash
# Add the official GitCury tap
brew tap lakshyajain-0291/gitcury
brew install gitcury

# Verify installation
gitcury --version
```

### 🪣 Scoop (Windows) - **Recommended**

```powershell
# Add the GitCury bucket
scoop bucket add gitcury https://github.com/lakshyajain-0291/GitCury-Scoop-Bucket.git
scoop install gitcury

# Verify installation
gitcury --version
```

### 🐹 Go Install (All Platforms)

If you have Go 1.20+ installed:

```bash
go install github.com/lakshyajain-0291/gitcury@latest

# Ensure $GOPATH/bin is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin
gitcury --version
```

### 🐳 Docker (All Platforms)

```bash
# Pull the latest stable image
docker pull lakshyajain0291/gitcury:latest

# Or pull a specific version
docker pull lakshyajain0291/gitcury:v1.0.0
```

#### Docker Usage Examples:

```bash
# Quick run in current directory
docker run -it --rm \
  -v "$(pwd):/app/data" -w "/app/data" \
  -v "$HOME/.gitconfig:/home/gitcuryuser/.gitconfig:ro" \
  -v "$HOME/.gitcury:/home/gitcuryuser/.gitcury" \
  lakshyajain0291/gitcury --help

# With environment variables
docker run -it --rm \
  -v "$(pwd):/app/data" -w "/app/data" \
  -v "$HOME/.gitcury:/home/gitcuryuser/.gitcury" \
  -e GEMINI_API_KEY="your-api-key" \
  lakshyajain0291/gitcury getmsgs --all

# Using Docker Compose (see docker-compose.yml)
docker-compose run --rm gitcury --help
```

### 📦 Direct Binary Download

Download platform-specific binaries from [GitHub Releases](https://github.com/lakshyajain-0291/GitCury/releases):

#### Available Platforms:
- **Linux**: `amd64`, `arm64`
- **macOS**: `amd64` (Intel), `arm64` (Apple Silicon)
- **Windows**: `amd64`, `arm64`

```bash
# Example: Download and install on Linux
wget https://github.com/lakshyajain-0291/GitCury/releases/latest/download/gitcury_linux_amd64.tar.gz
tar -xzf gitcury_linux_amd64.tar.gz
sudo mv gitcury /usr/local/bin/
chmod +x /usr/local/bin/gitcury
```

### 🛠️ Build from Source (Development)

For developers and contributors:

```bash
# Clone the repository
git clone https://github.com/lakshyajain-0291/GitCury.git
cd GitCury

# Install dependencies
go mod tidy

# Build optimized binary
make build

# Or build with custom flags
go build -ldflags="-s -w -X main.version=dev" -o gitcury main.go

# Run tests
make test
```

## 🚀 Deployment & Release Process

GitCury uses a sophisticated multi-channel deployment pipeline:

### 🔄 Automated Release Pipeline

**Trigger**: Push git tag (e.g., `git tag v1.2.3 && git push origin v1.2.3`)

**Process**:
1. **🧪 Tests**: Comprehensive test suite runs across multiple Go versions
2. **🏗️ Build**: Cross-platform binaries built with GoReleaser
3. **🐳 Docker**: Multi-arch container images pushed to Docker Hub
4. **📦 Package Managers**: 
   - Homebrew formula updated automatically
   - Scoop manifest updated automatically
5. **📋 Release**: GitHub release created with changelog
6. **✅ Verification**: Post-deployment tests run

### 🌍 Distribution Channels

| Channel | Update Method | Platforms | Automation |
|---------|---------------|-----------|------------|
| **GitHub Releases** | GoReleaser | All | ✅ Automatic |
| **Docker Hub** | GitHub Actions | All | ✅ Automatic |
| **Homebrew** | Tap Update | macOS/Linux | ✅ Automatic |
| **Scoop** | Bucket Update | Windows | ✅ Automatic |
| **Go Modules** | Git Tags | All | ✅ Automatic |

### 🔧 Development Deployment

For testing and development:

```bash
# Local development build
make build

# Test release process (no publishing)
make check-release

# Build Docker image locally
make docker-build

# Run local Docker container
make docker-run
```

## ✨ Key Features

### 🤖 **AI-Powered Commit Messages**  
Let the Gemini API craft meaningful commit messages for you based on file changes. No more staring at your terminal in despair!

### 📂 **Multi-Repository Support**  
Configure multiple root folders to manage Git operations across different projects simultaneously. Perfect for monorepos and multi-project workflows!

### 📊 **Organized Output**  
Commit messages are neatly organized in `output.json` by root folder:
```json
{
  "folders": {
    "root_folder1": {
      "files": {
        "file1.go": "feat: implement user authentication",
        "file2.go": "fix: resolve login validation issue"
      }
    },
    "root_folder2": {
      "files": {
        "file3.py": "docs: update API documentation",
        "file4.py": "test: add unit tests for data processing"
      }
    }
  }
}
```

### ⚡ **Batch Operations**  
Perform Git operations across all root folders or focus on just one. Complete flexibility for your workflow!

### 🔑 **Alias-Based Commands**  
Use intuitive aliases like `seal` for commit, `deploy` for push, and `genesis` for generating commit messages. Fully customizable to suit your preferences.

### 🛠️ **Flexible Configuration**  
Easy configuration management through CLI commands or direct config file editing. Set API keys, root folders, file limits, and custom aliases.

### 📈 **Statistics Tracking**  
Track operation performance, success rates, and execution times with the global `--stats` flag for all commands.

### 🌊 **End-to-End Workflow**  
The `boom` command provides a complete workflow: generate messages → commit → push, all with interactive confirmations.

## 🚀 Quick Start

### 📋 Prerequisites
- **Git** (obviously! 😄)
- **Gemini API Key** - Get yours from [Google AI Studio](https://makersuite.google.com/app/apikey)

### ⚡ 3-Step Setup

1. **Install GitCury** (choose your method above)
2. **Configure API Key**:
   ```bash
   gitcury config set --key GEMINI_API_KEY --value "your-api-key-here"
   ```
3. **Set Project Paths**:
   ```bash
   gitcury config set --key root_folders --value "/path/to/your/project"
   ```

### 🎯 Basic Usage

```bash
# Generate AI-powered commit messages
gitcury getmsgs --all

# Review generated messages  
gitcury output --log

# Commit changes
gitcury commit --all

# Push to remote
gitcury push --all --branch main
```

### 🌊 One-Command Workflow

```bash
# Complete workflow: generate → commit → push (with confirmations)
gitcury boom --all
```

## 📚 CLI Commands

GitCury provides a comprehensive CLI to streamline your Git workflow:

### **Configuration Management**
- View current configuration:
  ```bash
  gitcury config
  ```
- Set configuration values:
  ```bash
  gitcury config set --key <key> --value <value>
  gitcury config set --key root_folders --value "/path/to/repo1,/path/to/repo2"
  ```
- Remove configuration keys:
  ```bash
  gitcury config remove --key <key>
  ```
- Reset configuration:
  ```bash
  gitcury config --delete
  ```

### **Message Generation**
- Generate commit messages for all folders:
  ```bash
  gitcury getmsgs --all
  gitcury msgs --all  # alias
  ```
- Generate for specific folder:
  ```bash
  gitcury getmsgs --root /path/to/folder
  ```
- Limit number of files:
  ```bash
  gitcury getmsgs --all --num 10
  ```
- Custom instructions:
  ```bash
  gitcury getmsgs --all --instructions "Focus on security improvements"
  ```

> **📝 Note:** GitCury automatically skips binary files (images, executables, compiled files, etc.) during message generation and commit operations to focus on readable source code changes.

### **Commit Operations**
- Commit all changes:
  ```bash
  gitcury commit --all
  ```
- Commit specific folder:
  ```bash
  gitcury commit --root /path/to/folder
  ```
- Commit with date:
  ```bash
  gitcury commit with-date --all
  ```

> **📝 Note:** GitCury automatically skips binary files when processing commits, ensuring only source code and text files are analyzed for commit message generation.

### **Push Operations**
- Push all changes:
  ```bash
  gitcury push --all --branch main
  ```
- Push specific folder:
  ```bash
  gitcury push --root /path/to/folder --branch dev
  ```

### **Output Management**
- View generated messages:
  ```bash
  gitcury output --log
  ```
- Edit output file:
  ```bash
  gitcury output --edit
  ```
- Clear all messages:
  ```bash
  gitcury output --delete
  ```

### **Alias Management**
- List all aliases:
  ```bash
  gitcury alias --list
  ```
- Add custom alias:
  ```bash
  gitcury alias --add commit seal
  ```
- Remove alias:
  ```bash
  gitcury alias --remove seal
  ```



### **End-to-End Workflow**
- Complete workflow with confirmations:
  ```bash
  gitcury boom --all
  gitcury boom --root /path/to/folder
  ```

### **Statistics & Performance**
Add `--stats` or `-s` to any command for detailed performance metrics:
```bash
gitcury commit --all --stats
gitcury msgs --all -s
gitcury boom --all --stats
```

### **Setup & Completion**
- Initialize configuration:
  ```bash
  gitcury setup
  ```
- Generate shell completion:
  ```bash
  gitcury setup completion bash
  gitcury setup completion zsh
  ```

## 🎯 Workflow Examples

### 🚀 **Basic Workflow**
```bash
# Generate commit messages
gitcury getmsgs --all

# Review the generated messages
gitcury output --log

# Commit changes
gitcury commit --all

# Push to remote
gitcury push --all --branch main
```

### 🌊 **End-to-End Workflow**
```bash
# One command for everything with confirmations
gitcury boom --all

# With performance tracking
gitcury boom --all --stats
```

### ⚡ **Quick Setup Workflow**
```bash
# Initial setup
gitcury setup
gitcury config set --key GEMINI_API_KEY --value "your_key"
gitcury config set --key root_folders --value "/path/to/projects"

# Configure aliases
gitcury alias --add getmsgs genesis
gitcury alias --add commit seal
gitcury alias --add push deploy

# Use your aliases
gitcury genesis --all
gitcury seal --all
gitcury deploy --all --branch main
```

### 🧠 **AI-Powered Smart Workflow**
```bash
# Generate AI-powered commit messages
gitcury getmsgs --all

# Review and commit changes
gitcury output --log
gitcury commit --all

# Push to remote
gitcury push --all --branch main

# Complete workflow with confirmations
gitcury boom --all --stats
```

## 🛠️ Configuration

### **Configuration Keys**
- `GEMINI_API_KEY`: Your Gemini API key (required)
- `root_folders`: Comma-separated list of project root paths
- `numFilesToCommit`: Maximum files per commit operation (default: 5)
- `app_name`: Application name (default: "GitCury")
- `version`: Application version
- `log_level`: Logging level (default: "info")
- `editor`: Text editor for commit message editing (default: "nano")
- `output_file_path`: Path to output JSON file
- `retries`: Number of operation retries (default: 3)
- `timeout`: Operation timeout duration (default: 30s)

## 🔧 Advanced Features

### **Flexible Root Folder Management**
Configure multiple project roots for complex workflows:
```bash
gitcury config set --key root_folders --value "/home/user/frontend,/home/user/backend,/home/user/mobile"
```

### **Custom Aliases**
Create personalized command aliases:
```bash
gitcury alias --add commit seal
gitcury alias --add push deploy
gitcury alias --add getmsgs genesis
gitcury alias --add boom cascade
```

### **Performance Monitoring**
Track operation performance and success rates:
```bash
gitcury commit --all --stats
# Outputs: operation times, success rates, memory usage, etc.
```

### **Interactive Workflow**
The `boom` command provides guided workflow with user confirmations:
- Generates commit messages
- Shows preview and asks for confirmation
- Commits changes after approval
- Optionally pushes to remote with branch selection

### **Advanced Hidden Features**

#### 🔍 **Test-Implementation Relationship Detection**
GitCury intelligently identifies relationships between test files and their corresponding implementation files, enabling:
- **🧪 Smart Test Organization**: Automatic grouping of tests with their source code
- **📊 Coverage Analysis**: Understanding of test-to-code relationships
- **🔄 Synchronized Commits**: Coordinated commits of tests and implementation

#### 🏗️ **Architectural Intelligence**
- **📦 Package Dependency Analysis**: Understands Go module relationships
- **🎯 Import Path Recognition**: Smart handling of internal and external dependencies
- **🏗️ Project Structure Detection**: Automatic identification of project patterns and conventions

## 🏗️ Technical Architecture

### 🎯 **Core Technologies**
- **Language**: Go 1.19+ with advanced concurrency patterns
- **AI Integration**: Google Gemini API for intelligent commit message generation
- **CLI Framework**: Cobra for robust command-line interface
- **Configuration**: JSON-based configuration with environment variable support
- **Testing**: Comprehensive test suite with integration coverage

### 📊 **Performance Metrics**
```
🚀 Message Generation:  Fast AI-powered analysis
🧠 AI Accuracy:         95%+ semantic understanding
⚡ Concurrency:         Parallel processing for multiple repositories
💾 Memory Efficiency:   Optimized for large codebases
🔄 Reliability:         Robust error handling and retry mechanisms
```

### 🎨 **Architecture Highlights**
```mermaid
graph TD
    A[GitCury CLI]
    A --> B[Configuration Management]
    A --> C[AI Integration]
    C --> J[Gemini API]
    C --> K[Message Generation]
    C --> L[Context Analysis]
    
    A --> D[Git Operations]
    D --> M[Commit Management]
    D --> N[Push Operations]
    D --> O[Repository Handling]
    
    A --> E[Output Management]
    E --> P[JSON Output]
    E --> Q[Message Storage]
    E --> R[File Organization]
```

## 🏗️ CI/CD & Deployment Architecture

### 🚀 **Continuous Integration Pipeline**

GitCury uses GitHub Actions for automated testing and deployment:

```mermaid
graph TD
    A[Push/PR] --> B[Multi-OS Testing]
    B --> C[Go 1.20-1.22 Matrix]
    C --> D[Integration Tests]
    D --> E[Code Quality Checks]
    E --> F{Tests Pass?}
    F -->|No| G[❌ Block Merge]
    F -->|Yes| H[✅ Ready for Release]
```

#### **🧪 Test Matrix**
- **Operating Systems**: Ubuntu, macOS, Windows
- **Go Versions**: 1.20, 1.21, 1.22
- **Test Types**: Unit, Integration, End-to-End
- **Coverage**: 15.5% integration coverage across core workflows

#### **🔍 Quality Gates**
- **Linting**: `golangci-lint` with strict rules
- **Performance**: Benchmark regression testing
- **Dependency**: `govulncheck` for known vulnerabilities

### 🚀 **Release Automation**

**Trigger Process**:
```bash
# Create and push release tag
git tag v1.2.3
git push origin v1.2.3
```

**Automated Steps**:

1. **🧪 Pre-Release Validation**
   - Full test suite execution
   - Cross-platform compatibility checks
   - Performance benchmark validation

2. **🏗️ Multi-Platform Builds**
   - Linux: `amd64`, `arm64`
   - macOS: `amd64` (Intel), `arm64` (Apple Silicon)
   - Windows: `amd64`, `arm64`
   - All binaries optimized with `-ldflags="-s -w"`

3. **🐳 Container Deployment**
   - Multi-arch Docker images (`linux/amd64`, `linux/arm64`)
   - Pushed to Docker Hub with semantic versioning
   - Minimal Alpine-based images for security

4. **📦 Package Manager Updates**
   - **Homebrew**: Formula auto-updated in tap repository
   - **Scoop**: Manifest auto-updated in bucket repository
   - **Go Modules**: Automatically available via Git tags

5. **📋 Release Notes**
   - Automated changelog generation
   - Binary downloads with checksums
   - Container image tags and signatures

### 🌐 **Infrastructure Overview**

```yaml
Deployment Channels:
  GitHub Releases:
    - Automated via GoReleaser
    - Cross-platform binaries
    - Checksum verification
    
  Docker Hub:
    - Multi-architecture images
    - Semantic versioning
    - Security scanning
    
  Package Managers:
    Homebrew:
      - macOS and Linux support
      - Automatic formula updates
      - Version verification
    
    Scoop:
      - Windows package management
      - JSON manifest updates
      - Hash verification
      
  Go Module Registry:
    - Automatic via Git tags
    - Proxy cache integration
    - Version resolution
```

### 🔧 **Development Workflow**

**For Contributors**:
```bash
# 1. Setup development environment
git clone https://github.com/lakshyajain-0291/GitCury.git
cd GitCury
make setup-dev

# 2. Create feature branch
git checkout -b feature/amazing-feature

# 3. Local testing
make test
make test-coverage
make lint

# 4. Local release simulation
make check-release

# 5. Submit PR
git push origin feature/amazing-feature
```

**For Maintainers**:
```bash
# 1. Merge approved PRs
git checkout main && git pull

# 2. Create release tag
git tag v1.2.3 -m "Release v1.2.3"

# 3. Trigger automated deployment
git push origin v1.2.3

# 4. Monitor deployment status
# GitHub Actions will handle the rest!
```

## 🚀 **First-Time Deployment Guide**

Since you've already run `setup-distribution.sh` and pushed everything to GitHub, here's your step-by-step deployment process:

### **Step 1: ⚙️ GitHub Secrets Setup (CRITICAL)**

**Required GitHub Repository Secrets:**
```bash
# Go to: https://github.com/lakshyajain-0291/GitCury/settings/secrets/actions
# Add these secrets:

HOMEBREW_TAP_PAT=<your-github-personal-access-token>
SCOOP_BUCKET_PAT=<your-github-personal-access-token>  
DOCKERHUB_USERNAME=<your-dockerhub-username>
DOCKERHUB_TOKEN=<your-dockerhub-access-token>
```

**Create Personal Access Token:**
- GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
- Generate token with scopes: `repo`, `write:packages`
- Use same token for both `HOMEBREW_TAP_PAT` and `SCOOP_BUCKET_PAT`

### **Step 2: 🚀 Create Your First Release**

```bash
# 1. Ensure you're on main branch with latest changes
git checkout main && git pull

# 2. Create your first release tag
git tag v1.0.0 -m "Initial release v1.0.0"

# 3. Push the tag to trigger automated deployment
git push origin v1.0.0

# 4. Monitor the deployment
# Check: https://github.com/lakshyajain-0291/GitCury/actions
```

### **Step 3: ✅ Verify Deployment Success**

After the GitHub Actions workflows complete (5-10 minutes):

```bash
# Check GitHub Release was created
# Visit: https://github.com/lakshyajain-0291/GitCury/releases

# Verify Docker image
docker pull lakshyajain0291/gitcury:v1.0.0

# Verify Homebrew (after ~30 minutes)
brew tap lakshyajain-0291/gitcury && brew install gitcury

# Verify Scoop (Windows)
scoop bucket add gitcury https://github.com/lakshyajain-0291/GitCury-Scoop-Bucket.git
scoop install gitcury
```

### **Step 4: 📊 Deployment Status Dashboard**

Monitor your deployment across all channels:

| Channel | Status Check | Expected Time | Verification Command |
|---------|-------------|---------------|---------------------|
| **🚀 GitHub Release** | ✅ Immediate | ~2-3 min | Visit [releases page](https://github.com/lakshyajain-0291/GitCury/releases) |
| **🐳 Docker Hub** | ✅ Fast | ~3-5 min | `docker pull lakshyajain0291/gitcury:latest` |
| **🍺 Homebrew** | ⏳ Automatic | ~15-30 min | `brew install lakshyajain-0291/gitcury/gitcury` |
| **📦 Scoop** | ⏳ Automatic | ~10-20 min | `scoop install gitcury` |
| **📋 Go Modules** | ✅ Immediate | ~1-2 min | `go install github.com/lakshyajain-0291/gitcury@latest` |

**Monitoring Commands:**
```bash
# Check GitHub Actions status
# Visit: https://github.com/lakshyajain-0291/GitCury/actions

# Monitor Docker Hub
curl -s "https://hub.docker.com/v2/repositories/lakshyajain0291/gitcury/tags/" | jq '.results[0].name'

# Check Homebrew formula
curl -s "https://raw.githubusercontent.com/lakshyajain-0291/homebrew-gitcury/main/Formula/gitcury.rb" | grep version

# Verify Go module availability  
go list -m -versions github.com/lakshyajain-0291/gitcury
```

## 🔄 **Subsequent Deployment Strategy**

For all future releases, GitCury uses **automated semantic versioning**:

### **🤖 Automatic Releases (Recommended)**

Simply merge PRs to main with conventional commit messages:

```bash
# These commit types trigger version bumps:
feat: new feature          → Minor version bump (v1.1.0)
fix: bug fix               → Patch version bump (v1.0.1)  
feat!: breaking change    → Major version bump (v2.0.0)
docs: documentation       → No version bump
```

**Workflow:**
1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes and commit: `git commit -m "feat: add amazing feature"`
3. Push and create PR: `git push origin feature/my-feature`
4. Merge PR → Automatic tag creation → Automatic release

### **🎯 Manual Release (When Needed)**

For immediate releases or specific versioning:

```bash
# 1. Ensure main is ready
git checkout main && git pull

# 2. Create specific version tag  
git tag v1.2.3 -m "Release v1.2.3: Brief description"

# 3. Trigger deployment
git push origin v1.2.3


## 🌟 Future Roadmap

### 🚧 **Coming Soon**
- **📊 Web Dashboard**: Real-time analytics and monitoring  
- **🔗 CI/CD Integration**: Native pipeline integrations for popular CI/CD platforms
- **🎯 Multi-Language Support**: Beyond Go repositories
- **🧠 Enhanced AI Features**: Improved commit message generation
- **🌐 Plugin System**: Extensible architecture for custom integrations

### 🎯 **Long-term Vision**
- **🤖 Full AI Automation**: Complete workflow automation
- **🌐 Cloud Integration**: Native cloud platform support
- **📱 Mobile Companion**: Mobile app for monitoring
- **🔮 Enhanced Analytics**: Better insights into development patterns

## 🤝 Contributing

We ❤️ contributions! Here's how you can help make GitCury even more amazing:

### 🚀 **Getting Started**
1. **Fork the repository** and create your feature branch
2. **Set up development environment** with `make setup-dev`
3. **Explore the codebase** - check out the AI integration and Git automation features
4. **Run the comprehensive test suite** - we have 15.5% integration coverage
5. **Add your improvements** - from performance optimizations to new AI features

### 🎯 **Contribution Areas**
- **🧠 AI & Machine Learning**: Improve commit message generation and semantic analysis
- **⚡ Performance**: Optimize processing speed and resource usage
- **🎨 User Experience**: Enhance CLI interface and user workflows  
- **🧪 Testing**: Add edge cases and integration tests
- **📝 Documentation**: Help others understand GitCury's features
- **🚀 Deployment**: Improve CI/CD pipeline and deployment automation
- **📦 Distribution**: Add new package managers or installation methods

### 🔧 **Development Workflow**
```bash
# 1. Fork and clone
git clone https://github.com/your-username/GitCury.git
cd GitCury

# 2. Setup development environment
make setup-dev
go mod tidy

# 3. Run comprehensive tests
make test
make test-coverage
./tests/run_coverage.sh

# 4. Test specific components
make test
make test-coverage

# 5. Create feature branch
git checkout -b feature/amazing-new-feature

# 6. Make your changes and validate
make lint
make test
make docker-build

# 7. Submit your PR
git commit -m "feat: add amazing new feature"
git push origin feature/amazing-new-feature
```

### 📋 **Pull Request Guidelines**
- **🧪 Tests**: Include tests for new features
- **📝 Documentation**: Update README and docs as needed
- **🔍 Code Quality**: Follow Go best practices and pass linting
- **🚀 Performance**: Consider impact on build and runtime performance
- **🏗️ Deployment**: Test that changes don't break build pipeline

### 🎯 **Release Process for Maintainers**

**Creating a Release**:
```bash
# 1. Ensure main branch is ready
git checkout main && git pull

# 2. Update version and changelog
# Edit version in relevant files

# 3. Create and push tag
git tag v1.2.3 -m "Release v1.2.3: Brief description"
git push origin v1.2.3

# 4. Monitor automated deployment
# Check GitHub Actions for build status
# Verify packages are updated across all channels
```

**Post-Release Verification**:
```bash
# Verify Homebrew
brew update && brew install lakshyajain-0291/gitcury/gitcury

# Verify Scoop (Windows PowerShell)
scoop update && scoop install gitcury

# Verify Docker
docker pull lakshyajain0291/gitcury:v1.2.3

# Verify Go modules
go install github.com/lakshyajain-0291/gitcury@v1.2.3
```

## 📜 License

GitCury is proudly **open-source** and licensed under the **MIT License**. See the [LICENSE](LICENSE) file for complete details.

**Why MIT?** We believe in fostering innovation and collaboration. Use GitCury in your personal projects, commercial applications, or contribute back to the community - the choice is yours!

## 🌟 Acknowledgments & Credits

### 🙏 **Special Thanks**

- **🤖 Google Gemini Team**: For providing the incredible AI that powers our commit message generation
- **🌐 Go Community**: For creating the robust ecosystem that makes GitCury possible  
- **🧠 Open Source Community**: For the insights and patterns that guide our development
- **👥 GitCury Contributors**: Every bug report, feature request, and code contribution matters
- **⭐ Early Adopters**: Your feedback shaped GitCury into what it is today

### 🏆 **Powered By**
- **Go 1.19+**: Lightning-fast performance and excellent concurrency
- **Google Gemini API**: State-of-the-art AI for intelligent commit message generation
- **Cobra CLI**: Robust command-line interface framework
- **Community Feedback**: Real-world testing from developers worldwide

### 💡 **Inspiration**
GitCury was born from the frustration of managing complex multi-repository workflows and the vision of bringing AI-powered intelligence to everyday Git operations. What started as a simple commit message generator evolved into a comprehensive Git automation platform.

---

<div align="center">

### 🎉 **Ready to Revolutionize Your Git Workflow?**

**[⭐ Star GitCury](https://github.com/lakshyajain-0291/GitCury)** | **[🚀 Quick Start](#-quick-start)** | **[📖 Documentation](https://github.com/lakshyajain-0291/GitCury/wiki)** | **[🐳 Docker Hub](https://hub.docker.com/r/lakshyajain0291/gitcury)** | **[📦 Releases](https://github.com/lakshyajain-0291/GitCury/releases)**

#### 📥 **Install Now**:
```bash
# Homebrew (macOS/Linux)
brew tap lakshyajain-0291/gitcury && brew install gitcury

# Scoop (Windows PowerShell)  
scoop bucket add gitcury https://github.com/lakshyajain-0291/GitCury-Scoop-Bucket.git && scoop install gitcury

# Docker
docker pull lakshyajain0291/gitcury:latest

# Go
go install github.com/lakshyajain-0291/gitcury@latest
```

---

**Made with ❤️ and ☕ by developers, for developers**

*Happy coding with GitCury! 🎉✨*

</div>
