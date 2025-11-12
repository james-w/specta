# GitHub Actions CI Workflows

This directory contains the CI/CD workflows for the Specta project.

## ci.yml

The main CI workflow runs on every push to `master`/`main` and on all pull requests. It includes the following jobs:

### Jobs

#### Build
- Builds the project with Go 1.23 (the release version)
- Compiles the `specta` binary
- Uploads the binary as an artifact for use in integration tests
- **Purpose**: Ensures the code compiles and produces a distributable binary

#### Unit Tests
- Runs unit tests for the root package and cmd package
- Uses Go 1.23 (matching the build version)
- Generates code coverage reports
- Warns if coverage drops below 70%
- **Annotations**: Test failures are automatically annotated in PRs with line numbers
- **Coverage**: Only unit tests contribute to coverage metrics (not integration tests)

#### Verify Generated Files
- **Depends on**: Build job (needs the binary artifact)
- Downloads the `specta` binary built in the Build job
- Generates test fixtures for `example/` and `showcase/` directories
- Verifies generated files match committed versions
- Runs once (no matrix)
- **Purpose**: Ensures developers have committed the latest generated code

#### Integration Tests
- **Depends on**: Build job (needs the binary artifact)
- Tests against multiple Go versions (1.22, 1.23)
- Downloads the `specta` binary built in the Build job
- Generates test fixtures for `example/` and `showcase/` directories
- Runs integration tests to ensure generated code compiles and works across Go versions
- **Purpose**: Validates that generated code is compatible with supported Go versions

#### Lint
- Runs golangci-lint with comprehensive linters
- Uses Go 1.23
- Configuration in `.golangci.yml`
- Only shows new issues on PRs
- **Annotations**: Automatically adds inline annotations to PRs for any lint issues
- Includes: errcheck, gosimple, govet, ineffassign, staticcheck, goimports, misspell, and more

#### Format Check
- Verifies code is formatted with `gofmt -s`
- Runs `go vet` to catch suspicious constructs
- Uses Go 1.23
- **Annotations**: Shows which files need formatting

#### Security Scan
- Runs Gosec security scanner
- Uses Go 1.23
- Uploads results in SARIF format
- **Annotations**: Security issues are shown in the Security tab and as PR annotations

#### Dependency Check
- Uses govulncheck to scan for known vulnerabilities in dependencies
- Uses Go 1.23
- Fails if vulnerabilities are found

## Architecture

### Parallel Execution
Most jobs run in parallel for faster CI times:
- Build, Unit Tests, Lint, Format Check, Security Scan, and Dependency Check all run simultaneously
- Verify Generated Files and Integration Tests wait for Build (since they need the binary artifact)
- Verify Generated Files and Integration Tests run in parallel with each other

### Test Strategy
- **Unit Tests**: Test source code with Go 1.23, collect coverage
- **Integration Tests**: Test generated code compatibility across Go 1.22 and 1.23 using the built binary

This separation ensures:
1. Fast feedback on code changes (unit tests run immediately)
2. Binary validation (integration tests use the actual distributable binary)
3. Cross-version compatibility (generated code works with minimum supported Go version)

### Go Version Strategy
- **Go 1.23**: Used for building releases and running unit tests/linting
- **Go 1.22 & 1.23**: Integration tests verify generated code works across supported versions
- Rationale: Users download pre-built binaries (built with 1.23), but run them in projects with varying Go versions (1.22+)

## Features

### Caching
All jobs use Go module caching (`cache: true` in setup-go) which speeds up workflow runs by caching:
- Downloaded modules
- Build cache

Note: Caches are per-job and don't share between jobs, but this is optimal for parallel execution.

### Annotations
The following jobs provide inline annotations in PRs:
- **golangci-lint**: Lint issues are shown directly on the line of code
- **Test failures**: Stack traces are linked to source lines
- **gosec**: Security issues are annotated
- **Format check**: Shows which files need formatting
- **Verify Generated Files**: Shows diffs if generated code is out of date

### Binary Artifacts
The Build job uploads the `specta` binary as an artifact:
- Available for download from GitHub Actions UI
- Used by Verify Generated Files to check committed code
- Used by Integration Tests to generate fixtures
- Retained for 30 days

## Local Development

Run the same checks locally before pushing:

```bash
# Install the specta binary
go install ./cmd/...

# Run unit tests
go test -v -race . ./cmd/...

# Generate test fixtures (requires specta in PATH)
go generate ./...

# Run integration tests
go test -v -race ./example/... ./showcase/...

# Run linter (requires golangci-lint installation)
golangci-lint run

# Check formatting
gofmt -s -l .
go vet ./...

# Check for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### go:generate Directives
The project uses `go:generate` directives that assume `specta` is in your PATH:

```go
//go:generate specta -config testgen.yaml
```

To run these locally:
1. Install the tool: `go install ./cmd/...`
2. Ensure `$GOPATH/bin` is in your PATH
3. Run: `go generate ./...`

## Configuration Files

- `.golangci.yml` - golangci-lint configuration with enabled linters and settings
- This workflow file - `.github/workflows/ci.yml`
