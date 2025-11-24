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
		result := m.nameMatcher.Matches(u.Name)
		if !result.Matched {
			return specta.MatchResult{Matched: false, Message: "name did not match"}
		}
	}
	if m.emailMatcher != nil {
		result := m.emailMatcher.Matches(u.Email)
		if !result.Matched {
			return specta.MatchResult{Matched: false, Message: "email did not match"}
		}
	}
	return specta.MatchResult{Matched: true}
}

var (
	value    = 42
	expected = 42
	list     = []string{"a", "b", "c"}
	name     = "Alice"
	user     = User{Name: "Alice", Email: "alice@example.com", Age: 30}
)
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
package mypackage_test

import (
    "testing"
    "github.com/james-w/specta"
)

func TestUser(t *testing.T) {
    user := User{
        Name:  "Alice",
        Email: "alice@example.com",
        Age:   30,
    }

    // Compose matchers for readable assertions
    specta.AssertThat(t, user.Name, specta.Equal("Alice"))
    specta.AssertThat(t, user.Email, specta.Contains("example.com"))
    specta.AssertThat(t, user.Age, specta.GreaterThan(18))

    // Combine matchers with AllOf
    specta.AssertThat(t, user.Email, specta.AllOf(
        specta.HasPrefix("alice"),
        specta.HasSuffix(".com"),
        specta.Contains("@"),
    ))
}
```

## Core Concepts

### Matchers

Matchers are reusable predicates that test values and provide detailed failure messages:

```go
specta.AssertThat(t, value, specta.Equal(expected))
specta.AssertThat(t, list, specta.ContainsAllElements("a", "b", "c"))
specta.AssertThat(t, name, specta.Not(specta.Equal("")))
```

### Composition

Build complex matchers from simple ones:

```go
validEmail := specta.AllOf(
    specta.Contains("@"),
    specta.MatchesRegex(`^[a-z]+@[a-z]+\.[a-z]+$`),
)

specta.AssertThat(t, user.Email, validEmail)
```

### Partial Matching

Only assert what matters for each test:

```go
// Only care about the name and email, other fields can have any value
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
- **License**: Apache 2.0
