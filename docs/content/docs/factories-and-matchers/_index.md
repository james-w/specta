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
    specta.AssertThat(t, result, factory.MatchUser().
        WithName(specta.Equal("Alice")).
        WithEmail(specta.Equal("alice@example.com")).
        WithID(specta.Not(IsEmpty())))
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

    specta.AssertThat(t, user, factory.MatchUser().
        WithName(specta.Equal("Alice")))
}

// Test 2: Only care about email domain validation
func TestEmailDomain(t *testing.T) {
    p := specta.New()

    user := CreateUser(factory.User().
        Email("alice@company.com").
        Build(p))

    specta.AssertThat(t, user, factory.MatchUser().
        WithEmail(HaveSuffix("@company.com")))
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
    specta.AssertThat(t, created, factory.MatchUser().
        WithID(specta.Not(IsEmpty())).              // DB generated
        WithName(specta.Equal(userData.Name)).
        WithEmail(specta.Equal(userData.Email)).
        WithCreatedAt(specta.Not(IsZero())))        // DB timestamp
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
    specta.AssertThat(t, updated, factory.MatchUser().
        WithID(specta.Equal(user.ID)).                  // Same ID
        WithName(specta.Equal(updates.Name)).           // Updated
        WithCreatedAt(specta.Equal(user.CreatedAt)).    // Unchanged
        WithUpdatedAt(specta.GreaterThan(user.UpdatedAt))) // Changed
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
    specta.AssertThat(t, result, factory.MatchUser().
        WithID(specta.Equal(user.ID)).
        WithName(specta.Equal(user.Name)).
        WithEmail(specta.Equal(user.Email)))
}
```

**Note:** You could extract this matcher into a reusable function:

<!-- skip-test -->
```go
func MatchesDBUser(user User) factory.UserMatcher {
    return factory.MatchUser().
        WithID(specta.Equal(user.ID)).
        WithName(specta.Equal(user.Name)).
        WithEmail(specta.Equal(user.Email))
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
                specta.AssertThat(t, err, NotBeNil())
            } else {
                specta.AssertThat(t, err, BeNil())
            }
        })
    }
}
```

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
