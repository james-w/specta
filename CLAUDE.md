# specta - Testing Framework Development Guide

## Project Overview

**specta** is a Go testing library emphasizing composition and reuse through matchers and test data factories. Module: `github.com/james-w/specta` (Go 1.22+)

The core value: build reusable test components instead of duplicating assertions. Matchers compose, factories generate test data, and error messages include actual test expressions via AST parsing.

## Structure

Core library: `matchers.go`, `matchers_collection.go`, `matchers_specialized.go`, `matchers_diff.go`, `primitives.go`, `spec.go`

Code generator: `cmd/main.go` - reads YAML config, generates factories, matchers, and specs for types

Tests: `*_test.go` files with table-driven subtests, no test duplication

Generated: `example/` and `showcase/` have `factory/` directories with `*_gen.go` and `*_matcher_gen.go` files (excluded from linting). These act as integration tests.

## Code Generation (specta.yaml)

**Flow**: YAML config → `cmd/main.go` parses source → generates 3 files per type:
- `spec/*_gen.go` - Low-level API with `Maybe` types
- `factory/*_gen.go` - High-level Recipe fluent API
- `factory/*_matcher_gen.go` - Matcher builders

Generated code is type-checked before writing. All generated files have `//go:build !ignore_testgen` tags and are excluded from linting.

## Task Runner (pls)

This project uses `pls` (github.com/james-w/pls) as a task runner. **Always use pls commands instead of running tools directly** to ensure consistency and proper dependency management.

**Common tasks:**
- `pls build` - Build the code generator binary (./specta)
- `pls run test` - Test main module only
- `pls run test-example` - Generate and test example module
- `pls run test-showcase` - Generate and test showcase module
- `pls run test-all` - Test all modules (automatically generates code first)
- `pls run generate` - Regenerate code for all modules
- `pls run format` - Format all code with gofmt
- `pls run lint` - Run golangci-lint
- `pls run ci` - Full CI check (format + generate + test-all + lint)

**Passing extra arguments:**
You can pass additional flags to test commands:
- `pls run test -run TestFoo` - Run specific tests in main module
- `pls run test -v -race` - Run with verbose and race detector
- `pls run test-example -run TestUser` - Run specific tests in example module

The `{args}` placeholder in pls.toml allows passing extra flags directly to the underlying command.

**Why use pls:**
- Ensures code generation runs before tests (test-example/test-showcase depend on generate)
- The generator binary is built as an artifact (only rebuilds when source changes)
- Tasks declare dependencies explicitly (no manual ordering required)
- Consistent commands across the team

## Issue Tracking (Beads)

This project uses **Beads** (github.com/steveyegge/beads) for issue tracking. Issues are stored in `.beads/` directory with SQLite database + JSONL backup.

**Issue prefix:** `gomatchers-` (auto-detected from directory name)

### Common Commands

**Creating issues:**
```bash
bd new "Title"                              # Create P2 task (default)
bd create "Fix bug" -p 1 -t bug -d "Description"
bd create "Add feature" -l "backend,urgent"
```

**Listing & querying:**
```bash
bd list                    # All issues
bd ready                   # Issues ready to work (no blockers)
bd blocked                 # Blocked issues
bd show <issue-id>        # View details (e.g., bd show gomatchers-abc)
bd stale                   # Not updated recently
```

**Updating issues:**
```bash
bd update <id> --description "Detailed description"
bd update <id> --priority 1 --status in_progress
bd close <id> --reason "Fixed in PR #123"
bd comment <id> "Progress update"
```

**Dependencies:**
```bash
bd dep add <from-id> <to-id> --type blocks           # Blocking dependency
bd dep add <from-id> <to-id> --type after            # Sequential order
bd dep add <from-id> <to-id> --type discovered-from  # Found during work
bd dep tree <id>                                      # Visualize dependencies
```

**Priority levels:** P0 (critical), P1 (high), P2 (medium/default), P3 (low), P4 (backlog)

**Status values:** open, in_progress, closed

**Issue types:** bug, feature, task, epic, chore

### When to Create Issues

- Bugs or problems discovered during development
- Feature requests or enhancements
- Technical debt that should be tracked
- Large refactorings that need planning
- Documentation improvements

Issues are lightweight - create them freely. Use dependencies to show relationships and ordering.

### Integration with Development

The beads database auto-syncs to `.beads/issues.jsonl` which is git-tracked. Issues persist across sessions and can be queried by AI agents for context on ongoing work.

See `bd help` for complete command reference or https://github.com/steveyegge/beads for full documentation.

## Linting & Testing

**.golangci.yml**: errcheck, govet, ineffassign, staticcheck, unused, misspell, gocyclo, dupl, unconvert. Test files and generated files excluded.

**Testing**: 70% coverage threshold (warns in CI). Run: `go test -v -race ./...` after running code generation.

**Format**: `gofmt -s` required, verified in CI. `go vet` also required.

## Development Guidelines

**Matchers**: Implement `Matcher[T]` interface with detailed failure messages. Compose with `AllOf`, `AnyOf`, `Not`. Use `matchers_diff.go` for structured error output.

**Testing**: Use subtests (`t.Run`), test both success/failure, avoid duplicate test patterns. Simple and focused.

**Code Generation**: Type-check before writing. Use templates (embedded in cmd/main.go). Generated code must be committed when types change.

**Error Messages**: Include actual vs expected values, use structured diffs with symbols (✓ matched, ✗ failed, ~ unchecked), auto-detect terminal colors.

## Before Committing

**Quick check:** Run `pls run ci`

This single command will:
1. Format code (`gofmt -s -w .`)
2. Regenerate code for all modules (if needed)
3. Run all tests (main + example + showcase with `-v -race`)
4. Run linter (`golangci-lint run`)

**Then ask yourself:**
5. Should README.md be updated?
6. Does the change include anything important enough to update CLAUDE.md?

**Manual workflow (if needed):**
- Regenerate code: `pls run generate` (builds generator first, runs in all modules)
- Format: `pls run format`
- Test: `pls run test-all` (generates code automatically before testing)
- Lint: `pls run lint`

Generated files are automatically committed when pls regenerates them.

## Primitives System

`Primitives` interface provides deterministic test data generation:
- Counter-based (`Next()`)
- Strings with prefixes
- Times (increment from base)
- Deterministic UUIDs
- IDs and random values

`Gen` type implements it with configurable start value, base time, step, prefix.

## Key Patterns

**Composition**: Build matchers from simple pieces. Reuse them.
**Partial Matching**: Only assert fields that matter. Other fields get defaults. Prevents brittle tests when new fields are added.
**Factories vs Matchers**: Two sides of same coin. Factory builds test data with defaults. Matcher validates with partial field matching.

See README.md for user documentation and comprehensive examples.
