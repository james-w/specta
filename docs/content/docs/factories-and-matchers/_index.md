---
title: "Factories + Matchers Together"
weight: 5
---

# Factories + Matchers Together

Factories and matchers are two sides of the same coin. Together, they create a powerful testing pattern.

## The Pattern

**Factories**: Generate test data with defaults → "Build with what matters"

**Matchers**: Validate with partial matching → "Assert what matters"

```go
func TestUserCreation(t *testing.T) {
    p := specta.NewGen()

    // Build: specify only what we're setting
    input := factory.NewUser(p).
        WithName("Alice").
        WithEmail("alice@example.com").
        Build()

    result := CreateUser(input)

    // Match: assert only what we care about
    specta.AssertThat(t, result, factory.MatchUser().
        WithName(specta.Equal("Alice")).
        WithEmail(specta.Equal("alice@example.com")).
        WithID(specta.Not(BeEmpty())))
}
```

## Why This Works

### Prevents Brittle Tests

When you add a new field to `User`:

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
- Every test that creates a `User{}` must be updated
- Every assertion checking all fields breaks
- Brittle, unmaintainable tests

**With factories/matchers:**
- Factory provides a default for `LastLogin`
- Matchers only check fields they care about
- Tests continue working!

### Focus Tests on Intent

Each test asserts exactly what it's testing:

```go
// Test 1: Only care about name normalization
func TestNameNormalization(t *testing.T) {
    p := specta.NewGen()

    user := CreateUser(factory.NewUser(p).
        WithName("  ALICE  ").
        Build())

    specta.AssertThat(t, user, factory.MatchUser().
        WithName(specta.Equal("Alice")))
}

// Test 2: Only care about email domain validation
func TestEmailDomain(t *testing.T) {
    p := specta.NewGen()

    user := CreateUser(factory.NewUser(p).
        WithEmail("alice@company.com").
        Build())

    specta.AssertThat(t, user, factory.MatchUser().
        WithEmail(HaveSuffix("@company.com")))
}
```

## Practical Examples

### Testing CRUD Operations

```go
func TestCreateUser(t *testing.T) {
    p := specta.NewGen()
    db := setupTestDB(t)

    // Arrange: Build input
    userData := factory.NewUser(p).
        WithName("Alice").
        WithEmail("alice@example.com").
        Build()

    // Act
    created := db.CreateUser(userData)

    // Assert: Validate result
    specta.AssertThat(t, created, factory.MatchUser().
        WithID(specta.Not(BeEmpty())).              // DB generated
        WithName(specta.Equal("Alice")).
        WithEmail(specta.Equal("alice@example.com")).
        WithCreatedAt(specta.Not(BeZero())))        // DB timestamp
}

func TestUpdateUser(t *testing.T) {
    p := specta.NewGen()
    db := setupTestDB(t)

    // Create initial user
    user := db.CreateUser(factory.NewUser(p).Build())

    // Update name
    updates := factory.NewUser(p).
        WithName("New Name").
        Build()

    updated := db.UpdateUser(user.ID, updates)

    // Assert only what changed
    specta.AssertThat(t, updated, factory.MatchUser().
        WithID(specta.Equal(user.ID)).              // Same ID
        WithName(specta.Equal("New Name")).         // Updated
        WithCreatedAt(specta.Equal(user.CreatedAt))) // Unchanged
}
```

### API Response Validation

```go
func TestGetUserAPI(t *testing.T) {
    p := specta.NewGen()

    // Setup: Create user in DB
    user := factory.NewUser(p).
        WithName("Alice").
        WithEmail("alice@example.com").
        Build()
    db.CreateUser(user)

    // Make API request
    resp := httpGet(t, "/api/users/"+user.ID)

    // Parse response
    var result User
    json.Unmarshal(resp.Body, &result)

    // Validate response
    specta.AssertThat(t, result, factory.MatchUser().
        WithID(specta.Equal(user.ID)).
        WithName(specta.Equal("Alice")).
        WithEmail(specta.Equal("alice@example.com")))
}
```

### Database Entity Tests

```go
func TestUserRepository(t *testing.T) {
    p := specta.NewGen()
    repo := NewUserRepository(db)

    tests := []struct {
        name  string
        user  User
        check func(*testing.T, User)
    }{
        {
            name: "creates user with generated ID",
            user: factory.NewUser(p).WithName("Alice").Build(),
            check: func(t *testing.T, result User) {
                specta.AssertThat(t, result, factory.MatchUser().
                    WithID(specta.Not(BeEmpty())).
                    WithName(specta.Equal("Alice")))
            },
        },
        {
            name: "normalizes email",
            user: factory.NewUser(p).WithEmail("ALICE@EXAMPLE.COM").Build(),
            check: func(t *testing.T, result User) {
                specta.AssertThat(t, result, factory.MatchUser().
                    WithEmail(specta.Equal("alice@example.com")))
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := repo.Create(tt.user)
            tt.check(t, result)
        })
    }
}
```

## Advanced Patterns

### Testing Relationships

```go
func TestOrderWithItems(t *testing.T) {
    p := specta.NewGen()

    order := factory.NewOrder(p).
        WithUser(factory.NewUser(p).
            WithName("Alice").
            Build()).
        WithItems([]Item{
            factory.NewItem(p).WithPrice(1000).Build(),
            factory.NewItem(p).WithPrice(2000).Build(),
        }).
        Build()

    result := ProcessOrder(order)

    specta.AssertThat(t, result, factory.MatchOrder().
        WithStatus(specta.Equal("processed")).
        WithTotal(specta.Equal(3000)).
        WithUser(factory.MatchUser().
            WithName(specta.Equal("Alice"))))
}
```

### Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    p := specta.NewGen()

    tests := []struct {
        name      string
        user      User
        wantError bool
    }{
        {
            name: "valid user",
            user: factory.NewUser(p).
                WithEmail("valid@example.com").
                Build(),
            wantError: false,
        },
        {
            name: "invalid email",
            user: factory.NewUser(p).
                WithEmail("invalid").
                Build(),
            wantError: true,
        },
        {
            name: "missing name",
            user: factory.NewUser(p).
                WithName("").
                Build(),
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

- [Property-Based Testing](/docs/property-based-testing/) - Use factories for PBT
- [Advanced Topics](/docs/advanced/) - Custom matchers, patterns
- [Examples](/docs/examples/) - More real-world examples
