# GitHub Actions CI Workflows

This directory contains the CI/CD workflows for the Specta project.

## ci.yml

The main CI workflow runs on every push to `master`/`main` and on all pull requests. It includes the following jobs:

### Jobs

#### Test
- Runs tests with race detection
- Tests against multiple Go versions (1.22, 1.23)
- Generates code coverage reports
- Warns if coverage drops below 70%
- Uses Go module caching for faster builds
- **Annotations**: Test failures are automatically annotated in PRs with line numbers

#### Lint
- Runs golangci-lint with comprehensive linters
- Configuration in `.golangci.yml`
- Only shows new issues on PRs
- **Annotations**: Automatically adds inline annotations to PRs for any lint issues
- Includes: errcheck, gosimple, govet, ineffassign, staticcheck, goimports, misspell, and more

#### Format Check
- Verifies code is formatted with `gofmt -s`
- Runs `go vet` to catch suspicious constructs
- **Annotations**: Shows which files need formatting

#### Go Generate Check
- Runs `go generate ./...`
- Verifies no files are modified (all generated code should be committed)
- Fails if any changes are detected
- **Annotations**: Shows the diff if generate produces changes

#### Build
- Verifies the code builds successfully
- Builds both main package and cmd package

#### Security Scan
- Runs Gosec security scanner
- Uploads results in SARIF format
- **Annotations**: Security issues are shown in the Security tab and as PR annotations

#### Dependency Check
- Uses govulncheck to scan for known vulnerabilities in dependencies
- Fails if vulnerabilities are found

## Features

### Caching
All jobs use Go module caching (`cache: true` in setup-go) which dramatically speeds up workflow runs by caching:
- Downloaded modules
- Build cache

### Annotations
The following jobs provide inline annotations in PRs:
- **golangci-lint**: Lint issues are shown directly on the line of code
- **Test failures**: Stack traces are linked to source lines
- **gosec**: Security issues are annotated
- **Format check**: Shows which files need formatting
- **Generate check**: Shows diffs if generated code is out of date

### Matrix Testing
Tests run against multiple Go versions (1.22, 1.23) to ensure compatibility.

## Local Development

Run the same checks locally before pushing:

```bash
# Run tests
go test -v -race ./...

# Run linter (requires golangci-lint installation)
golangci-lint run

# Check formatting
gofmt -s -l .
go vet ./...

# Verify go generate is up to date
go generate ./...
git diff --exit-code

# Check for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Configuration Files

- `.golangci.yml` - golangci-lint configuration with enabled linters and settings
- This workflow file - `.github/workflows/ci.yml`
