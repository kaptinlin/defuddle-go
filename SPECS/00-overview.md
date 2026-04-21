# Defuddle Go Overview

## Overview

Define the library-level scope, default usage stories, and the boundary between the root parsing package, the public extractor extension surface, and the repo-local CLI.

Do not use this file for detailed type definitions, parse-step rules, or lint policy. See `20-api-specs.md`, `30-data-model-specs.md`, `40-architecture-specs.md`, and `50-coding-standards.md`.

## Library Scope

Defuddle Go owns one job: turn HTML or a fetchable URL into a normalized extraction result with metadata, cleaned content, optional Markdown, and optional site-specific extraction.

The repository ships three relevant surfaces:

- the root `defuddle` package for library callers
- the public `extractors` package for site-specific extractor registration
- `cmd/defuddle`, a thin CLI wrapper over the root package

> **Why:** One parsing engine keeps the library and CLI aligned. The `extractors` package stays public so callers can extend supported sites without forking the repository.
> **Rejected:** Separate library and CLI extraction engines because they drift; hiding extractor registration in `internal/` because it blocks supported extension.

## Default Usage Stories

### Parse existing HTML

1. Call `NewDefuddle(html, options)`.
2. Call `(*Defuddle).Parse(ctx)`.
3. Read `Result.Content`, `Result.Metadata`, and optionally `Result.ContentMarkdown`.

### Fetch and parse a URL

1. Call `ParseFromURL(ctx, url, options)`.
2. Let the root package fetch through `requests.Client`.
3. Read the same `Result` contract as direct HTML parsing.

### Extend a supported site

1. Implement `extractors.BaseExtractor`.
2. Register it through `extractors.Register(...)` or a custom `Registry`.
3. Let the root package prefer that extractor before the generic fallback path.

> **Why:** These are the three stable user stories present in the code today. More specialized workflows should layer on top of them instead of adding parallel entry points.

## Capability Map

| Capability | State | Notes |
| --- | --- | --- |
| HTML parsing through `NewDefuddle` + `Parse` | Shipped | Primary library path |
| URL parsing through `ParseFromURL` | Shipped | Uses a default `requests.Client` when `Options.Client` is nil |
| Optional Markdown output | Shipped | `ContentMarkdown` is populated only when requested and conversion succeeds |
| Site-specific extractor registration | Shipped | Root parse prefers a matching extractor before generic fallback |
| Debug diagnostics | Shipped | Returned through `Result.DebugInfo` when debug is enabled |
| Explicit element-processor toggles in `Options` | Intended, not yet fully implemented | Main parse path exports the flags but does not yet route them into `internal/elements/` |
| CSS media-query evaluation for mobile styles | Intended, not yet implemented | `evaluateMediaQueries()` currently returns no style changes |

> **Status**: `Options.ProcessCode`, `ProcessImages`, `ProcessHeadings`, `ProcessMath`, `ProcessFootnotes`, and `ProcessRoles` are exported API fields, but `parseInternal` does not currently consult them when building the main parse pipeline.
>
> **Status**: media-query evaluation is not yet implemented in `defuddle.go`; the current pipeline still extracts content without CSS stylesheet execution.

## Source-of-Truth Rules

- `SPECS/` is the canonical home for library contracts and intended-state notes.
- `CLAUDE.md` is the concise repo guide for agents.
- `README.md` should describe shipped usage and examples, not forward-looking contracts.

> **Why:** The repository needs one place for terse workflow guidance and one place for durable contract detail. Mixing them makes both stale faster.

## Terminology

| Term | Definition | Not |
| --- | --- | --- |
| **Root package** | The `defuddle` package that owns parsing orchestration and result construction | Not the CLI package |
| **Extractor** | A site-specific implementation that can replace the generic fallback path for matching URLs | Not a post-processing filter |
| **Fallback parser** | The generic extraction pipeline used when no site-specific extractor handles the document | Not an error path |
| **Normalized result** | The `Result` value returned after metadata extraction, cleanup, and optional Markdown conversion | Not raw DOM output |

## Forbidden

- Do not add a second parsing engine for the CLI.
- Do not document planned behavior in `README.md` as if it were shipped.
- Do not move extractor registration behind an internal-only package.
- Do not treat exported-but-unwired options as completed behavior; mark the gap explicitly.

## References

- [20-api-specs.md](20-api-specs.md) — public functions, methods, registry helpers, and compatibility rules
- [30-data-model-specs.md](30-data-model-specs.md) — `Options`, `Result`, metadata, and extractor payload shapes
- [40-architecture-specs.md](40-architecture-specs.md) — package boundaries and parse pipeline
- [50-coding-standards.md](50-coding-standards.md) — tooling, lint, test, and documentation rules

**Origin:** extracted from `CLAUDE.md` sections `Code Architecture`, `API Compatibility`, `Performance Features`, `Testing Strategy`, and `Configuration Notes`, then verified against `defuddle.go`, `types.go`, `extractors/base.go`, `extractors/registry.go`, `Taskfile.yml`, `.golangci.yml`, and `lefthook.yml`.
