---
title: "Test Data Factories"
weight: 4
---

<!-- setup
package doctest

import (
	"fmt"
	"time"
	"github.com/james-w/specta"
)

type User struct {
	ID        string
	Name      string
	Email     string
	Age       int
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Mock factory - Recipe pattern
type UserRecipe struct {
	name string
	role string
}

func (r UserRecipe) Name(name string) UserRecipe {
	r.name = name
	return r
}

func (r UserRecipe) Role(role string) UserRecipe {
	r.role = role
	return r
}

func (r UserRecipe) Build(s specta.Source) User {
	// Use Source for deterministic generation - simplified mock
	counter := s.DrawBits(32)

	name := r.name
	if name == "" {
		name = fmt.Sprintf("name-%d", counter)
	}
	role := r.role
	if role == "" {
		role = "user"
	}
	return User{
		ID:        fmt.Sprintf("user-%d", counter),
		Name:      name,
		Email:     fmt.Sprintf("email-%d", counter),
		Age:       30,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

type factoryType struct{}

func (f factoryType) User() UserRecipe {
	return UserRecipe{}
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

```go
func TestUserWorkflow(t *testing.T) {
    p := specta.New()

    // Crystal clear: this test cares about roles
    alice := factory.User().Role("admin").Build(p)
    bob := factory.User().Role("user").Build(p)

    // Everything else gets sensible defaults
    // IDs, emails, names, timestamps - all deterministic, not random

    // Verify we got users with the roles we specified
    specta.AssertThat(t, alice.Role, specta.Equal("admin"))
    specta.AssertThat(t, bob.Role, specta.Equal("user"))
}
```

**The Philosophy:**
- **Only specify what matters** - If the test doesn't care about the name, don't set it
- **Defaults aren't special** - They're unlikely to be magic values that make tests pass accidentally
- **Deterministic, not random** - Same test run produces same data (no flakiness)
- **Clear dependencies** - Reader immediately sees what the test depends on

With factories, `alice` and `bob` will have different IDs, emails, names (generated deterministically), but you only specified roles because **that's what the test actually cares about**.

## Source and Deterministic Generation

The `Source` interface provides deterministic test data generation. The main implementation is `PrimitivesGen`, which generates friendly, sequential values.

**Why deterministic?**
- **Reproducible tests** - Same test run produces same data (no flakiness)
- **Debuggable** - When a test fails, you can reproduce the exact scenario
- **Sequential values** - Each call generates a unique, sequential value

### Using `specta.New()`

```go
func TestFactories(t *testing.T) {
    p := specta.New()  // Returns a Source backed by PrimitivesGen

    // Factories use Source for deterministic generation
    user1 := factory.User().Build(p)
    user2 := factory.User().Build(p)

    // Each user gets unique values automatically
    // Use the generated values - don't predict what they'll be
    specta.AssertThat(t, user1.ID, specta.Not(specta.Equal("")))
    specta.AssertThat(t, user2.ID, specta.Not(specta.Equal("")))
    specta.AssertThat(t, user1.ID, specta.Not(specta.Equal(user2.ID)))  // Different users

    // Same order, same data every time - no test flakiness!
}
```

Values are deterministic and sequential - same order, same data every time. Use the generated values without predicting what they'll be.

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
// Result: All fields populated with deterministic values

// Override specific fields
alice := factory.User().
    Name("Alice").
    Email("alice@example.com").
    Build(p)
// Result:
//   Name: "Alice"         ← Specified
//   Email: "alice@example.com"  ← Specified
//   Other fields (ID, Age, CreatedAt) auto-generated deterministically
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

Build complex object graphs using the `FromRecipe` pattern:

<!-- skip-test -->
```go
func TestOrderProcessing(t *testing.T) {
    p := specta.New()

    // Build an entire order graph
    // Use FromRecipe to nest factory recipes without building intermediate values
    order := factory.Order().
        UserFromRecipe(factory.User().Name("Alice")).
        ShippingAddressFromRecipe(factory.Address().City("Seattle")).
        BillingAddressFromRecipe(factory.Address().City("Portland")).
        Build(p)

    // Test order processing logic
    ProcessOrder(order)
}
```

**Key pattern**: Use `FromRecipe` methods to compose nested factories. Pass recipes, not built objects. The parent factory builds everything together with the same Source.

## Determinism and Reproducibility

Factories with Source give you **reproducible test data**:

```go
func TestDeterminism(t *testing.T) {
    // Same starting point = same data every time
    p := specta.New()

    user1 := factory.User().Build(p)
    user2 := factory.User().Build(p)

    // Verify values are populated and unique
    specta.AssertThat(t, user1.ID, specta.Not(specta.Equal("")))
    specta.AssertThat(t, user2.ID, specta.Not(specta.Equal("")))
    specta.AssertThat(t, user1.ID, specta.Not(specta.Equal(user2.ID)))

    // Use generated values as inputs - don't predict what they'll be
    // Tests are reproducible - same inputs = same behavior every time!
}
```

This is crucial for:
- **Debugging flaky tests** - Reproduce exact failure scenarios
- **Consistent test environments** - Same data in CI and locally

## Next Steps

- [Factories + Matchers Together]({{< relref "/docs/factories-and-matchers/" >}}) - The complete pattern
- [Property-Based Testing]({{< relref "/docs/property-based-testing/" >}}) - Use factories for PBT
- [Examples]({{< relref "/docs/examples/" >}}) - Real-world factory usage
