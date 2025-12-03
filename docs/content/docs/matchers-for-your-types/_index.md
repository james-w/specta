---
title: "Matchers for Your Types"
weight: 3
---

<!-- setup
package doctest

import (
	"github.com/james-w/specta"
)

type User struct {
	ID    string
	Name  string
	Email string
	Age   int
}

// Simulate generated matcher (simplified for doc tests)
type UserMatcher struct {
	idMatcher    specta.Matcher[string]
	nameMatcher  specta.Matcher[string]
	emailMatcher specta.Matcher[string]
	ageMatcher   specta.Matcher[int]
}

func UserMatches() UserMatcher {
	return UserMatcher{}
}

func (m UserMatcher) ID(matcher specta.Matcher[string]) UserMatcher {
	m.idMatcher = matcher
	return m
}

func (m UserMatcher) Name(matcher specta.Matcher[string]) UserMatcher {
	m.nameMatcher = matcher
	return m
}

func (m UserMatcher) Email(matcher specta.Matcher[string]) UserMatcher {
	m.emailMatcher = matcher
	return m
}

func (m UserMatcher) Age(matcher specta.Matcher[int]) UserMatcher {
	m.ageMatcher = matcher
	return m
}

func (m UserMatcher) Matches(actual User) specta.MatchResult {
	// Simplified: just check each field if matcher is set
	if m.nameMatcher != nil {
		if result := m.nameMatcher.Matches(actual.Name); !result.Matched {
			return specta.MatchResult{Matched: false, Message: "Name: " + result.Message}
		}
	}
	if m.emailMatcher != nil {
		if result := m.emailMatcher.Matches(actual.Email); !result.Matched {
			return specta.MatchResult{Matched: false, Message: "Email: " + result.Message}
		}
	}
	if m.ageMatcher != nil {
		if result := m.ageMatcher.Matches(actual.Age); !result.Matched {
			return specta.MatchResult{Matched: false, Message: "Age: " + result.Message}
		}
	}
	return specta.MatchResult{Matched: true}
}

var (
	user = User{ID: "user123", Name: "Alice", Email: "alice@example.com", Age: 30}
)

func CreateUser(name, email string) User {
	return User{Name: name, Email: email, Age: 30}
}
-->

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
specta.AssertThat(t, user, UserMatches().
    Name(specta.Equal("Alice")).
    Email(specta.Contains("example.com")))
```

Clean, fluent, partial matching - only assert what matters!

## Option 1: Manual Custom Matchers

You can implement the `Matcher[T]` interface yourself:

<!-- skip-test -->
```go
type UserMatcher struct {
    nameMatcher  specta.Matcher[string]
    emailMatcher specta.Matcher[string]
}

func UserMatches() UserMatcher {
    return UserMatcher{}
}

func (m UserMatcher) Name(matcher specta.Matcher[string]) UserMatcher {
    m.nameMatcher = matcher
    return m
}

func (m UserMatcher) Email(matcher specta.Matcher[string]) UserMatcher {
    m.emailMatcher = matcher
    return m
}

// Implements the Matcher[User] interface
func (m UserMatcher) Matches(actual User) specta.MatchResult {
    // Only check fields that have matchers set (partial matching)
    if m.nameMatcher != nil {
        if result := m.nameMatcher.Matches(actual.Name); !result.Matched {
            return specta.MatchResult{
                Matched: false,
                Message: "Name: " + result.Message,
            }
        }
    }
    if m.emailMatcher != nil {
        if result := m.emailMatcher.Matches(actual.Email); !result.Matched {
            return specta.MatchResult{
                Matched: false,
                Message: "Email: " + result.Message,
            }
        }
    }
    return specta.MatchResult{Matched: true}
}
```

Now you can use it directly:

<!-- skip-test -->
```go
specta.AssertThat(t, user, UserMatches().
    Name(specta.Equal("Alice")).
    Email(validEmail))  // No wrapper needed - UserMatcher IS a Matcher[User]
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
  - package: .               # Current package
    types:
      include:
        - User               # Generate for User type
```

The generator will introspect your `User` struct and generate code for all exported (public) fields automatically.

### Step 2: Run the Generator

<!-- skip-test -->
```bash
go run github.com/james-w/specta/cmd -config specta.yaml
```

This generates **three files** in `factory/` directory:

1. **`factory/spec/user_gen.go`** - Low-level factory builders (covered in [Factories section]({{< relref "/docs/factories/" >}}))
2. **`factory/user_gen.go`** - Recipe fluent API for test data (covered in [Factories section]({{< relref "/docs/factories/" >}}))
3. **`factory/user_matcher_gen.go`** - Matcher builders **← We'll use this!**

All files include build tags:
<!-- skip-test -->
```go
//go:build !ignore_testgen
```

This allows excluding them from linting while keeping them in your tests.

### Step 3: Use Generated Matchers

```go
func TestCreateUser(t *testing.T) {
    user := CreateUser("Alice", "alice@example.com")

    // Clean, fluent, partial matching!
    specta.AssertThat(t, user, UserMatches().
        Name(specta.Equal("Alice")).
        Email(specta.Contains("example.com")))
}
```

**Key points**:
- Constructor: `UserMatches()` creates a new matcher
- Methods: `Name()`, `Email()`, etc. (field names, no "With" prefix)
- Direct usage: The matcher implements `Matcher[User]` interface directly
- Partial matching: Only specified fields are validated

**Benefits:**
- Only assert fields that matter for this test
- Other fields (ID, Age) are ignored
- Compose with core matchers
- Structured error messages
- Refactoring-safe: adding new fields doesn't break tests

## Constructor-Based Generation

For types created via constructor functions (not direct struct instantiation), use constructor-based generation.

### When to Use

**Use field-based** (previous section) when:
- Your type is a simple struct with public fields
- Tests create instances via struct literals: `User{Name: "Alice", ...}`

**Use constructor-based** when:
- Your type is created via a constructor function
- Constructor performs validation or initialization logic
- Example: `account := NewBankAccount("Alice", 1000)`

### Configuration

<!-- skip-test -->
```yaml
version: 1
targets:
  - package: .
    types:
      - name: BankAccount
        constructor: NewBankAccount    # Constructor function name
        defaults:                       # Default values for constructor params
          name:
            pattern: email              # Built-in pattern (email, url, uuid)
          balance:
            custom: specta.Int()        # Custom generator expression
```

**Built-in patterns**: `email`, `url`, `uuid`

**Custom generators**: Any `specta.Gen[T]` expression (e.g., `specta.Int()`, `specta.String()`, etc.)

### What Gets Generated

For constructor-based types, the generator:
1. Analyzes the constructor signature
2. Generates factory code using the constructor
3. **Does NOT generate field-based matchers** (fields may be private!)
4. Requires getter-based matchers (next section)

### Example: BankAccount

Given:
<!-- skip-test -->
```go
type BankAccount struct {
	name    string  // private fields
	balance int
}

func NewBankAccount(name string, balance int) BankAccount {
	// validation logic
	return BankAccount{name: name, balance: balance}
}

func (a BankAccount) GetName() string { return a.name }
func (a BankAccount) GetBalance() int { return a.balance }
```

Configuration:
<!-- skip-test -->
```yaml
types:
  - name: BankAccount
    constructor: NewBankAccount
    defaults:
      name:
        pattern: email
      balance:
        custom: specta.Map(specta.Int(), func(v int64) int { return int(v) })
```

**Next**: Define matchers via getter methods (next section).

## Getter-Based Matchers

For constructor-based types with private fields, define matchers based on getter methods.

### Configuration

<!-- skip-test -->
```yaml
types:
  - name: BankAccount
    constructor: NewBankAccount
    defaults:
      name:
        pattern: email
      balance:
        custom: specta.Int()
    matchers:                           # Define matchers via getters
      - name: Name                       # Matcher method name
        getter: GetName                  # Getter method name
      - name: Balance
        getter: GetBalance
```

### Generated Matcher

The generator creates type-safe matcher methods:

<!-- skip-test -->
```go
// Generated code
type BankAccountMatcher struct {
	nameMatcher    specta.Matcher[string]
	balanceMatcher specta.Matcher[int]
}

func BankAccountMatches() BankAccountMatcher {
	return BankAccountMatcher{}
}

// Matcher method calls the getter
func (m BankAccountMatcher) Name(matcher specta.Matcher[string]) BankAccountMatcher {
	m.nameMatcher = matcher
	return m
}

func (m BankAccountMatcher) Balance(matcher specta.Matcher[int]) BankAccountMatcher {
	m.balanceMatcher = matcher
	return m
}

func (m BankAccountMatcher) Matcher() specta.Matcher[BankAccount] {
	return specta.MatcherFunc[BankAccount](func(actual BankAccount) specta.MatchResult {
		// Calls GetName(), GetBalance() on actual value
		nameValue := actual.GetName()
		balanceValue := actual.GetBalance()
		// Evaluates matchers...
	})
}
```

### Usage

<!-- skip-test -->
```go
func TestBankAccount(t *testing.T) {
    account := NewBankAccount("alice@example.com", 1000)

    specta.AssertThat(t, account, factory.BankAccountMatches().
        Name(specta.Contains("@")).
        Balance(specta.GreaterThan(0)).
        Matcher())
}
```

**Key difference**: Matcher calls getter methods, not direct field access.

## Understanding Generated Files

### Build Tags

Generated files include:
<!-- skip-test -->
```go
//go:build !ignore_testgen
```

This allows excluding them from linting/analysis tools while keeping them in your tests.

### What Gets Generated

The generator creates **three files** for each type:

1. **`factory/{type}_matcher_gen.go`** - Matcher builders (**← This section focuses here**)
   - Contains `{Type}Matcher` struct with field matcher storage
   - Constructor: `{Type}Matches()` creates a new matcher builder
   - Fluent methods: one per field, named after the field (no "With" prefix)
   - Implements the `Matcher[T]` interface directly via `Matches(actual T) MatchResult` method
   - No wrapper needed - the struct IS the matcher, use it directly in `AssertThat`
   - For getter-based matchers: methods call getter functions instead of accessing fields directly

2. **`factory/spec/{type}_gen.go`** - Low-level factory builders
   - Contains `{Type}Spec` struct with `Maybe[T]` fields
   - Option functions like `With{Type}{Field}()`
   - Used internally by the Recipe API
   - Covered in [Test Data Factories](../factories/)

3. **`factory/{type}_gen.go`** - Recipe fluent API
   - High-level builder for creating test data
   - Fluent methods for setting field values
   - Generates deterministic test data
   - Covered in [Test Data Factories](../factories/) and [Factories + Matchers](../factories-and-matchers/)

All three files work together to provide factories (test data creation) and matchers (assertions), but this section focuses only on the matcher file.

## Using Generated Matchers

### Basic Field Matching

```go
func TestBasicMatching(t *testing.T) {
    specta.AssertThat(t, user, UserMatches().
        Name(specta.Equal("Alice")).
        Age(specta.GreaterThan(18)))
}
```

### Compose with Core Matchers

```go
func TestComposition(t *testing.T) {
    specta.AssertThat(t, user, UserMatches().
        Email(specta.AllOf(
            specta.Contains("@"),
            specta.HasSuffix(".com"),
        )).
        Name(specta.Not(specta.Equal(""))))
}
```

### Nested Struct Matching

If `User` has a nested `Address` struct:

<!-- skip-test -->
```go
specta.AssertThat(t, user, UserMatches().
    Address(AddressMatches().
        City(specta.Equal("Seattle")).
        ZipCode(specta.MatchesRegex(`^\d{5}$`)).
        Matcher()).
    Matcher())
```

### Partial Matching in Action

<!-- skip-test -->
```go
// Test 1: Only care about name
specta.AssertThat(t, user, UserMatches().
    Name(specta.Equal("Alice")).
    Matcher())

// Test 2: Only care about email domain
specta.AssertThat(t, user, UserMatches().
    Email(specta.HasSuffix("@company.com")).
    Matcher())

// Test 3: Validate age and creation time
specta.AssertThat(t, user, UserMatches().
    Age(specta.GreaterThan(0)).
    CreatedAt(specta.Not(specta.IsNil[time.Time]())).
    Matcher())
```

Each test asserts exactly what it cares about. Adding new fields to `User` won't break these tests.

## Reusable Matchers

Build higher-level matchers from the generated ones:

<!-- skip-test -->
```go
// Define reusable matchers for common patterns
func ValidUser() specta.Matcher[User] {
    return UserMatches().
        Name(specta.Not(specta.Equal(""))).
        Email(specta.Contains("@")).
        Age(specta.GreaterThan(0)).
        Matcher()
}

func AdminUser() specta.Matcher[User] {
    return UserMatches().
        Name(specta.Not(specta.Equal(""))).
        Email(specta.AllOf(
            specta.Contains("@"),
            specta.HasSuffix("@company.com"),
        )).
        Age(specta.GreaterThan(0)).
        Matcher()
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
   func ValidUser() specta.Matcher[User] {
       return UserMatches().
           Name(specta.Not(specta.Equal(""))).
           Email(specta.Contains("@")).
           Age(specta.GreaterThan(0)).
           Status(specta.Equal("active")).  // One update here
           Matcher()
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
