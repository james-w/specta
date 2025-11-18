---
title: "Introduction"
weight: 1
---

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
    . "github.com/james-w/specta"
)

func TestUser(t *testing.T) {
    user := User{
        Name:  "Alice",
        Email: "alice@example.com",
        Age:   30,
    }

    // Compose matchers for readable assertions
    AssertThat(t, user.Name, Equal("Alice"))
    AssertThat(t, user.Email, Contains("example.com"))
    AssertThat(t, user.Age, GreaterThan(18))

    // Combine matchers with AllOf
    AssertThat(t, user.Email, AllOf(
        HasPrefix("alice"),
        HasSuffix(".com"),
        Contains("@"),
    ))
}
```

## Core Concepts

### Matchers

Matchers are reusable predicates that test values and provide detailed failure messages:

```go
AssertThat(t, value, Equal(expected))
AssertThat(t, list, ContainsAllElements("a", "b", "c"))
AssertThat(t, name, Not(Equal("")))
```

### Composition

Build complex matchers from simple ones:

```go
validEmail := AllOf(
    Contains("@"),
    MatchesRegex(`^[a-z]+@[a-z]+\.[a-z]+$`),
)

AssertThat(t, user.Email, validEmail)
```

### Partial Matching

Only assert what matters for each test:

```go
// Only care about the name and email, other fields can have any value
AssertThat(t, user, MatchUser().
    WithName(Equal("Alice")).
    WithEmail(Contains("example.com")))
```

## Next Steps

- Learn about [Core Matchers](/docs/core-matchers/) - the building blocks
- Explore [Code Generation](/docs/matchers-for-your-types/) for custom types
- Set up [Test Data Factories](/docs/factories/) for deterministic test data

## Module Information

- **Module**: `github.com/james-w/specta`
- **Go Version**: 1.22+
- **License**: Apache 2.0
