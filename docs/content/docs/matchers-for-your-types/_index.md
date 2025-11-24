---
title: "Matchers for Your Types"
weight: 3
---

# Matchers for Your Types

Core matchers are great, but what about testing your own custom types? This section shows how to get fluent, composable matchers for your structs.

## The Problem

Suppose you have a `User` struct:

<!-- skip-test -->
```go
type User struct {
    ID        string
    Name      string
    Email     string
    Age       int
    CreatedAt time.Time
}
```

Testing without matchers is verbose and brittle:

<!-- skip-test -->
```go
func TestCreateUser(t *testing.T) {
    user := CreateUser("Alice", "alice@example.com")

    // Verbose and breaks when new fields are added
    if user.Name != "Alice" {
        t.Errorf("expected name Alice, got %s", user.Name)
    }
    if user.Email != "alice@example.com" {
        t.Errorf("expected email alice@example.com, got %s", user.Email)
    }
    if user.Age <= 0 {
        t.Error("expected positive age")
    }
    // What about ID? CreatedAt? Do we test them everywhere?
}
```

Using basic matchers improves error messages, but isn't composable:

<!-- skip-test -->
```go
func TestCreateUser(t *testing.T) {
    user := CreateUser("Alice", "alice@example.com")

    // Better, but can't reuse these assertions
    specta.AssertThat(t, user.Name, specta.Equal("Alice"))
    specta.AssertThat(t, user.Email, specta.Equal("alice@example.com"))
    specta.AssertThat(t, user.Age, specta.GreaterThan(0))
}

func TestUpdateUser(t *testing.T) {
    user := UpdateUser(existingUser, "Bob")

    // Have to repeat the same field assertions
    specta.AssertThat(t, user.Name, specta.Equal("Bob"))
    specta.AssertThat(t, user.Email, specta.Equal("alice@example.com"))
    // Can't express "same email as before" as a reusable matcher
}
```

**The Problem:**
- Can't build a reusable "valid user" matcher to use across tests
- Can't compose field matchers into higher-level concepts
- Every test duplicates the same field-by-field assertions

**What we want:**

<!-- skip-test -->
```go
specta.AssertThat(t, user, MatchUser().
    WithName(specta.Equal("Alice")).
    WithEmail(specta.Contains("example.com")))
```

Clean, fluent, partial matching - only assert what matters!

## Option 1: Manual Custom Matchers

You can implement the `Matcher[T]` interface yourself:

<!-- skip-test -->
```go
type userMatcher struct {
    nameMatcher  specta.Matcher[string]
    emailMatcher specta.Matcher[string]
}

func MatchUser() *userMatcher {
    return &userMatcher{}
}

func (m *userMatcher) WithName(matcher specta.Matcher[string]) *userMatcher {
    m.nameMatcher = matcher
    return m
}

func (m *userMatcher) WithEmail(matcher specta.Matcher[string]) *userMatcher {
    m.emailMatcher = matcher
    return m
}

func (m *userMatcher) Match(user User) specta.MatchResult {
    // Implementation details...
}
```

This works, but it's tedious and error-prone. There's a better way...

## Option 2: Code Generation (Recommended)

specta can automatically generate matchers (and factories) for your types!

### Step 1: Create `specta.yaml`

In your package directory, create a configuration file:

<!-- skip-test -->
```yaml
# specta.yaml
version: 1
targets:
  - package: .
    types:
      include:
        - User
```

The generator will introspect your `User` struct and generate matchers for all its exported (public) fields automatically.

### Step 2: Run the Generator

<!-- skip-test -->
```bash
go run github.com/james-w/specta/cmd -config specta.yaml
```

This generates three files in `factory/`:
- `user_matcher_gen.go` - Matcher builders ← We'll use this!
- `user_gen.go` - Factory recipes (covered in next section)
- `spec/user_gen.go` - Low-level Factory builders (covered in next section)

### Step 3: Use Generated Matchers

<!-- skip-test -->
```go
package mypackage_test

import (
    "github.com/james-w/specta"
    "testing"
    "mypackage/factory"
)

func TestCreateUser(t *testing.T) {
    user := CreateUser("Alice", "alice@example.com")

    // Clean, fluent, partial matching!
    specta.AssertThat(t, user, factory.MatchUser().
        WithName(specta.Equal("Alice")).
        WithEmail(specta.Contains("example.com")))
}
```

**Benefits:**
- Only assert fields that matter for this test
- Other fields (ID, CreatedAt) are ignored
- Compose with core matchers
- Structured error messages
- Refactoring-safe: adding new fields doesn't break tests

## Understanding Generated Files

### Build Tags

Generated files include:
<!-- skip-test -->
```go
//go:build !ignore_testgen
```

This allows excluding them from linting/analysis tools while keeping them in your tests.

### What Gets Generated

**`*_matcher_gen.go`**: Fluent matcher builders

<!-- skip-test -->
```go
// Generated matcher
func MatchUser() *UserMatcher {
    return &UserMatcher{}
}

func (m *UserMatcher) WithName(matcher specta.Matcher[string]) *UserMatcher {
    m.matchers = append(m.matchers, fieldMatcher{"Name", matcher})
    return m
}

func (m *UserMatcher) WithEmail(matcher specta.Matcher[string]) *UserMatcher {
    // ...
}

func (m *UserMatcher) Match(user User) specta.MatchResult {
    // Applies all configured field matchers
}
```

## Using Generated Matchers

### Basic Field Matching

<!-- skip-test -->
```go
specta.AssertThat(t, user, MatchUser().
    WithName(specta.Equal("Alice")).
    WithAge(GreaterThan(18)))
```

### Compose with Core Matchers

<!-- skip-test -->
```go
specta.AssertThat(t, user, MatchUser().
    WithEmail(AllOf(
        ContainString("@"),
        HaveSuffix(".com"),
    )).
    WithName(specta.Not(specta.Equal(""))))
```

### Nested Struct Matching

If `User` has a nested `Address` struct:

<!-- skip-test -->
```go
specta.AssertThat(t, user, MatchUser().
    WithAddress(MatchAddress().
        WithCity(specta.Equal("Seattle")).
        WithZipCode(specta.MatchesRegex(`^\d{5}$`))))
```

### Partial Matching in Action

<!-- skip-test -->
```go
// Test 1: Only care about name
specta.AssertThat(t, user, MatchUser().WithName(specta.Equal("Alice")))

// Test 2: Only care about email domain
specta.AssertThat(t, user, MatchUser().WithEmail(specta.HasSuffix("@company.com")))

// Test 3: Validate age and creation time
specta.AssertThat(t, user, MatchUser().
    WithAge(GreaterThan(0)).
    WithCreatedAt(Not(specta.IsNil[time.Time]())))
```

Each test asserts exactly what it cares about. Adding new fields to `User` won't break these tests.

### Reusable Matchers

Build higher-level matchers from the generated ones:

<!-- skip-test -->
```go
// Define reusable matchers for common patterns
func ValidUser() *factory.UserMatcher {
    return factory.MatchUser().
        WithName(specta.Not(specta.Equal(""))).
        WithEmail(specta.Contains("@")).
        WithAge(specta.GreaterThan(0))
}

func AdminUser() *factory.UserMatcher {
    return ValidUser().
        WithEmail(specta.HasSuffix("@company.com"))
}

// Use across tests
func TestCreateUser(t *testing.T) {
    user := CreateUser("Alice", "alice@company.com", 30)
    specta.AssertThat(t, user, AdminUser())
}

func TestPromoteUser(t *testing.T) {
    user := PromoteToAdmin(regularUser)
    specta.AssertThat(t, user, AdminUser())
}
```

**When you add a new field** (e.g., `Status string`):
1. Regenerate: `go run github.com/james-w/specta/cmd -config specta.yaml`
2. Update `ValidUser()` if the new field should be validated:
   <!-- skip-test -->
   ```go
   func ValidUser() *factory.UserMatcher {
       return factory.MatchUser().
           WithName(specta.Not(specta.Equal(""))).
           WithEmail(specta.Contains("@")).
           WithAge(specta.GreaterThan(0)).
           WithStatus(specta.Equal("active"))  // One update here
   }
   ```
3. All tests using `ValidUser()` now validate the new field - **zero test changes needed!**

## Regenerating After Changes

When you modify your types, regenerate:

<!-- skip-test -->
```bash
go run github.com/james-w/specta/cmd -config specta.yaml
```

The generator:
1. Parses your type definitions
2. Generates type-safe matchers
3. Type-checks the generated code
4. Only writes if successful

**Important:** Commit generated files with your source changes, or CI will fail!

## Next Steps

Now you have matchers for your types! Next, learn about:
- [Test Data Factories]({{< relref "/docs/factories/" >}}) - The flip side of matchers
- [Factories + Matchers Together]({{< relref "/docs/factories-and-matchers/" >}}) - The complete pattern
- [API Reference]({{< relref "/docs/api-reference/" >}}) - Full generator configuration options
