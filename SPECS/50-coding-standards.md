# Defuddle Go Coding Standards

## Overview

Define the tooling gates, testing expectations, documentation rules, and compatibility discipline for changing Defuddle Go.

Do not use this file for package topology or public field definitions. See `40-architecture-specs.md` for architecture and `30-data-model-specs.md` for public data shapes.

## Toolchain and Gates

| Command or tool | Contract |
| --- | --- |
| `task test` | Runs `go test -race ./...` |
| `task lint` | Runs `golangci-lint` plus tidy checks |
| `task verify` | Runs deps, fmt, vet, lint, test, and `govulncheck` |
| `lefthook` pre-commit | Runs trailing-whitespace cleanup, gitleaks, lint, test, markdownlint, and yamllint |
| `task markdownlint` | Lints Markdown, including `SPECS/**` |

> **Why:** The repository already defines a complete local gate. Work should go green locally before it is committed or pushed.
> **Rejected:** Treating CI as the first validation step because that is too slow for routine iteration; excluding `SPECS/**` from markdownlint because the canonical design docs would silently stop being checked.

## Must Follow Rules

- Use the Go version declared in `go.mod` for code changes.
- Keep public comments and documentation in English.
- Preserve the repository's TypeScript-compatibility intent on shared root-package contracts.
- Prefer a single obvious API path over compatibility shims or duplicate entry points.
- Keep `README.md` limited to shipped usage and examples.
- Put intended-state design rules and explicit code/spec gaps in `SPECS/` with `> **Status**:` notes.
- Keep `CLAUDE.md` concise; move durable contract detail into `SPECS/`.

## Lint Discipline

The repository's lint configuration is intentionally strict. Changes must respect the active checks, especially:

- error naming and error wrapping hygiene
- exhaustive handling
- context-aware I/O
- security checks
- no stale or vague `nolint` usage
- consistent formatting through `gofmt` and `goimports`

When a lint rule fails, fix the code or docs rather than weakening the rule for convenience.

## Testing Rules

- Run `task lint` and `task test` before every commit.
- Use root-package tests for public parse behavior.
- Use `extractors/` tests for registry or site-specific extraction changes.
- Use `internal/*` tests for implementation-package changes.
- Keep race detection green; it is part of the default `task test` contract.
- Treat compatibility-sensitive parser behavior as something that requires tests, not just README examples.

> **Why:** The project already has a layered test layout that matches package ownership. New work should strengthen that structure instead of bypassing it.

## Documentation Rules

- `README.md` documents shipped behavior.
- `SPECS/` documents current contracts and intended-state gaps.
- `CLAUDE.md` documents workflow and points to the right spec.
- `AGENTS.md` remains a symlink to `CLAUDE.md`.
- Markdown in `SPECS/` must pass markdownlint.

When code lags a better intended contract, keep the rule in `SPECS/` and mark the gap explicitly. Do not silently narrow the docs to whatever the code happens to do today.

## Compatibility Discipline

- Do not rename exported fields casually.
- Do not present parsed-but-unwired CLI flags as shipped behavior.
- Do not present exported-but-unwired `Options.Process*` toggles as completed behavior.
- Do not hide public extension points behind internal packages.

## Terminology

| Term | Definition | Not |
| --- | --- | --- |
| **Gate** | A local validation step that must pass before shipping changes | Not a suggestion |
| **Compatibility-sensitive change** | A change that affects exported symbols, field names, or parse semantics | Not a local refactor |
| **Status note** | A `> **Status**:` callout in `SPECS/` marking a known code/spec gap | Not a TODO comment |

## Forbidden

- Do not exclude `SPECS/**` from markdownlint.
- Do not add doc-layout `_test.go` files for `SPECS/`, `CLAUDE.md`, or `AGENTS.md`.
- Do not document intended behavior in `README.md` as if it were already shipped.
- Do not add compatibility shims when one API path is enough.
- Do not fix gate failures by loosening checks unless the repository owner explicitly asks for that change.

## References

- [00-overview.md](00-overview.md) — repository scope and contract boundaries
- [20-api-specs.md](20-api-specs.md) — compatibility-sensitive public APIs
- [30-data-model-specs.md](30-data-model-specs.md) — exported structs and JSON tags
- [40-architecture-specs.md](40-architecture-specs.md) — package ownership and parse pipeline

**Origin:** extracted from `CLAUDE.md` sections `Testing Strategy` and `Configuration Notes`, then verified against `Taskfile.yml`, `.golangci.yml`, `lefthook.yml`, and `.markdownlint-cli2.yaml`.
