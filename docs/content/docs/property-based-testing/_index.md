---
title: "Property-Based Testing"
weight: 6
---

# Property-Based Testing

Property-based testing (PBT) tests your code with many generated inputs, verifying properties that should always hold true.

## What is Property-Based Testing?

**Example-based testing** checks specific cases:

```go
func TestAdd(t *testing.T) {
    AssertThat(t, Add(2, 3), Equal(5))
    AssertThat(t, Add(0, 0), Equal(0))
    AssertThat(t, Add(-1, 1), Equal(0))
}
```

**Property-based testing** checks general properties across many inputs:

```go
func TestAddProperties(t *testing.T) {
    p := specta.NewGen()

    for i := 0; i < 100; i++ {
        a := p.Next()
        b := p.Next()

        // Property: addition is commutative
        AssertThat(t, Add(a, b), Equal(Add(b, a)))

        // Property: adding zero is identity
        AssertThat(t, Add(a, 0), Equal(a))
    }
}
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

## Using Factories for PBT

specta's factories and primitives are perfect for property-based testing:

### Deterministic Generation

```go
func TestUserSerialization(t *testing.T) {
    p := specta.NewGen()

    for i := 0; i < 100; i++ {
        // Generate varied test data
        user := factory.NewUser(p).Build()

        // Property: round-trip serialization
        json, _ := json.Marshal(user)
        var decoded User
        json.Unmarshal(json, &decoded)

        AssertThat(t, decoded, factory.MatchUser().
            WithID(Equal(user.ID)).
            WithName(Equal(user.Name)).
            WithEmail(Equal(user.Email)))
    }
}
```

### Varying Specific Fields

```go
func TestEmailValidation(t *testing.T) {
    p := specta.NewGen()

    validEmails := []string{
        "user@example.com",
        "test.user@example.org",
        "admin+tag@company.io",
    }

    for _, email := range validEmails {
        user := factory.NewUser(p).
            WithEmail(email).
            Build()

        err := ValidateUser(user)
        AssertThat(t, err, BeNil())
    }
}
```

## Property Examples

### Round-Trip Properties

Encode and decode should be inverses:

```go
func TestJSONRoundTrip(t *testing.T) {
    p := specta.NewGen()

    for i := 0; i < 100; i++ {
        original := factory.NewOrder(p).
            WithUser(factory.NewUser(p).Build()).
            WithItems([]Item{
                factory.NewItem(p).Build(),
                factory.NewItem(p).Build(),
            }).
            Build()

        // Encode
        data, err := json.Marshal(original)
        AssertThat(t, err, BeNil())

        // Decode
        var decoded Order
        err = json.Unmarshal(data, &decoded)
        AssertThat(t, err, BeNil())

        // Property: decoded == original
        AssertThat(t, decoded, DeepEqual(original))
    }
}
```

### Idempotence

Applying an operation twice should equal applying it once:

```go
func TestNormalizationIdempotent(t *testing.T) {
    p := specta.NewGen()

    for i := 0; i < 100; i++ {
        user := factory.NewUser(p).Build()

        normalized1 := NormalizeUser(user)
        normalized2 := NormalizeUser(normalized1)

        // Property: normalizing twice = normalizing once
        AssertThat(t, normalized2, DeepEqual(normalized1))
    }
}
```

### Invariants

Properties that must always hold:

```go
func TestOrderTotalInvariant(t *testing.T) {
    p := specta.NewGen()

    for i := 0; i < 100; i++ {
        // Generate orders with random items
        numItems := 1 + (p.Next() % 10)
        items := make([]Item, numItems)
        expectedTotal := 0

        for j := 0; j < numItems; j++ {
            price := p.Next()
            items[j] = factory.NewItem(p).WithPrice(price).Build()
            expectedTotal += price
        }

        order := factory.NewOrder(p).WithItems(items).Build()

        calculated := CalculateOrderTotal(order)

        // Invariant: total = sum of item prices
        AssertThat(t, calculated, Equal(expectedTotal))
    }
}
```

### Commutativity

Order of operations doesn't matter:

```go
func TestMergeCommutative(t *testing.T) {
    p := specta.NewGen()

    for i := 0; i < 100; i++ {
        user1 := factory.NewUser(p).Build()
        user2 := factory.NewUser(p).Build()

        // Property: Merge(a, b) should equal Merge(b, a)
        merged1 := MergeUsers(user1, user2)
        merged2 := MergeUsers(user2, user1)

        AssertThat(t, merged1, DeepEqual(merged2))
    }
}
```

## Advanced PBT Patterns

### Stateful Testing

Test sequences of operations:

```go
func TestUserStateMachine(t *testing.T) {
    p := specta.NewGen()

    for i := 0; i < 50; i++ {
        user := factory.NewUser(p).
            WithStatus("pending").
            Build()

        // Apply random state transitions
        numTransitions := 1 + (p.Next() % 5)
        for j := 0; j < numTransitions; j++ {
            transition := p.Next() % 3
            switch transition {
            case 0:
                user = ActivateUser(user)
            case 1:
                user = SuspendUser(user)
            case 2:
                user = ReactivateUser(user)
            }

            // Invariant: status is always valid
            AssertThat(t, user.Status, AnyOf(
                Equal("pending"),
                Equal("active"),
                Equal("suspended"),
            ))
        }
    }
}
```

### Shrinking

When a property fails, try to find the minimal failing case:

```go
func TestWithShrinking(t *testing.T) {
    p := specta.NewGen()

    for i := 0; i < 100; i++ {
        // Generate increasingly complex data
        numUsers := 1 + (p.Next() % 20)
        users := make([]User, numUsers)

        for j := 0; j < numUsers; j++ {
            users[j] = factory.NewUser(p).Build()
        }

        result := ProcessUsers(users)

        // If this fails, binary search to find minimal failing case
        if !ValidateResult(result) {
            // Shrink: try with fewer users
            low, high := 0, numUsers
            for low < high {
                mid := (low + high) / 2
                testResult := ProcessUsers(users[:mid])
                if ValidateResult(testResult) {
                    low = mid + 1
                } else {
                    high = mid
                }
            }
            t.Fatalf("minimal failing case: %d users", low)
        }
    }
}
```

### Testing with Constraints

Generate data that satisfies specific constraints:

```go
func TestAgeValidation(t *testing.T) {
    p := specta.NewGen()

    // Test with users of various ages
    for age := 0; age < 150; age++ {
        user := factory.NewUser(p).
            WithAge(age).
            Build()

        err := ValidateUser(user)

        if age < 18 {
            AssertThat(t, err, NotBeNil())
            AssertThat(t, err.Error(), ContainString("age"))
        } else if age > 120 {
            AssertThat(t, err, NotBeNil())
            AssertThat(t, err.Error(), ContainString("age"))
        } else {
            AssertThat(t, err, BeNil())
        }
    }
}
```

## Combining with Matchers

Use matchers to express properties clearly:

```go
func TestCacheInvariant(t *testing.T) {
    p := specta.NewGen()
    cache := NewCache()

    for i := 0; i < 100; i++ {
        key := p.String("key")
        value := factory.NewUser(p).Build()

        // Set
        cache.Set(key, value)

        // Get
        retrieved, found := cache.Get(key)

        // Properties
        AssertThat(t, found, Equal(true))
        AssertThat(t, retrieved, factory.MatchUser().
            WithID(Equal(value.ID)).
            WithName(Equal(value.Name)))
    }
}
```

## Best Practices

1. **Start with 100 iterations** - Enough to catch most issues
2. **Use deterministic generation** - Same seed = same tests
3. **Test properties, not implementations** - Focus on "what" not "how"
4. **Combine with example tests** - PBT finds edge cases, examples document behavior
5. **Keep properties simple** - Complex properties are hard to understand when they fail
6. **Use meaningful assertions** - Matchers help express intent

## Next Steps

- [Advanced Topics](/docs/advanced/) - Custom generators and matchers
- [Examples](/docs/examples/) - More PBT examples
- [API Reference](/docs/api-reference/) - Primitives configuration
