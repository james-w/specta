---
title: "Introduction"
weight: 1
---

<!-- setup
package doctest

import (
	"testing"
	"github.com/james-w/specta"
)

var _ = testing.Verbose

type User struct {
	Name  string
	Email string
	Age   int
}

// Mock matcher builder for User
type UserMatcher struct {
	nameMatcher  specta.Matcher[string]
	emailMatcher specta.Matcher[string]
}

func MatchUser() *UserMatcher {
	return &UserMatcher{}
}

func (m *UserMatcher) WithName(matcher specta.Matcher[string]) *UserMatcher {
	m.nameMatcher = matcher
	return m
}

func (m *UserMatcher) WithEmail(matcher specta.Matcher[string]) *UserMatcher {
	m.emailMatcher = matcher
	return m
}

func (m *UserMatcher) Matches(u User) specta.MatchResult {
	if m.nameMatcher != nil {
		if result := m.nameMatcher.Matches(u.Name); !result.Matched {
			return specta.MatchResult{Matched: false, Message: "name: " + result.Message}
		}
	}
	if m.emailMatcher != nil {
		if result := m.emailMatcher.Matches(u.Email); !result.Matched {
			return specta.MatchResult{Matched: false, Message: "email: " + result.Message}
		}
	}
	return specta.MatchResult{Matched: true}
}

var (
	value    = 42
	expected = 42
	age      = 25
	active   = true
	list     = []string{"a", "b", "c"}
	name     = "Alice"
	user     = User{Name: "Alice", Email: "alice@example.com", Age: 30}
	userPtr  = &User{Name: "Alice", Email: "alice@example.com", Age: 30}
)

func registerUser(email string, age int) User {
	return User{Name: "Alice", Email: email, Age: age}
}
-->

# Introduction to specta

specta is a Go testing library that emphasizes **composition and reuse** through matchers and test data factories.

## Why specta?

Traditional Go tests often suffer from:
- **Duplicated assertions** across test cases
- **Brittle tests** that break when adding new fields
- **Verbose error messages** that don't clearly show what failed
- **Inconsistent test data** generation

specta solves these problems by providing:
- **Composable matchers** - Build complex assertions from simple pieces
- **Partial matching** - Assert only the fields that matter
- **Clear error messages** - Structured diffs with visual indicators
- **Deterministic factories** - Generate reproducible test data

## Installation

```bash
go get github.com/james-w/specta
```

## Quick Start

Here's a simple example showing the power of matchers:

```go
func TestUserValidation(t *testing.T) {
    user := registerUser("alice@example.com", 30)

    // Match individual values with built-in matchers
    specta.AssertThat(t, user.Age, specta.GreaterThan(18))
    specta.AssertThat(t, user.Email, specta.Contains("@example.com"))

    // Combine matchers with AllOf
    specta.AssertThat(t, user.Email, specta.AllOf(
        specta.HasPrefix("alice"),
        specta.HasSuffix(".com"),
        specta.Contains("@"),
    ))
}
```

**When tests fail**, you see clear error messages:

```
user.Age: expected value > 18 but got 15
user.Email: expected string to contain "@example.com" but got "alice@test.org"
```

## Core Concepts

### Matchers

Matchers are reusable predicates that test values and provide detailed failure messages:

```go
specta.AssertThat(t, value, specta.Equal(expected))
specta.AssertThat(t, age, specta.GreaterThanOrEqual(21))
specta.AssertThat(t, active, specta.IsTrue())
specta.AssertThat(t, name, specta.Not(specta.Equal("")))
```

### AssertThat vs RequireThat

specta provides two assertion functions:

**`AssertThat`** - Reports failures but continues test execution:

```go
specta.AssertThat(t, user.Name, specta.Equal("Alice"))
specta.AssertThat(t, user.Age, specta.GreaterThan(18))
// Both assertions run, even if first fails
```

**`RequireThat`** - Stops test execution immediately on failure:

```go
specta.RequireThat(t, userPtr, specta.IsNotNil[User]())
// Test stops here if userPtr is nil - prevents panic below
specta.AssertThat(t, userPtr.Name, specta.Equal("Alice"))
```

**When to use RequireThat:**
- **Guard conditions**: When nil values would cause panics
- **Prerequisites**: When subsequent assertions depend on earlier conditions
- **Fatal errors**: When continuing after failure would produce misleading errors

**When to use AssertThat:**
- **Independent checks**: When failures don't affect other assertions
- **Multiple validations**: When you want to see all failures at once
- **Most cases**: AssertThat is the default choice

### Composition

Build complex matchers from simple ones:

```go
validEmail := specta.AllOf(
    specta.Contains("@"),
    specta.MatchesRegex(`^[a-z]+@[a-z]+\.[a-z]+$`),
)

specta.AssertThat(t, user.Email, validEmail)
```

### Available Built-in Matchers

specta includes matchers for common scenarios:

- **Equality**: `Equal`, `DeepEqual`, `Is`
- **Numeric**: `GreaterThan`, `LessThan`, `GreaterThanOrEqual`, `LessThanOrEqual`
- **Strings**: `Contains`, `HasPrefix`, `HasSuffix`, `MatchesRegex`
- **Boolean**: `IsTrue`, `IsFalse`
- **Composition**: `AllOf`, `AnyOf`, `Not`

See [Core Matchers]({{< relref "/docs/core-matchers/" >}}) for the complete list including error matchers, time matchers, collection matchers, and more.

### Partial Matching

Only assert what matters for each test:

```go
// Only care about the name and email, other fields can have any value
// This test won't break when new fields (like PhoneNumber, Address) are added
specta.AssertThat(t, user, MatchUser().
    WithName(specta.Equal("Alice")).
    WithEmail(specta.Contains("example.com")))
```

## Next Steps

- Learn about [Core Matchers]({{< relref "/docs/core-matchers/" >}}) - the building blocks
- Explore [Code Generation]({{< relref "/docs/matchers-for-your-types/" >}}) for custom types
- Set up [Test Data Factories]({{< relref "/docs/factories/" >}}) for deterministic test data

## Module Information

- **Module**: `github.com/james-w/specta`
- **Go Version**: 1.22+
- **License**: MIT
