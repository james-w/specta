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

func TestUserRegistration(t *testing.T) {
    // Register a new user
    user := registerUser("alice@example.com", 30)

    // Match multiple fields at once
    gomatchers.AssertThat(t, user,
        UserMatches().
            Name(gomatchers.Equal("Alice")).
            Age(gomatchers.GreaterThan(18)).
            Email(gomatchers.Contains("@example.com")).
            Matcher())
}
```

**When tests fail**, you see exactly what went wrong with structured diffs:

```
UserView {
  ✓ ID: "user-123"
  ✗ Name: expected "Alice" but got "Bob"
  ✓ Active: true
  ✗ Score: expected > 100 but got 50
  ~ Email: "test@example.com"
}
```

✓ = matched, ✗ = failed, ~ = not checked (shown for context). Colors automatically enabled in terminals.

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
var IsActiveAdmin = UserMatches().
    Active(gomatchers.IsTrue()).
    Role(gomatchers.Equal("admin")).
    Matcher()

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

Generate factories alongside matchers for consistent test data.

**Key insight**: Only specify what matters to your test. This makes tests clearer and avoids unintended dependencies on irrelevant test data details.

```go
p := gomatchers.New()

// Build with defaults - gets realistic random data for all fields
user := factory.User().Build(p)

// Override only what matters to THIS test
user := factory.User().
    Email("alice@example.com").  // This test cares about email format
    Build(p)
// Name, Age, ID, etc. get sensible defaults - we don't care about them

// Build nested structures - specify only relevant parts
order := factory.Order().
    UserFromRecipe(
        factory.User().Email("buyer@example.com"),  // Only email matters
    ).
    Total(99.99).  // Only total matters
    Build(p)
// Status, ID, CreatedAt, etc. get defaults
```

This prevents brittle tests: if you later realize `User` needs a `PhoneNumber` field, tests that don't care about phone numbers keep working without changes.

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

### Table-Driven Tests with Factories

Factories shine in table-driven tests by letting you vary inputs while keeping irrelevant fields consistent:

```go
func TestOrderDiscount(t *testing.T) {
    p := gomatchers.New()

    tests := []struct {
        name           string
        userRecipe     factory.UserRecipe
        orderTotal     float64
        expectedStatus string
        shouldDiscount bool
    }{
        {
            name:           "premium user gets discount",
            userRecipe:     factory.User().AccountType("premium"),
            orderTotal:     100.0,
            expectedStatus: "approved",
            shouldDiscount: true,
        },
        {
            name:           "regular user no discount",
            userRecipe:     factory.User().AccountType("regular"),
            orderTotal:     100.0,
            expectedStatus: "approved",
            shouldDiscount: false,
        },
        {
            name:           "new user no discount",
            userRecipe:     factory.User().AccountType("regular").CreatedAt(time.Now()),
            orderTotal:     100.0,
            expectedStatus: "pending_review",
            shouldDiscount: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Build test data - only varies what matters
            order := factory.Order().
                Total(tt.orderTotal).
                UserFromRecipe(tt.userRecipe).
                Build(p)

            result := applyDiscountRules(order)

            // Verify expected outcomes
            gomatchers.AssertThat(t, result,
                OrderMatches().
                    Status(gomatchers.Equal(tt.expectedStatus)).
                    DiscountApplied(gomatchers.Equal(tt.shouldDiscount)).
                    Matcher())
        })
    }
}
```

**Why this works well**: Each test case specifies only the user attributes that matter (account type, creation date), while all other user fields (name, email, address, etc.) get consistent defaults. This makes it crystal clear what's being varied in each test case.

### Interface-Based Matchers

When multiple concrete types share common behavior through interfaces or embedded fields, you can create matchers that work across all of them:

```go
// Notification is an interface implemented by Email, SMS, Push
type Notification interface {
    GetPriority() int
    IsSent() bool
    GetRecipient() string
}

// Create a matcher that works for ANY notification type
func IsUrgentNotification[T Notification]() gomatchers.Matcher[T] {
    return gomatchers.AllOf(
        gomatchers.Field("Priority",
            func(n T) int { return n.GetPriority() },
            gomatchers.GreaterThanOrEqual(5)),
        gomatchers.Field("Sent",
            func(n T) bool { return n.IsSent() },
            gomatchers.IsTrue()),
    )
}

// Same matcher works for all concrete types
func TestEmailNotification(t *testing.T) {
    email := EmailNotification{Priority: 5, Sent: true, ...}
    gomatchers.AssertThat(t, email, IsUrgentNotification[EmailNotification]())
}

func TestSMSNotification(t *testing.T) {
    sms := SMSNotification{Priority: 8, Sent: true, ...}
    gomatchers.AssertThat(t, sms, IsUrgentNotification[SMSNotification]())
}

func TestPushNotification(t *testing.T) {
    push := PushNotification{Priority: 7, Sent: true, ...}
    gomatchers.AssertThat(t, push, IsUrgentNotification[PushNotification]())
}
```

This lets you define matching logic once based on the interface contract, then apply it to all implementing types.

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
- When validation rules change (e.g., "active users must also have verified email"), update 50 places
- Easy to miss one, causing inconsistent test coverage and maintenance burden

With gomatchers:
- One `IsValidUser` matcher
- 50 tests use it
- Change validation rule → update matcher → all tests updated consistently
- If you miss updating the matcher, ALL 50 tests fail immediately, showing you exactly what needs fixing

This is the difference between *writing assertions* and *building a test framework*.

## Contributing

Contributions welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE) for details.
