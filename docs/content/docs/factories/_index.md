---
title: "Test Data Factories"
weight: 4
---

# Test Data Factories

Factories are the flip side of matchers. While matchers **validate** with partial matching, factories **generate** test data with sensible defaults.

## Why Factories?

### The Test Data Problem

```go
func TestUserWorkflow(t *testing.T) {
    // Ugh, need to specify everything every time
    user1 := User{
        ID:        "user-1",
        Name:      "Alice",
        Email:     "alice@example.com",
        Age:       30,
        Role:      "admin",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    user2 := User{
        ID:        "user-2",
        Name:      "Bob",
        Email:     "bob@example.com",
        Age:       25,
        Role:      "user",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Copy-paste exhaustion...
}
```

### The Factory Solution

```go
func TestUserWorkflow(t *testing.T) {
    p := specta.NewGen()

    // Specify only what matters, get sensible defaults for the rest
    alice := factory.NewUser(p).WithName("Alice").WithRole("admin").Build()
    bob := factory.NewUser(p).WithName("Bob").Build()

    // IDs, emails, timestamps, etc. are automatically generated
}
```

## Primitives System

The `Primitives` interface provides deterministic test data generation:

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

```go
p := specta.NewGen()

// Deterministic generation
id1 := p.ID("user")   // "user-0"
id2 := p.ID("user")   // "user-1"
email1 := p.String("user") + "@example.com"  // "user-0@example.com"
email2 := p.String("user") + "@example.com"  // "user-1@example.com"

// Times increment from base
t1 := p.Time()  // 2024-01-01 00:00:00
t2 := p.Time()  // 2024-01-01 00:00:01
```

### Configuring Gen

```go
// Custom configuration
p := &specta.Gen{
    Counter:  100,                          // Start counter at 100
    BaseTime: time.Date(2025, 1, 1, ...),  // Custom base time
    Step:     5 * time.Second,              // 5-second increments
    Prefix:   "test",                       // Default prefix
}
```

## Generated Factories

Remember the `specta.yaml` from the matchers section? It **also** generates factories!

```yaml
# specta.yaml
package: mypackage
output_dir: factory
types:
  - name: User
    fields:
      - name: ID
        type: string
      - name: Name
        type: string
      - name: Email
        type: string
      - name: Age
        type: int
      - name: CreatedAt
        type: time.Time
```

Running `go run github.com/james-w/specta/cmd/main.go -config specta.yaml` generates:

- `factory/user_gen.go` - Factory builders
- `factory/user_matcher_gen.go` - Matchers (we already covered these!)

## Using Generated Factories

### Basic Usage

```go
p := specta.NewGen()

// Build with all defaults
user := factory.NewUser(p).Build()
// Result:
//   ID: "user-0"
//   Name: "name-0"
//   Email: "email-0"
//   Age: 0
//   CreatedAt: <base time>

// Override specific fields
alice := factory.NewUser(p).
    WithName("Alice").
    WithEmail("alice@example.com").
    Build()
// Result:
//   ID: "user-1"          ← Auto-generated
//   Name: "Alice"         ← Specified
//   Email: "alice@example.com"  ← Specified
//   Age: 1                ← Auto-incremented
//   CreatedAt: <base time + 1s>  ← Auto-incremented
```

### Recipe Pattern

Build multiple related objects:

```go
func TestMultipleUsers(t *testing.T) {
    p := specta.NewGen()

    users := []User{
        factory.NewUser(p).WithName("Alice").WithRole("admin").Build(),
        factory.NewUser(p).WithName("Bob").Build(),
        factory.NewUser(p).WithName("Charlie").Build(),
    }

    // Each gets unique IDs, emails, timestamps automatically
}
```

### Nested Objects

```go
// Generate nested structures
order := factory.NewOrder(p).
    WithUser(factory.NewUser(p).
        WithName("Alice").
        Build()).
    WithItems([]Item{
        factory.NewItem(p).WithSKU("WIDGET-1").Build(),
        factory.NewItem(p).WithSKU("GADGET-2").Build(),
    }).
    Build()
```

## Factory Patterns

### Named Configurations

Create factory functions for common scenarios:

```go
// factory/helpers.go
func AdminUser(p Primitives) User {
    return NewUser(p).
        WithRole("admin").
        WithPermissions([]string{"read", "write", "delete"}).
        Build()
}

func GuestUser(p Primitives) User {
    return NewUser(p).
        WithRole("guest").
        WithPermissions([]string{"read"}).
        Build()
}

func ExpiredUser(p Primitives) User {
    return NewUser(p).
        WithExpiresAt(time.Now().Add(-24 * time.Hour)).
        Build()
}
```

Usage:

```go
func TestPermissions(t *testing.T) {
    p := specta.NewGen()

    admin := factory.AdminUser(p)
    guest := factory.GuestUser(p)

    // Test with clearly-named test data
}
```

### Partial Specification

Only set what matters for your test:

```go
// Test email validation - don't care about other fields
func TestEmailValidation(t *testing.T) {
    p := specta.NewGen()

    validUser := factory.NewUser(p).
        WithEmail("valid@example.com").
        Build()

    invalidUser := factory.NewUser(p).
        WithEmail("invalid-email").
        Build()

    // Everything else is auto-generated
}
```

### Composing Factories

Build complex object graphs:

```go
func TestOrderProcessing(t *testing.T) {
    p := specta.NewGen()

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

```go
func TestSomething(t *testing.T) {
    // Same starting point = same data every time
    p := specta.NewGen()

    user1 := factory.NewUser(p).Build()
    user2 := factory.NewUser(p).Build()

    // user1.ID == "user-0" (always)
    // user2.ID == "user-1" (always)

    // Tests are reproducible!
}
```

This is crucial for:
- Debugging flaky tests
- Property-based testing (coming next!)
- Consistent test environments

## Next Steps

- [Factories + Matchers Together](/docs/factories-and-matchers/) - The complete pattern
- [Property-Based Testing](/docs/property-based-testing/) - Use factories for PBT
- [Examples](/docs/examples/) - Real-world factory usage
