---
title: specta
description: Composable Testing for Go
---

# specta

**Composable Testing for Go**

specta is a Go testing library that emphasizes composition and reuse through matchers and test data factories. Build reusable test components instead of duplicating assertions.

## Quick Links

- [Get Started](/docs/introduction/)
- [GitHub Repository](https://github.com/james-w/specta)
- [API Reference](/docs/api-reference/)

## Key Features

- **Composable Matchers** - Build complex assertions from simple, reusable pieces
- **Type-Safe Factories** - Generate deterministic test data with sensible defaults
- **Partial Matching** - Assert only what matters, avoid brittle tests
- **Code Generation** - Automatically generate matchers and factories for your types
- **Property-Based Testing** - Use factories and matchers for powerful property-based tests

## Quick Example

```go
// Instead of brittle assertions like:
if user.Name != "Alice" || user.Email != "alice@example.com" {
    t.Fatal("user mismatch")
}

// Write composable, expressive matchers:
AssertThat(t, user, MatchUser().
    WithName(Equal("Alice")).
    WithEmail(Contains("example.com")))
```

[Get Started →](/docs/introduction/)
