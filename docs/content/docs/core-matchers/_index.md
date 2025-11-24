---
title: "Core Matchers"
weight: 2
---

<!-- setup
package doctest

import (
	"testing"
)

var _ = testing.Verbose

type User struct {
	Name  string
	Email string
	Age   int
}

type YourType struct{}

var (
	value        = 42
	user         = User{Name: "Alice", Email: "alice@example.com", Age: 30}
	expectedUser = User{Name: "Alice", Email: "alice@example.com", Age: 30}
	age          = 25
	score        = 85
	count        = 5
	ptr          *YourType
	nonNilPtr    = &YourType{}
	message      = "error occurred"
	filename     = "test_file.go"
	email        = "user@example.com"
	name         = "Alice"
	status       = "active"
	list         = []string{"apple", "banana", "cherry"}
	emptyList    = []string{}
	numbers      = []int{2, 4, 6, 8}
)
-->

# Core Matchers

Matchers are the foundation of specta. They test values and provide detailed, structured failure messages.

## Basic Matchers

### Equality

```go
specta.AssertThat(t, value, specta.Equal(42))
specta.AssertThat(t, user, specta.DeepEqual(expectedUser))
```

### Comparison

```go
specta.AssertThat(t, age, specta.GreaterThan(18))
specta.AssertThat(t, score, specta.LessThan(100))
specta.AssertThat(t, value, specta.GreaterThanOrEqual(0))
```

### Nil Checking

```go
specta.AssertThat(t, ptr, specta.IsNil[YourType]())
specta.AssertThat(t, nonNilPtr, specta.IsNotNil[YourType]())
```

## String Matchers

### Substring Matching

```go
specta.AssertThat(t, message, specta.Contains("error"))
specta.AssertThat(t, filename, specta.HasPrefix("test_"))
specta.AssertThat(t, email, specta.HasSuffix(".com"))
```

### Pattern Matching

```go
specta.AssertThat(t, email, specta.MatchesRegex(`^[a-z]+@[a-z]+\.[a-z]+$`))
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
```

### Membership

```go
specta.AssertThat(t, list, specta.ContainsElement("apple"))
specta.AssertThat(t, list, specta.ContainsAllElements("apple", "banana", "cherry"))
specta.AssertThat(t, list, specta.ContainsAnyElement("apple", "durian"))
```

### Element Matching

```go
// All elements must be greater than zero
numbers := []int{2, 4, 6, 8}
specta.AssertThat(t, numbers, specta.Every(specta.GreaterThan(0)))
```

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

Extract common matcher patterns:

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

## Next Steps

Now that you understand core matchers, learn how to:
- Generate matchers for [your custom types]({{< relref "/docs/matchers-for-your-types/" >}})
- Create [test data with factories]({{< relref "/docs/factories/" >}})
- Combine [factories and matchers]({{< relref "/docs/factories-and-matchers/" >}}) for powerful tests
