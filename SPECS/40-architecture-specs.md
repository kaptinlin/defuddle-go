# Defuddle Go Architecture Specs

## Overview

Define the package boundaries, parse pipeline, extractor topology, and explicit current implementation gaps for Defuddle Go.

Do not use this file for public field definitions or lint policy. See `30-data-model-specs.md` for data shapes and `50-coding-standards.md` for tooling and workflow rules.

## Package Boundaries

| Package or path | Owns | Does not own |
| --- | --- | --- |
| Root `defuddle` package | Parse orchestration, option merging, result construction, URL fetching entry points | Site-specific extractor registration internals, low-level standardization helpers |
| `extractors/` | Site-specific extractor interfaces, registry, built-in site registrations | Generic fallback extraction |
| `internal/metadata/` | Metadata extraction from document, schema.org payload, and meta tags | Main-content scoring and cleanup |
| `internal/scoring/` | Heuristic scoring and removal of non-content blocks | Final result assembly |
| `internal/standardize/` | Content cleanup and normalization after main-content selection | Site detection and metadata extraction |
| `internal/elements/` | Optional element processors for code, images, headings, math, footnotes, and roles | Main parse orchestration |
| `internal/debug/` | Debug timers, processing steps, and statistics | Parse decisions themselves |
| `cmd/defuddle/` | CLI flag parsing and output formatting | A second parsing implementation |

> **Why:** Each package should own one stage of the extraction story. The root package composes these stages; it should not absorb every algorithm directly.
> **Rejected:** Moving registry logic into the root package because that couples extension with orchestration; treating `cmd/defuddle` as a separate engine because that invites drift.

## Parse Pipeline

The generic parse path runs in this order:

1. Merge defaults, instance options, and override options.
2. Extract schema.org data.
3. Collect meta tags.
4. Extract metadata from the document and base URL.
5. Try a site-specific extractor.
6. Evaluate media-query-derived mobile styles.
7. Find main content through entry-point selectors, then table heuristics, then score-based fallback.
8. Remove small images and optionally all images.
9. Remove hidden elements, low-score content, and clutter selectors.
10. Standardize the chosen content subtree.
11. Count words and optionally convert to Markdown.
12. Attach debug information when enabled.

> **Why:** The extractor-first design gives site-specific implementations priority, while the fallback parser remains the common baseline for arbitrary HTML.

## Main-Content Selection Rules

### Selection order

1. First matching entry-point selector.
2. Highest-scoring table cell when the score exceeds the threshold.
3. Highest-scoring `div`, `section`, `article`, or `main` candidate above the threshold.
4. Fallback to raw `<body>` HTML when no main-content node is found.

### Cleanup order

After selecting the content subtree, the generic path must remove noise before standardization:

- small images discovered from the source document
- all images when `RemoveImages` is true
- hidden elements
- low-score elements removed by `internal/scoring`
- exact and partial selector matches when enabled

### Standardization order

`internal/standardize.Content` is responsible for:

- whitespace normalization
- comment removal semantics
- heading normalization
- footnote normalization
- embedded-element normalization
- wrapper flattening and empty-element cleanup outside debug mode

> **Why:** Content detection is only half of the contract. Output quality depends on applying cleanup and standardization in a stable order.

## Built-in Extractor Topology

The default registry initializes built-ins once and currently registers extractors for:

- Twitter / X
- YouTube
- Reddit
- Hacker News
- ChatGPT
- Claude
- Grok / x.ai
- Gemini
- GitHub issues and pull requests

Built-ins must be added by registering new `ExtractorMapping` values. Do not add site-specific conditionals to the root parser.

## Explicit Gap Contracts

### Media-query evaluation

The architecture intends to support a lightweight mobile-style evaluation phase before main-content detection.

> **Status**: `evaluateMediaQueries()` currently returns an empty slice, so the architecture stage exists but does not yet apply stylesheet-derived changes.

### Element processors

The architecture includes dedicated processors for code, images, headings, math, footnotes, and roles.

> **Status**: `internal/elements/` exists and has direct tests, but the main parse path does not yet wire the exported `Options.Process*` toggles and option structs into a dedicated post-processing phase.
> **Why:** These are still useful intended-state contracts. The repository already has the package boundaries and processor implementations, so the gap should stay visible instead of disappearing from working memory.

## Performance and Concurrency Rules

- Prefer one parse pipeline with explicit stages over parallel special cases.
- Keep object allocation pressure low through focused helpers and normalization passes.
- Treat parser instances as document-scoped values; create a new `Defuddle` for each source document.
- Preserve the ability to process multiple documents concurrently by avoiding shared mutable parse state outside the extractor registry cache.

## Terminology

| Term | Definition | Not |
| --- | --- | --- |
| **Pipeline stage** | A stable step in the generic parse flow | Not an exported API by itself |
| **Extractor-first** | Prefer a matching site-specific extractor before generic fallback | Not “extractors only” |
| **Standardization** | Cleanup and normalization after content selection | Not metadata extraction |

## Forbidden

- Do not put site-specific branching directly into the root parser when a registry mapping can own it.
- Do not document the media-query stage as complete while it remains a stub.
- Do not claim that `internal/elements/` processors are fully wired into `Parse` when the current root pipeline does not consult the exported toggles.
- Do not treat `cmd/defuddle` as a separate architecture layer with its own extraction semantics.

## References

- [00-overview.md](00-overview.md) — scope and shipped/intended capability map
- [20-api-specs.md](20-api-specs.md) — public entry points and extractor API
- [30-data-model-specs.md](30-data-model-specs.md) — option and result fields populated by the pipeline
- [50-coding-standards.md](50-coding-standards.md) — rules for changing pipeline behavior safely

**Origin:** extracted from `CLAUDE.md` sections `Code Architecture` and `Performance Features`, then verified against `defuddle.go`, `extractors/registry.go`, `internal/metadata/metadata.go`, `internal/scoring`, `internal/standardize/content.go`, and `internal/elements/`.
