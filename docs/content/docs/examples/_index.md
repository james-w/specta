---
title: "Examples & Recipes"
weight: 9
---

# Examples & Recipes

Real-world examples and common patterns.

## Complete CRUD Example

<!-- skip-test -->
```go
package users_test

import (
    "testing"
    . "github.com/james-w/specta"
    "myapp/factory"
    "myapp/users"
)

func TestUserCRUD(t *testing.T) {
    p := specta.New()
    db := setupTestDB(t)

    t.Run("Create", func(t *testing.T) {
        input := factory.NewUser(p).
            WithName("Alice").
            WithEmail("alice@example.com").
            Build()

        created := db.CreateUser(input)

        specta.AssertThat(t, created, factory.MatchUser().
            WithID(specta.Not(IsEmpty())).
            WithName(specta.Equal("Alice")).
            WithEmail(specta.Equal("alice@example.com")).
            WithCreatedAt(specta.Not(IsZero())))
    })

    t.Run("Read", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        db.CreateUser(user)

        found, err := db.GetUser(user.ID)

        specta.AssertThat(t, err, BeNil())
        specta.AssertThat(t, found, factory.MatchUser().
            WithID(specta.Equal(user.ID)).
            WithName(specta.Equal(user.Name)))
    })

    t.Run("Update", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        created := db.CreateUser(user)

        updates := factory.NewUser(p).
            WithName("Updated Name").
            Build()

        updated := db.UpdateUser(created.ID, updates)

        specta.AssertThat(t, updated, factory.MatchUser().
            WithID(specta.Equal(created.ID)).
            WithName(specta.Equal("Updated Name")).
            WithUpdatedAt(specta.GreaterThan(created.UpdatedAt)))
    })

    t.Run("Delete", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        created := db.CreateUser(user)

        err := db.DeleteUser(created.ID)
        specta.AssertThat(t, err, BeNil())

        _, err = db.GetUser(created.ID)
        specta.AssertThat(t, err, specta.Not(BeNil()))
    })
}
```

## REST API Testing

<!-- skip-test -->
```go
package api_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    . "github.com/james-w/specta"
    "myapp/factory"
)

func TestUserAPI(t *testing.T) {
    p := specta.New()
    server := setupTestServer(t)

    t.Run("GET /users/:id", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        server.DB.CreateUser(user)

        req := httptest.NewRequest("GET", "/users/"+user.ID, nil)
        resp := httptest.NewRecorder()
        server.ServeHTTP(resp, req)

        specta.AssertThat(t, resp.Code, specta.Equal(http.StatusOK))

        var result User
        json.NewDecoder(resp.Body).Decode(&result)

        specta.AssertThat(t, result, factory.MatchUser().
            WithID(specta.Equal(user.ID)).
            WithName(specta.Equal(user.Name)))
    })

    t.Run("POST /users", func(t *testing.T) {
        payload := factory.NewUser(p).
            WithName("Alice").
            WithEmail("alice@example.com").
            Build()

        body, _ := json.Marshal(payload)
        req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
        resp := httptest.NewRecorder()
        server.ServeHTTP(resp, req)

        specta.AssertThat(t, resp.Code, specta.Equal(http.StatusCreated))

        var created User
        json.NewDecoder(resp.Body).Decode(&created)

        specta.AssertThat(t, created, factory.MatchUser().
            WithID(specta.Not(IsEmpty())).
            WithName(specta.Equal("Alice")).
            WithEmail(specta.Equal("alice@example.com")))
    })

    t.Run("PUT /users/:id", func(t *testing.T) {
        user := factory.NewUser(p).Build()
        created := server.DB.CreateUser(user)

        updates := map[string]string{"name": "Updated"}
        body, _ := json.Marshal(updates)
        req := httptest.NewRequest("PUT", "/users/"+created.ID, bytes.NewReader(body))
        resp := httptest.NewRecorder()
        server.ServeHTTP(resp, req)

        specta.AssertThat(t, resp.Code, specta.Equal(http.StatusOK))

        var updated User
        json.NewDecoder(resp.Body).Decode(&updated)

        specta.AssertThat(t, updated, factory.MatchUser().
            WithID(specta.Equal(created.ID)).
            WithName(specta.Equal("Updated")))
    })
}
```

## Complex Object Graphs

<!-- skip-test -->
```go
func TestOrderProcessing(t *testing.T) {
    p := specta.New()

    // Build a complete order with user, items, and addresses
    order := factory.NewOrder(p).
        WithUser(factory.NewUser(p).
            WithName("Alice").
            WithEmail("alice@example.com").
            Build()).
        WithShippingAddress(factory.NewAddress(p).
            WithStreet("123 Main St").
            WithCity("Seattle").
            WithZipCode("98101").
            Build()).
        WithBillingAddress(factory.NewAddress(p).
            WithStreet("456 Oak Ave").
            WithCity("Portland").
            WithZipCode("97201").
            Build()).
        WithItems([]Item{
            factory.NewItem(p).
                WithSKU("WIDGET-1").
                WithPrice(1999).
                WithQuantity(2).
                Build(),
            factory.NewItem(p).
                WithSKU("GADGET-2").
                WithPrice(2999).
                WithQuantity(1).
                Build(),
        }).
        Build()

    result := ProcessOrder(order)

    specta.AssertThat(t, result, factory.MatchOrder().
        WithStatus(specta.Equal("processed")).
        WithTotal(specta.Equal(6997)). // 1999*2 + 2999*1
        WithUser(factory.MatchUser().
            WithName(specta.Equal("Alice"))).
        WithShippingAddress(factory.MatchAddress().
            WithCity(specta.Equal("Seattle"))))
}
```

## Validation Testing

<!-- skip-test -->
```go
func TestUserValidation(t *testing.T) {
    p := specta.New()

    tests := []struct {
        name      string
        user      User
        wantError bool
        errMatch  specta.Matcher[string]
    }{
        {
            name: "valid user",
            user: factory.NewUser(p).
                WithName("Alice").
                WithEmail("alice@example.com").
                WithAge(30).
                Build(),
            wantError: false,
        },
        {
            name: "missing name",
            user: factory.NewUser(p).
                WithName("").
                Build(),
            wantError: true,
            errMatch:  ContainString("name"),
        },
        {
            name: "invalid email",
            user: factory.NewUser(p).
                WithEmail("not-an-email").
                Build(),
            wantError: true,
            errMatch:  ContainString("email"),
        },
        {
            name: "negative age",
            user: factory.NewUser(p).
                WithAge(-1).
                Build(),
            wantError: true,
            errMatch:  ContainString("age"),
        },
        {
            name: "age too high",
            user: factory.NewUser(p).
                WithAge(200).
                Build(),
            wantError: true,
            errMatch:  ContainString("age"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUser(tt.user)

            if tt.wantError {
                specta.AssertThat(t, err, NotBeNil())
                specta.AssertThat(t, err.Error(), tt.errMatch)
            } else {
                specta.AssertThat(t, err, BeNil())
            }
        })
    }
}
```

## Property-Based Testing Example

<!-- skip-test -->
```go
func TestUserSerializationRoundTrip(t *testing.T) {
    p := specta.New()

    for i := 0; i < 100; i++ {
        // Generate varied test data
        original := factory.NewOrder(p).
            WithUser(factory.NewUser(p).Build()).
            WithItems(generateRandomItems(p, 1+p.Next()%10)).
            Build()

        // Serialize
        data, err := json.Marshal(original)
        specta.AssertThat(t, err, BeNil())

        // Deserialize
        var decoded Order
        err = json.Unmarshal(data, &decoded)
        specta.AssertThat(t, err, BeNil())

        // Property: round-trip preserves data
        specta.AssertThat(t, decoded, Deepspecta.Equal(original))
    }
}

func generateRandomItems(p Primitives, count int) []Item {
    items := make([]Item, count)
    for i := 0; i < count; i++ {
        items[i] = factory.NewItem(p).
            WithPrice(100 + p.Next()%10000).
            WithQuantity(1 + p.Next()%10).
            Build()
    }
    return items
}
```

## Reusable Matchers Pattern

<!-- skip-test -->
```go
// matchers/common.go
package matchers

import . "github.com/james-w/specta"

// Email matchers
var (
    ValidEmail = specta.AllOf(
        ContainString("@"),
        MatchRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
    )

    CompanyEmail = specta.AllOf(
        ValidEmail,
        HaveSuffix("@mycompany.com"),
    )

    PersonalEmail = specta.AllOf(
        ValidEmail,
        specta.Not(HaveSuffix("@mycompany.com")),
    )
)

// User matchers
var (
    ActiveUser = factory.MatchUser().
        WithStatus(specta.Equal("active"))

    AdminUser = factory.MatchUser().
        WithRole(specta.Equal("admin")).
        WithPermissions(ContainAll("read", "write", "delete"))

    ValidUser = factory.MatchUser().
        WithEmail(ValidEmail).
        WithAge(specta.AllOf(specta.GreaterThan(0), specta.LessThan(150)))
)

// Usage in tests
func TestUser(t *testing.T) {
    user := CreateUser("alice@mycompany.com")

    specta.AssertThat(t, user.Email, CompanyEmail)
    specta.AssertThat(t, user, ValidUser)
}
```

## Testing State Machines

<!-- skip-test -->
```go
func TestUserStatusTransitions(t *testing.T) {
    p := specta.New()

    tests := []struct {
        name        string
        initial     string
        transition  string
        expected    string
        shouldError bool
    }{
        {
            name:       "pending to active",
            initial:    "pending",
            transition: "activate",
            expected:   "active",
        },
        {
            name:       "active to suspended",
            initial:    "active",
            transition: "suspend",
            expected:   "suspended",
        },
        {
            name:        "pending to suspended (invalid)",
            initial:     "pending",
            transition:  "suspend",
            shouldError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user := factory.NewUser(p).
                WithStatus(tt.initial).
                Build()

            result, err := TransitionUser(user, tt.transition)

            if tt.shouldError {
                specta.AssertThat(t, err, NotBeNil())
            } else {
                specta.AssertThat(t, err, BeNil())
                specta.AssertThat(t, result, factory.MatchUser().
                    WithStatus(specta.Equal(tt.expected)))
            }
        })
    }
}
```

## Migration Pattern

### Before: Traditional Go Testing

<!-- skip-test -->
```go
func TestOldWay(t *testing.T) {
    user := User{
        ID:        "user-1",
        Name:      "Alice",
        Email:     "alice@example.com",
        Age:       30,
        Role:      "admin",
        CreatedAt: time.Now(),
    }

    result := ProcessUser(user)

    if result.ID != user.ID {
        t.Errorf("expected ID %s, got %s", user.ID, result.ID)
    }
    if result.Name != "Alice" {
        t.Errorf("expected name Alice, got %s", result.Name)
    }
    if result.Status != "processed" {
        t.Errorf("expected status processed, got %s", result.Status)
    }
}
```

### After: With specta

<!-- skip-test -->
```go
func TestNewWay(t *testing.T) {
    p := specta.New()

    user := factory.NewUser(p).
        WithName("Alice").
        Build()

    result := ProcessUser(user)

    specta.AssertThat(t, result, factory.MatchUser().
        WithID(specta.Equal(user.ID)).
        WithName(specta.Equal("Alice")).
        WithStatus(specta.Equal("processed")))
}
```

## Common Recipes

### Setup Helpers

<!-- skip-test -->
```go
// helpers/test_helpers.go
package helpers

func SetupUserWithOrders(p Primitives, orderCount int) (User, []Order) {
    user := factory.NewUser(p).Build()

    orders := make([]Order, orderCount)
    for i := 0; i < orderCount; i++ {
        orders[i] = factory.NewOrder(p).
            WithUser(user).
            WithItems(factory.RandomItems(p, 1+p.Next()%5)).
            Build()
    }

    return user, orders
}

func AdminUserWithPermissions(p Primitives, permissions ...string) User {
    return factory.NewUser(p).
        WithRole("admin").
        WithPermissions(permissions).
        Build()
}
```

### Custom Primitives

<!-- skip-test -->
```go
// Create specialized primitives for your domain
type ProductPrimitives struct {
    specta.Source
}

func (p *ProductPrimitives) SKU() string {
    return fmt.Sprintf("SKU-%06d", p.Next())
}

func (p *ProductPrimitives) Price() int {
    // Prices between $10 and $10000
    return 1000 + (p.Next() % 9000) * 100
}

// Usage
func TestProducts(t *testing.T) {
    p := &ProductPrimitives{Gen: specta.New()}

    product := factory.NewProduct(p).
        WithSKU(p.SKU()).
        WithPrice(p.Price()).
        Build()

    // Test with domain-specific generated data
}
```

## Next Steps

- Explore the [showcase directory](https://github.com/james-w/specta/tree/main/showcase) in the repo
- Check out [example directory](https://github.com/james-w/specta/tree/main/example) for more patterns
- Read the [API Reference]({{< relref "/docs/api-reference/" >}}) for complete documentation
- Visit the [GitHub repository](https://github.com/james-w/specta) to contribute
