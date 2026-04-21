# Defuddle Go API Specs

## Overview

Define the public contracts for the root `defuddle` package, the public `extractors` extension package, and the repo-local CLI surface where it forwards root-package behavior.

Do not use this file for field-by-field data-model detail or package topology. See `30-data-model-specs.md` for payload shapes and `40-architecture-specs.md` for pipeline and package boundaries.

## Root Package Entry Points

| Symbol | Contract |
| --- | --- |
| `NewDefuddle(html string, options *Options) (*Defuddle, error)` | Parse caller-supplied HTML into a reusable parser instance |
| `(*Defuddle).Parse(ctx context.Context) (*Result, error)` | Extract metadata and main content from the configured document |
| `ParseFromURL(ctx context.Context, url string, options *Options) (*Result, error)` | Fetch a URL, build a parser, and return the same `Result` contract as direct HTML parsing |
| `ParseFromString(ctx context.Context, html string, options *Options) (*Result, error)` | Convenience wrapper for one-shot HTML parsing |

> **Why:** The root package should read as a small, obvious surface: construct, parse, or fetch-and-parse. More specialized behavior belongs in options or extractor registration, not in new top-level entry points.
> **Rejected:** Separate sync and async APIs because `context.Context` already handles cancellation; a builder-only API because it adds ceremony to the common path.

## Parse Semantics

### `NewDefuddle`

- Accepts raw HTML as a string and stores a parsed `goquery.Document`.
- Returns an error when the HTML cannot be parsed into a document.
- Enables debug diagnostics only when `options != nil && options.Debug`.

### `(*Defuddle).Parse`

- Runs the standard parse pipeline.
- If the first pass returns `WordCount < 200`, retries once with `RemovePartialSelectors` disabled.
- Returns the retry result only when the retry produces more content.

> **Why:** A second pass without partial-selector removal recovers overly aggressive cleanup on sparse pages without exposing another public method.

### `ParseFromURL`

- Initializes `options` when the caller passes `nil`.
- Writes `options.URL = url` when `options.URL` is empty.
- Uses `options.Client` when provided.
- Otherwise creates a default `requests.Client` with the Defuddle user agent and a 30s timeout.

> **Why:** The root package keeps URL parsing usable with zero setup while still allowing callers to inject a custom client.
> **Rejected:** Requiring a caller-supplied HTTP client for all URL parsing because that adds too much ceremony; hiding the fetched URL from `Options.URL` because that breaks downstream metadata extraction.

### `ParseFromString`

- Exists only as a one-shot convenience wrapper.
- Must remain behaviorally equivalent to `NewDefuddle(html, options)` followed by `Parse(ctx)`.

## Compatibility Rules

- Keep the root-package field names and broad result shape aligned with the TypeScript Defuddle surface where the root parser overlaps it.
- Preserve the returned `Result` structure for successful parses: metadata plus `Content`, optional `ContentMarkdown`, optional `ExtractorType`, optional `MetaTags`, and optional `DebugInfo`.
- Prefer extending behavior through `Options` or the `extractors` package instead of adding new top-level parse functions.

> **Why:** TypeScript compatibility is part of the repository promise, but Go callers still need an idiomatic extension story.

## Extractor Extension Surface

### Required contracts

| Symbol | Contract |
| --- | --- |
| `extractors.BaseExtractor` | Implement `CanExtract() bool`, `Extract() *ExtractorResult`, and `Name() string` |
| `extractors.ExtractorResult` | Return cleaned text/HTML plus optional extracted content and variables |
| `extractors.ExtractorMapping` | Bind patterns to an extractor constructor |
| `extractors.Registry` | Own extractor registration, lookup, and cache invalidation |

### Default-registry helpers

- `extractors.Register(mapping)` registers against the default registry.
- `extractors.FindExtractor(document, url, schemaOrgData)` initializes built-ins and resolves the first matching extractor.
- `extractors.ClearCache()` clears the default-registry domain cache.

### Extractor-resolution rules

- Built-in extractors are initialized exactly once.
- URL resolution may match by hostname string or regular expression.
- `FindExtractor` returns `nil` when the URL is empty or cannot be parsed.
- The root parser only uses a resolved extractor when `CanExtract()` returns true.

> **Why:** Extractor lookup must stay predictable and cheap. The API supports extension without forcing callers to reimplement built-in registration.

## Built-in Extractor Contract

The default registry currently ships built-ins for the following families:

- Twitter/X
- YouTube
- Reddit
- Hacker News
- ChatGPT
- Claude
- Grok / x.ai
- Gemini
- GitHub issues and pull requests

Adding a new built-in extractor must extend the registry rather than introducing special-case dispatch in the root package.

## CLI Parse Contract

`cmd/defuddle` exposes one public subcommand: `defuddle parse <source>`.

Current forwarded behavior:

- `--json`
- `--markdown` and `--md`
- `--property`
- `--output`
- `--timeout`
- `--debug`

> **Status**: `--header`, `--proxy`, and `--user-agent` are parsed by the CLI flag layer, but the current command implementation does not wire them into the HTTP client used by `ParseFromURL`.
> **Why:** The CLI should stay a thin adapter over the root package. A flag is not part of the shipped contract until it affects runtime behavior.

## Terminology

| Term | Definition | Not |
| --- | --- | --- |
| **Entry point** | A public function or method that callers invoke directly | Not an internal helper |
| **Compatibility rule** | A public contract that constrains future changes to shape or semantics | Not a one-off implementation detail |
| **Default registry** | The shared extractor registry used by package-level convenience helpers | Not the only possible registry |

## Forbidden

- Do not add new top-level parse functions when an option or extractor extension solves the problem.
- Do not change exported JSON field names in `Result`, `Metadata`, `MetaTag`, or `Options` without a compatibility review.
- Do not add root-package switch dispatch for built-in extractors; register them through `extractors.Registry`.
- Do not document CLI flags as shipped behavior when the command layer does not forward them to runtime behavior.

## References

- [00-overview.md](00-overview.md) — library scope and capability map
- [30-data-model-specs.md](30-data-model-specs.md) — public structs and payload invariants
- [40-architecture-specs.md](40-architecture-specs.md) — extractor-first pipeline and package ownership
- [50-coding-standards.md](50-coding-standards.md) — compatibility, test, and documentation rules

**Origin:** extracted from `CLAUDE.md` sections `Code Architecture` and `API Compatibility`, then verified against `defuddle.go`, `types.go`, `extractors/base.go`, `extractors/registry.go`, `internal/metadata/metadata.go`, and `cmd/defuddle/main.go`.
