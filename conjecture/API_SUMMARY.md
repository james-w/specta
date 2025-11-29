# Conjecture Go API - Summary and Hypothesis Mapping

## File Overview

| File | Purpose | Hypothesis Equivalent |
|------|---------|----------------------|
| `choice.go` | Choice types, sequence, comparison | `conjecture/data.py` IRNode, ChoiceSequence |
| `data.go` | ConjectureData, primitive draws | `conjecture/data.py` ConjectureData |
| `gen.go` | Gen[T] and combinators | `strategies/*.py` |
| `shrink.go` | Shrinking passes | `conjecture/shrinker.py` |
| `engine.go` | Test orchestration | `conjecture/engine.py`, `core.py` |
| `testing.go` | Go testing integration | N/A (Go-specific) |

## Key Design Decisions

### 1. Choice Sequence as Foundation

Like Hypothesis, all generation flows through typed choices:

```go
type ChoiceType uint8

const (
    ChoiceBoolean ChoiceType = iota
    ChoiceInteger
    ChoiceFloat
    ChoiceString  // Hypothesis's primitive string type
    ChoiceBytes
)
```

This matches Hypothesis's modern IR with `draw_boolean`, `draw_integer`, `draw_float`, `draw_string`, `draw_bytes`.

### 2. Shrinking Operates on Choice Sequence

The shrinker **never sees generated values**—only the choice sequence:

```go
func Shrink(ctx context.Context, initial *ChoiceSequence, test func(*ConjectureData) bool) ShrinkResult
```

This preserves Hypothesis's key insight: shrink the inputs to the parser, and outputs shrink automatically.

### 3. Shortlex Ordering

```go
func (cs *ChoiceSequence) Compare(other *ChoiceSequence) int {
    // Shorter is simpler
    if len(cs.choices) != len(other.choices) {
        return cmp.Compare(len(cs.choices), len(other.choices))
    }
    // Same length: lexicographic
    for i := range cs.choices {
        if c := compareChoices(cs.choices[i], other.choices[i]); c != 0 {
            return c
        }
    }
    return 0
}
```

### 4. Biased Generation, Ignored During Shrinking

Generation uses biased sampling (small values, boundaries, powers of 2), but this bias is **not stored** in the choice sequence. During replay/shrinking, only the value matters:

```go
func (d *ConjectureData) drawIntegerBiased(params IntegerParams) int64 {
    probe := d.random.Uint64()
    
    switch {
    case probe < 0x2000000000000000: // ~12.5% - shrinkToward
        // ...
    case probe < 0x4000000000000000: // ~12.5% - small values
        // ...
    case probe < 0x6000000000000000: // ~12.5% - boundaries
        // ...
    default: // ~50% - uniform
        // ...
    }
}
```

### 5. Span Markers for Hierarchical Shrinking

```go
func (d *ConjectureData) StartSpan(label string) int
func (d *ConjectureData) EndSpan(spanID int, discard bool)
```

This enables targeted deletion of semantic units, like Hypothesis's `start_example()`/`stop_example()`.

### 6. Collection Shrinking via Continuation Booleans

Lists use continuation booleans rather than upfront length:

```go
func Slice[T any](elem Gen[T], minLen, maxLen int) Gen[[]T] {
    return Gen[[]T]{
        draw: func(d *ConjectureData) ([]T, error) {
            result := make([]T, 0, minLen)
            
            // Required elements
            for i := 0; i < minLen; i++ {
                v, err := elem.Draw(d)
                // ...
            }
            
            // Optional with continuation booleans
            for len(result) < maxLen {
                cont, _ := d.DrawBoolean(0.9 - ...)
                if !cont { break }
                // ...
            }
        },
    }
}
```

When shrinking flips booleans to false, elements disappear naturally.

### 7. ShrinkToward for Integers

```go
type IntegerParams struct {
    Min          int64
    Max          int64
    ShrinkToward int64  // Target for shrinking
}
```

This preserves Hypothesis's approach where choice value 0 produces `ShrinkToward`, making lexicographic reduction move toward the target.

## Shrinking Passes

The shrinker implements passes similar to Hypothesis's ECOOP 2020 paper:

| Pass | Purpose |
|------|---------|
| `passDeleteSpans` | Delete marked regions (Hypothesis's intervals) |
| `passDeleteChunks` | Delete arbitrary chunks |
| `passZeroChunks` | Replace with zero values |
| `passMinimizeIndividual` | Binary search toward targets |
| `passSwapAdjacent` | Swap for lexicographic improvement |
| `passRedistribute` | Move value between adjacent choices |

## Usage Example

```go
func TestReverseReverse(t *testing.T) {
    ct := conjecture.NewT(t)
    
    conjecture.Check(ct, "reverse_identity",
        conjecture.Slice(conjecture.Integer(-100, 100), 0, 50),
        func(xs []int64) error {
            original := clone(xs)
            reverse(xs)
            reverse(xs)
            if !equal(xs, original) {
                return errors.New("reverse(reverse(xs)) != xs")
            }
            return nil
        })
}
```

## Differences from Hypothesis

| Aspect | Hypothesis | Conjecture (Go) |
|--------|-----------|-----------------|
| Type safety | Dynamic typing | Go generics `Gen[T]` |
| Error handling | Exceptions | Explicit `error` returns |
| Cancellation | N/A | `context.Context` |
| Configuration | `@settings` decorator | Functional options |
| Test framework | pytest/unittest | `testing.T` integration |

## Your Problem: Mode Bits

**Your approach**: Initial bits select mode (small/boundary/uniform), stored in sequence.

**Problem**: Mode bytes shrink independently of value bytes, and shorter ≠ simpler.

**Hypothesis solution** (and ours):
1. Mode selection happens during generation only
2. Only the **value** is stored in the choice sequence  
3. Constraints (min, max, shrinkToward, intervals) are metadata
4. Shrinker operates on values using shortlex, ignoring how they were generated

**Key insight**: If you store `[mode=2, value=50]` vs `[value=50]`, the shrinker sees different sequences. By storing only values with constraints as metadata, the shrinker naturally finds minimal values regardless of which generation "mode" produced them.

## Recommended Changes to Your Library

1. **Don't encode mode in the bitstream**—use it only during generation
2. **Store typed choices** with constraints as metadata (not part of the shrinkable sequence)
3. **Ensure monotonicity**: smaller choice values → simpler outputs
4. **Use continuation booleans** for collections, not upfront length
5. **Implement span markers** for targeted shrinking of semantic units
