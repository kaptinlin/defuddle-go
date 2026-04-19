# Release Process

This document outlines the release process for Defuddle Go, including both the Go library and CLI tool.

## Overview

Defuddle Go uses automated releases through GitHub Actions and GoReleaser. The process is triggered by pushing a version tag to the repository.

## Release Workflow

### 1. Automated Release Pipeline

The release process consists of:

1. **Testing**: Runs comprehensive tests on Go 1.25
2. **Building**: Cross-compiles binaries for multiple platforms (Linux, macOS, Windows)
3. **Packaging**: Creates archives with appropriate formats for each platform
4. **Publishing**: Releases to GitHub with generated changelog and assets
5. **Checksums**: Generates checksums for all release artifacts

### 2. Supported Platforms

The automated build creates binaries for:

- **Linux**: 386, amd64, arm, arm64
- **macOS (Darwin)**: amd64, arm64
- **Windows**: 386, amd64

## Version Management

### Semantic Versioning

Defuddle Go follows [Semantic Versioning](https://semver.org/) (SemVer):

- **MAJOR** version: Incompatible API changes
- **MINOR** version: Backward-compatible functionality additions
- **PATCH** version: Backward-compatible bug fixes

### Version Format

- Release versions: `v1.2.3`
- Pre-release versions: `v1.2.3-alpha.1`, `v1.2.3-beta.1`, `v1.2.3-rc.1`

## Creating a Release

### Prerequisites

1. **Repository Access**: Write access to the repository
2. **Clean State**: All changes committed and pushed to main branch
3. **Tests Passing**: All CI tests must pass
4. **Version Updated**: Update version in `cmd/main.go` if needed

### Release Steps

#### 1. Prepare the Release

```bash
# Ensure you're on the main branch and up to date
git checkout main
git pull origin main

# Run tests locally
task test

# Test CLI build
task build-cli
./bin/defuddle --version
```

#### 2. Update Version (if needed)

Update the version in `cmd/main.go`:

```go
var (
    version = "0.2.0"  // Update this
)
```

Commit the version change:

```bash
git add cmd/main.go
git commit -m "feat: bump version to v0.2.0"
git push origin main
```

#### 3. Create and Push Tag

```bash
# Create the release tag
make tag VERSION=v0.2.0
```

This command will:
- Create an annotated tag
- Push the tag to GitHub
- Trigger the automated release workflow

#### 4. Monitor Release

1. Check the [Actions tab](https://github.com/kaptinlin/defuddle-go/actions) for the release workflow
2. Verify the release appears in [Releases](https://github.com/kaptinlin/defuddle-go/releases)
3. Test one of the pre-built binaries

### Alternative Manual Tag Creation

If you prefer manual tag creation:

```bash
# Create annotated tag
git tag -a v0.2.0 -m "Release v0.2.0"

# Push tag to trigger release
git push origin v0.2.0
```

## Testing Releases

### Local Release Testing

Test the release build locally without publishing:

```bash
# Install GoReleaser if not already installed
make install-goreleaser

# Test release configuration
make release-test

# Create a snapshot release
make release-snapshot
```

### Pre-Release Testing

For major releases, consider creating a pre-release:

```bash
# Create pre-release tag
git tag -a v1.0.0-rc.1 -m "Release candidate v1.0.0-rc.1"
git push origin v1.0.0-rc.1
```

GoReleaser will automatically mark versions with `-alpha`, `-beta`, `-rc` as pre-releases.

## Release Checklist

### Before Release

- [ ] All tests passing locally (`task test`)
- [ ] CLI builds successfully (`task build-cli`)
- [ ] Version updated in `cmd/main.go` (if needed)
- [ ] CHANGELOG.md updated (optional, auto-generated)
- [ ] README.md examples tested
- [ ] No breaking changes (or properly documented)

### After Release

- [ ] Release workflow completed successfully
- [ ] GitHub release created with proper assets
- [ ] Release notes generated correctly
- [ ] Pre-built binaries work on different platforms
- [ ] `go install` works with new version
- [ ] Documentation updated if needed

## Troubleshooting

### Release Workflow Fails

1. **Check Action Logs**: Review the GitHub Actions logs for errors
2. **Test Locally**: Run `make release-test` to check GoReleaser configuration
3. **Tag Issues**: Ensure tag follows `v*` pattern and is properly annotated

### Missing Assets

If release assets are missing:

1. Check GoReleaser configuration in `.goreleaser.yml`
2. Verify build targets in the `builds` section
3. Re-run release workflow if needed

### Version Conflicts

If version conflicts occur:

1. Delete the problematic tag: `git tag -d v0.2.0 && git push origin :refs/tags/v0.2.0`
2. Fix version issues
3. Recreate the tag

## Emergency Procedures

### Rollback Release

If a release needs to be rolled back:

1. **Mark as Pre-release**: Edit the GitHub release and mark as pre-release
2. **Delete Release**: Delete the release from GitHub (this won't delete the tag)
3. **Delete Tag**: Remove the tag if necessary
4. **Fix Issues**: Address the problems
5. **Create New Release**: Follow normal release process

### Hotfix Release

For urgent fixes:

1. Create hotfix branch from the problematic release tag
2. Apply minimal fixes
3. Update patch version
4. Create new release tag
5. Follow normal release process

## Contact

For questions about the release process, please:

- Open an issue on GitHub
- Contact the maintainers
- Check the GitHub Actions documentation

---

**Note**: This release process is automated and designed to minimize manual intervention. Always test locally before creating releases.
