# GitCury Release & Versioning Strategy

This document outlines how GitCury handles versioning, releases, and deployment.

## Semantic Versioning

GitCury follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** version increases when we make incompatible API changes
- **MINOR** version increases when we add functionality in a backward-compatible manner
- **PATCH** version increases when we make backward-compatible bug fixes

## Automated Release Process

Releases are automatically managed through our CI/CD pipeline:

1. **Pull Requests** are automatically tested, linted, and validated against our release configuration
2. **Merged PRs** to the main branch trigger automatic version bumping based on commit messages
3. **Version tags** (e.g., v1.2.3) trigger the full release process

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` - Adds a new feature (MINOR version bump)
- `fix:` - Fixes a bug (PATCH version bump)
- `docs:` - Documentation changes (no version bump)
- `style:` - Code style changes (no version bump)
- `refactor:` - Code refactoring without functionality changes (PATCH version bump)
- `perf:` - Performance improvements (PATCH version bump)
- `test:` - Adding/updating tests (no version bump)
- `chore:` - Build process or tooling changes (no version bump)

A commit with `BREAKING CHANGE:` in the footer or with `!` after the type (e.g., `feat!:`) triggers a MAJOR version bump.

## Distribution Channels

GitCury is distributed through multiple channels:

1. **GitHub Releases** - Direct binary downloads for all supported platforms
2. **Homebrew** - For macOS and Linux users
3. **Scoop** - For Windows users
4. **Docker Hub** - Containerized version
5. **Go Install** - For Go developers

## Release Artifacts

Each release includes:

- Compiled binaries for Windows, macOS, and Linux (ARM64 and AMD64)
- Checksums for security verification
- Docker images tagged with version and latest
- Updated package manager formulas

## Manual Release (if needed)

For manual releases (rarely needed):

```bash
# 1. Update version in relevant files if necessary
# 2. Create and push a tag
git tag v1.2.3
git push origin v1.2.3
```

## Maintenance Policy

- We maintain the latest MAJOR version with security and bug fixes
- Older MAJOR versions receive critical security fixes for 6 months after a new MAJOR release
