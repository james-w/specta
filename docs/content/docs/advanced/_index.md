---
title: "Advanced Topics"
weight: 7
bookCollapseSection: false
---

# Advanced Topics

Deep dives into advanced patterns and techniques.

## Custom Matchers

While code generation handles most cases, sometimes you need a custom matcher.

### Implementing Matcher[T]

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

```go
type containsAnyMatcher struct {
    substrings []string
}

func ContainsAny(substrings ...string) *containsAnyMatcher {
    return &containsAnyMatcher{substrings: substrings}
}

func (m *containsAnyMatcher) Match(value string) MatchResult {
    for _, substr := range m.substrings {
        if strings.Contains(value, substr) {
            return MatchResult{
                Matched: true,
                Message: fmt.Sprintf("string contains '%s'", substr),
            }
        }
    }

    return MatchResult{
        Matched: false,
        Message: fmt.Sprintf(
            "expected string to contain any of %v, got: %s",
            m.substrings, value,
        ),
    }
}
```

Usage:

```go
AssertThat(t, message, ContainsAny("error", "warning", "failure"))
```

### Example: Complex Struct Matcher

```go
type validUserMatcher struct{}

func BeValidUser() *validUserMatcher {
    return &validUserMatcher{}
}

func (m *validUserMatcher) Match(user User) MatchResult {
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
        return MatchResult{
            Matched: false,
            Message: "User validation failed:\n  " +
                     strings.Join(failures, "\n  "),
        }
    }

    return MatchResult{
        Matched: true,
        Message: "User is valid",
    }
}
```

Usage:

```go
AssertThat(t, user, BeValidUser())
```

## Testing Patterns

### Table-Driven Tests with Matchers

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        matcher Matcher[string]
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
            matcher: Not(BeEmpty()),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            AssertThat(t, tt.input, tt.matcher)
        })
    }
}
```

### Reusable Matcher Compositions

Define matchers as package-level variables:

```go
var (
    // Email matchers
    ValidEmail = AllOf(
        ContainString("@"),
        MatchRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
    )

    CompanyEmail = AllOf(
        ValidEmail,
        HaveSuffix("@company.com"),
    )

    // User matchers
    ActiveUser = MatchUser().
        WithStatus(Equal("active"))

    AdminUser = MatchUser().
        WithRole(Equal("admin")).
        WithPermissions(ContainAll("read", "write", "delete"))
)
```

Usage:

```go
func TestUserCreation(t *testing.T) {
    user := CreateUser("alice@company.com")

    AssertThat(t, user.Email, CompanyEmail)
    AssertThat(t, user, ActiveUser)
}
```

### Error Testing Patterns

```go
func TestErrorConditions(t *testing.T) {
    tests := []struct {
        name    string
        input   User
        wantErr Matcher[error]
    }{
        {
            name:    "missing email",
            input:   User{Name: "Alice"},
            wantErr: Not(BeNil()),
        },
        {
            name:    "invalid age",
            input:   User{Name: "Alice", Email: "alice@example.com", Age: -1},
            wantErr: AllOf(
                Not(BeNil()),
                ErrorContains("age"),
            ),
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
            AssertThat(t, err, tt.wantErr)
        })
    }
}
```

## Performance Considerations

### Matcher Composition Overhead

Matchers are lightweight, but deeply nested compositions can add overhead:

```go
// Fine for most tests
matcher := AllOf(
    Equal(x),
    GreaterThan(y),
    LessThan(z),
)

// Consider simplifying if performance-critical
// Or write a custom matcher
```

### Factory Generation Performance

Factories are fast, but building large object graphs has cost:

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

```go
func BenchmarkUserValidation(b *testing.B) {
    p := specta.NewGen()
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

```go
func TestSomething(t *testing.T) {
    // Regular assertions
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }

    // specta matchers
    AssertThat(t, value, Equal(expected))

    // Mix and match as needed
}
```

### Subtests Organization

```go
func TestUserWorkflow(t *testing.T) {
    p := specta.NewGen()

    t.Run("creation", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        AssertThat(t, user.ID, Not(BeEmpty()))
    })

    t.Run("validation", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        err := ValidateUser(user)
        AssertThat(t, err, BeNil())
    })
}
```

### Coverage Considerations

specta matchers count toward test coverage:

```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Generated code (`*_gen.go`) is excluded from coverage by build tags.

## Code Generation Advanced

### Custom Templates

The generator uses embedded templates. For custom behavior, fork and modify `cmd/main.go`.

### Type Checking

Generated code is type-checked before writing:

```go
// Generator ensures this compiles:
func (m *UserMatcher) WithName(matcher Matcher[string]) *UserMatcher {
    // ...
}
```

If generation fails, it's usually because:
- Type in YAML doesn't match source
- Import paths are incorrect
- Unsupported type (fix: add to generator)

### Incremental Generation

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

- [API Reference](/docs/api-reference/) - Complete API documentation
- [Examples](/docs/examples/) - Real-world patterns
- [GitHub Repository](https://github.com/james-w/specta) - Source code and issues
