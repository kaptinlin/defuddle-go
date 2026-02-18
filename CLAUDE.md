# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building and Installation
```bash
# Build the CLI binary
make build-cli

# Install CLI to system
make install-cli

# Build with Go directly
go build -o bin/defuddle ./cmd/defuddle
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run tests with race detector
go test -race ./...

# Run unit tests only
make test-unit

# Run benchmarks
make bench

# Verbose test output
make test-verbose
```

### Code Quality
```bash
# Run all linters
make lint

# Format code
make fmt

# Run go vet
make vet

# Complete verification (format, vet, lint, test)
make verify

# Quick development verification
make dev
```

### Dependencies
```bash
# Download and tidy dependencies
make deps

# Initialize git submodules (required for reference library)
make submodules
```

### Development Workflow
```bash
# Quick development cycle
make dev && make test

# Full verification before commit
make verify
```

## Code Architecture

### Core Structure
- **`defuddle.go`** - Main entry point and orchestration logic
- **`types.go`** - Core data structures and type definitions
- **`cmd/main.go`** - CLI application entry point
- **`extractors/`** - Site-specific content extractors (ChatGPT, GitHub, Reddit, Twitter/X, YouTube, etc.)
- **`internal/`** - Internal implementation packages

### Key Architectural Components

#### Extraction Pipeline
1. **Schema.org Processing** - JSON-LD structured data extraction
2. **Site-Specific Detection** - Specialized extractors for platforms
3. **Content Scoring** - Algorithm to identify main content areas
4. **Clutter Removal** - Removes ads, navigation, sidebars
5. **Element Processing** - Handles code blocks, images, math, footnotes
6. **Standardization** - Normalizes HTML structure
7. **Markdown Conversion** - Optional HTML-to-Markdown transformation

#### Internal Packages
- **`internal/elements/`** - Element-specific processing (code, images, headings, math, footnotes, roles)
- **`internal/scoring/`** - Content relevance scoring algorithms
- **`internal/metadata/`** - Metadata extraction from HTML
- **`internal/standardize/`** - HTML structure normalization
- **`internal/markdown/`** - HTML to Markdown conversion
- **`internal/debug/`** - Debug logging and processing information
- **`internal/pool/`** - Object pooling for performance optimization

#### Extractor System
Site-specific extractors implement the `BaseExtractor` interface:
- **Registry-based** - Extractors register themselves for URL patterns
- **Modular Design** - Each extractor handles platform-specific content
- **Fallback Mechanism** - Falls back to general extraction if no specific extractor matches
- **Built-in Support** - ChatGPT, GitHub (issues), Reddit, Twitter/X, YouTube, Hacker News, Grok, Claude, and Gemini

### Key Dependencies
- **goquery** - HTML parsing and DOM manipulation
- **requests** - HTTP client for URL fetching (mandatory for compatibility)
- **html-to-markdown** - HTML to Markdown conversion
- **json-gold** - JSON-LD processing for Schema.org data
- **cobra** - CLI framework

## API Compatibility
This Go implementation maintains complete compatibility with the original TypeScript Defuddle library:
- Identical method signatures and return structures
- Same input produces same output across platforms
- Field names aligned with JavaScript version

## Performance Features
- Object pooling with `sync.Pool` to minimize allocations
- Optimized string building with `strings.Builder`
- Concurrent-safe processing for multiple documents
- Structured logging with `slog`

## Testing Strategy
- Uses `testify` for test assertions
- Target coverage >90%
- Benchmark tests with allocation reporting
- Race condition detection enabled
- TypeScript compatibility validation with identical inputs

## Configuration Notes
- Cursor rules available in `.cursor/defuddle-go-rules.mdc`
- Follows strict Go best practices and TypeScript API compatibility
- All comments must be in English with original JavaScript code included for reference


## Agent Skills

This package indexes agent skills from its own .agents/skills directory (defuddle-go/.agents/skills/):

| Skill | When to Use |
|-------|-------------|
| [agent-md-creating](.agents/skills/agent-md-creating/) | Create or update CLAUDE.md and AGENTS.md instructions for this Go package. |
| [code-simplifying](.agents/skills/code-simplifying/) | Refine recently changed Go code for clarity and consistency without behavior changes. |
| [committing](.agents/skills/committing/) | Prepare conventional commit messages for this Go package. |
| [dependency-selecting](.agents/skills/dependency-selecting/) | Evaluate and choose Go dependencies with alternatives and risk tradeoffs. |
| [go-best-practices](.agents/skills/go-best-practices/) | Apply Google Go style and architecture best practices to code changes. |
| [linting](.agents/skills/linting/) | Configure or run golangci-lint and fix lint issues in this package. |
| [modernizing](.agents/skills/modernizing/) | Adopt newer Go language and toolchain features safely. |
| [ralphy-initializing](.agents/skills/ralphy-initializing/) | Initialize or repair the .ralphy workflow configuration. |
| [ralphy-todo-creating](.agents/skills/ralphy-todo-creating/) | Generate or refine TODO tracking via the Ralphy workflow. |
| [readme-creating](.agents/skills/readme-creating/) | Create or rewrite README.md for this package. |
| [releasing](.agents/skills/releasing/) | Prepare release and semantic version workflows for this package. |
| [testing](.agents/skills/testing/) | Design or update tests (table-driven, fuzz, benchmark, and edge-case coverage). |
