---
title: "Test Data Factories"
weight: 4
---

<!-- setup
package doctest

import (
	"testing"
	"time"
	"github.com/james-w/specta"
)

var _ = testing.Verbose

type User struct {
	ID        string
	Name      string
	Email     string
	Age       int
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Mock factory
type factoryType struct{}

type userBuilder struct {
	p    specta.Source
	name string
	role string
}

func (b *userBuilder) WithName(name string) *userBuilder {
	b.name = name
	return b
}

func (b *userBuilder) WithRole(role string) *userBuilder {
	b.role = role
	return b
}

func (b *userBuilder) Build() User {
	name := b.name
	if name == "" {
		name = b.p.StringWith("user")
	}
	role := b.role
	if role == "" {
		role = "user"
	}
	return User{
		ID:        b.p.ID(),
		Name:      name,
		Email:     b.p.StringWith("user") + "@example.com",
		Age:       30,
		Role:      role,
		CreatedAt: b.p.Time(),
		UpdatedAt: b.p.Time(),
	}
}

func (f factoryType) NewUser(p specta.Source) *userBuilder {
	return &userBuilder{p: p}
}

var factory = factoryType{}
-->

# Test Data Factories

Factories are the flip side of matchers. While matchers **validate** with partial matching, factories **generate** test data with sensible defaults.

## Why Factories?

### The Test Data Problem

<!-- skip-test -->
```go
func TestUserWorkflow(t *testing.T) {
    // What does this test actually care about?
    user1 := User{
        ID:        "user-1",
        Name:      "Alice",
        Email:     "alice@example.com",
        Age:       30,
        Role:      "admin",           // This matters
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    user2 := User{
        ID:        "user-2",
        Name:      "Bob",
        Email:     "bob@example.com",
        Age:       25,
        Role:      "user",             // This matters
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Test logic here...
}
```

**The Problem:** It's unclear what the test depends on. Does it care about:
- The specific names "Alice" and "Bob"?
- The exact email addresses?
- The ages being different?
- The timestamps?

You have to read the whole test to figure out what actually matters. Most of these fields are just **noise** - required by the struct but irrelevant to the test.

### The Factory Solution

<!-- skip-test -->
```go
func TestUserWorkflow(t *testing.T) {
    p := specta.New()

    // Crystal clear: this test cares about roles
    alice := factory.NewUser(p).WithRole("admin").Build()
    bob := factory.NewUser(p).WithRole("user").Build()

    // Everything else gets sensible defaults
    // IDs, emails, names, timestamps - all deterministic, not random
}
```

**The Philosophy:**
- **Only specify what matters** - If the test doesn't care about the name, don't set it
- **Defaults aren't special** - They're unlikely to be magic values that make tests pass accidentally
- **Deterministic, not random** - Same test run produces same data (no flakiness)
- **Clear dependencies** - Reader immediately sees what the test depends on

With factories, `alice` and `bob` will have different IDs, emails, names (generated deterministically), but you only specified roles because **that's what the test actually cares about**.

## Primitives System

The `Primitives` interface provides deterministic test data generation:

<!-- skip-test -->
```go
type Primitives interface {
    Next() int                    // Counter-based: 0, 1, 2, 3...
    String(prefix string) string  // "prefix-0", "prefix-1", ...
    Time() time.Time              // Base + increments
    UUID() string                 // Deterministic UUIDs
    ID(prefix string) string      // "prefix-0", "prefix-1", ...
}
```

### Using `Gen`

`Gen` is the standard implementation:

<!-- skip-test -->
```go
p := specta.New()

// Deterministic generation
id1 := p.ID()   // "user-0"
id2 := p.ID()   // "user-1"
email1 := p.StringWith("user") + "@example.com"  // "user-0@example.com"
email2 := p.StringWith("user") + "@example.com"  // "user-1@example.com"

// Times increment from base
t1 := p.Time()  // 2024-01-01 00:00:00
t2 := p.Time()  // 2024-01-01 00:00:01
```

## Generated Factories

Remember the `specta.yaml` from the matchers section? It **also** generates factories!

```yaml
# specta.yaml
version: 1
targets:
  - package: .
    types:
      include:
        - User
```

Running `go run github.com/james-w/specta/cmd -config specta.yaml` generates:

- `factory/user_gen.go` - Factory recipes (fluent builders)
- `factory/user_matcher_gen.go` - Matchers (we already covered these!)
- `factory/spec/user_gen.go` - Low-level factory builders

## Using Generated Factories

### Basic Usage

<!-- skip-test -->
```go
p := specta.New()

// Build with all defaults
user := factory.User().Build(p)
// Result:
//   ID: "user-0"
//   Name: "name-0"
//   Email: "email-0"
//   Age: 0
//   CreatedAt: <base time>

// Override specific fields
alice := factory.User().
    Name("Alice").
    Email("alice@example.com").
    Build(p)
// Result:
//   ID: "user-1"          ← Auto-generated
//   Name: "Alice"         ← Specified
//   Email: "alice@example.com"  ← Specified
//   Age: 1                ← Auto-incremented
//   CreatedAt: <base time + 1s>  ← Auto-incremented
```

### Recipe Pattern

Build multiple related objects:

<!-- skip-test -->
```go
func TestMultipleUsers(t *testing.T) {
    p := specta.New()

    users := []User{
        factory.User().Name("Alice").Role("admin").Build(p),
        factory.User().Name("Bob").Build(p),
        factory.User().Name("Charlie").Build(p),
    }

    // Each gets unique IDs, emails, timestamps automatically
}
```

### Nested Objects

<!-- skip-test -->
```go
p := specta.New()

// Generate nested structures
order := factory.Order().
    UserFromRecipe(factory.User().Name("Alice")).
    Items([]Item{
        factory.Item().SKU("WIDGET-1").Build(p),
        factory.Item().SKU("GADGET-2").Build(p),
    }).
    Build(p)
```

## Factory Patterns

### Named Configurations

Create factory functions that return recipes for common scenarios:

<!-- skip-test -->
```go
// factory/helpers.go
func AdminUser() UserRecipe {
    return factory.User().
        Role("admin").
        Permissions([]string{"read", "write", "delete"})
}

func GuestUser() UserRecipe {
    return factory.User().
        Role("guest").
        Permissions([]string{"read"})
}

func ExpiredUser() UserRecipe {
    return factory.User().
        ExpiresAt(time.Now().Add(-24 * time.Hour))
}
```

Usage:

<!-- skip-test -->
```go
func TestPermissions(t *testing.T) {
    p := specta.New()

    // Build directly
    admin := factory.AdminUser().Build(p)
    guest := factory.GuestUser().Build(p)

    // Or further customize before building
    superAdmin := factory.AdminUser().
        Name("Super Admin").
        Build(p)
}
```

**Why return recipes?** They're composable - you can further customize them before building, or use them in nested structures with `FromRecipe` methods.

### Composing Factories

Build complex object graphs:

<!-- skip-test -->
```go
func TestOrderProcessing(t *testing.T) {
    p := specta.New()

    // Build an entire order graph
    order := factory.NewOrder(p).
        WithUser(factory.NewUser(p).WithName("Alice").Build()).
        WithShippingAddress(factory.NewAddress(p).
            WithCity("Seattle").
            Build()).
        WithBillingAddress(factory.NewAddress(p).
            WithCity("Portland").
            Build()).
        WithItems([]Item{
            factory.NewItem(p).WithPrice(1999).Build(),
            factory.NewItem(p).WithPrice(2999).Build(),
        }).
        Build()

    // Test order processing logic
    ProcessOrder(order)
}
```

## Determinism and Reproducibility

Factories with primitives give you **reproducible test data**:

<!-- skip-test -->
```go
func TestSomething(t *testing.T) {
    // Same starting point = same data every time
    p := specta.New()

    user1 := factory.NewUser(p).Build()
    user2 := factory.NewUser(p).Build()

    // user1.ID == "user-0" (always)
    // user2.ID == "user-1" (always)

    // Tests are reproducible!
}
```

This is crucial for:
- Debugging flaky tests
- Consistent test environments

## Next Steps

- [Factories + Matchers Together]({{< relref "/docs/factories-and-matchers/" >}}) - The complete pattern
- [Property-Based Testing]({{< relref "/docs/property-based-testing/" >}}) - Use factories for PBT
- [Examples]({{< relref "/docs/examples/" >}}) - Real-world factory usage
