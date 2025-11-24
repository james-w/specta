---
title: "API Reference"
weight: 8
---

# API Reference

Complete reference for specta's API.

## Core Functions

### AssertThat

<!-- skip-test -->
```go
func AssertThat[T any](t *testing.T, value T, matcher Matcher[T])
```

Asserts that `value` matches the given matcher. Fails the test with a structured error message if not.

**Example:**
<!-- skip-test -->
```go
AssertThat(t, user.Name, Equal("Alice"))
```

## Matcher Interface

<!-- skip-test -->
```go
type Matcher[T any] interface {
    Match(value T) MatchResult
}

type MatchResult struct {
    Matched bool
    Message string
}
```

All matchers implement this interface.

## Basic Matchers

### Equal

<!-- skip-test -->
```go
func Equal[T comparable](expected T) Matcher[T]
```

Matches if value equals expected using `==`.

### DeepEqual

<!-- skip-test -->
```go
func DeepEqual[T any](expected T) Matcher[T]
```

Matches if value deeply equals expected using `reflect.DeepEqual`.

### BeNil

<!-- skip-test -->
```go
func BeNil[T any]() Matcher[T]
```

Matches if value is nil.

### NotBeNil

<!-- skip-test -->
```go
func NotBeNil[T any]() Matcher[T]
```

Matches if value is not nil.

## Comparison Matchers

### GreaterThan

<!-- skip-test -->
```go
func GreaterThan[T constraints.Ordered](threshold T) Matcher[T]
```

Matches if value > threshold.

### LessThan

<!-- skip-test -->
```go
func LessThan[T constraints.Ordered](threshold T) Matcher[T]
```

Matches if value < threshold.

### GreaterThanOrEqual

<!-- skip-test -->
```go
func GreaterThanOrEqual[T constraints.Ordered](threshold T) Matcher[T]
```

Matches if value >= threshold.

### LessThanOrEqual

<!-- skip-test -->
```go
func LessThanOrEqual[T constraints.Ordered](threshold T) Matcher[T]
```

Matches if value <= threshold.

## String Matchers

### ContainString

<!-- skip-test -->
```go
func ContainString(substring string) Matcher[string]
```

Matches if string contains the substring.

### HavePrefix

<!-- skip-test -->
```go
func HavePrefix(prefix string) Matcher[string]
```

Matches if string starts with prefix.

### HaveSuffix

<!-- skip-test -->
```go
func HaveSuffix(suffix string) Matcher[string]
```

Matches if string ends with suffix.

### MatchRegex

<!-- skip-test -->
```go
func MatchRegex(pattern string) Matcher[string]
```

Matches if string matches the regular expression pattern.

### BeEmpty

<!-- skip-test -->
```go
func IsEmpty() Matcher[string]
```

Matches if string is empty.

## Collection Matchers

### HaveLength

<!-- skip-test -->
```go
func HaveLength[T any](expected int) Matcher[[]T]
```

Matches if slice/array has the expected length.

### Contain

<!-- skip-test -->
```go
func Contain[T comparable](item T) Matcher[[]T]
```

Matches if slice/array contains the item.

### ContainAll

<!-- skip-test -->
```go
func ContainAll[T comparable](items ...T) Matcher[[]T]
```

Matches if slice/array contains all specified items.

### ContainAny

<!-- skip-test -->
```go
func ContainAny[T comparable](items ...T) Matcher[[]T]
```

Matches if slice/array contains at least one of the specified items.

### EachMatch

<!-- skip-test -->
```go
func EachMatch[T any](predicate func(T) bool) Matcher[[]T]
```

Matches if all elements in the slice/array satisfy the predicate.

**Example:**
<!-- skip-test -->
```go
AssertThat(t, numbers, EachMatch(func(n int) bool {
    return n > 0
}))
```

## Composition Matchers

### AllOf

<!-- skip-test -->
```go
func AllOf[T any](matchers ...Matcher[T]) Matcher[T]
```

Matches if all matchers pass (AND logic).

**Example:**
<!-- skip-test -->
```go
AssertThat(t, email, AllOf(
    ContainString("@"),
    HaveSuffix(".com"),
))
```

### AnyOf

<!-- skip-test -->
```go
func AnyOf[T any](matchers ...Matcher[T]) Matcher[T]
```

Matches if at least one matcher passes (OR logic).

**Example:**
<!-- skip-test -->
```go
AssertThat(t, status, AnyOf(
    Equal("active"),
    Equal("pending"),
))
```

### Not

<!-- skip-test -->
```go
func Not[T any](matcher Matcher[T]) Matcher[T]
```

Inverts a matcher (negation).

**Example:**
<!-- skip-test -->
```go
AssertThat(t, name, Not(IsEmpty()))
```

## Primitives Interface

<!-- skip-test -->
```go
type Primitives interface {
    Next() int
    String(prefix string) string
    Time() time.Time
    UUID() string
    ID(prefix string) string
}
```

Provides deterministic test data generation.

### Gen Implementation

<!-- skip-test -->
```go
type Gen struct {
    Counter  int           // Current counter value
    BaseTime time.Time     // Base time for Time()
    Step     time.Duration // Increment for Time()
    Prefix   string        // Default prefix
}

func New() *Gen
```

**Example:**
<!-- skip-test -->
```go
p := specta.New()
id := p.ID("user")        // "user-0"
name := p.StringWith("name")  // "name-0"
t := p.Time()             // Base time
```

### Methods

#### Next

<!-- skip-test -->
```go
func (g *Gen) Next() int
```

Returns the current counter and increments it.

#### String

<!-- skip-test -->
```go
func (g *Gen) String(prefix string) string
```

Returns `prefix-N` where N is the current counter, then increments.

#### Time

<!-- skip-test -->
```go
func (g *Gen) Time() time.Time
```

Returns `BaseTime + (Counter * Step)`, then increments counter.

#### UUID

<!-- skip-test -->
```go
func (g *Gen) UUID() string
```

Returns a deterministic UUID based on the current counter, then increments.

#### ID

<!-- skip-test -->
```go
func (g *Gen) ID(prefix string) string
```

Alias for `String(prefix)`.

## Code Generation

### Configuration File (specta.yaml)

<!-- skip-test -->
```yaml
package: <package-name>
output_dir: <output-directory>
types:
  - name: <type-name>
    fields:
      - name: <field-name>
        type: <field-type>
```

**Example:**
<!-- skip-test -->
```yaml
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

### Running the Generator

<!-- skip-test -->
```bash
go run github.com/james-w/specta/cmd/main.go -config specta.yaml
```

### Generated Files

For each type `TypeName`, generates:

- `<output_dir>/typename_gen.go` - Low-level spec API
- `<output_dir>/typename_matcher_gen.go` - Matcher builders
- `<output_dir>/factory/typename_gen.go` - Factory builders (if factories enabled)

All generated files include:
<!-- skip-test -->
```go
//go:build !ignore_testgen
```

## Generated Matcher API

For a type `User` with fields `Name` and `Email`:

<!-- skip-test -->
```go
func MatchUser() *UserMatcher

func (m *UserMatcher) WithName(matcher Matcher[string]) *UserMatcher
func (m *UserMatcher) WithEmail(matcher Matcher[string]) *UserMatcher

func (m *UserMatcher) Match(user User) MatchResult
```

**Example:**
<!-- skip-test -->
```go
AssertThat(t, user, MatchUser().
    WithName(Equal("Alice")).
    WithEmail(ContainString("example.com")))
```

## Generated Factory API

For a type `User`:

<!-- skip-test -->
```go
func NewUser(p Primitives) *UserRecipe

func (r *UserRecipe) WithName(name string) *UserRecipe
func (r *UserRecipe) WithEmail(email string) *UserRecipe

func (r *UserRecipe) Build() User
```

**Example:**
<!-- skip-test -->
```go
p := specta.New()
user := NewUser(p).
    WithName("Alice").
    WithEmail("alice@example.com").
    Build()
```

## Error Messages

Matcher failures produce structured error messages:

<!-- skip-test -->
```
Value does not match:
  ✗ AllOf:
    ✓ ContainString("@")
    ✗ HaveSuffix(".com")
      Expected: string ending with ".com"
      Got:      "user@example.org"
    ✓ HavePrefix("user")
```

**Symbols:**
- `✓` - Matcher succeeded
- `✗` - Matcher failed
- `~` - Matcher not evaluated

## Build Tags

Exclude generated files from linting:

<!-- skip-test -->
```go
//go:build !ignore_testgen
```

To exclude from analysis:
<!-- skip-test -->
```bash
go vet -tags=ignore_testgen ./...
```

## Module Information

**Import Path:** `github.com/james-w/specta`

**Go Version:** 1.22+

**Installation:**
<!-- skip-test -->
```bash
go get github.com/james-w/specta
```
