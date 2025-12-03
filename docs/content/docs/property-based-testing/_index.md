---
title: "Property-Based Testing"
weight: 6
---

# Property-Based Testing

We've been using factories to specify only what matters for each test. But how do you know your code works with *different* values for the fields you didn't specify? Maybe it only works for strings up to a certain length, and the test string you are specifying happens to be OK?

**Property-based testing** verifies that properties hold across many varied inputs, without you having to think of examples for each edge case.

It's similar to fuzzing in that you are testing with random data, but it is a little more controlled.

## From Example-Based to Property-Based

Consider testing a function. With example-based testing, you'd write:

<!-- skip-test -->
```go
func TestReverse(t *testing.T) {
    result := Reverse("hello")
    specta.AssertThat(t, result, specta.Equal("olleh"))
}
```

This tests one specific case. But you want to verify a **property** that should hold for all strings:

> **Property**: Reversing a string twice returns the original

Instead of writing many individual test cases, use property-based testing:

<!-- skip-test -->
```go
func TestReverse(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Generate a random string
        s := specta.Draw(pt, specta.String(), "s")

        // Test the property
        reversed := Reverse(Reverse(s))
        specta.AssertThat(pt, reversed, specta.Equal(s))
    })
}
```

**What happens:**
1. `Property()` runs your test function 100 times (by default)
2. Each time, generators produce different values (`String()` generates varied strings)
3. The failure report shows the exact input that failed

This is property-based testing: verify properties hold across many generated inputs.

It does not replace example-based testing, but it supplements it. You still want to have example-based tests to verify specific behaviour, ensure that edge cases are covered, but some tests are more readable as property based tests as the intent is clearer, plus you test more examples and find edge cases automatically.

The failure that specta shows you for a property goes through a process called "shrinking". This attempts to find a simpler version of the inputs that caused the property to fail. As the input is random the propoerty may fail for a very complex input, and it's not clear what matters. Through shrinking specta aims to make it easier for you to debug the problem. The process isn't perfect, so it's not necessarily the simplest input possible, and in fact that's not even something that can be defined for all properties, but it aims to be as useful as possible. For example, if you have a bug in a string function when the input is over 10 characters long, specta will hopefully show you a failing input that is 11 characters long, and has simple printable characters, rather than obscure unicode.

## Generators

Generators produce values for property-based tests. specta provides built-in generators for common types:

<!-- skip-test -->
```go
specta.Property(t, func(pt *specta.T) {
    // Generate integers between 1 and 100
    n := specta.Draw(pt, specta.Int().Range(1, 100), "n")

    // Generate alpha-numeric ascii strings, between 1 and 50 characters long
    name := specta.Draw(pt, specta.String().AlphaNum().MinLen(1).MaxLen(50), "name")

    // Generate booleans
    flag := specta.Draw(pt, specta.Bool(), "flag")

    // Generate slices of non-negative integers with between 1 and 10 elements
    numbers := specta.Draw(pt, specta.Slice(specta.Int().NonNegative()).MinLen(1).MaxLen(10), "numbers")

    // Test your code with these generated values
})
```

### Available Generators

**Basic Types:**
- **`Int()`** - Integers with constraints: `.Range(min, max)`, `.Min(n)`, `.Max(n)`, `.Positive()`, `.NonNegative()`, `.Negative()`
- **`String()`** - Strings with constraints: `.AlphaNum()`, `.Alpha()`, `.Printable()`, `.ASCII()`, `.MinLen()`, `.MaxLen()`, `.Len(n)`, `.NonEmpty()`, `.Prefix()`, `.Suffix()`, `.ExampleHint()`
- **`Bool()`** - Boolean values
- **`Float64(min, max)`** - Floating point numbers in range

**Time/Data:**
- **`Time()`** - Time values (Unix epoch to 9999-12-31)
- **`Duration()`** - Time durations (0 to 24 hours)
- **`Bytes()`** - Byte slices (0-100 bytes)
- **`BytesLen(min, max)`** - Byte slices with specific length range
- **`UUID()`** - UUID v4 values (string format)
- **`Email()`** - Generated email addresses (user_N@example.com)
- **`URL()`** - Generated URLs (https://example.com/path_N)

**Collections:**
- **`Slice(elementGen)`** - Slices of any type: `.MinLen()`, `.MaxLen()`, `.Len(n)`, `.NonEmpty()`
- **`MapOf(keyGen, valueGen)`** - Maps with generated keys and values: `.MinLen()`, `.MaxLen()`, `.NonEmpty()`

**Combinators:**
- **`Choice(gen1, gen2, ...)`** - Uniform choice from generators: `.Or(gen)`, `.OrWeighted(gen, weight)`
- **`Optional(gen)`** - Generate present (80%) or absent (20%) values: `Optional(gen, probability)`
- **`Pair(genA, genB)`** - Generate tuples of two values
- **`Triple(genA, genB, genC)`** - Generate tuples of three values

### Generator Constraints

Chain methods to constrain generated values:

<!-- skip-test -->
```go
// Positive integers between 1 and 100
age := specta.Draw(pt, specta.Int().Range(1, 100), "age")

// Non-empty alphanumeric strings
username := specta.Draw(pt, specta.String().AlphaNum().NonEmpty().MaxLen(20), "username")

// Slices with 5-10 elements
items := specta.Draw(pt, specta.Slice(specta.Int().Positive()).MinLen(5).MaxLen(10), "items")

// Choice between multiple generators
status := specta.Draw(pt, specta.Choice(
    specta.String().Const("active"),
    specta.String().Const("inactive"),
    specta.String().Const("pending"),
), "status")

// Optional values (80% present by default)
middleName := specta.Draw(pt, specta.Optional(specta.String().Alpha()), "middleName")
// middleName is *string - nil when absent, &value when present

// Tuples of values
coords := specta.Draw(pt, specta.Pair(
    specta.Int().Range(0, 100),
    specta.Int().Range(0, 100),
), "coords")
// coords is struct{A int64; B int64}
```

### Understanding Draw()

All the examples above use `specta.Draw(pt, generator, "label")` to generate values. This is the **canonical pattern** for property-based testing in specta.

**Why use Draw()?**

<!-- skip-test -->
```go
// ALWAYS use this pattern:
s := specta.Draw(pt, specta.String(), "s")

// DON'T call .Draw() directly on generators:
s := specta.String().Draw(pt.Data, "s")  // ❌ Don't do this
```

The `Draw()` helper provides three critical benefits:

**1. Automatic Error Handling**

When generators filter values (via `.Filter()`) or when you use `pt.Assume()`, some test iterations need to be skipped. `Draw()` handles this automatically:

<!-- skip-test -->
```go
specta.Property(t, func(pt *specta.T) {
    // Generator might fail to produce a value after filtering
    x := specta.Draw(pt, specta.Int().Filter(isPrime), "x")
    // Draw() automatically skips the test iteration if filtering fails

    // No need to handle errors - Draw() does it for you
})
```

**2. Better Error Messages**

Labels in `Draw()` calls appear in failure messages, showing you exactly which generated values caused the failure:

```
Property failed after 23 tests with seed 1234567890

Generated values:
  s: "hello"
  n: 42
  numbers: [1, 2, 3]

Assertion failed:
  expected: 100
  got: 84
```

Without labels, you'd see cryptic failures without knowing what inputs caused them.

**3. Value Tracking for Shrinking**

When a test fails, specta automatically tries to find the **minimal failing input** (shrinking). `Draw()` tracks generated values so shrinking can work:

<!-- skip-test -->
```go
// Test fails with a large string
s := specta.Draw(pt, specta.String(), "s")  // Generated: "abcdefghijklmnopqrstuvwxyz"

// Shrinking finds minimal failure
// Shrunk to: "ab"

// Final error shows both:
// Property failed with shrunk input:
//   s: "ab"
```

**Best Practice:** Always use `specta.Draw(pt, generator, "label")` in property tests. Choose descriptive labels that make error messages clear.

## When to Use Property-Based Testing

**Use PBT for:**
- Testing invariants (properties that always hold)
- Round-trip properties (encode/decode, serialize/deserialize)
- Idempotence (calling twice = calling once)
- Commutativity, associativity
- Edge cases you might not think of

**Use example-based tests for:**
- Specific business logic
- Known edge cases
- Regression tests for specific bugs

**Use both!** They complement each other.

## Common Types of Properties

Here are common patterns for property-based tests:

### Round-Trip Operations

Operations that reverse each other (encode/decode, serialize/deserialize):

<!-- skip-test -->
```go
func TestJSONRoundTrip(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        original := specta.Draw(pt, specta.MapOf(
            specta.String().AlphaNum(),
            specta.Int(),
        ), "data")

        // Encode to JSON
        jsonData, err := json.Marshal(original)
        specta.AssertThat(pt, err, specta.NoErr())

        // Decode from JSON
        var decoded map[string]int
        err = json.Unmarshal(jsonData, &decoded)
        specta.AssertThat(pt, err, specta.NoErr())

        // Property: decoded == original
        specta.AssertThat(pt, decoded, specta.Equal(original))
    })
}
```

Similar patterns: encrypt/decrypt, compress/decompress, parse/print, rotate left/right.

Properties like this will catch any cases that don't round-trip correctly, but don't require
knowing what the intermediate state is. Given any struct you can check that you get the
same thing if you encode/decode it to JSON, without having to say anything about what
the JSON looks like.

### Operations Maintain Invariants

Verify that operations preserve important properties:

<!-- skip-test -->
```go
func TestSortPreservesLength(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        numbers := specta.Draw(pt, specta.Slice(specta.Int()), "numbers")
        sorted := Sort(numbers)

        // Invariants that must hold
        specta.AssertThat(pt, len(sorted), specta.Equal(len(numbers)))
        for i := 1; i < len(sorted); i++ {
            specta.AssertThat(pt, sorted[i], specta.GreaterThanOrEqual(sorted[i-1]))
        }
    })
}
```

Other examples:
- Balance stays same in money transfers: `before.total == after.total`
- Shopping cart total matches items: `cart.total == sum(item.price)`
- Database foreign keys stay valid after updates
- Reference counts stay consistent
- Function doesn't panic

### Comparing with Simpler Implementation

Test optimized code against simple reference version:

<!-- skip-test -->
```go
func TestFastSumMatchesSimpleSum(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        numbers := specta.Draw(pt, specta.Slice(specta.Int()).MaxLen(100), "numbers")

        // Compare optimized vs simple
        specta.AssertThat(pt, FastSum(numbers), specta.Equal(SimpleSum(numbers)))
    })
}
```

Other examples:
- Optimized algorithm vs straightforward version
- Parallel version vs sequential version
- Custom data structure vs standard library equivalent
- New regex engine vs known-good implementation

### Order Doesn't Matter (Commutativity)

Operations that should work the same regardless of input order:

- `Add(x, y) == Add(y, x)` - addition works either way
- `distance(a, b) == distance(b, a)` - distance is symmetric
- `set.union(a, b) == set.union(b, a)` - set operations
- Search results should be same regardless of query term order

### Applying Multiple Times = Applying Once (Idempotence)

Operations where repeating has no additional effect:

- Normalize string: `normalize(normalize(s)) == normalize(s)`
- Add to set: `set.add(x).add(x) == set.add(x)`
- HTTP PUT requests: multiple identical PUTs = one PUT
- Close file: `close(close(f))` safe to call multiple times
- Database migrations: running twice = running once

### Predictable Transformations

When you transform the input, outputs have predictable relationships even if you don't know exact values:

- More specific search returns fewer results: `search("ABC") ⊆ search("AB")`
- Processing larger input takes more time: `time(process(data + data)) > time(process(data))`
- Scaling image 2x then 0.5x returns approximately original size
- If you insert item X into collection, searching for X should find it

This is useful when you can't compute the exact output, but know how different inputs relate to each other

### Operations Compose Correctly (Associativity)

Multiple operations combine in expected ways:

- `append(a, append(b, c)) == append(append(a, b), c)` - grouping doesn't matter
- `filter(f, filter(g, xs)) == filter(f and g, xs)` - filters combine
- `map(f, map(g, xs)) == map(f∘g, xs)` - maps combine

**Start simple:** Begin with "doesn't panic" or "result length matches input length", then add more specific properties as you understand the behavior.

## Advanced Patterns

### Filtering with Assume

Skip test iterations that don't meet preconditions:

<!-- skip-test -->
```go
func TestDivision(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        x := specta.Draw(pt, specta.Int(), "x")
        y := specta.Draw(pt, specta.Int(), "y")

        // Skip if y is zero
        pt.Assume(y != 0)

        result := x / y
        // Test division properties
        specta.AssertThat(pt, result*y, specta.Equal(x-x%y))
    })
}
```

It's better to constrain a generator to prevent the value being generated in the first place as skips waste iterations, but it's not always possible or convient to do that. specta will warn if your property is skipping too many iterations as it's not giving the coverage that you think it is and is wasteful. In those cases try and rewrite the tests to constrain a generator instead.


### Constraining Generated Values

There are three ways to ensure generated values meet requirements:

#### 1. Generator Constraints (Best)

Use built-in constraints when possible - they're fast and always succeed:

<!-- skip-test -->
```go
// Good: constrain the generator directly
age := specta.Draw(pt, specta.Int().Range(0, 120), "age")
username := specta.Draw(pt, specta.String().AlphaNum().MinLen(3).MaxLen(20), "username")
```

#### 2. Filtering (When Constraints Don't Suffice)

Use `.Filter()` for conditions that can't be expressed as generator constraints:

<!-- skip-test -->
```go
// Generate an even number (50% of integers pass - good for filtering)
n := specta.Draw(pt, specta.Int().Range(0, 100).Filter(func(x int64) bool {
    return x % 2 == 0
}), "n")
```

**Caveat**: The generator retries up to 100 times, then skips the test iteration. If your filter has a low success rate, generation will be slower and you'll skip more tests. Keep the success rate reasonably high (aim for >10%).

#### 3. Assume (For Complex Preconditions)

Use `pt.Assume()` for complex preconditions, especially those involving relationships between values:

<!-- skip-test -->
```go
x := specta.Draw(pt, specta.Int().Range(0, 100), "x")
y := specta.Draw(pt, specta.Int().Range(0, 100), "y")

// Skip this test if x >= y (correlation between two generated values)
pt.Assume(x < y)
```

**Caveat**: Skipped tests don't count toward your test total. If too many tests are skipped (>90%), specta will warn you.

**Better approach**: When possible, restructure generation to avoid `Assume()`:

<!-- skip-test -->
```go
// Generate x, then generate y > x
x := specta.Draw(pt, specta.Int().Range(0, 99), "x")
y := specta.Draw(pt, specta.Int().Range(x+1, 100), "y")
// Now x < y is always true, no Assume() needed
```

### Configuring Property Tests

<!-- skip-test -->
```go
func TestExpensive(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Test code
    },
        specta.MaxTests(1000),      // Run 1000 iterations (default: 100)
        specta.Seed(42),             // Use specific seed for reproducibility
        specta.MaxShrinks(500),      // Limit shrinking attempts (default: 1000)
    )
}
```

### Understanding Randomness and Seeds

**By default, property tests use random seeds** (`time.Now().UnixNano()`). This means each test run explores different parts of the input space:

<!-- skip-test -->
```go
func TestStringReverse(t *testing.T) {
    // No explicit seed - uses time.Now().UnixNano()
    specta.Property(t, func(pt *specta.T) {
        s := specta.Draw(pt, specta.String(), "s")
        // Different random strings on each test run
    })
}
```

**This is intentional and beneficial:**
- Each CI run tests different inputs
- Over time, you get broader coverage
- Non-determinism helps find edge cases you didn't think of

**When to use explicit seeds:**

1. **Reproducing failures** - Copy the seed from the failure message:
   <!-- skip-test -->
   ```go
   // Property failed with seed 1234567890
   specta.Property(t, func(pt *specta.T) { ... }, specta.Seed(1234567890))
   ```

2. **Documentation examples** - For consistent output in docs:
   <!-- skip-test -->
   ```go
   specta.Property(t, func(pt *specta.T) { ... }, specta.Seed(42))
   ```

3. **Framework/library tests** - When you need deterministic behavior for meta-testing

**Don't use explicit seeds for regular tests** - random seeds provide better long-term coverage. If you find a bug, reproduce it with the seed from the failure message, fix it, then remove the explicit seed.

### Combining with Factory Recipes

Factory recipes work seamlessly with property-based testing to build complex types with deterministic generation:

<!-- skip-test -->
```go
func TestUserProcessing(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Use pt as the Source for factory recipes
        user := factory.User().
            NameFromGenerator(specta.String().Alpha().MinLen(1).MaxLen(50)).
            AgeFromGenerator(specta.Int().Range(18, 120)).
            Build(pt)

        // Test property with the generated user
        processed := ProcessUser(user)
        specta.AssertThat(pt, processed.Valid, specta.Equal(true))
    })
}
```

**Why this works**: `specta.T` implements the `Source` interface, so you can pass `pt` directly to `Build()`. This ensures all generation (both from generators and from factory defaults) uses the same random source, enabling proper shrinking when tests fail.

#### Factories with Fixed and Variable Fields

Mix literal values with generated ones to test specific scenarios across varied inputs:

<!-- skip-test -->
```go
func TestAdminPermissions(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Fixed: user is always an admin
        // Variable: other fields are generated
        user := factory.User().
            Role("admin").  // Fixed literal
            NameFromGenerator(specta.String().AlphaNum().NonEmpty()).
            Build(pt)

        // Property: admins can always access admin panel
        specta.AssertThat(pt, CanAccessAdminPanel(user), specta.Equal(true))
    })
}
```

#### Nested Complex Types

Use `FromRecipe` to generate nested structures:

<!-- skip-test -->
```go
func TestNestedStructures(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Generate nested user with variable address
        user := factory.User().
            NameFromRecipe(
                factory.Name().
                    FirstFromGenerator(specta.String().Alpha().MinLen(1)).
                    LastFromGenerator(specta.String().Alpha().MinLen(1)),
            ).
            Build(pt)

        // Test property with nested structure
        fullName := user.Name.First + " " + user.Name.Last
        specta.AssertThat(pt, FormatName(user), specta.Equal(fullName))
    })
}
```

#### Customizing Default Generators

> **Note:** This section covers customizing generators in the context of property-based testing. For more details on factory customization and additional patterns, see the [Factories documentation]({{< relref "/docs/factories/" >}}).

Generated factories create default generators for each field using the `factory/spec/*_gen.go` files. For example:

<!-- skip-test -->
```go
// Default field generators in generated code
var (
    UserAgeGenerator specta.Generator[int] = specta.GeneratorFromProvider(
        func(s specta.Source) int { return int(specta.Int().Draw(s, "")) },
    )
    UserEmailGenerator specta.Generator[string] = specta.String().ExampleHint("email_")
)
```

These generators are **composable** - you can extend them with constraints. To customize defaults, override them in a custom file (`factory/spec/user_defaults.go`):

<!-- skip-test -->
```go
package spec

import (
    "fmt"
    "github.com/james-w/specta"
)

func init() {
    // Override default to add age constraints - now composable!
    UserAgeGenerator = specta.Int().Range(18, 120)

    // Override with a constrained email generator
    UserEmailGenerator = specta.String().
        AlphaNum().
        MinLen(5).
        MaxLen(20).
        Suffix("@test.example.com")
}
```

**Benefits for PBT:**

When using factories in property-based tests, custom default generators ensure generated values are always valid:

<!-- skip-test -->
```go
func TestUserValidation(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Age defaults to Range(18, 120), so we never get invalid ages
        user := factory.User().Build(pt)

        // Property: all generated users pass validation
        specta.AssertThat(pt, ValidateUser(user), specta.BeNil())
    })
}
```

**Composing generators**: You can also extend default generators directly in tests:

<!-- skip-test -->
```go
func TestSpecificEmailDomain(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Extend the default email generator with a filter
        customEmail := spec.UserEmailGenerator.Filter(func(email string) bool {
            return strings.HasSuffix(email, "@example.com")
        })

        user := factory.User().EmailFromGenerator(customEmail).Build(pt)

        specta.AssertThat(pt, user.Email, specta.HasSuffix("@example.com"))
    })
}
```

This approach lets you specify constraints once in the factory defaults, or compose them further when needed. These custom files are preserved when regenerating factories - only `*_gen.go` files are overwritten.

#### Adding Domain-Specific Recipe Methods

You can also extend generated recipes with convenience methods (`factory/user.go`):

<!-- skip-test -->
```go
package factory

// WithAdminRole configures a user with admin privileges
func (r UserRecipe) WithAdminRole() UserRecipe {
    return r.FirstName("Admin").LastName("User").
        Email("admin@example.com").Active(true)
}

// Convenience function for creating admin users
func AdminUser() UserRecipe {
    return User().WithAdminRole()
}
```

And custom matchers for common assertions (`factory/user_matcher.go`):

<!-- skip-test -->
```go
package factory

// IsActive matches users where Active is true
func IsActive() specta.Matcher[User] {
    return UserMatches().Active(specta.Equal(true))
}
```

**Using these customizations in property tests:**

<!-- skip-test -->
```go
func TestAdminPermissions(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Use custom recipe method
        admin := factory.AdminUser().Build(pt)

        // Test property with custom matcher
        specta.AssertThat(pt, admin, factory.IsActive())
        specta.AssertThat(pt, CanAccessAdminPanel(admin), specta.Equal(true))
    })
}
```

### Combining with Matchers

Use matchers to express properties clearly:

<!-- skip-test -->
```go
func TestStringProperties(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        s := specta.Draw(pt, specta.String().Printable(), "s")

        upper := strings.ToUpper(s)

        // Property: uppercase has same length
        specta.AssertThat(pt, len(upper), specta.Equal(len(s)))

        // Property: uppercase contains no lowercase letters
        specta.AssertThat(pt, upper, specta.Not(specta.MatchesRegex(`[a-z]`)))
    })
}
```

## Shrinking

When a property fails, specta automatically shrinks the input to find the minimal failing case:

<!-- skip-test -->
```go
func TestBuggyReverse(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        s := specta.Draw(pt, specta.String(), "s")

        reversed := BuggyReverse(s) // Has a bug with certain inputs

        specta.AssertThat(pt, Reverse(reversed), specta.Equal(s))
    })
}
```

**If this test fails**, specta will:
1. Detect the failure with some complex string
2. Try progressively simpler strings
3. Report the minimal failing input (e.g., `"a"` or `""`)

### Understanding Failure Messages

When a property test fails, specta provides a detailed error message showing exactly what went wrong:

```
=== Property Test Failed ===
Seed: 7
Attempts: 1/100

Generated values:
  a = 0
  b = 0
  c = 0

Failure:
  buggySum([]int64{a, b, c}): expected 0 but got 1

Reproduce: Property(t, check, Seed(7))
```

**What this tells you:**

1. **Seed** - Use this to reproduce the exact failure: `specta.Property(t, check, Seed(7))`

2. **Attempts** - How many iterations ran before failure (1/100 means it failed on the first test)

3. **Generated values** - All values from `Draw()` calls with their labels. These are the **shrunk** values - specta automatically found the minimal failing case.

4. **Failure** - The assertion that failed, showing expected vs actual values

**Example with struct formatting:**

```
=== Property Test Failed ===
Seed: 42
Attempts: 1/1

Generated values:
  parent = Parent{
    Child: UserView{
      ID: "a",
      Name: "a",
      Active: false,
      Score: 0,
    },
  }

Failure:
  parent.Child.Name: expected "deliberately-wrong" but got "a"

Reproduce: Property(t, check, Seed(42))
```

Structs are pretty-printed with proper indentation for readability.

**Why labels matter:** Without descriptive labels in `Draw()` calls, you'd see unhelpful output. Always use meaningful labels that describe what the value represents

## Best Practices

1. **Start with 100 iterations** - Enough to catch most issues
2. **Use specific generator constraints** - Narrow the input space to valid values
3. **Test properties, not implementations** - Focus on "what" not "how"
4. **Combine with example tests** - PBT finds edge cases, examples document behavior
5. **Keep properties simple** - Complex properties are hard to understand when they fail
6. **Use meaningful labels** - `.Draw(pt, "age")` not `.Draw(pt, "x")`

## Next Steps

- [Advanced Topics]({{< relref "/docs/advanced/" >}}) - Custom generators
- [Examples]({{< relref "/docs/examples/" >}}) - More PBT examples
- [API Reference]({{< relref "/docs/api-reference/" >}}) - Property configuration
