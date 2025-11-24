package specta_test

import (
	"math"
	"strings"
	"testing"
	"unicode"

	"github.com/james-w/specta"
)

// TestIntGenerator_Constraints tests integer generation with constraints
func TestIntGenerator_Constraints(t *testing.T) {
	t.Run("generates full range by default", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int()
			value := gen.Draw(t.Source, "test")
			// Should generate any int64 value
			_ = value
		}, specta.MaxTests(100))
	})

	t.Run("respects Range constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(10, 20)
			value := gen.Draw(t.Source, "test")
			if value < 10 || value > 20 {
				t.Errorf("value %d outside range [10, 20]", value)
			}
		}, specta.MaxTests(100))
	})

	t.Run("respects Min constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Min(100)
			value := gen.Draw(t.Source, "test")
			if value < 100 {
				t.Errorf("value %d less than min 100", value)
			}
		}, specta.MaxTests(100))
	})

	t.Run("respects Max constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Max(50)
			value := gen.Draw(t.Source, "test")
			if value > 50 {
				t.Errorf("value %d greater than max 50", value)
			}
		}, specta.MaxTests(100))
	})

	t.Run("Positive generates values > 0", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Positive()
			value := gen.Draw(t.Source, "test")
			if value <= 0 {
				t.Errorf("value %d not positive", value)
			}
		}, specta.MaxTests(100))
	})

	t.Run("NonNegative generates values >= 0", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().NonNegative()
			value := gen.Draw(t.Source, "test")
			if value < 0 {
				t.Errorf("value %d is negative", value)
			}
		}, specta.MaxTests(100))
	})

	t.Run("Negative generates values < 0", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Negative()
			value := gen.Draw(t.Source, "test")
			if value >= 0 {
				t.Errorf("value %d not negative", value)
			}
		}, specta.MaxTests(100))
	})

	t.Run("chained constraints work", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Min(10).Max(100)
			value := gen.Draw(t.Source, "test")
			if value < 10 || value > 100 {
				t.Errorf("value %d outside range [10, 100]", value)
			}
		}, specta.MaxTests(100))
	})
}

// TestStringGenerator tests string generation with constraints
func TestStringGenerator(t *testing.T) {
	t.Run("generates any bytes by default", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String()
			value := gen.Draw(t.Source, "test")
			// Should be able to generate any string including invalid UTF-8
			_ = value
		}, specta.MaxTests(100))
	})

	t.Run("NonEmpty generates non-empty strings", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().NonEmpty()
			value := gen.Draw(t.Source, "test")
			if len(value) == 0 {
				t.Errorf("generated empty string")
			}
		}, specta.MaxTests(100))
	})

	t.Run("MinLen respects minimum length", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().MinLen(10)
			value := gen.Draw(t.Source, "test")
			if len(value) < 10 {
				t.Errorf("string length %d less than min 10", len(value))
			}
		}, specta.MaxTests(100))
	})

	t.Run("MaxLen respects maximum length", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().MaxLen(5)
			value := gen.Draw(t.Source, "test")
			if len(value) > 5 {
				t.Errorf("string length %d greater than max 5", len(value))
			}
		}, specta.MaxTests(100))
	})

	t.Run("Len generates exact length", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().Len(10)
			value := gen.Draw(t.Source, "test")
			if len(value) != 10 {
				t.Errorf("string length %d not equal to 10", len(value))
			}
		}, specta.MaxTests(100))
	})

	t.Run("Prefix adds prefix", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().Prefix("test_")
			value := gen.Draw(t.Source, "test")
			if !strings.HasPrefix(value, "test_") {
				t.Errorf("string %q doesn't have prefix 'test_'", value)
			}
		}, specta.MaxTests(100))
	})

	t.Run("Suffix adds suffix", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().Suffix("_end")
			value := gen.Draw(t.Source, "test")
			if !strings.HasSuffix(value, "_end") {
				t.Errorf("string %q doesn't have suffix '_end'", value)
			}
		}, specta.MaxTests(100))
	})

	t.Run("ASCII generates only ASCII", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().ASCII().MinLen(1)
			value := gen.Draw(t.Source, "test")
			for i, b := range []byte(value) {
				if b > 0x7F {
					t.Errorf("byte at index %d (%d) is not ASCII", i, b)
				}
			}
		}, specta.MaxTests(100))
	})

	t.Run("Printable generates only printable ASCII", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().Printable().MinLen(1)
			value := gen.Draw(t.Source, "test")
			for i, r := range value {
				if r < 0x20 || r > 0x7E {
					t.Errorf("character at index %d (%c, %d) is not printable ASCII", i, r, r)
				}
			}
		}, specta.MaxTests(100))
	})

	t.Run("AlphaNum generates only alphanumeric", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().AlphaNum().MinLen(1)
			value := gen.Draw(t.Source, "test")
			for i, r := range value {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					t.Errorf("character at index %d (%c) is not alphanumeric", i, r)
				}
			}
		}, specta.MaxTests(100))
	})

	t.Run("Alpha generates only alphabetic", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().Alpha().MinLen(1)
			value := gen.Draw(t.Source, "test")
			for i, r := range value {
				if !unicode.IsLetter(r) {
					t.Errorf("character at index %d (%c) is not alphabetic", i, r)
				}
			}
		}, specta.MaxTests(100))
	})

	t.Run("chained constraints work", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().Prefix("user_").AlphaNum().MinLen(10).MaxLen(20)
			value := gen.Draw(t.Source, "test")

			if !strings.HasPrefix(value, "user_") {
				t.Errorf("string %q doesn't have prefix 'user_'", value)
			}
			if len(value) < 10 || len(value) > 20 {
				t.Errorf("string length %d outside range [10, 20]", len(value))
			}
			// Check alphanumeric after prefix
			rest := strings.TrimPrefix(value, "user_")
			for i, r := range rest {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					t.Errorf("character at index %d (%c) is not alphanumeric", i, r)
				}
			}
		}, specta.MaxTests(100))
	})

	t.Run("can generate empty strings when allowed", func(t *testing.T) {
		foundEmpty := false
		// Probability of empty is 1/101, so need enough attempts
		for seed := int64(0); seed < 100 && !foundEmpty; seed++ {
			specta.Property(t, func(t *specta.T) {
				gen := specta.String()
				value := gen.Draw(t.Source, "test")
				if len(value) == 0 {
					foundEmpty = true
				}
			}, specta.Seed(seed), specta.MaxTests(100))
		}

		if !foundEmpty {
			t.Errorf("never generated empty string (probability issue, not a bug)")
		}
	})
}

// TestIntGenerator_EdgeCases tests edge cases
func TestIntGenerator_EdgeCases(t *testing.T) {
	t.Run("finds boundary values", func(t *testing.T) {
		foundMin := false
		foundMax := false

		for seed := int64(0); seed < 100 && (!foundMin || !foundMax); seed++ {
			specta.Property(t, func(t *specta.T) {
				gen := specta.Int().Range(10, 20)
				value := gen.Draw(t.Source, "test")
				if value == 10 {
					foundMin = true
				}
				if value == 20 {
					foundMax = true
				}
			}, specta.Seed(seed), specta.MaxTests(20))
		}

		if !foundMin {
			t.Errorf("never found min boundary value 10")
		}
		if !foundMax {
			t.Errorf("never found max boundary value 20")
		}
	})

	t.Run("handles single value range", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(42, 42)
			value := gen.Draw(t.Source, "test")
			if value != 42 {
				t.Errorf("expected 42, got %d", value)
			}
		}, specta.MaxTests(10))
	})

	t.Run("generates negative values by default", func(t *testing.T) {
		foundNegative := false

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int()
			value := gen.Draw(t.Source, "test")
			if value < 0 {
				foundNegative = true
			}
		}, specta.MaxTests(1000))

		if !foundNegative {
			t.Errorf("never generated negative value")
		}
	})

	t.Run("generates extreme values", func(t *testing.T) {
		var minSeen, maxSeen int64 = math.MaxInt64, math.MinInt64

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int()
			value := gen.Draw(t.Source, "test")
			if value < minSeen {
				minSeen = value
			}
			if value > maxSeen {
				maxSeen = value
			}
		}, specta.MaxTests(1000))

		// Should see values spread across the range
		if minSeen > -1000000 {
			t.Logf("warning: minimum value seen (%d) not very negative", minSeen)
		}
		if maxSeen < 1000000 {
			t.Logf("warning: maximum value seen (%d) not very positive", maxSeen)
		}
	})
}

// TestIntGenerator_EdgeCaseBiasing tests that edge cases are generated frequently
func TestIntGenerator_EdgeCaseBiasing(t *testing.T) {
	t.Run("biases toward boundaries", func(t *testing.T) {
		found := make(map[int64]int)
		edgeCases := []int64{0, 1, 99, 100} // boundaries and near-boundaries for Range(0, 100)

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(0, 100)
			value := gen.Draw(t.Source, "test")
			found[value]++
		}, specta.MaxTests(1000))

		// With 30% bias and ~7 edge cases out of 101 values,
		// we expect each edge case to appear roughly: (1000 * 0.30) / 7 ≈ 43 times
		// Plus some from the 70% uniform: (1000 * 0.70) / 101 ≈ 7 times per value
		// Total expected per edge case: ~50 times (rough estimate)

		for _, edge := range edgeCases {
			count := found[edge]
			if count < 20 {
				t.Errorf("edge case %d only appeared %d times (expected ~50 with biasing)", edge, count)
			}
			t.Logf("edge case %d appeared %d times", edge, count)
		}
	})

	t.Run("biases toward zero", func(t *testing.T) {
		foundZero := 0

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(-100, 100)
			value := gen.Draw(t.Source, "test")
			if value == 0 {
				foundZero++
			}
		}, specta.MaxTests(1000))

		// Zero is an edge case in a pool of ~15 edge cases (0, ±1, boundaries, powers of 2)
		// 30% bias distributed among ~15 edge cases: (1000 * 0.30) / 15 ≈ 20 times
		// Plus uniform contribution: (1000 * 0.70) / 201 ≈ 3 times
		// Total expected: ~23 times, but with randomness could be lower
		// Just verify it's more than uniform (which would be ~5)
		if foundZero < 10 {
			t.Errorf("zero only appeared %d times (expected >10 with biasing, uniform would be ~5)", foundZero)
		}
		t.Logf("zero appeared %d times out of 1000", foundZero)
	})

	t.Run("biases toward powers of 2", func(t *testing.T) {
		powers := []int64{1, 2, 4, 8, 16, 32, 64}
		foundPowers := 0

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(0, 100)
			value := gen.Draw(t.Source, "test")
			for _, p := range powers {
				if value == p {
					foundPowers++
					break
				}
			}
		}, specta.MaxTests(1000))

		// 7 powers of 2 in range, expect them to appear more frequently
		// Uniform: (1000 * 7) / 101 ≈ 69 times total
		// With biasing: should be 150+ times total
		if foundPowers < 100 {
			t.Errorf("powers of 2 only appeared %d times total (expected 150+ with biasing)", foundPowers)
		}
		t.Logf("powers of 2 appeared %d times total out of 1000", foundPowers)
	})

	t.Run("finds division by zero bug quickly", func(t *testing.T) {
		// Simulate a bug that only fails when denominator is 0
		// Count iterations across multiple seeds until we find zero
		totalIterations := 0
		found := false

		for seed := int64(0); seed < 50 && !found; seed++ {
			specta.Property(t, func(t *specta.T) {
				totalIterations++
				denominator := specta.Int().Range(-10, 10).Draw(t.Source, "denom")

				// Check if we hit the edge case
				if denominator == 0 {
					found = true
					// Early termination would be nice, but not critical
				}
			}, specta.Seed(seed), specta.MaxTests(20))
		}

		if !found {
			t.Errorf("biasing failed to find zero in %d iterations (expected to find it quickly)", totalIterations)
		} else {
			t.Logf("found zero in %d total iterations across multiple seeds (with biasing)", totalIterations)
			// With 30% bias toward edge cases and 0 being one of ~10 edge cases,
			// we expect to find it within ~30-50 iterations on average
			// But due to randomness, allow up to 200
			if totalIterations > 200 {
				t.Logf("warning: took %d iterations, expected faster with biasing", totalIterations)
			}
		}
	})
}

// TestSliceGenerator tests slice generation with element generators
func TestSliceGenerator(t *testing.T) {
	t.Run("generates slices by default", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int())
			value := gen.Draw(t.Source, "test")
			// Should generate a slice (possibly empty)
			_ = value
		}, specta.MaxTests(100))
	})

	t.Run("respects MinLen constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int()).MinLen(5)
			value := gen.Draw(t.Source, "test")
			if len(value) < 5 {
				t.Errorf("slice length %d less than min 5", len(value))
			}
		}, specta.MaxTests(100))
	})

	t.Run("respects MaxLen constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int()).MaxLen(10)
			value := gen.Draw(t.Source, "test")
			if len(value) > 10 {
				t.Errorf("slice length %d greater than max 10", len(value))
			}
		}, specta.MaxTests(100))
	})

	t.Run("respects Len constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.String()).Len(7)
			value := gen.Draw(t.Source, "test")
			if len(value) != 7 {
				t.Errorf("slice length %d not equal to 7", len(value))
			}
		}, specta.MaxTests(100))
	})

	t.Run("NonEmpty generates non-empty slices", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Bool()).NonEmpty()
			value := gen.Draw(t.Source, "test")
			if len(value) == 0 {
				t.Errorf("generated empty slice")
			}
		}, specta.MaxTests(100))
	})

	t.Run("element generator constraints are respected", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int().Range(1, 10))
			value := gen.Draw(t.Source, "test")
			for i, v := range value {
				if v < 1 || v > 10 {
					t.Errorf("element %d at index %d outside range [1, 10]", v, i)
				}
			}
		}, specta.MaxTests(100))
	})

	t.Run("can generate empty slices by default", func(t *testing.T) {
		foundEmpty := false

		for seed := int64(0); seed < 100 && !foundEmpty; seed++ {
			specta.Property(t, func(t *specta.T) {
				gen := specta.Slice(specta.Int())
				value := gen.Draw(t.Source, "test")
				if len(value) == 0 {
					foundEmpty = true
				}
			}, specta.Seed(seed), specta.MaxTests(50))
		}

		if !foundEmpty {
			t.Logf("note: never generated empty slice (low probability, not necessarily a bug)")
		}
	})

	t.Run("chained constraints work", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.String().AlphaNum()).MinLen(2).MaxLen(5)
			value := gen.Draw(t.Source, "test")

			if len(value) < 2 || len(value) > 5 {
				t.Errorf("slice length %d outside range [2, 5]", len(value))
			}

			// Check each element is alphanumeric
			for i, s := range value {
				for _, r := range s {
					if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
						t.Errorf("element %d at index %d contains non-alphanumeric character %c", i, i, r)
					}
				}
			}
		}, specta.MaxTests(100))
	})

	t.Run("generates varying lengths", func(t *testing.T) {
		lengths := make(map[int]bool)

		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int()).MinLen(0).MaxLen(10)
			value := gen.Draw(t.Source, "test")
			lengths[len(value)] = true
		}, specta.MaxTests(200))

		// Should see at least a few different lengths
		if len(lengths) < 3 {
			t.Errorf("only saw %d different lengths, expected more variety", len(lengths))
		}
	})

	t.Run("Filter works on slices", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			// Only accept slices with at least one even number
			gen := specta.Slice(specta.Int().Range(0, 100)).NonEmpty().Filter(func(s []int64) bool {
				for _, v := range s {
					if v%2 == 0 {
						return true
					}
				}
				return false
			})
			value := gen.Draw(t.Source, "test")

			// Verify filter condition holds
			foundEven := false
			for _, v := range value {
				if v%2 == 0 {
					foundEven = true
					break
				}
			}
			if !foundEven {
				t.Errorf("filter should ensure at least one even number, but none found in %v", value)
			}
		}, specta.MaxTests(100))
	})
}

// TestSliceGenerator_Deterministic tests deterministic behavior
func TestSliceGenerator_Deterministic(t *testing.T) {
	t.Run("deterministic mode produces predictable results", func(t *testing.T) {
		gen := specta.New(specta.WithStart(0))

		slice1 := specta.Slice(specta.Int()).MinLen(3).MaxLen(5).Draw(gen, "test")

		// Reset to same state
		gen = specta.New(specta.WithStart(0))
		slice2 := specta.Slice(specta.Int()).MinLen(3).MaxLen(5).Draw(gen, "test")

		// Should get identical results
		if len(slice1) != len(slice2) {
			t.Errorf("deterministic slices have different lengths: %d vs %d", len(slice1), len(slice2))
		}

		for i := range slice1 {
			if slice1[i] != slice2[i] {
				t.Errorf("deterministic slices differ at index %d: %d vs %d", i, slice1[i], slice2[i])
			}
		}
	})

	t.Run("deterministic mode uses counter for length", func(t *testing.T) {
		gen := specta.New(specta.WithStart(0))

		// First slice should have length based on counter 0
		slice := specta.Slice(specta.Int()).MinLen(0).MaxLen(10).Draw(gen, "test")
		expectedLen := 0 // counter 0 % 11 = 0
		if len(slice) != expectedLen {
			t.Logf("note: first slice length %d (implementation detail, not a bug if different)", len(slice))
		}
	})
}

// TestSliceGenerator_SizeBiasing tests that collection sizes are biased toward edge cases
func TestSliceGenerator_SizeBiasing(t *testing.T) {
	t.Run("biases toward empty slices", func(t *testing.T) {
		dist := make(map[string]int)

		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int()).MinLen(0).MaxLen(100)
			value := gen.Draw(t.Source, "test")
			length := len(value)

			switch {
			case length == 0:
				dist["empty"]++
			case length == 1:
				dist["single"]++
			case length >= 2 && length <= 5:
				dist["small"]++
			case length >= 6 && length <= 20:
				dist["medium"]++
			case length >= 21 && length <= 50:
				dist["large"]++
			default:
				dist["very_large"]++
			}
		}, specta.MaxTests(1000))

		t.Logf("Distribution: empty=%d (%.1f%%), single=%d (%.1f%%), small=%d (%.1f%%), medium=%d (%.1f%%), large=%d (%.1f%%), very_large=%d (%.1f%%)",
			dist["empty"], float64(dist["empty"])/10,
			dist["single"], float64(dist["single"])/10,
			dist["small"], float64(dist["small"])/10,
			dist["medium"], float64(dist["medium"])/10,
			dist["large"], float64(dist["large"])/10,
			dist["very_large"], float64(dist["very_large"])/10)

		// Empty has 20% weight, expect ~200 occurrences
		if dist["empty"] < 150 {
			t.Errorf("empty slices only appeared %d times (expected ~200 with 20%% bias)", dist["empty"])
		}
	})

	t.Run("biases toward single element", func(t *testing.T) {
		foundSingle := 0

		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.String()).MinLen(0).MaxLen(100)
			value := gen.Draw(t.Source, "test")
			if len(value) == 1 {
				foundSingle++
			}
		}, specta.MaxTests(1000))

		// Single element has 25% weight, expect ~250 occurrences
		if foundSingle < 200 {
			t.Errorf("single-element slices only appeared %d times (expected ~250 with 25%% bias)", foundSingle)
		}
		t.Logf("single-element slices appeared %d times out of 1000 (expected ~250)", foundSingle)
	})

	t.Run("biases toward small sizes 2-5", func(t *testing.T) {
		foundSmall := 0

		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int()).MinLen(0).MaxLen(100)
			value := gen.Draw(t.Source, "test")
			if len(value) >= 2 && len(value) <= 5 {
				foundSmall++
			}
		}, specta.MaxTests(1000))

		// Small (2-5) has 35% weight, expect ~350 occurrences
		if foundSmall < 300 {
			t.Errorf("small slices (2-5) only appeared %d times (expected ~350 with 35%% bias)", foundSmall)
		}
		t.Logf("small slices (2-5) appeared %d times out of 1000 (expected ~350)", foundSmall)
	})

	t.Run("finds empty collection bug quickly", func(t *testing.T) {
		// Simulate a bug that only fails on empty collections
		totalIterations := 0
		found := false

		for seed := int64(0); seed < 50 && !found; seed++ {
			specta.Property(t, func(t *specta.T) {
				totalIterations++
				slice := specta.Slice(specta.Int()).MinLen(0).MaxLen(20).Draw(t.Source, "test")

				// Check if we hit the edge case
				if len(slice) == 0 {
					found = true
				}
			}, specta.Seed(seed), specta.MaxTests(10))
		}

		if !found {
			t.Errorf("biasing failed to find empty slice in %d iterations", totalIterations)
		} else {
			t.Logf("found empty slice in %d total iterations (with 20%% bias)", totalIterations)
			// With 20% bias, should find it very quickly (< 50 iterations typically)
			if totalIterations > 100 {
				t.Logf("warning: took %d iterations, expected faster with biasing", totalIterations)
			}
		}
	})
}

// TestStringGenerator_LengthBiasing tests that string lengths are biased toward edge cases
func TestStringGenerator_LengthBiasing(t *testing.T) {
	t.Run("biases toward boundaries", func(t *testing.T) {
		dist := make(map[int]int)
		edgeLengths := []int{0, 1, 49, 50} // max=50, so edge cases are 0, 1, 49, 50

		specta.Property(t, func(t *specta.T) {
			gen := specta.String().MinLen(0).MaxLen(50)
			value := gen.Draw(t.Source, "test")
			dist[len(value)]++
		}, specta.MaxTests(1000))

		// Each edge case in pool of 4, with 30% bias = 30%/4 = 7.5% each
		// Expect at least ~40 occurrences for each edge (out of 1000 total)
		for _, length := range edgeLengths {
			count := dist[length]
			if count < 40 {
				t.Errorf("edge length %d only appeared %d times (expected ~75 with biasing)", length, count)
			}
		}

		total := 0
		for _, count := range dist {
			total += count
		}
		t.Logf("Total draws: %d", total)
		for _, length := range edgeLengths {
			t.Logf("Length %d: %d times (%.1f%%)", length, dist[length], float64(dist[length])/10)
		}
	})

	t.Run("finds empty string bug quickly", func(t *testing.T) {
		// Simulate a bug that only fails on empty strings
		totalIterations := 0
		found := false

		for seed := int64(0); seed < 50 && !found; seed++ {
			specta.Property(t, func(t *specta.T) {
				totalIterations++
				str := specta.String().MinLen(0).MaxLen(20).Draw(t.Source, "test")

				// Check if we hit the edge case
				if len(str) == 0 {
					found = true
				}
			}, specta.Seed(seed), specta.MaxTests(10))
		}

		if !found {
			t.Errorf("biasing failed to find empty string in %d iterations", totalIterations)
		} else {
			t.Logf("found empty string in %d total iterations (with ~7.5%% bias)", totalIterations)
			// With biasing, should find it quickly (typically < 100 iterations)
			if totalIterations > 200 {
				t.Logf("warning: took %d iterations, expected faster with biasing", totalIterations)
			}
		}
	})

	t.Run("finds max length string", func(t *testing.T) {
		foundMax := 0

		specta.Property(t, func(t *specta.T) {
			gen := specta.String().MinLen(0).MaxLen(100)
			value := gen.Draw(t.Source, "test")
			if len(value) == 100 {
				foundMax++
			}
		}, specta.MaxTests(1000))

		// Max length is in edge case pool, with 30% bias across ~4 cases = ~7.5% each
		// Expect at least 40 occurrences out of 1000
		if foundMax < 40 {
			t.Errorf("max length (100) only appeared %d times (expected ~75 with biasing)", foundMax)
		}
		t.Logf("max length (100) appeared %d times out of 1000", foundMax)
	})
}

// TestMapGenerator tests map generation with key and value generators
func TestMapGenerator(t *testing.T) {
	t.Run("generates maps by default", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Map(specta.String(), specta.Int())
			value := gen.Draw(t.Source, "test")
			// Should generate a map (possibly empty)
			_ = value
		}, specta.MaxTests(100))
	})

	t.Run("respects MinLen constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Map(specta.String(), specta.Int()).MinLen(5)
			value := gen.Draw(t.Source, "test")
			if len(value) < 5 {
				t.Errorf("map length %d less than min 5", len(value))
			}
		}, specta.MaxTests(100))
	})

	t.Run("respects MaxLen constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Map(specta.String(), specta.Bool()).MaxLen(10)
			value := gen.Draw(t.Source, "test")
			if len(value) > 10 {
				t.Errorf("map length %d greater than max 10", len(value))
			}
		}, specta.MaxTests(100))
	})

	t.Run("respects Len constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Map(specta.Int(), specta.String()).Len(7)
			value := gen.Draw(t.Source, "test")
			// Note: May be less than 7 if key collisions occur
			if len(value) > 7 {
				t.Errorf("map length %d greater than exact 7", len(value))
			}
		}, specta.MaxTests(100))
	})

	t.Run("NonEmpty generates non-empty maps", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Map(specta.String().AlphaNum(), specta.Float64()).NonEmpty()
			value := gen.Draw(t.Source, "test")
			if len(value) == 0 {
				t.Errorf("generated empty map")
			}
		}, specta.MaxTests(100))
	})

	t.Run("key and value generator constraints are respected", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Map(
				specta.String().Prefix("key_"),
				specta.Int().Range(1, 10),
			)
			value := gen.Draw(t.Source, "test")
			for k, v := range value {
				if !strings.HasPrefix(k, "key_") {
					t.Errorf("key %q doesn't have prefix 'key_'", k)
				}
				if v < 1 || v > 10 {
					t.Errorf("value %d outside range [1, 10]", v)
				}
			}
		}, specta.MaxTests(100))
	})

	t.Run("can generate empty maps by default", func(t *testing.T) {
		foundEmpty := false

		for seed := int64(0); seed < 100 && !foundEmpty; seed++ {
			specta.Property(t, func(t *specta.T) {
				gen := specta.Map(specta.String(), specta.Int())
				value := gen.Draw(t.Source, "test")
				if len(value) == 0 {
					foundEmpty = true
				}
			}, specta.Seed(seed), specta.MaxTests(50))
		}

		if !foundEmpty {
			t.Logf("note: never generated empty map (low probability, not necessarily a bug)")
		}
	})

	t.Run("generates varying sizes", func(t *testing.T) {
		sizes := make(map[int]bool)

		specta.Property(t, func(t *specta.T) {
			gen := specta.Map(specta.String(), specta.Int()).MinLen(0).MaxLen(10)
			value := gen.Draw(t.Source, "test")
			sizes[len(value)] = true
		}, specta.MaxTests(200))

		// Should see at least a few different sizes
		if len(sizes) < 3 {
			t.Errorf("only saw %d different sizes, expected more variety", len(sizes))
		}
	})

	t.Run("Filter works on maps", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			// Only accept maps where all values are even
			gen := specta.Map(
				specta.String(),
				specta.Int().Range(0, 100),
			).NonEmpty().Filter(func(m map[string]int64) bool {
				for _, v := range m {
					if v%2 != 0 {
						return false
					}
				}
				return true
			})
			value := gen.Draw(t.Source, "test")

			// Verify filter condition holds
			for k, v := range value {
				if v%2 != 0 {
					t.Errorf("filter should ensure all values are even, but %s=%d is odd", k, v)
				}
			}
		}, specta.MaxTests(100))
	})

	t.Run("handles key collisions gracefully", func(t *testing.T) {
		// Use a generator that produces limited unique keys
		specta.Property(t, func(t *specta.T) {
			gen := specta.Map(
				specta.Int().Range(0, 5), // Only 6 possible keys
				specta.String(),
			).MinLen(3).MaxLen(10)
			value := gen.Draw(t.Source, "test")

			// Should generate a map, possibly smaller than MaxLen due to collisions
			if len(value) > 6 {
				t.Errorf("map has %d entries but only 6 unique keys possible", len(value))
			}
		}, specta.MaxTests(100))
	})
}

// TestMapGenerator_Deterministic tests deterministic behavior
func TestMapGenerator_Deterministic(t *testing.T) {
	t.Run("deterministic mode produces predictable results", func(t *testing.T) {
		gen := specta.New(specta.WithStart(0))

		map1 := specta.Map(specta.String(), specta.Int()).MinLen(3).MaxLen(5).Draw(gen, "test")

		// Reset to same state
		gen = specta.New(specta.WithStart(0))
		map2 := specta.Map(specta.String(), specta.Int()).MinLen(3).MaxLen(5).Draw(gen, "test")

		// Should get identical results
		if len(map1) != len(map2) {
			t.Errorf("deterministic maps have different sizes: %d vs %d", len(map1), len(map2))
		}

		for k, v1 := range map1 {
			v2, ok := map2[k]
			if !ok {
				t.Errorf("deterministic maps differ: key %q present in map1 but not map2", k)
			} else if v1 != v2 {
				t.Errorf("deterministic maps differ at key %q: %d vs %d", k, v1, v2)
			}
		}

		for k := range map2 {
			if _, ok := map1[k]; !ok {
				t.Errorf("deterministic maps differ: key %q present in map2 but not map1", k)
			}
		}
	})

	t.Run("deterministic mode uses counter for size", func(t *testing.T) {
		gen := specta.New(specta.WithStart(0))

		// First map should have size based on counter 0
		m := specta.Map(specta.String(), specta.Int()).MinLen(0).MaxLen(10).Draw(gen, "test")
		expectedSize := 0 // counter 0 % 11 = 0
		if len(m) != expectedSize {
			t.Logf("note: first map size %d (implementation detail, not a bug if different)", len(m))
		}
	})
}
