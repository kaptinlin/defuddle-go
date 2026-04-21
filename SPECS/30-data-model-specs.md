# Defuddle Go Data Model Specs

## Overview

Define the public data shapes used by the root `defuddle` package and the public extractor extension surface: `Options`, `Result`, metadata payloads, and extractor payload models.

Do not use this file for top-level function semantics or package topology. See `20-api-specs.md` for API behavior and `40-architecture-specs.md` for the parse pipeline.

## `Options`

### Core execution fields

| Field | Type | Contract |
| --- | --- | --- |
| `Debug` | `bool` | Enables debug diagnostics and debug-oriented processing behavior |
| `URL` | `string` | Supplies the source URL for metadata extraction and extractor matching |
| `Markdown` | `bool` | Requests Markdown conversion |
| `SeparateMarkdown` | `bool` | Keeps HTML content and optionally adds `ContentMarkdown` |
| `Client` | `*requests.Client` | Injects the HTTP client used by `ParseFromURL`; excluded from JSON |

### Cleanup fields

| Field | Type | Default | Contract |
| --- | --- | --- | --- |
| `RemoveExactSelectors` | `bool` | `true` | Enables exact-selector clutter removal |
| `RemovePartialSelectors` | `bool` | `true` | Enables attribute-pattern clutter removal |
| `RemoveImages` | `bool` | `false` | Removes images from extracted content |

### Element-processing fields

| Field family | Contract |
| --- | --- |
| `ProcessCode`, `ProcessImages`, `ProcessHeadings`, `ProcessMath`, `ProcessFootnotes`, `ProcessRoles` | Intended per-feature toggles for post-processing stages |
| `CodeOptions`, `ImageOptions`, `HeadingOptions`, `MathOptions`, `FootnoteOptions`, `RoleOptions` | Intended per-feature configuration payloads |

> **Status**: the element-processing booleans and nested option structs are exported on `Options`, but the main parse path does not yet consult them when constructing `Result`.
> **Why:** The options bag keeps the TypeScript-shaped configuration surface in one place while allowing a Go-only HTTP client injection point.
> **Rejected:** Splitting the public config into many small structs because that makes it harder to pass and mirror across entry points; serializing `Client` into JSON because transport clients are runtime dependencies, not data.

## `Result`

| Field | Type | Contract |
| --- | --- | --- |
| Embedded `Metadata` | `Metadata` | Always present on success |
| `Content` | `string` | Cleaned HTML content |
| `ContentMarkdown` | `*string` | Present only when Markdown was requested and conversion succeeded |
| `ExtractorType` | `*string` | Present only when a site-specific extractor produced the result |
| `MetaTags` | `[]MetaTag` | Optional collected meta tags from the source document |
| `DebugInfo` | `*debug.Info` | Optional diagnostic payload when debug is enabled |

### Result invariants

- `Content` is the canonical content field. Markdown never replaces it.
- `ParseTime` is measured in milliseconds.
- `WordCount` is derived from the HTML content emitted into `Content`.
- `ExtractorType`, when present, is the extractor name lowercased with the `Extractor` suffix removed.
- `DebugInfo` is diagnostic output, not a stable construction API for external packages.

## `Metadata`

| Field | Type | Contract |
| --- | --- | --- |
| `Title` | `string` | Best title found from metadata and extraction rules |
| `Description` | `string` | Best available description or summary |
| `Domain` | `string` | Hostname-derived domain when available |
| `Favicon` | `string` | Best available favicon URL |
| `Image` | `string` | Best available primary image URL |
| `ParseTime` | `int64` | Elapsed parse time in milliseconds |
| `Published` | `string` | Best available published timestamp string |
| `Author` | `string` | Best available author string |
| `Site` | `string` | Site name from metadata or extractor variables |
| `SchemaOrgData` | `any` | Extracted schema.org payload |
| `WordCount` | `int` | Word count computed from emitted content |

`SchemaOrgData` is intentionally opaque. Callers may inspect or serialize it, but the root package does not promise a narrower static shape.

## `MetaTag`

| Field | Type | Contract |
| --- | --- | --- |
| `Name` | `*string` | Optional `name` attribute |
| `Property` | `*string` | Optional `property` attribute |
| `Content` | `*string` | Optional `content` attribute payload |

Use pointers so the serialized result can distinguish absent values from empty strings.

## Extractor Payload Models

### `extractors.ExtractorResult`

| Field | Type | Contract |
| --- | --- | --- |
| `Content` | `string` | Extracted text representation |
| `ContentHTML` | `string` | Extracted HTML used as the root parser's `Result.Content` |
| `ExtractedContent` | `map[string]any` | Optional extractor-specific payload |
| `Variables` | `map[string]string` | Optional override values such as `title`, `author`, or `published` |

### Root-package compatibility helpers

| Symbol | Contract |
| --- | --- |
| `ExtractorVariables` | Alias for `map[string]string` |
| `ExtractedContent` | Compatibility helper type with optional title/author/published/content fields |

> **Why:** The root package keeps the main result model stable while allowing extractor-specific detail to remain flexible.

## Terminology

| Term | Definition | Not |
| --- | --- | --- |
| **Options bag** | The single public config struct used by all root entry points | Not a builder |
| **Opaque field** | A field whose concrete runtime shape may vary and is not narrowed by this spec | Not undocumented data |
| **Override variable** | A string value returned by an extractor that can replace root metadata fields | Not a new top-level result field |

## Forbidden

- Do not serialize runtime-only dependencies such as `Client`.
- Do not make optional result fields required without a compatibility review.
- Do not rename exported JSON tags casually; field names are part of the contract.
- Do not claim that `Process*` toggles already change parse output; the current main path does not wire them in.

## References

- [00-overview.md](00-overview.md) — scope and capability map
- [20-api-specs.md](20-api-specs.md) — public entry points and extractor extension API
- [40-architecture-specs.md](40-architecture-specs.md) — where each field is populated in the parse pipeline
- [50-coding-standards.md](50-coding-standards.md) — compatibility and test expectations for public structs

**Origin:** extracted from `CLAUDE.md` sections `API Compatibility` and `Code Architecture`, then verified against `types.go`, `internal/metadata/metadata.go`, `defuddle.go`, and `extractors/base.go`.
