---
title: "Core Matchers"
weight: 2
---

<!-- setup
package doctest

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"
)

var _ = testing.Verbose

type User struct {
	Name  string
	Email string
	Age   int
}

type YourType struct{}

var (
	// Value matchers
	value        = 42
	expected     = 42
	age          = 25
	score        = 85
	count        = 5
	zeroInt      = 0

	// Boolean
	active       = true
	verified     = false

	// String matchers
	message      = "error occurred"
	filename     = "test_file.go"
	email        = "user@example.com"
	name         = "Alice"
	status       = "active"

	// Collection matchers
	list         = []string{"apple", "banana", "cherry"}
	emptyList    = []string{}
	numbers      = []int{2, 4, 6, 8}

	// Map matchers
	userMap      = map[string]int{"alice": 30, "bob": 25}
	emptyMap     = map[string]int{}

	// Pointer matchers
	ptr          *YourType
	nonNilPtr    = &YourType{}

	// Error matchers
	err          = errors.New("operation failed")
	nilErr       error
	baseErr      = errors.New("base error")
	wrappedErr   = fmt.Errorf("wrapped: %w", baseErr)

	// Time matchers
	now          = time.Now()
	past         = time.Now().Add(-1 * time.Hour)
	future       = time.Now().Add(30 * time.Minute)
	oneHour      = time.Hour

	// User examples
	user         = User{Name: "Alice", Email: "alice@example.com", Age: 30}
	expectedUser = User{Name: "Alice", Email: "alice@example.com", Age: 30}

	// Compiled patterns
	emailPattern = regexp.MustCompile(`^[a-z]+@[a-z]+\.[a-z]+$`)
)
-->

# Core Matchers

Matchers are the foundation of specta. They test values and provide detailed, structured failure messages.

## Value Matchers

### Equality

Test if values are equal:

```go
specta.AssertThat(t, value, specta.Equal(42))
specta.AssertThat(t, user, specta.DeepEqual(expectedUser))
specta.AssertThat(t, value, specta.Is(42))  // Alias for Equal
```

**Use `Equal` for comparable types, `DeepEqual` for structs/slices/any type.**

### Comparison

Compare numeric values:

```go
specta.AssertThat(t, age, specta.GreaterThan(18))
specta.AssertThat(t, score, specta.LessThan(100))
specta.AssertThat(t, age, specta.GreaterThanOrEqual(21))
specta.AssertThat(t, score, specta.LessThanOrEqual(100))
```

Works with all numeric types (int, float, uint variants).

### Zero Value

Check if a value is its type's zero value:

```go
specta.AssertThat(t, zeroInt, specta.IsZero[int]())
specta.AssertThat(t, name, specta.Not(specta.IsZero[string]()))
```

## Boolean Matchers

Test boolean values:

```go
specta.AssertThat(t, active, specta.IsTrue())
specta.AssertThat(t, verified, specta.IsFalse())
```

Simpler and clearer than `Equal(true)` or `Equal(false)`.

## String Matchers

### Substring Matching

```go
specta.AssertThat(t, message, specta.Contains("error"))
specta.AssertThat(t, filename, specta.HasPrefix("test_"))
specta.AssertThat(t, email, specta.HasSuffix(".com"))
```

### Pattern Matching

Match against regular expressions:

```go
// Compile pattern on each call
specta.AssertThat(t, email, specta.MatchesRegex(`^[a-z]+@[a-z]+\.[a-z]+$`))

// Use pre-compiled pattern (more efficient for repeated use)
specta.AssertThat(t, email, specta.MatchesPattern(emailPattern))
```

### Empty Strings

```go
specta.AssertThat(t, name, specta.Not(specta.Equal("")))
```

## Collection Matchers

### Size and Emptiness

```go
specta.AssertThat(t, list, specta.HasSize[string](3))
specta.AssertThat(t, emptyList, specta.IsEmpty[string]())
specta.AssertThat(t, list, specta.IsNotEmpty[string]())
```

### Membership

```go
specta.AssertThat(t, list, specta.ContainsElement("apple"))
specta.AssertThat(t, list, specta.ContainsAllElements("apple", "banana", "cherry"))
specta.AssertThat(t, list, specta.ContainsAnyElement("apple", "durian"))
```

### Element Matching

Match patterns across collection elements:

```go
// All elements must match
specta.AssertThat(t, numbers, specta.Every(specta.GreaterThan(0)))

// At least one element must match
specta.AssertThat(t, numbers, specta.Any(specta.GreaterThan(5)))

// No elements should match
specta.AssertThat(t, numbers, specta.None(specta.LessThan(0)))
```

## Map Matchers

Test map contents:

```go
// Check for key existence
specta.AssertThat(t, userMap, specta.HasKey[string, int]("alice"))

// Check for value existence
specta.AssertThat(t, userMap, specta.HasValue[string, int](30))

// Check for specific key-value pair
specta.AssertThat(t, userMap, specta.HasEntry[string, int]("alice", 30))

// Check map size
specta.AssertThat(t, userMap, specta.MapHasSize[string, int](2))
specta.AssertThat(t, emptyMap, specta.MapHasSize[string, int](0))
```

## Pointer Matchers

### Nil Checks

```go
specta.AssertThat(t, ptr, specta.IsNil[YourType]())
specta.AssertThat(t, nonNilPtr, specta.IsNotNil[YourType]())
```

### Value Matching

Match the value a pointer points to:

```go
value := 42
ptr := &value

specta.AssertThat(t, ptr, specta.PointsTo(specta.Equal(42)))
specta.AssertThat(t, ptr, specta.PointsTo(specta.GreaterThan(40)))
```

## Error Matchers

### Basic Error Checks

```go
specta.AssertThat(t, err, specta.IsError())
specta.AssertThat(t, nilErr, specta.NoErr())
```

### Error Messages

Check error message content:

```go
specta.AssertThat(t, err, specta.ErrorContains("operation failed"))
specta.AssertThat(t, err, specta.ErrorContains("failed"))
```

### Error Chain Matching

Use Go's error chain utilities:

```go
// Check if error is in the error chain (uses errors.Is)
specta.AssertThat(t, wrappedErr, specta.ErrorIs(baseErr))
```

## Time Matchers

### Time Comparison

```go
specta.AssertThat(t, future, specta.After(now))
specta.AssertThat(t, past, specta.Before(now))
specta.AssertThat(t, now, specta.Between(past, future))
```

### Time Tolerance

Check if times are close enough (useful for testing timestamps with slight variations):

```go
// future is 30 minutes from now, check it's within 1 hour
specta.AssertThat(t, future, specta.WithinDuration(now, oneHour))
```

## Field Extraction

Extract and match computed or derived values:

```go
// Match on a computed field
specta.AssertThat(t, user,
    specta.Field("FullName",
        func(u User) string { return u.Name },
        specta.Equal("Alice")))

// Match on method result
specta.AssertThat(t, user,
    specta.Field("Age",
        func(u User) int { return u.Age },
        specta.GreaterThan(18)))

// Combine with composition
specta.AssertThat(t, user,
    specta.AllOf(
        specta.Field("Name", func(u User) string { return u.Name }, specta.Contains("A")),
        specta.Field("Age", func(u User) int { return u.Age }, specta.GreaterThan(18)),
    ))
```

Useful for:
- Testing unexported fields (via getter methods)
- Computed properties
- Complex validation logic

## Composing Matchers

### AllOf (AND logic)

All matchers must pass:

```go
specta.AssertThat(t, email, specta.AllOf(
    specta.Contains("@"),
    specta.HasSuffix(".com"),
    specta.HasPrefix("user"),
))
```

### AnyOf (OR logic)

At least one matcher must pass:

```go
specta.AssertThat(t, status, specta.AnyOf(
    specta.Equal("active"),
    specta.Equal("pending"),
    specta.Equal("processing"),
))
```

### Not (Negation)

Inverts a matcher:

```go
specta.AssertThat(t, name, specta.Not(specta.Equal("")))
specta.AssertThat(t, list, specta.Not(specta.ContainsElement("forbidden")))
```

## Understanding Matcher Errors

When a matcher fails, specta provides structured error messages with visual indicators:

```
Value does not match:
  ✗ AllOf:
    ✓ Contains("@")
    ✗ HasSuffix(".com")
      Expected: string ending with ".com"
      Got:      "user@example.org"
    ✓ HasPrefix("user")
```

**Symbols:**
- `✓` - Matcher passed
- `✗` - Matcher failed
- `~` - Matcher not evaluated (due to short-circuit logic)

## Building Reusable Matchers

Building reusable matchers lets you compose test assertions and create domain-specific validation logic that you can use across your test suite.

### Simple Value Matchers

Extract common patterns for primitive types:

```go
import "github.com/james-w/specta"

// Define reusable matchers
var (
    validEmail = specta.AllOf(
        specta.Contains("@"),
        specta.MatchesRegex(`^[^@]+@[^@]+\.[^@]+$`),
    )

    positiveInteger = specta.GreaterThan(0)
)

// Use in tests
func TestUser(t *testing.T) {
    specta.AssertThat(t, user.Email, validEmail)
    specta.AssertThat(t, user.Age, positiveInteger)
}
```

### Custom Type Matchers

You can implement the `Matcher[T]` interface to create matchers for your own types:

<!-- skip-test -->
```go
// User type from your application
type User struct {
    Name  string
    Email string
    Age   int
}

// UserMatcher allows partial matching on User fields
type UserMatcher struct {
    nameMatcher  specta.Matcher[string]
    emailMatcher specta.Matcher[string]
    ageMatcher   specta.Matcher[int]
}

// UserMatches creates a new matcher builder for User
func UserMatches() UserMatcher {
    return UserMatcher{}
}

// Name sets the matcher for the Name field
func (m UserMatcher) Name(matcher specta.Matcher[string]) UserMatcher {
    m.nameMatcher = matcher
    return m
}

// Email sets the matcher for the Email field
func (m UserMatcher) Email(matcher specta.Matcher[string]) UserMatcher {
    m.emailMatcher = matcher
    return m
}

// Age sets the matcher for the Age field
func (m UserMatcher) Age(matcher specta.Matcher[int]) UserMatcher {
    m.ageMatcher = matcher
    return m
}

// Matches implements the Matcher[User] interface
func (m UserMatcher) Matches(actual User) specta.MatchResult {
    // Only check fields that have matchers set (partial matching)
    if m.nameMatcher != nil {
        if result := m.nameMatcher.Matches(actual.Name); !result.Matched {
            return specta.MatchResult{
                Matched: false,
                Message: "Name: " + result.Message,
            }
        }
    }

    if m.emailMatcher != nil {
        if result := m.emailMatcher.Matches(actual.Email); !result.Matched {
            return specta.MatchResult{
                Matched: false,
                Message: "Email: " + result.Message,
            }
        }
    }

    if m.ageMatcher != nil {
        if result := m.ageMatcher.Matches(actual.Age); !result.Matched {
            return specta.MatchResult{
                Matched: false,
                Message: "Age: " + result.Message,
            }
        }
    }

    return specta.MatchResult{Matched: true}
}
```

Now you can use it to create reusable, composable matchers:

<!-- skip-test -->
```go
func TestUserCreation(t *testing.T) {
    user := CreateUser("Alice", "alice@example.com", 30)

    // Partial matching - only assert what matters
    specta.AssertThat(t, user, UserMatches().
        Name(specta.Equal("Alice")).
        Email(validEmail))  // Compose with reusable matcher!
}

func TestUserValidation(t *testing.T) {
    user := ValidateUser(input)

    // Different test, different assertions
    specta.AssertThat(t, user, UserMatches().
        Email(validEmail).
        Age(positiveInteger))
}
```

**Key benefits:**
- **Partial matching**: Only specified fields are validated
- **Composable**: Mix custom matchers with built-in ones
- **Reusable**: Define once, use across all User tests
- **Type-safe**: Compile-time checking for field types

### Building a Matcher Library

As your codebase grows, you'll want a library of reusable matchers:

<!-- skip-test -->
```go
// matchers/user.go
package matchers

import "github.com/james-w/specta"

// Common user matchers
var (
    ValidEmail = specta.AllOf(
        specta.Contains("@"),
        specta.MatchesRegex(`^[^@]+@[^@]+\.[^@]+$`),
    )

    Adult = specta.GreaterThanOrEqual(18)

    ValidUser = UserMatches().
        Email(ValidEmail).
        Age(Adult)

    AdminUser = UserMatches().
        Email(specta.HasSuffix("@company.com")).
        Age(Adult)
)

// Use across tests
func TestCreateUser(t *testing.T) {
    user := CreateUser(input)
    specta.AssertThat(t, user, matchers.ValidUser)
}

func TestPromoteToAdmin(t *testing.T) {
    admin := PromoteToAdmin(user)
    specta.AssertThat(t, admin, matchers.AdminUser)
}
```

This pattern—building reusable, composable matchers for your domain types—is powerful. But there's a problem...

### Why Code Generation?

The manual approach works, but it's **tedious and error-prone**:

**For each type, you need to write:**
1. A matcher struct with optional field matchers
2. A constructor function
3. Fluent builder methods for each field (with correct return type)
4. The `Matches()` implementation with partial matching logic
5. Proper error messages for each field

**For a 5-field struct, that's ~50-70 lines of boilerplate** — all mechanical, repetitive code.

**And when you add a field to your type:**
- Update the matcher struct
- Add a new builder method
- Update the Matches() logic

There's a better way: **automatic code generation**.

The next section shows how specta can generate all this boilerplate for you, keeping your matchers in sync with your types automatically.

## Next Steps

Now that you understand how matchers work—including how to build them manually—learn how to:
- [Automatically generate matchers]({{< relref "/docs/matchers-for-your-types/" >}}) for your types (recommended!)
- Create [test data with factories]({{< relref "/docs/factories/" >}})
- Combine [factories and matchers]({{< relref "/docs/factories-and-matchers/" >}}) for powerful tests
