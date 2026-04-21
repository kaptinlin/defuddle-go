# defuddle-go

Go library and CLI for extracting readable content from HTML documents and web pages. The root package owns parsing orchestration, `extractors/` owns site-specific extension points, and `cmd/defuddle` is a thin CLI adapter over the same engine.

For installation and usage examples, see [README.md](README.md). For monorepo-wide workflow rules, see [../CLAUDE.md](../CLAUDE.md).

## Commands

Run from this directory.

```bash
task test          # Run go test -race ./...
task lint          # Run golangci-lint and tidy checks
task markdownlint  # Run markdownlint on Markdown files
task verify        # Run deps + fmt + vet + lint + test + vuln
```

## Architecture

```text
defuddle-go/
├── defuddle.go             # Root parser orchestration and URL fetch helpers
├── types.go                # Public options, results, and compatibility helpers
├── extractors/             # Public site-specific extractor interfaces and registry
├── internal/metadata/      # Metadata extraction
├── internal/scoring/       # Heuristic scoring and content removal
├── internal/standardize/   # Content normalization and cleanup
├── internal/elements/      # Optional element processors and experiments
├── internal/debug/         # Debug timers and statistics
├── cmd/defuddle/           # CLI wrapper over the root package
└── SPECS/                  # Canonical design and contract documents
```

## Agent Workflow

### Read Code First, Then the Relevant Spec

When changing code, read the target implementation and tests first. Use the matching `SPECS/` file when the task affects public contracts, data shapes, pipeline behavior, or repository rules.

If `SPECS/` and the current code diverge, verify whether the spec describes shipped behavior or intended behavior. Shipped behavior in `README.md` must match the code. Intended behavior may stay in `SPECS/` only when it is marked with an explicit status note.

### Documentation Boundaries

- `README.md` documents shipped usage.
- `SPECS/` documents contracts and explicit intended-state gaps.
- This file stays concise and points to the right spec.
- `AGENTS.md` is a symlink to this file.

## SPECS Index

| File | Purpose | Status |
| --- | --- | --- |
| `SPECS/00-overview.md` | Library scope, default usage stories, and spec map | Current |
| `SPECS/20-api-specs.md` | Root-package API, extractor extension API, and CLI contract notes | Current with explicit gap notes |
| `SPECS/30-data-model-specs.md` | `Options`, `Result`, metadata, and extractor payload shapes | Current with explicit gap notes |
| `SPECS/40-architecture-specs.md` | Package boundaries, parse pipeline, and extractor topology | Current with explicit gap notes |
| `SPECS/50-coding-standards.md` | Tooling, testing, compatibility, and documentation rules | Current |

## Design Notes

- Keep the root API small: construct, parse, or fetch-and-parse.
- Prefer extractor registration over root-package special cases.
- Keep TypeScript-compatible field names where the root parser mirrors the original Defuddle contract.
- Do not present parsed-but-unwired flags or exported-but-unwired options as shipped behavior.

## Testing

- Run `task lint` and `task test` before committing.
- Keep markdownlint green for `SPECS/**`, `README.md`, and this file.
- Add or update tests when parser behavior, extractor routing, or exported contracts change.

## Agent Skills

This package indexes agent skills from its own `.agents/skills` directory. Use the matching local skill when one explicitly applies to the task; otherwise follow the repository rules above.
