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
	p specta.Source
}

func (b userBuilder) Build() User {
	return User{ID: "test-id", Name: "Test", Email: "test@example.com", Age: 30, Active: true}
}

func (f factoryType) NewUser(p specta.Source) userBuilder {
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

// Order types for table-driven test example
type Order struct {
	Items int
	Total float64
}

type OrderResult struct {
	Status   string
	Discount float64
}

func ProcessOrder(o Order) OrderResult {
	if o.Items == 0 {
		return OrderResult{Status: "rejected"}
	}
	if o.Total >= 100 {
		return OrderResult{Status: "confirmed", Discount: o.Total * 0.1}
	}
	return OrderResult{Status: "confirmed"}
}

type OrderResultMatcher struct{}

func MatchOrderResult() *OrderResultMatcher { return &OrderResultMatcher{} }

func (m *OrderResultMatcher) WithStatus(matcher specta.Matcher[string]) *OrderResultMatcher {
	return m
}

func (m *OrderResultMatcher) WithDiscount(matcher specta.Matcher[float64]) *OrderResultMatcher {
	return m
}

func (m *OrderResultMatcher) Matches(r OrderResult) specta.MatchResult {
	return specta.MatchResult{Matched: true}
}

// Error types for error testing example
var ErrNotFound = fmt.Errorf("not found")

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s", e.Field)
}

var err error = &ValidationError{Field: "email"}
-->

# Advanced Topics

Deep dives into advanced patterns and techniques.

## Custom Matchers

While code generation handles most cases, sometimes you need a custom matcher.

### Implementing specta.Matcher[T]

The `Matcher` interface is defined as:

<!-- compile-only -->
```go
type Matcher[T any] interface {
    Matches(actual T) MatchResult
}

type MatchResult struct {
    Matched  bool     // Whether the match succeeded
    Message  string   // Human-readable description
    Details  []string // Additional failure details
    Expected any      // Expected value (for diffs)
    Actual   any      // Actual value (for diffs)
    Path     string   // Field path for nested failures
}
```

For simple matchers, you only need to set `Matched` and `Message`. The other fields enhance error output:
- `Details` for multi-line failure reasons
- `Expected`/`Actual` for structured diffs
- `Path` for nested struct field errors

### Using MatcherFunc for Simple Matchers

For one-off matchers, you can use `MatcherFunc` instead of defining a struct:

<!-- compile-only -->
```go
// Simple inline matcher
containsHello := specta.MatcherFunc(func(s string) specta.MatchResult {
    if strings.Contains(s, "hello") {
        return specta.MatchResult{Matched: true}
    }
    return specta.MatchResult{
        Matched: false,
        Message: fmt.Sprintf("expected string to contain 'hello', got: %s", s),
    }
})

specta.AssertThat(t, "hello world", containsHello)
```

Use `MatcherFunc` for simple, one-off matchers. For reusable matchers, define a proper type as shown in the examples below.

### Example: Custom String Matcher

<!-- compile-only -->
```go
type containsAnyMatcher struct {
    substrings []string
}

func ContainsAny(substrings ...string) *containsAnyMatcher {
    return &containsAnyMatcher{substrings: substrings}
}

func (m *containsAnyMatcher) Matches(actual string) specta.MatchResult {
    for _, substr := range m.substrings {
        if strings.Contains(actual, substr) {
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
            m.substrings, actual,
        ),
    }
}
```

Usage:

<!-- compile-only -->
```go
specta.AssertThat(t, message, ContainsAny("error", "warning", "failure"))
```

### Example: Complex Struct Matcher

<!-- compile-only -->
```go
// BeValidUser returns a matcher that validates multiple user fields.
// This demonstrates composing matchers - we manually extract field values
// but use matchers for the actual checks instead of if statements.
func BeValidUser() specta.Matcher[User] {
    return specta.MatcherFunc(func(actual User) specta.MatchResult {
        // Define matchers for each field
        idMatcher := specta.Not(specta.IsEmpty[string]())
        emailMatcher := specta.Contains("@")
        ageMatcher := specta.AllOf(
            specta.GreaterThanOrEqual(0),
            specta.LessThanOrEqual(150),
        )

        // Apply matchers to extracted field values
        if result := idMatcher.Matches(actual.ID); !result.Matched {
            result.Path = "ID"
            return result
        }

        if result := emailMatcher.Matches(actual.Email); !result.Matched {
            result.Path = "Email"
            return result
        }

        if result := ageMatcher.Matches(actual.Age); !result.Matched {
            result.Path = "Age"
            return result
        }

        return specta.MatchResult{
            Matched: true,
            Message: "User is valid",
        }
    })
}
```

Usage:

<!-- compile-only -->
```go
specta.AssertThat(t, user, BeValidUser())
```

### Field() Matcher for Specific Fields

Instead of writing a full custom matcher, use `Field()` to test specific fields:

<!-- skip-test -->
```go
// Test that user's age is positive
specta.AssertThat(t, user,
    specta.Field("Age", func(u User) int { return u.Age }, specta.GreaterThan(0)))

// Test computed values
specta.AssertThat(t, account,
    specta.Field("balance*2", func(a Account) int { return a.Balance * 2 }, specta.Equal(2000)))

// Test nested fields
specta.AssertThat(t, order,
    specta.Field("User.Email", func(o Order) string { return o.User.Email },
        specta.Contains("@example.com")))
```

The first parameter is just a label for error messages - you can describe what you're testing.

Assign Field matchers to variables for reuse:

<!-- compile-only -->
```go
var (
    PositiveAge = specta.Field("Age", func(u User) int { return u.Age }, specta.GreaterThan(0))
    ValidEmail  = specta.Field("Email", func(u User) string { return u.Email }, specta.Contains("@"))
)

// Reuse across tests
specta.AssertThat(t, user1, PositiveAge)
specta.AssertThat(t, user2, specta.AllOf(PositiveAge, ValidEmail))
```

## Testing Patterns

### Table-Driven Tests with Matchers

The matcher column lets you express different expected behaviors per test case:

<!-- compile-only -->
```go
func TestProcessOrder(t *testing.T) {
    tests := []struct {
        name    string
        order   Order
        matcher specta.Matcher[OrderResult]
    }{
        {
            name:    "standard order",
            order:   Order{Items: 3, Total: 50.00},
            matcher: MatchOrderResult().WithStatus(specta.Equal("confirmed")),
        },
        {
            name:    "large order gets discount",
            order:   Order{Items: 10, Total: 500.00},
            matcher: MatchOrderResult().
                WithDiscount(specta.GreaterThan(0.0)).
                WithStatus(specta.Equal("confirmed")),
        },
        {
            name:    "empty order rejected",
            order:   Order{Items: 0},
            matcher: MatchOrderResult().WithStatus(specta.Equal("rejected")),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ProcessOrder(tt.order)
            specta.AssertThat(t, result, tt.matcher)
        })
    }
}
```

### Reusable Matcher Compositions

Define matchers as package-level variables:

<!-- compile-only -->
```go
var (
    // Email matchers
    ValidEmail = specta.AllOf(
        specta.Contains("@"),
        specta.MatchesRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
    )

    CompanyEmail = specta.AllOf(
        ValidEmail,
        specta.HasSuffix("@company.com"),
    )

    // User matchers
    ActiveUser = MatchUser().
        WithStatus(specta.Equal("active"))

    AdminUser = MatchUser().
        WithRole(specta.Equal("admin")).
        WithPermissions(specta.ContainsAllElements[string]("read", "write", "delete"))
)
```

Usage:

<!-- compile-only -->
```go
func TestUserCreation(t *testing.T) {
    user := CreateUser("alice@company.com")

    specta.AssertThat(t, user.Email, CompanyEmail)
    specta.AssertThat(t, user, ActiveUser)
}
```

### Error Testing Patterns

<!-- compile-only -->
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
            wantErr: specta.IsError(),  // Expect an error
        },
        {
            name:    "invalid age",
            input:   User{Name: "Alice", Email: "alice@example.com", Age: -1},
            wantErr: specta.IsError(),  // Expect an error
        },
        {
            name:    "valid user",
            input:   User{Name: "Alice", Email: "alice@example.com", Age: 30},
            wantErr: specta.NoErr(),  // No error expected
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

For more specific error checks beyond presence, use `ErrorContains`, `ErrorIs`, or `ErrorAs`:

<!-- compile-only -->
```go
// Check error message
specta.AssertThat(t, err, specta.ErrorContains("invalid"))

// Check error identity (errors.Is)
specta.AssertThat(t, err, specta.ErrorIs(ErrNotFound))

// Check error type (errors.As)
var validationErr *ValidationError
specta.AssertThat(t, err, specta.ErrorAs(&validationErr))
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

<!-- compile-only -->
```go
func TestUserWorkflow(t *testing.T) {
    p := specta.New()

    t.Run("creation", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        specta.AssertThat(t, user.ID, specta.Not(specta.IsEmpty[string]()))
    })

    t.Run("validation", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        err := ValidateUser(user)
        specta.AssertThat(t, err, specta.NoErr())
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

### Extending Generated Code

The generator uses embedded templates that cover common use cases. You can extend generated matchers and factories with custom code in the same package - see [Combining with Factory Recipes]({{< relref "/docs/property-based-testing/#combining-with-factory-recipes" >}}) for examples of adding custom matchers and factory methods alongside generated code.

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
- Custom defaults or extensions are out of sync with generated code

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
