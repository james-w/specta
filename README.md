# gomatchers

A Go testing library that emphasizes **composition** and **reuse** through matchers and test data factories.

Unlike traditional assertion libraries that simply provide syntactic sugar for comparisons, gomatchers enables you to build reusable, composable test components that scale with your test suite.

## Core Philosophy

Testing isn't just about asserting values—it's about building maintainable test suites. gomatchers provides:

1. **Composable Matchers**: Build complex assertions from simple, reusable pieces
2. **Generated Factories**: Automatically generate test data builders from your types
3. **Partial Matching**: Match only what matters, ignore the rest
4. **Rich Error Messages**: See exactly what failed with structured, colored diffs

## Installation

```bash
go get github.com/james-w/gomatchers
```

## Quick Start

```go
import "github.com/james-w/gomatchers"

func TestUser(t *testing.T) {
    user := User{Name: "Alice", Age: 30, Email: "alice@example.com"}

    gomatchers.AssertThat(t, user,
        gomatchers.AllOf(
            UserMatches().Name(gomatchers.Equal("Alice")),
            UserMatches().Age(gomatchers.GreaterThan(18)),
            UserMatches().Email(gomatchers.Contains("@example.com")),
        ))
}
```

## Why Composition Matters

### Traditional Approach: Repeated Assertions

```go
func TestActiveUser(t *testing.T) {
    user := createUser()
    if !user.Active {
        t.Error("Expected user to be active")
    }
    if user.Role != "admin" {
        t.Error("Expected admin role")
    }
}

func TestAnotherActiveUser(t *testing.T) {
    user := createAnotherUser()
    // Copy-paste the same checks...
    if !user.Active {
        t.Error("Expected user to be active")
    }
    if user.Role != "admin" {
        t.Error("Expected admin role")
    }
}
```

### gomatchers Approach: Reusable Matchers

```go
// Define once
var IsActiveAdmin = gomatchers.AllOf(
    UserMatches().Active(gomatchers.IsTrue()),
    UserMatches().Role(gomatchers.Equal("admin")),
)

// Reuse everywhere
func TestActiveUser(t *testing.T) {
    user := createUser()
    gomatchers.AssertThat(t, user, IsActiveAdmin)
}

func TestAnotherActiveUser(t *testing.T) {
    user := createAnotherUser()
    gomatchers.AssertThat(t, user, IsActiveAdmin) // Same check, zero duplication
}

func TestBulkOperation(t *testing.T) {
    users := bulkCreateUsers()
    for _, user := range users {
        gomatchers.AssertThat(t, user, IsActiveAdmin) // Consistent everywhere
    }
}
```

**The difference**: When requirements change (e.g., "admins must also have verified email"), you update the matcher **once** instead of hunting through dozens of test files.

## Core Features

### 1. Basic Matchers

```go
// Equality
gomatchers.Equal(42)
gomatchers.DeepEqual(expected)

// Numeric comparisons
gomatchers.GreaterThan(18)
gomatchers.LessThan(100)
gomatchers.GreaterThanOrEqual(21)

// String matchers
gomatchers.Contains("@example.com")
gomatchers.HasPrefix("user_")
gomatchers.HasSuffix(".json")

// Boolean matchers
gomatchers.IsTrue()
gomatchers.IsFalse()

// Zero value checking
gomatchers.IsZero[int]()
```

### 2. Composable Matchers

Matchers can be combined to express complex conditions:

```go
// All conditions must match
gomatchers.AllOf(
    gomatchers.GreaterThan(0),
    gomatchers.LessThan(100),
)

// At least one condition must match
gomatchers.AnyOf(
    gomatchers.Equal("admin"),
    gomatchers.Equal("moderator"),
)

// Invert a matcher
gomatchers.Not(gomatchers.Contains("test"))
```

**Real-world example: Valid email check**

```go
// Define once, reuse everywhere
var IsValidEmail = gomatchers.AllOf(
    gomatchers.Contains("@"),
    gomatchers.Not(gomatchers.Contains(" ")),
    gomatchers.Not(gomatchers.HasPrefix("@")),
)

// Use in multiple contexts
gomatchers.AssertThat(t, user.Email, IsValidEmail)
gomatchers.AssertThat(t, admin.ContactEmail, IsValidEmail)
gomatchers.AssertThat(t, invoice.BillingEmail, IsValidEmail)
```

### 3. Structured Matchers (Generated)

Instead of manually writing matchers for your structs, generate them:

**testgen.yaml:**
```yaml
targets:
  - name: myapp
    types:
      - name: User
        getters:
          - name: Name
            getter: GetName
          - name: Email
            getter: GetEmail
          - name: Age
            getter: GetAge
```

**Generated API:**
```go
// Fluent matcher builder
UserMatches().
    Name(gomatchers.Equal("Alice")).
    Age(gomatchers.GreaterThan(18)).
    Email(gomatchers.Contains("@example.com"))
```

### 4. Partial Matching: The Killer Feature

Partial matching lets you assert only what's relevant to each test:

```go
func TestUserRegistration(t *testing.T) {
    user := registerUser("alice@example.com")

    // Only care about email being set correctly
    gomatchers.AssertThat(t, user,
        UserMatches().
            Email(gomatchers.Equal("alice@example.com")).
            Matcher())
    // Don't care about ID, CreatedAt, etc.
}

func TestUserActivation(t *testing.T) {
    user := activateUser(existingUser)

    // Only care about activation status
    gomatchers.AssertThat(t, user,
        UserMatches().
            Active(gomatchers.IsTrue()).
            Matcher())
    // Don't care about name, email, etc.
}
```

**Without partial matching**, you'd need to either:
- Check every field (brittle, breaks when adding fields)
- Write custom assertions for each scenario (duplication)
- Use multiple individual assertions (verbose, unclear intent)

### 5. Nested Matchers

Matchers compose naturally for nested structures:

```go
gomatchers.AssertThat(t, order,
    OrderMatches().
        Total(gomatchers.GreaterThan(100.0)).
        Status(gomatchers.Equal("shipped")).
        // Match nested user fields
        UserMatches(
            UserMatches().
                Email(gomatchers.Contains("@premium.com")).
                AccountType(gomatchers.Equal("premium")),
        ).
        Matcher())
```

### 6. Custom Field Extraction

For computed values or custom logic:

```go
// Match based on a computed property
gomatchers.Field("FullName",
    func(u User) string {
        return u.FirstName + " " + u.LastName
    },
    gomatchers.Equal("Alice Smith"))

// Match based on method result
gomatchers.Field("IsExpired",
    func(token Token) bool {
        return token.ExpiresAt.Before(time.Now())
    },
    gomatchers.IsFalse())
```

## Test Data Factories

Generate factories alongside matchers for consistent test data:

```go
p := gomatchers.New()

// Build with defaults
user := factory.User().Build(p)

// Override specific fields
user := factory.User().
    Name("Alice").
    Email("alice@example.com").
    Age(30).
    Build(p)

// Build nested structures
order := factory.Order().
    UserFromRecipe(
        factory.User().Email("buyer@example.com"),
    ).
    Total(99.99).
    Build(p)
```

### Factory + Matcher Integration

The real power comes from using both together:

```go
func TestOrderProcessing(t *testing.T) {
    p := gomatchers.New()

    // Create test order with specific properties
    order := factory.Order().
        Status("pending").
        Total(150.0).
        UserFromRecipe(
            factory.User().AccountType("premium"),
        ).
        Build(p)

    // Process it
    processed := processOrder(order)

    // Verify only what changed
    gomatchers.AssertThat(t, processed,
        OrderMatches().
            Status(gomatchers.Equal("completed")).
            ProcessedAt(gomatchers.Not(gomatchers.IsZero[time.Time]())).
            Matcher())
}
```

## Rich Error Messages

When assertions fail, gomatchers shows you exactly what went wrong with structured diffs:

```
UserView {
  ✓ ID: "user-123"
  ✗ Name: expected "Alice" but got "Bob"
  ✓ Active: true
  ✗ Score: expected > 100 but got 50
  ~ Email: "test@example.com"
}
```

Legend:
- ✓ Field matched
- ✗ Field failed (shows expected vs actual)
- ~ Field not checked (shown for context)

Colors are automatically enabled when output is a terminal.

## Real-World Example: E-commerce Testing

```go
// Define reusable matchers for business rules
var (
    IsPremiumUser = UserMatches().
        AccountType(gomatchers.Equal("premium")).
        Active(gomatchers.IsTrue()).
        Matcher()

    IsValidOrder = OrderMatches().
        Total(gomatchers.GreaterThan(0.0)).
        Status(gomatchers.AnyOf(
            gomatchers.Equal("pending"),
            gomatchers.Equal("processing"),
            gomatchers.Equal("completed"),
        )).
        Matcher()

    IsShippedOrder = gomatchers.AllOf(
        IsValidOrder,
        OrderMatches().
            Status(gomatchers.Equal("shipped")).
            ShippedAt(gomatchers.Not(gomatchers.IsZero[time.Time]())).
            Matcher(),
    )
)

func TestPremiumUserDiscount(t *testing.T) {
    p := gomatchers.New()

    // Create premium user
    user := factory.User().AccountType("premium").Active(true).Build(p)
    gomatchers.AssertThat(t, user, IsPremiumUser) // Verify setup

    // Create order
    order := factory.Order().
        UserFromRecipe(factory.User().ID(user.ID)).
        Total(100.0).
        Build(p)

    // Apply discount
    discounted := applyDiscount(order)

    // Verify discount applied
    gomatchers.AssertThat(t, discounted,
        OrderMatches().
            Total(gomatchers.LessThan(100.0)).
            DiscountApplied(gomatchers.IsTrue()).
            Matcher())
}

func TestOrderShipment(t *testing.T) {
    p := gomatchers.New()

    // Create pending order
    order := factory.Order().Status("pending").Build(p)

    // Ship it
    shipped := shipOrder(order)

    // Verify using reusable matcher
    gomatchers.AssertThat(t, shipped, IsShippedOrder)
}

func TestBulkOrderProcessing(t *testing.T) {
    p := gomatchers.New()

    // Create multiple orders
    orders := []Order{
        factory.Order().Status("pending").Build(p),
        factory.Order().Status("pending").Build(p),
        factory.Order().Status("pending").Build(p),
    }

    // Process in bulk
    results := bulkProcess(orders)

    // Verify all results are valid
    for i, result := range results {
        gomatchers.AssertThat(t, result, IsValidOrder,
            "Order %d should be valid", i)
    }
}
```

**Key benefits demonstrated:**
1. **Reusable matchers** (`IsPremiumUser`, `IsValidOrder`, `IsShippedOrder`) encode business rules once
2. **Composable** (`IsShippedOrder` builds on `IsValidOrder`)
3. **Consistent** (same checks across different test scenarios)
4. **Maintainable** (change business rule in one place)
5. **Readable** (test intent is clear)

## Advanced Patterns

### Table-Driven Tests with Matchers

```go
func TestAgeValidation(t *testing.T) {
    tests := []struct {
        name    string
        age     int
        matcher gomatchers.Matcher[int]
    }{
        {"child", 5, gomatchers.LessThan(13)},
        {"teen", 15, gomatchers.AllOf(
            gomatchers.GreaterThanOrEqual(13),
            gomatchers.LessThan(20),
        )},
        {"adult", 25, gomatchers.GreaterThanOrEqual(18)},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            gomatchers.AssertThat(t, tt.age, tt.matcher)
        })
    }
}
```

### Polymorphic Matchers

```go
// Match any notification type that meets criteria
func IsUrgentNotification(n Notification) gomatchers.Matcher[Notification] {
    return gomatchers.AllOf(
        NotificationMatches().
            Priority(gomatchers.GreaterThanOrEqual(5)).
            Sent(gomatchers.IsTrue()).
            Matcher(),
    )
}

// Works with any notification type
gomatchers.AssertThat(t, emailNotification, IsUrgentNotification)
gomatchers.AssertThat(t, smsNotification, IsUrgentNotification)
gomatchers.AssertThat(t, pushNotification, IsUrgentNotification)
```

## Code Generation

Generate matchers and factories for your types:

```bash
# Create testgen.yaml configuration
# Run generator
go run github.com/james-w/gomatchers/cmd/main.go
```

See [Configuration Guide](docs/configuration.md) for full details.

## Comparison with Other Libraries

| Feature | gomatchers | testify | gomega |
|---------|------------|---------|--------|
| Composable matchers | ✅ Core feature | ❌ | ✅ |
| Reusable matchers | ✅ First-class | ⚠️ Via functions | ⚠️ Via functions |
| Generated matchers | ✅ From config | ❌ | ❌ |
| Partial matching | ✅ Built-in | ❌ | ⚠️ Manual |
| Test data factories | ✅ Generated | ❌ | ❌ |
| Factory-matcher integration | ✅ Seamless | ❌ | ❌ |
| Structured diffs | ✅ With symbols/colors | ⚠️ Basic | ⚠️ Basic |
| Type-safe | ✅ Generics | ⚠️ Interface{} | ⚠️ Interface{} |

**gomatchers is for teams that want:**
- Tests that scale with growing test suites
- Reusable test components (not just assertions)
- Consistent test patterns across the codebase
- Less test maintenance burden

**Other libraries are for:**
- Quick assertions in simple test cases
- Teams that prefer writing assertions from scratch each time
- Smaller codebases where duplication isn't painful

## Philosophy: Beyond Assertions

Most assertion libraries ask: *"How can we make this comparison easier to write?"*

gomatchers asks: *"How can we make our test suite maintainable at scale?"*

The answer: **composition and reuse**.

- **Matchers are values**: Store them, pass them, combine them
- **Factories encode patterns**: Generate realistic test data consistently
- **Partial matching scales**: Add fields without breaking tests
- **Composition enables abstraction**: Build high-level test vocabulary

Example: Imagine you have 50 tests checking "valid users". With traditional assertions:
- 50 places checking `user.Email != ""`, `user.Active == true`, etc.
- When validation changes, update 50 places
- Easy to miss one, causing flaky tests

With gomatchers:
- One `IsValidUser` matcher
- 50 tests use it
- Change validation rule → update matcher → all tests updated

This is the difference between *writing assertions* and *building a test framework*.

## Contributing

Contributions welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE) for details.
