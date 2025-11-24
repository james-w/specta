---
title: "Advanced Topics"
weight: 7
---

<!-- setup
package doctest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/james-w/specta"
)

var _ = fmt.Sprint
var _ = strings.Contains
var _ = testing.Verbose

// User type for examples
type User struct {
	ID          string
	Name        string
	Email       string
	Age         int
	Active      bool
	Status      string
	Role        string
	Permissions []string
}

// Stub functions for examples
func ValidateUser(u User) error {
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if u.Age < 0 {
		return fmt.Errorf("age must be non-negative")
	}
	return nil
}

func CreateUser(email string) User {
	return User{Email: email, Status: "active"}
}

// Stub matchers referenced in examples
func MatchRegex(pattern string) specta.Matcher[string] {
	return specta.Equal("") // stub
}

func ContainString(substr string) specta.Matcher[string] {
	return specta.Equal("") // stub
}

func IsEmpty() specta.Matcher[string] {
	return specta.Equal("") // stub
}

func HaveSuffix(suffix string) specta.Matcher[string] {
	return specta.Equal("") // stub
}

func BeNil() specta.Matcher[error] {
	return specta.Equal[error](nil)
}

func ContainAll(items ...string) specta.Matcher[[]string] {
	return nil // stub
}

// Mock matcher builder
type UserMatcher struct{}

func MatchUser() *UserMatcher {
	return &UserMatcher{}
}

func (m *UserMatcher) WithStatus(matcher specta.Matcher[string]) *UserMatcher {
	return m
}

func (m *UserMatcher) WithRole(matcher specta.Matcher[string]) *UserMatcher {
	return m
}

func (m *UserMatcher) WithPermissions(matcher specta.Matcher[[]string]) *UserMatcher {
	return m
}

func (m *UserMatcher) Matches(u User) specta.MatchResult {
	return specta.MatchResult{Matched: true}
}

// Mock factory type
type factoryType struct{}

type userBuilder struct {
	p *specta.Gen
}

func (b userBuilder) Build() User {
	return User{ID: "test-id", Name: "Test", Email: "test@example.com", Age: 30, Active: true}
}

func (f factoryType) NewUser(p *specta.Gen) userBuilder {
	return userBuilder{p: p}
}

var (
	factory  = factoryType{}
	message  = "test message"
	user     = User{ID: "user-1", Name: "Alice", Email: "alice@example.com", Age: 30}
	value    = 42
	expected = 42
	x        = 10
	y        = 5
	z        = 20
)
-->

# Advanced Topics

Deep dives into advanced patterns and techniques.

## Custom Matchers

While code generation handles most cases, sometimes you need a custom matcher.

### Implementing specta.Matcher[T]

The `Matcher` interface is defined as:

<!-- skip-test -->
```go
type Matcher[T any] interface {
    Match(value T) MatchResult
}

type MatchResult struct {
    Matched bool
    Message string
}
```

### Example: Custom String Matcher

<!-- compile-only -->
```go
type containsAnyMatcher struct {
    substrings []string
}

func ContainsAny(substrings ...string) *containsAnyMatcher {
    return &containsAnyMatcher{substrings: substrings}
}

func (m *containsAnyMatcher) Match(value string) specta.MatchResult {
    for _, substr := range m.substrings {
        if strings.Contains(value, substr) {
            return specta.MatchResult{
                Matched: true,
                Message: fmt.Sprintf("string contains '%s'", substr),
            }
        }
    }

    return specta.MatchResult{
        Matched: false,
        Message: fmt.Sprintf(
            "expected string to contain any of %v, got: %s",
            m.substrings, value,
        ),
    }
}
```

Usage:

<!-- skip-test -->
```go
specta.AssertThat(t, message, ContainsAny("error", "warning", "failure"))
```

### Example: Complex Struct Matcher

<!-- compile-only -->
```go
type validUserMatcher struct{}

func BeValidUser() *validUserMatcher {
    return &validUserMatcher{}
}

func (m *validUserMatcher) Match(user User) specta.MatchResult {
    var failures []string

    if user.ID == "" {
        failures = append(failures, "ID is empty")
    }

    if !strings.Contains(user.Email, "@") {
        failures = append(failures, "Email is invalid")
    }

    if user.Age < 0 || user.Age > 150 {
        failures = append(failures, "Age is out of range")
    }

    if len(failures) > 0 {
        return specta.MatchResult{
            Matched: false,
            Message: "User validation failed:\n  " +
                     strings.Join(failures, "\n  "),
        }
    }

    return specta.MatchResult{
        Matched: true,
        Message: "User is valid",
    }
}
```

Usage:

<!-- skip-test -->
```go
specta.AssertThat(t, user, BeValidUser())
```

## Testing Patterns

### Table-Driven Tests with Matchers

<!-- skip-test -->
```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        matcher specta.Matcher[string]
    }{
        {
            name:    "valid email",
            input:   "user@example.com",
            matcher: MatchRegex(`^[^@]+@[^@]+\.[^@]+$`),
        },
        {
            name:    "contains @",
            input:   "user@example.com",
            matcher: ContainString("@"),
        },
        {
            name:    "not empty",
            input:   "hello",
            matcher: specta.Not(IsEmpty()),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            specta.AssertThat(t, tt.input, tt.matcher)
        })
    }
}
```

### Reusable Matcher Compositions

Define matchers as package-level variables:

<!-- skip-test -->
```go
var (
    // Email matchers
    ValidEmail = specta.AllOf(
        ContainString("@"),
        MatchRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
    )

    CompanyEmail = specta.AllOf(
        ValidEmail,
        HaveSuffix("@company.com"),
    )

    // User matchers
    ActiveUser = MatchUser().
        WithStatus(specta.Equal("active"))

    AdminUser = MatchUser().
        WithRole(specta.Equal("admin")).
        WithPermissions(ContainAll("read", "write", "delete"))
)
```

Usage:

<!-- skip-test -->
```go
func TestUserCreation(t *testing.T) {
    user := CreateUser("alice@company.com")

    specta.AssertThat(t, user.Email, CompanyEmail)
    specta.AssertThat(t, user, ActiveUser)
}
```

### Error Testing Patterns

<!-- skip-test -->
```go
func TestErrorConditions(t *testing.T) {
    tests := []struct {
        name    string
        input   User
        wantErr specta.Matcher[error]
    }{
        {
            name:    "missing email",
            input:   User{Name: "Alice"},
            wantErr: specta.Not(BeNil()),
        },
        {
            name:    "invalid age",
            input:   User{Name: "Alice", Email: "alice@example.com", Age: -1},
            wantErr: specta.Not(BeNil()),
        },
        {
            name:    "valid user",
            input:   User{Name: "Alice", Email: "alice@example.com", Age: 30},
            wantErr: BeNil(),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUser(tt.input)
            specta.AssertThat(t, err, tt.wantErr)
        })
    }
}
```

## Performance Considerations

### Matcher Composition Overhead

Matchers are lightweight, but deeply nested compositions can add overhead:

<!-- skip-test -->
```go
// Fine for most tests
matcher := specta.AllOf(
    specta.Equal(x),
    specta.GreaterThan(y),
    specta.LessThan(z),
)

// Consider simplifying if performance-critical
// Or write a custom matcher
```

### Factory Generation Performance

Factories are fast, but building large object graphs has cost:

<!-- skip-test -->
```go
// Fine: a few hundred objects
users := make([]User, 100)
for i := range users {
    users[i] = factory.NewUser(p).Build()
}

// Consider caching for thousands of objects
// Or use a more efficient approach for bulk generation
```

### Benchmarking

Use Go's built-in benchmarking with matchers:

<!-- skip-test -->
```go
func BenchmarkUserValidation(b *testing.B) {
    p := specta.New()
    user := factory.NewUser(p).Build()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ValidateUser(user)
    }
}
```

## Integration with Standard Testing

### Using with testing.T

specta works seamlessly with `testing.T`:

<!-- skip-test -->
```go
func TestSomething(t *testing.T) {
    // Regular assertions
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }

    // specta matchers
    specta.AssertThat(t, value, specta.Equal(expected))

    // Mix and match as needed
}
```

### Subtests Organization

<!-- skip-test -->
```go
func TestUserWorkflow(t *testing.T) {
    p := specta.New()

    t.Run("creation", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        specta.AssertThat(t, user.ID, specta.Not(IsEmpty()))
    })

    t.Run("validation", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        err := ValidateUser(user)
        specta.AssertThat(t, err, BeNil())
    })
}
```

### Coverage Considerations

specta matchers count toward test coverage:

<!-- skip-test -->
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Generated code (`*_gen.go`) is excluded from coverage by build tags.

## Code Generation Advanced

### Custom Templates

The generator uses embedded templates. For custom behavior, fork and modify `cmd/main.go`.

### Type Checking

Generated code is type-checked before writing. For example, the generator ensures that methods like `WithName` compile correctly:

<!-- skip-test -->
```go
func (m *UserMatcher) WithName(matcher specta.Matcher[string]) *UserMatcher {
    // ...
}
```

If generation fails, it's usually because:
- Type in YAML doesn't match source
- Import paths are incorrect
- Unsupported type (fix: add to generator)

### Incremental Generation

<!-- skip-test -->
```bash
# Generate for specific types
go run github.com/james-w/specta/cmd/main.go -config specta.yaml -types User,Order

# Full regeneration
go run github.com/james-w/specta/cmd/main.go -config specta.yaml
```

## Best Practices Summary

1. **Use generated matchers** for custom types (don't write manually)
2. **Extract common matchers** as package variables
3. **Partial matching** - only assert what each test cares about
4. **Deterministic factories** - use `Primitives` for reproducibility
5. **Table-driven tests** - combine with matchers for clarity
6. **Regenerate after changes** - keep generated code in sync
7. **Commit generated files** - required for CI

## Next Steps

- [API Reference]({{< relref "/docs/api-reference/" >}}) - Complete API documentation
- [Examples]({{< relref "/docs/examples/" >}}) - Real-world patterns
- [GitHub Repository](https://github.com/james-w/specta) - Source code and issues
