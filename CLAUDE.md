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

1. Regenerate code if types change: `go run ./cmd/main.go -config specta.yaml` in each target directory
2. Format: `gofmt -s -w .`
3. Test: `go test -v -race ./...`
4. Lint: `golangci-lint run`
5. Commit generated files with source changes or CI will fail
6. Should README.md be updated?
7. Does the change include anything important enough to update CLAUDE.md?

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
