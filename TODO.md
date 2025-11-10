# TODO: gomatchers Improvements

## High Priority

### 1. Better Error Messages (IN PROGRESS)
**Goal:** Improve matcher failure messages with actual vs expected values and structured diffs

**Completed:**
- [x] Add actual/expected fields to MatchResult
- [x] Update all basic matchers (Equal, GreaterThan, etc.) to include both values
- [x] Show field paths for nested failures (e.g., "User.Address.City")
- [x] Update all generated matchers to show actual values

**Remaining - Structured Diff Output:**
- [ ] Add structured diff with status symbols (✓/✗/~)
- [ ] Implement indented block format for nested structs
- [ ] Add terminal color detection (auto-detect TTY)
- [ ] Smart truncation for collections (show first 5, summarize rest)
- [ ] Handle both equality and constraint matchers appropriately

**Desired Output Format:**
```
User {
  ✓ FirstName: "Alice"                         // Matched
  ✗ Email: expected "alice@..." got "bob@..."  // Failed equality
  ✗ Age: expected > 18 but got 15              // Failed constraint
  ~ Active: true                                // Not checked (no matcher)
    Address: {
      ✓ Street: "123 Main"
      ✗ City: expected "NYC" but got "LA"
      ~ ZipCode: "10001"
    }
    Tags: [
      ✓ [0]: "verified"
      ✗ [1]: expected "premium" but got "basic"
      ~ [2]: "active"
      ... and 15 more items
    ]
}
```

**Status Symbols:**
- `✓` = matcher passed (green in terminal)
- `✗` = matcher failed (red in terminal)
- `~` = no matcher provided, shown for context (grey/dim in terminal)

**Implementation Details:**
- Hybrid approach: generated matchers use matcher-based diffs, DeepEqual uses reflection
- Detect equality vs constraint via `result.Expected != nil && result.Actual != nil`
- For equality: show both expected and actual values
- For constraints: show actual value + constraint message
- Auto-detect terminal for ANSI colors (colorGreen=32, colorRed=31, colorGrey=90)

**Files to modify:**
- `TODO.md` - this file
- `matchers_diff.go` - new file with reflection-based diff logic
- `cmd/main.go` - update matcher template for structured output
- `matchers.go` - update DeepEqual to use structDiff
- `matchers_test.go` - add tests for diff output

**Commit points:**
- "feat: add structured diff output with status symbols"

---

### 2. Improved Default Providers for Error-Returning Constructors
**Goal:** Generate valid defaults for constructors with validation

**Current Problem:**
```go
EmailDefaultAddress = func(p Primitives) string { return p.StringWith("address_") }
// Returns "address_" which fails NewEmail validation
```

**Solution Options:**

#### Option A: Config-based overrides
```yaml
types:
  - name: Email
    constructor: NewEmail
    defaults:
      Address: 'func(p Primitives) string { return p.StringWith("user") + "@example.com" }'
```

#### Option B: Smart pattern detection
- Detect validation patterns in constructor
- Generate appropriate defaults (email format, URLs, etc.)

#### Option C: Validation-aware generators
```yaml
types:
  - name: Email
    constructor: NewEmail
    validation:
      Address:
        pattern: email  # Built-in patterns: email, url, uuid, etc.
```

**Tasks:**
- [ ] Design config schema for default overrides
- [ ] Implement config parsing for custom defaults
- [ ] Add validation pattern detection (optional)
- [ ] Generate appropriate string-based defaults
- [ ] Update template to use configured defaults
- [ ] Add tests with various validation patterns

**Commit point:** "feat: support custom default providers for validated constructors"

---

### 3. Matcher Generation for Constructor-Based Types
**Goal:** Generate proper partial matchers instead of using DeepEqual

**Current Limitation:**
```go
func (r EmailRecipe) AsEqualMatcher() Matcher[Email] {
    // Constructor-based types: matchers are not yet fully supported
    expected := spec.BuildEmail(p, s)
    return testgen.DeepEqual(expected)  // ❌ No partial matching
}
```

**Desired Behavior:**
```go
func (r EmailRecipe) AsEqualMatcher() Matcher[Email] {
    s := spec.NewEmailSpec()
    for _, opt := range r.opts {
        opt(&s)
    }
    p := testgen.New()

    m := EmailMatches()
    if s.Address.IsSet() {
        m = m.Address(testgen.Equal(s.Address.Value(p)))
    }
    return m.Matcher()
}
```

**Tasks:**
- [ ] Check if type has both constructor AND matchers configured
- [ ] Generate AsEqualMatcher using matcher builder instead of DeepEqual
- [ ] For each spec field, if set, add to matcher
- [ ] Handle error-returning constructors (may need to panic on AsEqualMatcher)
- [ ] Add tests for partial matching with constructor types
- [ ] Test nested partial matching

**Commit point:** "feat: generate partial matchers for constructor-based types"

---

### 4. Collection Matchers
**Goal:** Add common collection matchers for slices, maps, and arrays

**Matchers to implement:**

#### Slice Matchers
```go
// Size matchers
HasSize[T any](size int) Matcher[[]T]
IsEmpty[T any]() Matcher[[]T]
IsNotEmpty[T any]() Matcher[[]T]

// Content matchers
Contains[T comparable](item T) Matcher[[]T]
ContainsAll[T comparable](items ...T) Matcher[[]T]
ContainsAny[T comparable](items ...T) Matcher[[]T]

// Predicate matchers
Every[T any](matcher Matcher[T]) Matcher[[]T]  // All elements match
Any[T any](matcher Matcher[T]) Matcher[[]T]    // At least one matches
None[T any](matcher Matcher[T]) Matcher[[]T]   // No elements match
```

#### Map Matchers
```go
HasKey[K comparable, V any](key K) Matcher[map[K]V]
HasValue[K comparable, V comparable](value V) Matcher[map[K]V]
HasEntry[K comparable, V comparable](key K, value V) Matcher[map[K]V]
MapHasSize[K comparable, V any](size int) Matcher[map[K]V]
```

**Tasks:**
- [ ] Implement HasSize, IsEmpty, IsNotEmpty
- [ ] Implement Contains, ContainsAll, ContainsAny
- [ ] Implement Every, Any, None with element matchers
- [ ] Implement map matchers
- [ ] Add comprehensive tests for each
- [ ] Update documentation with examples

**Commit point:** "feat: add collection matchers for slices and maps"

---

### 5. Nested Recipe Tracking for Constructor Types
**Goal:** Support FromRecipe methods for constructor parameters

**Current Limitation:**
```go
// Constructor-based types skip nested recipe tracking
type EmailRecipe struct {
    opts []testgen.Opt[spec.EmailSpec]
    // ❌ No nested recipe fields
}
```

**Desired Behavior:**
```go
type OrderRecipe struct {
    opts []testgen.Opt[spec.OrderSpec]
    userRecipe *UserRecipe  // ✅ Track nested recipe for partial matching
}

func (r OrderRecipe) UserFromRecipe(v UserRecipe) OrderRecipe {
    r.opts = append(r.opts, spec.WithOrderUserFromProvider(v.Provider()))
    r.userRecipe = &v
    return r
}
```

**Requirements:**
- Constructor param must be a custom type
- That custom type must have a Recipe
- Need to track recipe for AsEqualMatcher

**Tasks:**
- [ ] Detect custom types in constructor params
- [ ] Generate nested recipe fields for custom params
- [ ] Generate FromRecipe methods for custom params
- [ ] Update AsEqualMatcher to use nested recipes for partial matching
- [ ] Handle error-returning constructors with nested types
- [ ] Add tests for nested partial matching with constructors

**Considerations:**
- Does the custom type have a factory?
- Is it in the same package or imported?
- Error handling if nested type returns error

**Commit point:** "feat: support nested recipes for constructor parameters"

---

### 6. Documentation Generation
**Goal:** Generate helpful documentation in generated files

**What to generate:**

#### Example Usage in Generated Files
```go
// Example usage:
//
//	p := gomatchers.New()
//	user := factory.User().
//	    FirstName("Alice").
//	    Email("alice@example.com").
//	    Build(p)
//
//	// Matching
//	factory.UserMatches().
//	    FirstName(gomatchers.Equal("Alice")).
//	    Email(gomatchers.Contains("@example.com"))
```

#### Field Documentation
```go
// Name sets the name parameter.
// This maps to the 'name' parameter of NewBankAccount(name string, balance int).
func (r BankAccountRecipe) Name(v string) BankAccountRecipe
```

#### Constructor Documentation
```go
// BankAccount creates a new BankAccountRecipe.
// The underlying constructor is NewBankAccount(name string, balance int).
//
// Default providers:
//   - Name: generates "name_<N>"
//   - Balance: generates random int
func BankAccount() BankAccountRecipe
```

#### Matcher Documentation
```go
// BankAccountMatches creates a matcher for BankAccount instances.
// Matchers are generated from configured getters:
//   - Name: via GetName()
//   - Balance: via GetBalance()
//
// Example:
//   factory.BankAccountMatches().
//       Name(gomatchers.Equal("Alice")).
//       Balance(gomatchers.GreaterThan(1000))
func BankAccountMatches() BankAccountMatcher
```

**Tasks:**
- [ ] Add example usage to recipe constructors
- [ ] Document constructor signatures
- [ ] List default providers with their behavior
- [ ] Document getter mappings in matchers
- [ ] Add package-level documentation
- [ ] Generate README for factory package
- [ ] Add inline examples for complex patterns

**Commit point:** "feat: generate comprehensive documentation in factory code"

---

## Medium Priority

### 7. Additional Matcher Types
**Specific matchers beyond collections**

#### Nil/Pointer Matchers
```go
IsNil[T any]() Matcher[*T]
IsNotNil[T any]() Matcher[*T]
PointsTo[T any](matcher Matcher[T]) Matcher[*T]
```

#### Error Matchers
```go
IsError() Matcher[error]
ErrorContains(substr string) Matcher[error]
ErrorIs(target error) Matcher[error]
ErrorAs[T error](target *T) Matcher[error]
```

#### Time Matchers
```go
After(t time.Time) Matcher[time.Time]
Before(t time.Time) Matcher[time.Time]
Between(start, end time.Time) Matcher[time.Time]
WithinDuration(t time.Time, delta time.Duration) Matcher[time.Time]
```

#### Regex Matchers
```go
MatchesRegex(pattern string) Matcher[string]
MatchesPattern(regex *regexp.Regexp) Matcher[string]
```

**Tasks:**
- [ ] Implement nil/pointer matchers
- [ ] Implement error matchers
- [ ] Implement time matchers
- [ ] Implement regex matchers
- [ ] Add comprehensive tests
- [ ] Document with examples

**Commit point:** "feat: add specialized matchers (nil, error, time, regex)"

---

### 8. Convention-Based Auto-Detection
**Goal:** Generate matchers from getters without config

**Detection Rules:**
- `Get<Name>() T` → `Name(matcher Matcher[T])`
- `<Name>() T` (no params) → `Name(matcher Matcher[T])`
- `Is<Name>() bool` → `Name(matcher Matcher[bool])`
- `Has<Name>() bool` → `Name(matcher Matcher[bool])`

**Precedence:**
1. Explicit config overrides
2. Auto-detected methods
3. Fields (if no constructor)

**Edge Cases:**
- Methods with parameters (skip)
- Methods returning multiple values (skip or handle)
- Conflicting names (config wins)

**Tasks:**
- [ ] Implement method scanning
- [ ] Pattern matching for getter conventions
- [ ] Merge with configured matchers (config takes precedence)
- [ ] Add opt-out mechanism
- [ ] Test with various naming patterns
- [ ] Document conventions

**Commit point:** "feat: auto-detect getters for matcher generation"

---

## Low Priority

### 9. Performance Optimizations

#### AST Caching
- Cache parsed AST between generator runs
- Detect file changes with mtime/hash
- Reuse cached AST for unchanged files

#### Parallel Generation
- Generate specs/recipes/matchers concurrently
- Use worker pool for multiple types
- Proper error aggregation

#### Incremental Regeneration
- Track dependencies between types
- Only regenerate changed types
- Smart detection of config changes

**Tasks:**
- [ ] Implement AST caching layer
- [ ] Add parallel generation with goroutines
- [ ] Build dependency graph for types
- [ ] Implement incremental regeneration
- [ ] Benchmark improvements
- [ ] Add profiling support

**Commit point:** "perf: optimize code generation with caching and parallelization"

---

### 10. Enhanced Field Helper

**Additional helpers:**

```go
// Multiple extractors at once
Fields[T any](extractors map[string]func(T) any) Matcher[T]

// Transform before matching
Transform[T, V any](name string, transform func(T) V, matcher Matcher[V]) Matcher[T]

// Predicate-based
Where[T any](description string, predicate func(T) bool) Matcher[T]
```

**Usage:**
```go
Fields(map[string]func(BankAccount) any{
    "name": func(b BankAccount) any { return b.GetName() },
    "balance": func(b BankAccount) any { return b.GetBalance() },
})

Transform("name uppercase",
    func(b BankAccount) string { return strings.ToUpper(b.GetName()) },
    Equal("ALICE"),
)
```

**Tasks:**
- [ ] Implement Fields multi-extractor
- [ ] Implement Transform helper
- [ ] Implement Where predicate helper
- [ ] Add tests for each
- [ ] Document use cases

**Commit point:** "feat: add enhanced field extraction helpers"

---

## Future Considerations

### 11. Plugin System
- Custom matcher generators
- Custom default providers
- Custom template functions

### 12. IDE Integration
- Language server for config files
- Jump-to-definition from generated code
- Autocomplete for config

### 13. Validation
- Validate config against code
- Detect missing getters
- Warn about unreachable code paths

### 14. Migration Tools
- Migrate from other testing libraries
- Convert existing matchers to gomatchers
- Generate config from existing code

---

## Notes

- Each feature should have its own commit
- Add tests before marking complete
- Update showcase examples for each feature
- Keep backward compatibility
- Document breaking changes in commit messages
