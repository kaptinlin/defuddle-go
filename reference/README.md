# Defuddle Reference

This directory contains reference materials for the Go implementation of Defuddle, including the original TypeScript source code and development utilities.

## Directory Structure

```
reference/
├── defuddle/          # Original TypeScript source (git submodule)
├── scripts/           # Development and maintenance scripts
│   ├── update-reference.sh      # Update submodule to latest version
│   └── check-api-changes.sh     # Check for API compatibility issues
└── README.md          # This file
```

## Quick Start

### Setup Reference Code

Initialize the TypeScript reference submodule:

```bash
git submodule update --init --recursive
```

### Update Reference Code

Update to the latest TypeScript version:

```bash
./reference/scripts/update-reference.sh
```

### Check API Compatibility

Verify API compatibility between versions:

```bash
./reference/scripts/check-api-changes.sh
```

## Development Scripts

### `update-reference.sh`
Automatically updates the Defuddle TypeScript submodule to the latest version and commits the changes.

**Usage:**
```bash
./reference/scripts/update-reference.sh [version]
```

**Examples:**
```bash
# Update to latest main branch
./reference/scripts/update-reference.sh

# Update to specific version
./reference/scripts/update-reference.sh v0.6.4
```

### `check-api-changes.sh`
Analyzes API changes between TypeScript versions and identifies potential compatibility issues for the Go implementation.

**Usage:**
```bash
./reference/scripts/check-api-changes.sh [from-version] [to-version]
```

**Examples:**
```bash
# Check changes from current to latest
./reference/scripts/check-api-changes.sh

# Check changes between specific versions
./reference/scripts/check-api-changes.sh v0.6.3 v0.6.4
```

## Development Workflow

### 1. Regular Reference Updates

Keep the TypeScript reference up to date:

```bash
# Weekly or before implementing new features
./reference/scripts/update-reference.sh

# Check for any breaking changes
./reference/scripts/check-api-changes.sh
```

### 2. Source Code Analysis

Compare TypeScript implementation when developing Go features:

```bash
# View TypeScript project structure
tree reference/defuddle/src

# Search for specific implementations
grep -r "function parseContent" reference/defuddle/src/
```

### 3. API Compatibility Verification

Before releasing new Go versions:

```bash
# Check for API changes
./reference/scripts/check-api-changes.sh

# Review exported interfaces
grep -r "export.*interface" reference/defuddle/src/
```

## Source Code Mapping

### Key TypeScript Files → Go Implementation

| TypeScript Source | Go Implementation | Purpose |
|------------------|-------------------|----------|
| `src/defuddle.ts` | `defuddle.go` | Main class and parsing logic |
| `src/types.ts` | `types.go` | Type definitions and interfaces |
| `src/metadata.ts` | `internal/metadata/` | Metadata extraction algorithms |
| `src/scoring.ts` | `internal/scoring/` | Content scoring and selection |
| `src/standardize.ts` | `internal/standardize/` | HTML normalization |
| `src/constants.ts` | `internal/constants/` | Selectors and configuration |
| `src/extractors/` | `extractors/` | Site-specific content extractors |

### API Compatibility Matrix

| Feature | TypeScript | Go | Status | Notes |
|---------|------------|-------|--------|-------|
| Core parsing | ✅ | ✅ | ✅ Complete | Full compatibility |
| Metadata extraction | ✅ | ✅ | ✅ Complete | Enhanced with schema.org |
| Site extractors | ✅ | ✅ | ✅ Complete | Twitter, Reddit, etc. |
| Markdown conversion | ✅ | ✅ | ✅ Complete | Using html-to-markdown |
| Debug mode | ✅ | ✅ | ✅ Complete | Enhanced with Go features |

## Version Tracking

### Current Implementation Status

| Component | TypeScript Version | Go Implementation | Compatibility |
|-----------|-------------------|------------------|---------------|
| Core API | v0.6.4 | ✅ Complete | 100% |
| Options | v0.6.4 | ✅ Complete | 100% |
| Response Format | v0.6.4 | ✅ Complete | 100% |
| Site Extractors | v0.6.4 | ✅ Complete | 100% |

### Upgrade Path

When new TypeScript versions are released:

1. **Update Reference**: `./reference/scripts/update-reference.sh`
2. **Check Changes**: `./reference/scripts/check-api-changes.sh`
3. **Analyze Impact**: Review reported changes for Go implementation
4. **Implement Updates**: Update Go code to maintain compatibility

## Contributing

When contributing to the Go implementation:

1. **Check Latest Reference**: Update TypeScript source before implementing
2. **Maintain API Compatibility**: Ensure public API matches exactly
3. **Document Changes**: Update this mapping when adding new features
4. **Test Thoroughly**: Verify behavior matches TypeScript implementation

## Resources

- [Defuddle TypeScript Repository](https://github.com/kepano/defuddle)
- [API Documentation](https://kepano.github.io/defuddle/)
- [Go Implementation Rules](../.cursor/defuddle-go-rules.mdc)
- [Development Scripts](./scripts/) 