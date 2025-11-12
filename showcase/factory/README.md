# Factory Package

This package contains generated test factories and matchers for the showcase types.

## Overview

The factory package provides a fluent API for:
- **Building test data** with sensible defaults and easy customization
- **Matching values** with partial matching support
- **Nested object creation** with recipe composition

## Quick Start

### Building Test Data

```go
import (
    "testing"
    testgen "github.com/james-w/specta"
    "github.com/james-w/specta/showcase/factory"
)

func TestExample(t *testing.T) {
    p := testgen.New()

    // Build with defaults
    user := factory.User().Build(p)

    // Customize specific fields
    user = factory.User().
        FirstName("Alice").
        Email("alice@example.com").
        Active(true).
        Build(p)

    // Build many instances with unique values
    users := factory.User().
        Active(true).
        Many(10, p)
}
```

### Matching Values

```go
func TestMatching(t *testing.T) {
    user := factory.User().
        FirstName("Alice").
        Build(testgen.New())

    // Full field matching
    matcher := factory.UserMatches().
        FirstName(testgen.Equal("Alice")).
        Email(testgen.Contains("@example.com")).
        Active(testgen.IsTrue()).
        Matcher()

    testgen.AssertThat(t, user, matcher)
}
```

### Partial Matching

```go
func TestPartialMatching(t *testing.T) {
    user := factory.User().
        FirstName("Alice").
        Email("alice@example.com").
        Build(testgen.New())

    // Only check FirstName - other fields ignored
    matcher := factory.User().
        FirstName("Alice").
        AsEqualMatcher()

    testgen.AssertThat(t, user, matcher)
}
```

### Nested Objects

```go
func TestNested(t *testing.T) {
    p := testgen.New()

    // Build with nested recipe
    user := factory.User().
        AddressFromRecipe(
            factory.Address().
                City("New York").
                ZipCode("10001"),
        ).
        Build(p)

    // Partial matching works recursively
    matcher := factory.User().
        AddressFromRecipe(
            factory.Address().City("New York"),
        ).
        AsEqualMatcher()

    testgen.AssertThat(t, user, matcher)
}
```

### Constructor-Based Types

Some types use constructors (e.g., `NewBankAccount`, `NewEmail`). These work the same way:

```go
func TestConstructor(t *testing.T) {
    p := testgen.New()

    // BankAccount uses NewBankAccount(name string, balance int)
    account := factory.BankAccount().
        Name("Alice's Account").
        Balance(1000).
        Build(p)

    // Email has custom default provider for email addresses
    email := factory.Email().Build(p)  // generates valid email

    // Matching works the same
    matcher := factory.BankAccountMatches().
        Name(testgen.Equal("Alice's Account")).
        Balance(testgen.GreaterThan(500)).
        Matcher()

    testgen.AssertThat(t, account, matcher)
}
```

## Type-Specific Features

### Custom Default Providers

Some constructor parameters have custom default providers:

- **Email.Address**: Generates valid email addresses (e.g., `"user_1@example.com"`)

See each type's documentation for details on available custom providers.

### Available Matchers

Each generated type has two main functions:

- `TypeName()` - Creates a recipe for building instances
- `TypeNameMatches()` - Creates a matcher for verifying instances

Recipes have these methods:
- `FieldName(value)` - Set a specific field value
- `FieldNameFromRecipe(recipe)` - Set field using another recipe (for nested types)
- `Build(p)` - Build a single instance
- `Many(n, p)` - Build n instances with unique values
- `AsEqualMatcher()` - Convert recipe to partial matcher
- `Provider()` - Get a provider for lazy evaluation

Matchers have:
- `FieldName(matcher)` - Add a matcher for a field
- `Matcher()` - Get the composed matcher
- `FieldNameMatches(typeMatcher)` - Convenience method for nested type matchers (if applicable)

## Generated Types

This package includes factories and matchers for:

- **Address** - Struct-based type
- **BankAccount** - Constructor-based (`NewBankAccount`)
- **Email** - Constructor-based (`NewEmail`) with custom email provider
- **Account** - Constructor-based (`NewAccount`) with nested User
- **User** - Struct-based with nested Address
- **Product** - Struct-based
- **Order** - Struct-based with nested items
- **OrderItem** - Struct-based
- **BlogPost** - Struct-based with nested comments
- **Comment** - Struct-based

## Tips

1. **Use `Build(p)` for single instances**: Pass the same `testgen.Primitives` instance for consistent IDs
2. **Use `Many(n, p)` for collections**: Automatically generates unique values
3. **Partial matching is powerful**: Only specify the fields you care about in `AsEqualMatcher()`
4. **Nest recipes for complex objects**: Use `FromRecipe` methods for better composition
5. **Check type docs**: Each generated type has documentation showing constructors, defaults, and getters

## Regenerating

This package is generated by `go generate`. To regenerate:

```bash
cd showcase
go generate ./...
```

Configuration is in `testgen.yaml`.
