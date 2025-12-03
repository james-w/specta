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
        input := factory.User().
            Name("Alice").
            Email("alice@example.com").
            Build(p)

        created := db.CreateUser(input)

        specta.AssertThat(t, created, factory.UserMatches().
            ID(specta.Not(specta.IsEmpty[string]())).
            Name(specta.Equal("Alice")).
            Email(specta.Equal("alice@example.com")).
            CreatedAt(specta.Not(specta.IsZero[time.Time]())))
    })

    t.Run("Read", func(t *testing.T) {
        user := factory.User().Build(p)
        db.CreateUser(user)

        found, err := db.GetUser(user.ID)

        specta.AssertThat(t, err, specta.NoErr())
        specta.AssertThat(t, found, factory.UserMatches().
            ID(specta.Equal(user.ID)).
            Name(specta.Equal(user.Name)))
    })

    t.Run("Update", func(t *testing.T) {
        user := factory.User().Build(p)
        created := db.CreateUser(user)

        updates := factory.User().
            Name("Updated Name").
            Build(p)

        updated := db.UpdateUser(created.ID, updates)

        specta.AssertThat(t, updated, factory.UserMatches().
            ID(specta.Equal(created.ID)).
            Name(specta.Equal("Updated Name")).
            UpdatedAt(specta.GreaterThan(created.UpdatedAt)))
    })

    t.Run("Delete", func(t *testing.T) {
        user := factory.User().Build(p)
        created := db.CreateUser(user)

        err := db.DeleteUser(created.ID)
        specta.AssertThat(t, err, specta.NoErr())

        _, err = db.GetUser(created.ID)
        specta.AssertThat(t, err, specta.IsError())
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
        user := factory.User().Build(p)
        server.DB.CreateUser(user)

        req := httptest.NewRequest("GET", "/users/"+user.ID, nil)
        resp := httptest.NewRecorder()
        server.ServeHTTP(resp, req)

        specta.AssertThat(t, resp.Code, specta.Equal(http.StatusOK))

        var result User
        json.NewDecoder(resp.Body).Decode(&result)

        specta.AssertThat(t, result, factory.UserMatches().
            ID(specta.Equal(user.ID)).
            Name(specta.Equal(user.Name)))
    })

    t.Run("POST /users", func(t *testing.T) {
        payload := factory.User().
            Name("Alice").
            Email("alice@example.com").
            Build(p)

        body, _ := json.Marshal(payload)
        req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
        resp := httptest.NewRecorder()
        server.ServeHTTP(resp, req)

        specta.AssertThat(t, resp.Code, specta.Equal(http.StatusCreated))

        var created User
        json.NewDecoder(resp.Body).Decode(&created)

        specta.AssertThat(t, created, factory.UserMatches().
            ID(specta.Not(specta.IsEmpty[string]())).
            Name(specta.Equal("Alice")).
            Email(specta.Equal("alice@example.com")))
    })

    t.Run("PUT /users/:id", func(t *testing.T) {
        user := factory.User().Build(p)
        created := server.DB.CreateUser(user)

        updates := map[string]string{"name": "Updated"}
        body, _ := json.Marshal(updates)
        req := httptest.NewRequest("PUT", "/users/"+created.ID, bytes.NewReader(body))
        resp := httptest.NewRecorder()
        server.ServeHTTP(resp, req)

        specta.AssertThat(t, resp.Code, specta.Equal(http.StatusOK))

        var updated User
        json.NewDecoder(resp.Body).Decode(&updated)

        specta.AssertThat(t, updated, factory.UserMatches().
            ID(specta.Equal(created.ID)).
            Name(specta.Equal("Updated")))
    })
}
```

## Complex Object Graphs

<!-- skip-test -->
```go
func TestOrderProcessing(t *testing.T) {
    p := specta.New()

    // Build a complete order with user, items, and addresses
    order := factory.Order().
        User(factory.User().
            Name("Alice").
            Email("alice@example.com").
            Build(p)).
        ShippingAddress(factory.Address().
            Street("123 Main St").
            City("Seattle").
            ZipCode("98101").
            Build(p)).
        BillingAddress(factory.Address().
            Street("456 Oak Ave").
            City("Portland").
            ZipCode("97201").
            Build(p)).
        Items([]Item{
            factory.Item().
                SKU("WIDGET-1").
                Price(1999).
                Quantity(2).
                Build(p),
            factory.Item().
                SKU("GADGET-2").
                Price(2999).
                Quantity(1).
                Build(p),
        }).
        Build(p)

    result := ProcessOrder(order)

    specta.AssertThat(t, result, factory.OrderMatches().
        Status(specta.Equal("processed")).
        Total(specta.Equal(6997)). // 1999*2 + 2999*1
        User(factory.UserMatches().
            Name(specta.Equal("Alice"))).
        ShippingAddress(factory.AddressMatches().
            City(specta.Equal("Seattle"))))
}
```

## Validation Testing

<!-- skip-test -->
```go
func TestUserValidation(t *testing.T) {
    p := specta.New()

    tests := []struct {
        name      string
        recipe    UserRecipe
        wantError bool
        errMatch  specta.Matcher[string]
    }{
        {
            name: "valid user",
            recipe: factory.User().
                Name("Alice").
                Email("alice@example.com").
                Age(30),
            wantError: false,
        },
        {
            name: "missing name",
            recipe: factory.User().
                Name(""),
            wantError: true,
            errMatch:  specta.Contains("name"),
        },
        {
            name: "invalid email",
            recipe: factory.User().
                Email("not-an-email"),
            wantError: true,
            errMatch:  specta.Contains("email"),
        },
        {
            name: "negative age",
            recipe: factory.User().
                Age(-1),
            wantError: true,
            errMatch:  specta.Contains("age"),
        },
        {
            name: "age too high",
            recipe: factory.User().
                Age(200),
            wantError: true,
            errMatch:  specta.Contains("age"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user := tt.recipe.Build(p)
            err := ValidateUser(user)

            if tt.wantError {
                specta.AssertThat(t, err, specta.IsError())
                specta.AssertThat(t, err.Error(), tt.errMatch)
            } else {
                specta.AssertThat(t, err, specta.NoErr())
            }
        })
    }
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
        specta.Contains("@"),
        specta.MatchesRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
    )

    CompanyEmail = specta.AllOf(
        ValidEmail,
        specta.HasSuffix("@mycompany.com"),
    )

    PersonalEmail = specta.AllOf(
        ValidEmail,
        specta.Not(specta.HasSuffix("@mycompany.com")),
    )
)

// User matchers
var (
    ActiveUser = factory.UserMatches().
        Status(specta.Equal("active"))

    AdminUser = factory.UserMatches().
        Role(specta.Equal("admin")).
        Permissions(specta.ContainsAllElements[string]("read", "write", "delete"))

    ValidUser = factory.UserMatches().
        Email(ValidEmail).
        Age(specta.AllOf(specta.GreaterThan(0), specta.LessThan(150)))
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
            user := factory.User().
                Status(tt.initial).
                Build(p)

            result, err := TransitionUser(user, tt.transition)

            if tt.shouldError {
                specta.AssertThat(t, err, specta.IsError())
            } else {
                specta.AssertThat(t, err, specta.NoErr())
                specta.AssertThat(t, result, factory.UserMatches().
                    Status(specta.Equal(tt.expected)))
            }
        })
    }
}
```

## Common Recipes

### Setup Helpers

<!-- skip-test -->
```go
// helpers/test_helpers.go
package helpers

func SetupUserWithOrders(p specta.Source, orderCount int) (User, []Order) {
    user := factory.User().Build(p)

    // Use Many() to generate multiple orders with unique values
    orders := factory.Order().
        User(user).
        Many(orderCount, p)

    return user, orders
}

func AdminUserWithPermissions(p specta.Source, permissions ...string) User {
    return factory.User().
        Role("admin").
        Permissions(permissions).
        Build(p)
}
```

## Next Steps

- Explore the [showcase directory](https://github.com/james-w/specta/tree/main/showcase) in the repo
- Check out [example directory](https://github.com/james-w/specta/tree/main/example) for more patterns
- Read the [API Reference]({{< relref "/docs/api-reference/" >}}) for complete documentation
- Visit the [GitHub repository](https://github.com/james-w/specta) to contribute
