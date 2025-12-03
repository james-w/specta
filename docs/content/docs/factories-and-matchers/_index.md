---
title: "Factories + Matchers Together"
weight: 5
---

# Factories + Matchers Together

Factories and matchers are two sides of the same coin. Together, they create a powerful testing pattern.

## The Pattern

**Factories**: Generate test data with defaults → "Build with what matters"

**Matchers**: Validate with partial matching → "Assert what matters"

<!-- skip-test -->
```go
func TestUserCreation(t *testing.T) {
    p := specta.New()

    // Build: specify only what we're setting
    input := factory.User().
        Name("Alice").
        Email("alice@example.com").
        Build(p)

    result := CreateUser(input)

    // Match: assert only what we care about
    specta.AssertThat(t, result, factory.UserMatches().
        Name(specta.Equal("Alice")).
        Email(specta.Equal("alice@example.com")).
        ID(specta.Not(specta.IsEmpty[string]())))
}
```

## Why This Works

### Clear Test Dependencies

When you add a new field to `User`:

<!-- skip-test -->
```go
type User struct {
    ID        string
    Name      string
    Email     string
    Age       int
    CreatedAt time.Time
    UpdatedAt time.Time
    LastLogin time.Time  // NEW FIELD
}
```

**Without factories/matchers:**
- Tests creating `User{}` get zero value for `LastLogin`
- Unclear if that's intentional or an oversight
- Hard to tell what each test actually depends on
- If the new field needs a non-zero default, you must update every test

**With factories/matchers:**
- Factory provides a sensible default for `LastLogin` in one place
- Tests continue working with reasonable data
- Clear which fields each test cares about (only those in the matcher)
- One update to the factory helper if needed, not N test updates

### Focus Tests on Intent

Each test asserts exactly what it's testing:

<!-- skip-test -->
```go
// Test 1: Only care about name normalization
func TestNameNormalization(t *testing.T) {
    p := specta.New()

    user := CreateUser(factory.User().
        Name("  ALICE  ").
        Build(p))

    specta.AssertThat(t, user, factory.UserMatches().
        Name(specta.Equal("Alice")))
}

// Test 2: Only care about email domain validation
func TestEmailDomain(t *testing.T) {
    p := specta.New()

    user := CreateUser(factory.User().
        Email("alice@company.com").
        Build(p))

    specta.AssertThat(t, user, factory.UserMatches().
        Email(specta.HasSuffix("@company.com")))
}
```

## Practical Examples

### Testing CRUD Operations

<!-- skip-test -->
```go
func TestCreateUser(t *testing.T) {
    p := specta.New()
    db := setupTestDB(t)

    // Arrange: Build input
    userData := factory.User().
        Name("Alice").
        Email("alice@example.com").
        Build(p)

    // Act
    created := db.CreateUser(userData)

    // Assert: Validate result
    specta.AssertThat(t, created, factory.UserMatches().
        ID(specta.Not(specta.IsEmpty[string]())).              // DB generated
        Name(specta.Equal(userData.Name)).
        Email(specta.Equal(userData.Email)).
        CreatedAt(specta.Not(specta.IsZero[time.Time]())))
}

func TestUpdateUser(t *testing.T) {
    p := specta.New()
    db := setupTestDB(t)

    // Create initial user
    user := db.CreateUser(factory.User().Build(p))

    // Update name
    updates := factory.User().
        Name("New Name").
        Build(p)

    updated := db.UpdateUser(user.ID, updates)

    // Assert only what changed
    specta.AssertThat(t, updated, factory.UserMatches().
        ID(specta.Equal(user.ID)).                  // Same ID
        Name(specta.Equal(updates.Name)).           // Updated
        CreatedAt(specta.Equal(user.CreatedAt)).    // Unchanged
        UpdatedAt(specta.GreaterThan(user.UpdatedAt)))
}
```

### API Response Validation

<!-- skip-test -->
```go
func TestGetUserAPI(t *testing.T) {
    p := specta.New()

    // Setup: Create user in DB
    user := factory.User().
        Name("Alice").
        Email("alice@example.com").
        Build(p)
    db.CreateUser(user)

    // Make API request
    resp := httpGet(t, "/api/users/"+user.ID)

    // Parse response
    var result User
    json.Unmarshal(resp.Body, &result)

    // Validate response
    specta.AssertThat(t, result, factory.UserMatches().
        ID(specta.Equal(user.ID)).
        Name(specta.Equal(user.Name)).
        Email(specta.Equal(user.Email)))
}
```

**Note:** You could extract this matcher into a reusable function:

<!-- skip-test -->
```go
func MatchesDBUser(user User) specta.Matcher[User] {
    return factory.UserMatches().
        ID(specta.Equal(user.ID)).
        Name(specta.Equal(user.Name)).
        Email(specta.Equal(user.Email))
}

// Usage
specta.AssertThat(t, result, MatchesDBUser(user))
```

This is especially useful when testing multiple API endpoints that return the same user representation.

## Advanced Patterns

### Table-Driven Tests

<!-- skip-test -->
```go
func TestValidation(t *testing.T) {
    p := specta.New()

    tests := []struct {
        name      string
        user      User
        wantError bool
    }{
        {
            name: "valid user",
            user: factory.User().
                Email("valid@example.com").
                Build(p),
            wantError: false,
        },
        {
            name: "invalid email",
            user: factory.User().
                Email("invalid").
                Build(p),
            wantError: true,
        },
        {
            name: "missing name",
            user: factory.User().
                Name("").
                Build(p),
            wantError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUser(tt.user)
            if tt.wantError {
                specta.AssertThat(t, err, specta.IsError())
            } else {
                specta.AssertThat(t, err, specta.NoErr())
            }
        })
    }
}
```

### Recursive Partial Matching with AsEqualMatcher

One of the most powerful patterns is using `AsEqualMatcher()` with nested recipes. Partial matching works recursively - if you set specific fields on nested types, only those nested fields will be checked:

<!-- skip-test -->
```go
func TestOrderWithPartialUserMatch(t *testing.T) {
    p := specta.New()

    // Create test data - full objects with all fields
    order := factory.Order().
        Total(150.0).
        Status("pending").
        UserFromRecipe(
            factory.User().
                FirstName("Alice").
                Email("alice@example.com"),
        ).
        Build(p)

    // Act
    result := processOrder(order)

    // Assert with partial matching - only check what matters
    // This checks:
    // - Order.Status
    // - Order.User.FirstName (ONLY this user field)
    // Everything else (Order.Total, Order.ID, User.Email, User.LastName, etc.) is ignored
    specta.AssertThat(t, result, factory.Order().
        Status("completed").
        UserFromRecipe(
            factory.User().FirstName("Alice"),  // Only FirstName is checked
        ).
        AsEqualMatcher())
}
```

**How it differs from DeepEqual:**

<!-- skip-test -->
```go
// DeepEqual checks EVERYTHING
user := factory.User().FirstName("Alice").Build(p)
specta.AssertThat(t, result, specta.DeepEqual(user))
// Fails if ANY field differs (ID, Email, LastName, CreatedAt, etc.)

// AsEqualMatcher checks ONLY what you set
matcher := factory.User().FirstName("Alice").AsEqualMatcher()
specta.AssertThat(t, result, matcher)
// Only checks FirstName, ignores all other fields
```

This recursive partial matching is especially useful for:

1. **Testing data transformations** - Only assert what the transformation should change
2. **API response validation** - Check critical fields, ignore server-generated metadata
3. **Integration tests** - Focus on business logic, ignore infrastructure details

**When to use each approach:**

- **Explicit matchers** (`factory.UserMatches().FirstName(...)`) - When you want fine-grained control over matching logic (e.g., `Contains`, `GreaterThan`)
- **AsEqualMatcher** - When you want partial equality checking with nested objects
- **DeepEqual** - When you need exact equality of entire objects

## Benefits Summary

**Factories + Matchers provide:**

1. **Less duplication** - Reusable test data and assertions
2. **Focused tests** - Each test asserts its specific concern
3. **Resilient to change** - Adding fields doesn't break tests
4. **Readable** - Clear intent, less noise
5. **Deterministic** - Reproducible test data
6. **Composable** - Build complex scenarios from simple pieces

## Next Steps

- [Property-Based Testing]({{< relref "/docs/property-based-testing/" >}}) - Use factories for PBT
- [Advanced Topics]({{< relref "/docs/advanced/" >}}) - Custom matchers, patterns
- [Examples]({{< relref "/docs/examples/" >}}) - More real-world examples
