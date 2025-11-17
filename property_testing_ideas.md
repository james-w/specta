# Property-Based Testing Ideas for Go Test Generation Framework

## Progress Update (Current as of commit 5c150eb)

### ✅ Completed Work

**Stage 1: Core Foundation** - **COMPLETE**
- ✅ Implemented `Source` interface with dual-mode support:
  - `Gen` type for deterministic generation (friendly test values)
  - `randomSource` type for property testing (random byte stream)
  - Methods: `DrawBits(n int) uint64`, `WriteLog(msg string)`, `IsDeterministic() bool`
- ✅ Implemented `Generator[T]` interface with signature: `Draw(s Source, label string) T`
- ✅ Built shrinking algorithm (Hypothesis-style: zero blocks, reduce bytes, binary search on length)
- ✅ Created 8 basic generators:
  - `Int()` with `Range()` and `Filter()` constraints
  - `String()` with `Length()`, `Charset()`, `Prefix()`, `ExampleHint()`, `Filter()` constraints
  - `Bool()` (returns false in deterministic mode, random in property mode)
  - `Float64()` (basic implementation)
  - `Time()` (basic implementation)
  - `Duration()` (basic implementation)
  - `Bytes()` (basic implementation)
  - `UUID()` (basic implementation)
- ✅ Built `Property(t, check, opts...)` test runner with options:
  - `MaxTests(n)` - number of iterations
  - `Seed(seed)` - deterministic reproduction
  - `MaxShrinks(n)` - shrinking budget
- ✅ Deterministic reproduction working: failures show seed for reproduction
- ✅ Shrinking working: failures shrink to minimal failing case

**Stage 2: Primitives Migration** - **COMPLETE (different approach)**
- ✅ **Removed Primitives interface entirely** instead of bridging
  - Deleted `property_primitives.go` (241 lines)
  - Deleted `property_primitives_test.go` (379 lines)
  - Removed `T.Primitives()` method
- ✅ Migrated all code to use `Source + Generators` directly
- ✅ Updated code generator templates in `cmd/main.go`:
  - Changed `Provider[V]` from `func(p Primitives) V` to `func(s Source) V`
  - Changed `BuildWithSpec[T, S]` signature to use Source
  - Updated all generated factory code to use generators
  - Fixed variable shadowing issues in templates
- ✅ Regenerated and verified all example and showcase projects
- ✅ All tests passing (main package, example, showcase)
- Net result: -441 lines of code, cleaner architecture

**Stage 3: Essential Generators** - **PARTIALLY COMPLETE**
- ✅ Core primitive generators implemented (8 types)
- ✅ Dual-mode operation (deterministic friendly values vs random exploration)
- ✅ Basic constraint support (Range, Length, Charset, Prefix, Filter)
- ❌ Slice generator - **NOT DONE**
- ❌ Map generator - **NOT DONE**
- ❌ Advanced constraints (pattern matching, weighted booleans, etc.) - **NOT DONE**

**Stage 5: Polish** - **PARTIALLY COMPLETE**
- ✅ Error messages show:
  - Seed for reproduction
  - Generated values with labels
  - Shrunk failing case
  - Skip rate warnings
  - Reproduction command
- ✅ Integration with testing.T and custom TestingT interface
- ✅ Property test T type with `Errorf`, `Fatalf`, `Helper`, `Assume(condition)`
- ✅ Skip tracking and warnings for high skip rates
- ❌ Performance optimization - not yet profiled
- ❌ Comprehensive documentation - needs expansion

### ❌ Not Started

**Stage 4: Biasing System** - **NOT STARTED**
- No BiasLevel implementation
- No edge case biasing (0, ±1, boundaries, powers of 2)
- No Unicode biasing (ASCII/Latin-1/Emoji/control chars)
- No collection size biasing
- Currently using uniform random generation in property mode

**Stage 6: Advanced Features** - **NOT STARTED**
- No corpus persistence
- No regex-based string generation (direct construction)
- No dependent value generation
- No stateful/model-based testing
- No custom generator composition helpers beyond Filter()

### Key Architectural Decisions Made

1. **No Primitives**: Removed the abstraction entirely rather than bridging it
2. **Dual-mode generators**: Same Generator[T] works in both deterministic (example) and random (property) contexts
3. **Draw() takes Source**: Unified interface allows generators to work anywhere
4. **Filter() for constraints**: Simple retry-based filtering with configurable max attempts
5. **Integrated shrinking**: Shrink byte stream, not generated values (Hypothesis approach)
6. **Simple error messages**: Text-based output showing seed, values, and reproduction command

### What Changed From Original Plan

**Original Plan**: Bridge Primitives to Source so existing factories work unchanged
**Actual Implementation**: Removed Primitives entirely, migrated everything to Source + Generators

**Rationale**: Cleaner architecture with one way to do things rather than maintaining two parallel systems. Code generator can easily produce Source-based code, so migration was straightforward.

## Overview

This document explores ideas for extending the existing Go test generation framework to support property-based testing with automatic shrinking. The goal is to create something that feels natural in Go while providing the power of property-based testing, inspired by Hypothesis's approach but adapted to Go's idioms.

## Core Concepts

### The Problem We're Solving

Currently, the framework generates deterministic test values through the `Primitives` interface, which is great for debugging but doesn't help find edge cases. Property-based testing would:
- Automatically find edge cases that break our assumptions
- Shrink failing cases to minimal reproductions
- Bias toward values that commonly cause bugs (boundaries, special characters, etc.)

### Key Insight: Unified Interface

We could potentially keep the existing `Primitives` interface and make it work for both deterministic examples AND property-based testing. This would allow gradual migration and reuse of existing factory code.

## Architecture Ideas

### Following Hypothesis's Integrated Shrinking Approach

Instead of the traditional QuickCheck approach (separate generation and shrinking), we could follow Hypothesis's "Conjecture" strategy:

```go
// Generator produces values with integrated shrinking
type Generator[T any] interface {
    Generate(source *Source) T
    Example() T  // For debugging - produces simple, deterministic value
}

// Source provides random data and tracks generation choices for shrinking
type Source struct {
    data    []byte      // Random byte stream (like Hypothesis)
    index   int         // Current position
    choices []Choice    // Decisions made for shrinking
    bias    BiasLevel   // How much to bias toward edge cases
}
```

The key idea: track the byte stream used for generation, then shrink by trying simpler byte streams.

### Biasing for Better Bug Finding

One powerful feature we could add is automatic biasing toward problematic values:

```go
type BiasLevel int
const (
    NoBias    BiasLevel = iota
    LowBias             // 10% special values
    MediumBias          // 30% special values (recommended default)
    HighBias            // 50% special values
    ChaosBias           // 80% special values (stress testing)
)
```

## Core Generators

### Integer Generation with Biasing

```go
// Integer generator that biases toward edge cases
type IntGen struct {
    min      *int64
    max      *int64
    bias     BiasLevel
    specials []int64  // Values to bias toward (0, 1, -1, boundaries, powers of 2)
}

// Usage would be clean and chainable:
Int().Range(0, 100)      // Basic range
Int().Positive()          // Common constraint
Int().PowerOf(2)          // Special constraint
```

When generating, we'd bias toward:
- Boundaries (min, max, min+1, max-1)
- Zero, one, negative one
- Powers of two in range
- Previously failing values (learning from history)

### String Generation with Smart Biasing

Strings are where bugs often hide, especially with Unicode. Ideas for biasing:

```go
// String generation with character class biasing
type StringGen struct {
    minLen   *int
    maxLen   *int
    charset  CharSet
    bias     BiasLevel
}

// When biased, Unicode strings would generate:
// - 70% ASCII (most common)
// - 15% Latin-1 (extended characters)
// - 5% Emoji (modern systems)
// - 5% Control characters (\n, \r, \0)
// - 3% Combining marks (é as e + ´)
// - 1% RTL text (Arabic, Hebrew)
// - 1% Zalgo text (ḩ̸̺̃ë̴́ͅl̷̰̾l̶̰̾ö̶́̂)
```

For length, we'd bias toward:
- Empty strings (edge case)
- Single character (edge case)
- Very long strings (stress test)

### Float Generation with Special Values

```go
type FloatGen struct {
    min       *float64
    max       *float64
    allowNaN  bool
    allowInf  bool
    precision *int  // Decimal places
}

// Biasing would include:
// - NaN, +Inf, -Inf (if allowed)
// - Positive and negative zero
// - Boundary values
// - Common values (0.0, 1.0, -1.0)
```

### Collection Generation with Size Biasing

Real-world collections are usually small:

```go
// Size biasing for collections
sizeBias := Distribution{
    0:      0.15,  // Empty collections (edge case)
    1:      0.20,  // Single element (edge case)  
    2-5:    0.35,  // Small (most common)
    6-20:   0.20,  // Medium
    21-100: 0.08,  // Large
    101+:   0.02,  // Stress test
}
```

## Constraint Handling

### Direct Construction vs. Filtering

For complex constraints like regex patterns, we have several options:

1. **Random + Filter** (naive): Generate randomly and check - terribly slow for complex patterns
2. **SAT/SMT Solving**: Overkill for testing, complex to implement
3. **Direct Construction** (recommended): Parse the constraint and generate valid values directly

Example for regex patterns:
```go
// Instead of generating random strings and checking against regex,
// parse the regex and generate matching strings directly:

// Pattern: ^[A-Z]{3}-\d{4}$
// Becomes: 
// 1. Generate 3 uppercase letters
// 2. Add literal "-"
// 3. Generate 4 digits

type RegexGenerator struct {
    segments []segment  // Each segment knows how to generate its part
}
```

### Dependent Value Generation

For values that depend on each other (like transfer amount ≤ sender balance):

```go
// Context tracks generated values for dependencies
type GenContext struct {
    values map[string]any
    deps   []dependency  // For shrinking
}

// Generate with explicit dependencies
amount := ctx.GenerateDependent("amount", []string{"sender"}, 
    func(deps map[string]any) Generator {
        senderBalance := deps["sender"].(BankAccount).Balance()
        return Int().Range(0, senderBalance)
    })
```

## Integration with Existing Framework

### Bridging Primitives and Property Testing

The key insight: we can implement `Primitives` using our property testing `Source`:

```go
// Implement Primitives using property testing source
type sourceBasedPrimitives struct {
    source   *Source
    prefix   string
    baseTime time.Time
    // ... other fields
}

func NewPrimitivesFromSource(s *Source, opts ...Option) Primitives {
    // Returns a Primitives implementation that uses the Source
    // for random generation with biasing and shrinking support
}
```

This means existing factories could work in property tests without modification!

### Natural API for Property Tests

Several possible APIs to make property testing feel natural:

```go
// Option 1: WithPrimitives - reuse existing factories
func TestBankAccountProperty(t *testing.T) {
    Property(t).WithPrimitives(func(p Primitives) bool {
        account := BankAccountFactory{primitives: p}.Build()
        return account.Balance() >= 0
    })
}

// Option 2: WithContext - mix generators and factories
func TestMixedProperty(t *testing.T) {
    Property(t).WithContext(func(s *Source, p Primitives) bool {
        // Use factory for complex objects
        account := BankAccountFactory{primitives: p}.Build()
        
        // Use generators for constrained values
        amount := Int().Range(0, account.Balance()).Generate(s)
        
        return account.Withdraw(amount)
    })
}

// Option 3: WithFactory - automatic primitive injection
func TestWithFactory(t *testing.T) {
    Property(t).WithFactory(func(fb FactoryBuilder) bool {
        sender := fb.BankAccount().WithMinBalance(100).Build()
        receiver := fb.BankAccount().Build()
        // ... test properties
    })
}
```

## Shrinking Strategy

Following Hypothesis's approach, shrinking would work on the byte stream:

```go
type Shrinker struct {
    original []byte
    choices  []Choice
}

// Main shrinking strategies:
// 1. Zero out blocks (Hypothesis's main strategy)
// 2. Reduce individual byte values
// 3. Remove blocks (for shrinking collections)
// 4. Replace complex Unicode with ASCII
// 5. Simplify to boundary values
```

The beauty of this approach: shrinking is automatic and preserves constraints because we're shrinking the input byte stream, not the generated values directly.

## Common Types and Constraints to Support

### Primitives
- **Integers**: Range, positive/negative, even/odd, multiples, powers, prime
- **Floats**: Range, precision, NaN/Inf handling
- **Strings**: Length, charset, pattern, prefix/suffix, encoding
- **Booleans**: Weighted probability
- **Time/Duration**: Ranges, business hours, timezone constraints

### Collections
- **Slices/Arrays**: Length, uniqueness, sorted, sum constraints
- **Maps**: Size, key/value relationships

### Domain-Specific (future extensions)
- **IDs/UUIDs**: Format validation, sequential generation
- **Emails**: Domain constraints, RFC compliance levels
- **URLs**: Scheme, domain, path patterns
- **Phone numbers**: Country codes, formats
- **Money**: Currency, precision, positive amounts
- **IP addresses**: v4/v6, private/public, CIDR ranges
- **JSON**: Schema validation, depth limits

## Benefits of This Approach

1. **Unified Interface**: Same `Primitives` interface works for both examples and property tests
2. **Gradual Migration**: Existing factory code can be used in property tests
3. **Go-Idiomatic**: Feels natural to Go developers, not like a Haskell port
4. **Automatic Edge Cases**: Biasing finds bugs faster than uniform random
5. **Minimal Reproductions**: Integrated shrinking produces small failing cases
6. **Learning System**: Could track previously failing values across runs

## Open Questions

1. Should biasing be configurable per-test or globally?
2. How much magic is too much? (e.g., automatic primitive injection vs explicit)
3. Should we support stateful/model-based testing from the start?
4. How to handle flaky tests due to randomness?
5. Should we provide a way to reproduce failures deterministically?
6. What's the right default bias level?

## Updated Next Steps (Post-Migration)

### Immediate Next Steps (High Priority)

1. **Complete Stage 3: Essential Generators**
   - Implement `Slice[T]()` generator with:
     - Length constraints (MinLen, MaxLen)
     - Element generator composition
     - Unique element support
     - Sorted/Unsorted variants
   - Implement `Map[K,V]()` generator with:
     - Size constraints
     - Key/Value generator composition
   - Add more constraints to existing generators:
     - Float64: Precision, allow/disallow NaN/Inf, boundary constraints
     - Time: Range constraints, business hours, timezone handling
     - Duration: Range constraints, common durations (second, minute, hour)
     - String: Pattern matching (email, URL, phone)

2. **Performance Profiling and Optimization**
   - Profile property test execution with 100+ iterations
   - Target: <1ms per iteration for simple generators
   - Optimize hot paths in Source.DrawBits() and shrinking
   - Consider adding benchmark tests

3. **Documentation Expansion**
   - Add comprehensive README section on property testing
   - Document each generator with examples
   - Show migration guide from traditional tests to property tests
   - Add common patterns and anti-patterns
   - Document dual-mode behavior clearly

### Medium Priority

4. **Stage 4: Biasing System**
   - Design and implement `BiasLevel` configuration
   - Add integer biasing (0, ±1, boundaries, powers of 2)
   - Add string biasing (empty, single char, special chars)
   - Add Unicode biasing for strings (ASCII/Latin-1/Emoji/control/RTL)
   - Add collection size biasing (empty, single element, small/medium/large)
   - Benchmark: does biasing find seeded bugs faster?

5. **Generator Composition Helpers**
   - `OneOf(generators...)` - choose randomly from multiple generators
   - `Weighted(weight int)` - for Bool() and choosing between alternatives
   - `Maybe[T](gen Generator[T])` - generates Some(value) or None
   - Chain combinators for common patterns

### Lower Priority

6. **Stage 6: Advanced Features** (defer until proven need)
   - Corpus persistence (save interesting examples across runs)
   - Regex-based string generation (direct construction from pattern)
   - Dependent value generation (amount ≤ balance)
   - Stateful/model-based testing
   - Custom generator composition DSL

7. **Quality of Life Improvements**
   - Better visualization of shrinking progress
   - Statistics on generated values (distribution, coverage)
   - Parallel property test execution
   - Custom Assume() messages
   - Configurable shrinking strategies

### Deferred/Optional

- Integration with fuzzing (Go 1.18+ native fuzzing)
- Hypothesis-style example database
- Property test failure replay UI
- Cross-property correlation detection
- Mutation-based testing

## Conclusion

The goal is to make property-based testing feel natural in Go while providing powerful features like automatic shrinking and intelligent biasing. By integrating generators directly with the Source interface and factory pattern, we've created a unified system that works seamlessly in both deterministic test scenarios (with friendly example values) and property-based testing scenarios (with full type-space exploration).

The key is to embrace Go's imperative nature rather than forcing functional programming patterns, making the API feel like a natural extension of how Go developers already write tests.

**Current Status:** The core foundation is complete and working. Property tests can be written today using the available generators (Int, String, Bool, Float64, Time, Duration, Bytes, UUID), and failures shrink to minimal reproductions with seed-based deterministic replay. The next major work items are completing collection generators (Slice, Map), adding biasing for better bug-finding, and expanding documentation with comprehensive examples.

---

## Implementation Plan

### Stage 1: Core Foundation (2-3 weeks)
**Goal:** Get basic property testing working with integrated shrinking

**Tasks:**
- Implement `Source` type (byte stream with tracking)
- Implement basic `Generator[T]` interface
- Build simple shrinking algorithm (zero blocks, reduce bytes)
- Create `Int()` generator with range constraints
- Build `Property(t)` test runner with seed handling
- Add deterministic reproduction via seed
- Prove the concept works end-to-end

**Success criteria:** Can write property test with Int generator, failures shrink to minimal case, can reproduce with seed

**Key risks:** Shrinking algorithm complexity - budget extra time, consider copying Hypothesis algorithm exactly if simple approach doesn't work well

### Stage 2: Bridge to Primitives (1-2 weeks)
**Goal:** Make existing factories work in property tests

**Tasks:**
- Implement `Primitives` interface backed by `Source`
- Handle counter-based generation (Next(), IDs, UUIDs)
- Implement string generation with prefixes
- Test existing factories in property tests
- Document behavioral differences (deterministic vs random)

**Success criteria:** Existing `BankAccount` factory works in property test without modification

**Decision point:** Are factories working smoothly in property tests? Are behavioral differences acceptable?

### Stage 3: Essential Generators (2-3 weeks)
**Goal:** Cover common test data types

**Tasks:**
- String generator (length, charset constraints)
- Float generator (range, precision, NaN/Inf)
- Boolean generator (weighted)
- Slice generator (with element generator)
- Map generator
- Time/Duration generators

**Success criteria:** Can write realistic property tests for business logic without manual data generation

**Note:** Profile performance early - target <1ms per simple iteration

### Stage 4: Biasing System (2-3 weeks)
**Goal:** Find bugs faster with intelligent value selection

**Tasks:**
- Implement `BiasLevel` configuration
- Add edge case biasing to Int (0, ±1, boundaries, powers of 2)
- Add edge case biasing to String (empty, single char, long, special chars)
- Add Unicode biasing (ASCII/Latin-1/Emoji/control chars)
- Add collection size biasing
- Benchmark biased vs uniform bug finding

**Success criteria:** Biased generation finds seeded bugs 2-3x faster than uniform in benchmarks

**Decision point:** Is biasing actually helping? If not, may need to adjust distributions or strategy

### Stage 5: Polish & Ergonomics (1-2 weeks)
**Goal:** Make it production-ready

**Tasks:**
- Excellent error messages (failing property, shrunk input, seed, shrinking path)
- Configuration options (iterations, max shrink steps, timeout)
- Integration with existing test output
- Performance optimization
- Comprehensive documentation and examples

**Success criteria:** Developer can debug property test failure without asking for help

### Stage 6: Advanced Features (Future)
**Defer until proven need:**

- Corpus persistence (save interesting examples across runs)
- Regex-based string generation (direct construction from pattern)
- Dependent value generation (amount ≤ balance)
- Stateful/model-based testing
- Custom generator composition helpers

### Critical Requirements

**Must-haves for v1:**
- Seed-based deterministic reproduction (non-negotiable for CI)
- Clear error messages showing shrunk input and reproduction seed
- Performance budget met (<1ms per iteration for simple generators)

**Nice-to-haves for later:**
- Corpus persistence
- Advanced constraint solving
- Stateful testing

### Open Design Questions (to resolve during implementation)

1. **Code location:** Same package as matchers or new `property` subpackage?
2. **Shrinking precision:** Start simple (greedy) or copy Hypothesis exactly?
3. **API preference:** WithPrimitives (Option 1) seems best - confirm during Stage 2
4. **Biasing defaults:** MediumBias (30%) as default - validate with benchmarks in Stage 4
5. **Error message format:** Design early in Stage 1 as it will drive API decisions