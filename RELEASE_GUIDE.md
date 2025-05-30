# GitCury Release Guide

This document provides comprehensive instructions for releasing new versions of GitCury.

## Quick Release

For a standard release, use the manual release script:

```bash
# Test release (dry run)
./manual-release.sh v1.4.0 --dry-run

# Actual release
./manual-release.sh v1.4.0
```

## Manual Release Script

The `manual-release.sh` script provides a user-friendly way to create releases with proper validation and error handling.

### Features

- ✅ Version format validation (semantic versioning)
- ✅ Git repository status checks
- ✅ GoReleaser configuration validation
- ✅ Dry-run support for testing
- ✅ Force option for overwriting existing tags
- ✅ Colored output and progress indicators
- ✅ Environment variable validation
- ✅ Post-release checklist

### Usage Examples

```bash
# Standard release
./manual-release.sh v1.4.0

# Test release without publishing
./manual-release.sh v1.4.0 --dry-run

# Force release (overwrite existing tag)
./manual-release.sh v1.4.0 --force

# Beta release
./manual-release.sh v2.0.0-beta.1

# Release candidate
./manual-release.sh v1.4.0-rc.1

# Show help
./manual-release.sh --help
```

## Traditional GitHub Workflow Release

GitCury also supports automatic releases via GitHub Actions when tags are pushed:

```bash
# Create and push a tag
git tag -a v1.4.0 -m "Release v1.4.0"
git push origin v1.4.0
```

This triggers the `.github/workflows/release.yml` workflow automatically.

## Pre-Release Checklist

Before creating a release, ensure:

1. **Code Quality**
   - [ ] All tests pass: `go test ./...`
   - [ ] No linting errors: `golangci-lint run`
   - [ ] Code coverage is acceptable
   - [ ] Build succeeds: `go build ./...`

2. **Documentation**
   - [ ] README.md is up to date
   - [ ] CHANGELOG.md contains new version
   - [ ] API documentation is current
   - [ ] Version compatibility notes are added

3. **Configuration**
   - [ ] GoReleaser config is valid: `goreleaser check`
   - [ ] `config.json.example` includes all new options
   - [ ] Docker configuration is updated if needed

4. **Version Management**
   - [ ] Version follows semantic versioning
   - [ ] Breaking changes are documented
   - [ ] Migration guide exists for major versions

## Environment Setup

### Required Tools

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Verify installation
goreleaser --version
```

### Required Environment Variables

For actual releases (not dry-run):

```bash
export GITHUB_TOKEN="your_github_token"
export DOCKERHUB_USERNAME="your_dockerhub_username"  # Optional
export DOCKERHUB_TOKEN="your_dockerhub_token"        # Optional
```

## Release Artifacts

Each release creates the following artifacts:

### Binaries
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

### Archives
- `.tar.gz` for Linux/macOS
- `.zip` for Windows
- Source code archive

### Package Managers
- Homebrew formula (macOS/Linux)
- Scoop manifest (Windows)

### Container Images
- Docker images for multiple architectures
- Published to GitHub Container Registry

## Version Numbering

GitCury follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** version: incompatible API changes
- **MINOR** version: backwards-compatible functionality additions
- **PATCH** version: backwards-compatible bug fixes

### Examples

```
v1.0.0      # Initial release
v1.1.0      # New feature, backward compatible
v1.1.1      # Bug fix, backward compatible
v2.0.0      # Breaking changes
v2.0.0-beta.1   # Beta release
v2.0.0-rc.1     # Release candidate
```

## Troubleshooting

### Common Issues

1. **GoReleaser validation fails**
   ```bash
   goreleaser check
   ```

2. **Missing environment variables**
   ```bash
   echo $GITHUB_TOKEN
   ```

3. **Tag already exists**
   ```bash
   # Delete local tag
   git tag -d v1.4.0
   
   # Delete remote tag
   git push origin :refs/tags/v1.4.0
   ```

4. **Working directory not clean**
   ```bash
   git status
   git stash  # or commit changes
   ```

### Manual Cleanup

If a release fails partway through:

```bash
# Delete the tag locally and remotely
git tag -d v1.4.0
git push origin :refs/tags/v1.4.0

# Clean GoReleaser artifacts
rm -rf dist/

# Restart the release
./manual-release.sh v1.4.0
```

## Binary Installation Verification

After release, verify that users can install the binary correctly:

```bash
# Test installation
go install github.com/lakshyajain-0291/gitcury@latest

# Verify binary name (should be lowercase 'gitcury')
which gitcury
gitcury --version
```

## GitHub Release Page

After release, the GitHub release page should contain:

- ✅ Release notes from CHANGELOG.md
- ✅ Binary downloads for all platforms
- ✅ Source code archives
- ✅ Docker installation instructions
- ✅ Package manager installation instructions

Check: https://github.com/lakshyajain-0291/gitcury/releases

## Monitoring

After release, monitor:

1. **Download Statistics**: GitHub release page
2. **Installation Issues**: GitHub issues
3. **Docker Pulls**: Docker Hub/GitHub Container Registry
4. **Package Manager Stats**: Homebrew/Scoop analytics

## Rolling Back

To roll back a release:

1. **Hide the GitHub release** (don't delete to preserve download stats)
2. **Create a patch release** with fixes
3. **Update package managers** if necessary
4. **Communicate the issue** in release notes

## Next Steps

After a successful release:

1. Update the main branch if needed
2. Create a new milestone for the next version
3. Update project roadmap
4. Announce the release on relevant channels
5. Monitor for user feedback and issues

---

**Note**: This guide assumes you have appropriate permissions to create releases on the GitCury repository and access to required secrets/tokens.
