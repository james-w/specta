---
title: "Matchers for Your Types"
weight: 3
---

# Matchers for Your Types

Core matchers are great, but what about testing your own custom types? This section shows how to get fluent, composable matchers for your structs.

## The Problem

Suppose you have a `User` struct:

```go
type User struct {
    ID        string
    Name      string
    Email     string
    Age       int
    CreatedAt time.Time
}
```

Testing with basic matchers becomes verbose and brittle:

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

**Problems:**
- Repetitive field-by-field assertions
- Tests break when adding new fields (even optional ones)
- No composition or reusability
- Error messages aren't structured

**What we want:**

```go
specta.AssertThat(t, user, MatchUser().
    WithName(specta.Equal("Alice")).
    WithEmail(specta.Contains("example.com")))
```

Clean, fluent, partial matching - only assert what matters!

## Option 1: Manual Custom Matchers

You can implement the `Matcher[T]` interface yourself:

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

### Step 2: Run the Generator

```bash
go run github.com/james-w/specta/cmd/main.go -config specta.yaml
```

This generates three files in `factory/`:
- `user_gen.go` - Low-level spec API
- `user_matcher_gen.go` - Matcher builders ← We'll use this!
- `factory/user_gen.go` - Factory builders (covered in next section)

### Step 3: Use Generated Matchers

```go
package mypackage_test

import (
    "github.com/james-w/specta"
    "testing"
    . "github.com/james-w/specta"
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
```go
//go:build !ignore_testgen
```

This allows excluding them from linting/analysis tools while keeping them in your tests.

### What Gets Generated

**`*_matcher_gen.go`**: Fluent matcher builders

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

```go
specta.AssertThat(t, user, MatchUser().
    WithName(specta.Equal("Alice")).
    WithAge(GreaterThan(18)))
```

### Compose with Core Matchers

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

```go
specta.AssertThat(t, user, MatchUser().
    WithAddress(MatchAddress().
        WithCity(specta.Equal("Seattle")).
        WithZipCode(specta.MatchesRegex(`^\d{5}$`))))
```

### Partial Matching in Action

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

Each test asserts exactly what it cares about. Adding new fields to `User` won't break these tests!

## Example: Before & After

**Before (verbose, brittle):**

```go
func TestUserCreation(t *testing.T) {
    user := CreateUser("Alice", "alice@example.com", 30)

    if user.ID == "" {
        t.Error("ID should not be empty")
    }
    if user.Name != "Alice" {
        t.Errorf("expected name Alice, got %s", user.Name)
    }
    if user.Email != "alice@example.com" {
        t.Errorf("expected email alice@example.com, got %s", user.Email)
    }
    if user.Age != 30 {
        t.Errorf("expected age 30, got %d", user.Age)
    }
    if user.CreatedAt.IsZero() {
        t.Error("CreatedAt should be set")
    }
}
```

**After (clean, composable, focused):**

```go
func TestUserCreation(t *testing.T) {
    user := CreateUser("Alice", "alice@example.com", 30)

    specta.AssertThat(t, user, MatchUser().
        WithID(Not(specta.Equal(""))).
        WithName(specta.Equal("Alice")).
        WithEmail(specta.Contains("alice")).
        WithAge(Equal(30)).
        WithCreatedAt(Not(specta.IsZero[time.Time]())))
}
```

Or, focus on what matters:

```go
func TestUserCreation(t *testing.T) {
    user := CreateUser("Alice", "alice@example.com", 30)

    // Only assert the fields this test cares about
    specta.AssertThat(t, user, MatchUser().
        WithName(specta.Equal("Alice")).
        WithEmail(specta.Equal("alice@example.com")))
}
```

## Regenerating After Changes

When you modify your types, regenerate:

```bash
go run github.com/james-w/specta/cmd/main.go -config specta.yaml
```

The generator:
1. Parses your type definitions
2. Generates type-safe matchers
3. Type-checks the generated code
4. Only writes if successful

**Important:** Commit generated files with your source changes, or CI will fail!

## Next Steps

Now you have matchers for your types! Next, learn about:
- [Test Data Factories](/docs/factories/) - The flip side of matchers
- [Factories + Matchers Together](/docs/factories-and-matchers/) - The complete pattern
- [API Reference](/docs/api-reference/) - Full generator configuration options
