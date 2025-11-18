---
title: "Core Matchers"
weight: 2
bookCollapseSection: false
---

# Core Matchers

Matchers are the foundation of specta. They test values and provide detailed, structured failure messages.

## Basic Matchers

### Equality

```go
AssertThat(t, value, Equal(42))
AssertThat(t, user, DeepEqual(expectedUser))
```

### Comparison

```go
AssertThat(t, age, GreaterThan(18))
AssertThat(t, score, LessThan(100))
AssertThat(t, value, GreaterThanOrEqual(0))
AssertThat(t, count, LessThanOrEqual(10))
```

### Nil Checking

```go
AssertThat(t, ptr, BeNil())
AssertThat(t, value, NotBeNil())
```

## String Matchers

### Substring Matching

```go
AssertThat(t, message, ContainString("error"))
AssertThat(t, filename, HavePrefix("test_"))
AssertThat(t, email, HaveSuffix(".com"))
```

### Pattern Matching

```go
AssertThat(t, email, MatchRegex(`^[a-z]+@[a-z]+\.[a-z]+$`))
```

### Empty Strings

```go
AssertThat(t, name, Not(BeEmpty()))
```

## Collection Matchers

### Size and Emptiness

```go
AssertThat(t, list, HaveLength(5))
AssertThat(t, emptyList, BeEmpty())
```

### Membership

```go
AssertThat(t, list, Contain("apple"))
AssertThat(t, list, ContainAll("apple", "banana", "cherry"))
AssertThat(t, list, ContainAny("apple", "durian"))
```

### Element Matching

```go
// All elements must match the matcher
numbers := []int{2, 4, 6, 8}
AssertThat(t, numbers, EachMatch(func(n int) bool {
    return n%2 == 0
}))
```

## Composing Matchers

### AllOf (AND logic)

All matchers must pass:

```go
AssertThat(t, email, AllOf(
    ContainString("@"),
    HaveSuffix(".com"),
    HavePrefix("user"),
))
```

### AnyOf (OR logic)

At least one matcher must pass:

```go
AssertThat(t, status, AnyOf(
    Equal("active"),
    Equal("pending"),
    Equal("processing"),
))
```

### Not (Negation)

Inverts a matcher:

```go
AssertThat(t, name, Not(BeEmpty()))
AssertThat(t, list, Not(Contain("forbidden")))
```

## Understanding Matcher Errors

When a matcher fails, specta provides structured error messages with visual indicators:

```
Value does not match:
  ✗ AllOf:
    ✓ ContainString("@")
    ✗ HaveSuffix(".com")
      Expected: string ending with ".com"
      Got:      "user@example.org"
    ✓ HavePrefix("user")
```

**Symbols:**
- `✓` - Matcher passed
- `✗` - Matcher failed
- `~` - Matcher not evaluated (due to short-circuit logic)

## Building Reusable Matchers

Extract common matcher patterns:

```go
// Define reusable matchers
var (
    validEmail = AllOf(
        ContainString("@"),
        MatchRegex(`^[^@]+@[^@]+\.[^@]+$`),
    )

    positiveInteger = AllOf(
        GreaterThan(0),
        // Add custom matcher for integer check
    )
)

// Use in tests
func TestUser(t *testing.T) {
    AssertThat(t, user.Email, validEmail)
    AssertThat(t, user.Age, positiveInteger)
}
```

## Next Steps

Now that you understand core matchers, learn how to:
- Generate matchers for [your custom types](/docs/matchers-for-your-types/)
- Create [test data with factories](/docs/factories/)
- Combine [factories and matchers](/docs/factories-and-matchers/) for powerful tests
