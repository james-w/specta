---
title: "Property-Based Testing"
weight: 6
---

# Property-Based Testing

We've been using factories to specify only what matters for each test. But how do you know your code works with *different* values for the fields you didn't specify?

**Property-based testing** verifies that properties hold across many varied inputs, automatically.

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
        s := specta.String().Draw(pt, "s")

        // Test the property
        reversed := Reverse(Reverse(s))
        specta.AssertThat(pt, reversed, specta.Equal(s))
    })
}
```

**What happens:**
1. `Property()` runs your test function 100 times (by default)
2. Each time, generators produce different values (`String()` generates varied strings)
3. If a test fails, specta automatically **shrinks** to find the minimal failing input
4. The failure report shows the exact input that failed

This is property-based testing: verify properties hold across many generated inputs, with automatic shrinking.

## Generators

Generators produce values for property-based tests. specta provides built-in generators for common types:

<!-- skip-test -->
```go
specta.Property(t, func(pt *specta.T) {
    // Generate integers
    n := specta.Int().Range(1, 100).Draw(pt, "n")

    // Generate strings
    name := specta.String().AlphaNum().MinLen(1).MaxLen(50).Draw(pt, "name")

    // Generate booleans
    flag := specta.Bool().Draw(pt, "flag")

    // Generate slices
    numbers := specta.Slice(specta.Int().NonNegative()).MinLen(1).MaxLen(10).Draw(pt, "numbers")

    // Test your code with these generated values
})
```

### Available Generators

- **`Int()`** - Integers with constraints: `.Range(min, max)`, `.Positive()`, `.NonNegative()`, `.Negative()`
- **`String()`** - Strings with constraints: `.AlphaNum()`, `.Alpha()`, `.Printable()`, `.MinLen()`, `.MaxLen()`, `.Prefix()`, `.Suffix()`
- **`Bool()`** - Boolean values
- **`Float64()`** - Floating point numbers
- **`Time()`** - Time values with `.BaseTime()`
- **`Duration()`** - Time durations
- **`Bytes()`** - Byte slices with `.Len()`
- **`UUID()`** - UUID values
- **`Slice(elementGen)`** - Slices of any type: `.MinLen()`, `.MaxLen()`, `.NonEmpty()`
- **`Map(keyGen, valueGen)`** - Maps with generated keys and values

### Generator Constraints

Chain methods to constrain generated values:

<!-- skip-test -->
```go
// Positive integers between 1 and 100
age := specta.Int().Range(1, 100).Draw(pt, "age")

// Non-empty alphanumeric strings
username := specta.String().AlphaNum().NonEmpty().MaxLen(20).Draw(pt, "username")

// Slices with 5-10 elements
items := specta.Slice(specta.Int().Positive()).MinLen(5).MaxLen(10).Draw(pt, "items")
```

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

## Property Examples

### Commutativity

Test that operations are order-independent:

<!-- skip-test -->
```go
func TestAddCommutative(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        x := specta.Int().Draw(pt, "x")
        y := specta.Int().Draw(pt, "y")

        // Property: x + y == y + x
        specta.AssertThat(pt, Add(x, y), specta.Equal(Add(y, x)))
    })
}
```

### Idempotence

Test that applying an operation multiple times equals applying it once:

<!-- skip-test -->
```go
func TestNormalizeIdempotent(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        s := specta.String().Draw(pt, "s")

        once := Normalize(s)
        twice := Normalize(Normalize(s))

        // Property: normalizing twice == normalizing once
        specta.AssertThat(pt, twice, specta.Equal(once))
    })
}
```

### Round-Trip Properties

Encode and decode should be inverses:

<!-- skip-test -->
```go
func TestJSONRoundTrip(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Generate a map
        original := specta.Map(
            specta.String().AlphaNum(),
            specta.Int(),
        ).Draw(pt, "data")

        // Encode to JSON
        jsonData, err := json.Marshal(original)
        specta.AssertThat(pt, err, specta.BeNil())

        // Decode from JSON
        var decoded map[string]int
        err = json.Unmarshal(jsonData, &decoded)
        specta.AssertThat(pt, err, specta.BeNil())

        // Property: decoded == original
        specta.AssertThat(pt, decoded, specta.Equal(original))
    })
}
```

### Invariants

Properties that must always hold:

<!-- skip-test -->
```go
func TestSortedSlice(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Generate a slice of integers
        numbers := specta.Slice(specta.Int()).Draw(pt, "numbers")

        sorted := Sort(numbers)

        // Invariant 1: Sorted slice has same length
        specta.AssertThat(pt, len(sorted), specta.Equal(len(numbers)))

        // Invariant 2: Elements are in order
        for i := 1; i < len(sorted); i++ {
            specta.AssertThat(pt, sorted[i], specta.GreaterThanOrEqual(sorted[i-1]))
        }
    })
}
```

## Advanced Patterns

### Filtering with Assume

Skip test iterations that don't meet preconditions:

<!-- skip-test -->
```go
func TestDivision(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        x := specta.Int().Draw(pt, "x")
        y := specta.Int().Draw(pt, "y")

        // Skip if y is zero
        pt.Assume(y != 0)

        result := x / y
        // Test division properties
        specta.AssertThat(pt, result*y, specta.Equal(x-x%y))
    })
}
```

### Constraining Generated Values

There are three ways to ensure generated values meet requirements:

#### 1. Generator Constraints (Best)

Use built-in constraints when possible - they're fast and always succeed:

<!-- skip-test -->
```go
// Good: constrain the generator directly
age := specta.Int().Range(0, 120).Draw(pt, "age")
username := specta.String().AlphaNum().MinLen(3).MaxLen(20).Draw(pt, "username")
```

#### 2. Filtering (When Constraints Don't Suffice)

Use `.Filter()` for conditions that can't be expressed as generator constraints:

<!-- skip-test -->
```go
// Generate an even number (50% of integers pass - good for filtering)
n := specta.Int().Range(0, 100).Filter(func(x int64) bool {
    return x % 2 == 0
}).Draw(pt, "n")
```

**Caveat**: The generator retries up to 100 times, then skips the test iteration. If your filter has a low success rate, generation will be slower and you'll skip more tests. Keep the success rate reasonably high (aim for >10%).

#### 3. Assume (For Complex Preconditions)

Use `pt.Assume()` for complex preconditions, especially those involving relationships between values:

<!-- skip-test -->
```go
x := specta.Int().Range(0, 100).Draw(pt, "x")
y := specta.Int().Range(0, 100).Draw(pt, "y")

// Skip this test if x >= y (correlation between two generated values)
pt.Assume(x < y)
```

**Caveat**: Skipped tests don't count toward your test total. If too many tests are skipped (>90%), specta will warn you.

**Better approach**: When possible, restructure generation to avoid `Assume()`:

<!-- skip-test -->
```go
// Generate x, then generate y > x
x := specta.Int().Range(0, 99).Draw(pt, "x")
y := specta.Int().Range(x+1, 100).Draw(pt, "y")
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

### Combining with Factory Recipes

Factory recipes work seamlessly with property-based testing to build complex types with deterministic generation:

<!-- skip-test -->
```go
func TestUserProcessing(t *testing.T) {
    specta.Property(t, func(pt *specta.T) {
        // Use pt as the Source for factory recipes
        user := factory.User().
            Name(specta.String().Alpha().MinLen(1).MaxLen(50).Draw(pt, "name")).
            Age(specta.Int().Range(18, 120).Draw(pt, "age")).
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
            Name(specta.String().AlphaNum().NonEmpty().Draw(pt, "name")).
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
                    First(specta.String().Alpha().MinLen(1).Draw(pt, "first")).
                    Last(specta.String().Alpha().MinLen(1).Draw(pt, "last")),
            ).
            Build(pt)

        // Test property with nested structure
        fullName := user.Name.First + " " + user.Name.Last
        specta.AssertThat(pt, FormatName(user), specta.Equal(fullName))
    })
}
```

#### Customizing Default Generators

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
    return UserMatches().Active(specta.Equal(true)).Matcher()
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
        s := specta.String().Printable().Draw(pt, "s")

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
        s := specta.String().Draw(pt, "s")

        reversed := BuggyReverse(s) // Has a bug with certain inputs

        specta.AssertThat(pt, Reverse(reversed), specta.Equal(s))
    })
}
```

**If this test fails**, specta will:
1. Detect the failure with some complex string
2. Try progressively simpler strings
3. Report the minimal failing input (e.g., `"a"` or `""`)

The failure message shows:
- The seed for reproducibility
- The shrunk input that caused the failure
- All generated values with their labels

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
