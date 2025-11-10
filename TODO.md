# TODO: gomatchers Improvements

## High Priority

### 1. Better Error Messages (COMPLETED)
**Goal:** Improve matcher failure messages with actual vs expected values and structured diffs

**Completed:**
- [x] Add actual/expected fields to MatchResult
- [x] Update all basic matchers (Equal, GreaterThan, etc.) to include both values
- [x] Show field paths for nested failures (e.g., "User.Address.City")
- [x] Update all generated matchers to show actual values
- [x] Add structured diff with status symbols (✓/✗/~)
- [x] Implement indented block format for nested structs
- [x] Add terminal color detection (auto-detect TTY)
- [x] Smart truncation for collections (show first 5, summarize rest)
- [x] Handle both equality and constraint matchers appropriately

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

**Implementation Notes:**
- Created `matchers_term.go` for terminal detection and ANSI color support
- Created `matchers_diff.go` with two diff strategies:
  - `BuildMatcherStructDiff()` for generated matchers (uses cached field values)
  - `buildReflectionStructDiff()` for DeepEqual (uses reflection)
- Updated generated matcher template in `cmd/main.go`:
  - Extracts all field/getter values upfront (calls each exactly once)
  - Stores values in map for structured diff
  - Tracks match results per field
- Updated `DeepEqual` in `matchers.go` to use structured diff for struct types
- Added comprehensive tests in `matchers_test.go`
- Smart collection truncation shows max 5 items with "... and N more"
- Each getter called exactly once (cached for both matching and diff display)

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

## Investigation: Matcher-to-Generator Conversion

### Background

The use case: Define validation rules as matchers, then generate test data that satisfies those rules.

**Current state**: We have Recipe → Matcher (AsEqualMatcher), but not Matcher → Recipe.

**Example of desired pattern**:
```go
// Define validation rules
var IsValidUser = UserMatches().
    Email(gomatchers.Contains("@")).
    Age(gomatchers.GreaterThan(18)).
    Matcher()

// Generate test data satisfying those rules
recipe := factory.User().SatisfyingMatcher(IsValidUser)
user := recipe.Build(p)  // Will have email with @ and age > 18
```

### Investigation Results

**Matcher Inversion Feasibility** (16 total matchers analyzed):

**EASY (6 matchers)** - Single obvious satisfying value:
- `Equal(v)` → generate v
- `DeepEqual(v)` → generate v
- `Is(v)` → generate v (alias for Equal)
- `IsZero()` → generate zero value
- `IsTrue()` → generate true
- `IsFalse()` → generate false

**MEDIUM (6 matchers)** - Multiple options, reasonable default exists:
- `GreaterThan(n)` → generate n+1
- `LessThan(n)` → generate n-1
- `GreaterThanOrEqual(n)` → generate n
- `Contains(s)` → generate s (or "prefix" + s)
- `HasPrefix(p)` → generate p (or p + "suffix")
- `HasSuffix(s)` → generate s (or "prefix" + s)

**HARD (1 matcher)** - Requires special handling:
- `Field(extractor, matcher)` → Can't invert arbitrary extractors. Only feasible for generated matchers where we control the extractor and know it maps to a specific field.

**IMPOSSIBLE (3 matchers)** - Without major caveats:
- `Not(m)` → Infinite non-matching values, fundamentally ambiguous
- `AllOf(...)` → Constraint satisfaction problem, may be contradictory
- `AnyOf(...)` → Ambiguous which branch to satisfy

**Key obstacle**: Matchers are closures - their captured values aren't accessible for extraction.

### Proposed Solution

**Phase 1: Make Matchers Introspectable**

Add optional introspection without breaking existing API:

```go
type IntrospectableMatcher[T any] interface {
    Matcher[T]
    Introspect() MatcherConstraints
}

type MatcherConstraints struct {
    Type        string  // "equal", "greater_than", "allof", etc.
    Value       any     // Captured value (for Equal, GreaterThan, etc.)
    SubMatchers []any   // For composite matchers
}
```

Refactor matchers to implement this interface while keeping backward compatibility.

**Phase 2: Add Generator Inversion API**

```go
// Option A: Recipe method
recipe := factory.User().
    SatisfyingMatcher(someUserMatcher)

// Option B: Standalone function
recipe := factory.FromMatcher(someUserMatcher, factory.User())
```

**Phase 3: Implement Constraint Inversion**

- EASY matchers: Extract value, return it
- MEDIUM matchers: Apply heuristic (n+1, n-1, etc.)
- HARD matchers: Special case for generated matchers, error otherwise
- IMPOSSIBLE: Clear error messages explaining limitation

**Phase 4: Handle Composite Matchers**

- `AllOf`: Merge constraints if compatible (e.g., GreaterThan(5) + LessThan(10) → pick 7)
- Detect contradictions (e.g., Equal(5) + Equal(6) → error)
- `AnyOf`: Pick first matcher by convention (document this)
- `Not`: Not supported (error with explanation)

### Alternative: Recipe-First Pattern (Current)

Instead of Matcher → Recipe, use Recipe → Matcher (already implemented):

```go
// Define pattern as recipe
var ValidUser = factory.User().
    Email("user@example.com").
    Active(true)

// Generate data
user := ValidUser.Build(p)

// Validate outputs
gomatchers.AssertThat(t, result, ValidUser.AsEqualMatcher())
```

This works well but requires thinking in "generators" rather than "validators". The recipe becomes the single source of truth.

### Decision Points

1. **Pursue Matcher → Recipe inversion?**
   - Pros: More intuitive for "define rules, generate conforming data" use case
   - Cons: Requires refactoring, won't work for all matchers
   - Coverage: 75% of matchers invertible (12/16)

2. **Or advocate for Recipe → Matcher pattern?**
   - Pros: Already implemented, works well, no refactoring needed
   - Cons: Requires thinking generators-first rather than validators-first
   - Works: 100% of cases

3. **Or do both?**
   - Support both directions for maximum flexibility
   - Recipe → Matcher for partial matching (current)
   - Matcher → Recipe for generating conforming data (new)

### Recommendation

**Pursue the full solution** (Matcher → Recipe) because:
- 75% matcher coverage is good enough for common cases
- The use case is compelling (define validation rules, generate conforming data)
- Clear error messages can handle unsupported matchers
- It complements the existing Recipe → Matcher pattern
- Creates true symmetry: matchers and generators mirror each other

**Implementation order**:
1. Add introspection to basic matchers (EASY + MEDIUM)
2. Implement SatisfyingMatcher for generated recipes
3. Add constraint inversion logic
4. Handle AllOf with constraint merging
5. Document limitations clearly
6. Add comprehensive tests and examples

**Files to create/modify**:
- `matchers_introspect.go` - New file for introspection interface
- `matchers_generate.go` - New file for constraint inversion
- `matchers.go` - Refactor to implement introspection
- `cmd/main.go` - Add SatisfyingMatcher to recipe template
- `README.md` - Document the pattern with examples

---

## Notes

- Each feature should have its own commit
- Add tests before marking complete
- Update showcase examples for each feature
- Keep backward compatibility
- Document breaking changes in commit messages
