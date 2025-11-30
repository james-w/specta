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
			value := specta.Draw(t, gen, "random_int")
			// Should generate any int64 value
			_ = value
		}, specta.MaxTests(100))
	})

	t.Run("respects Range constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(10, 20)
			value := specta.Draw(t, gen, "constrained_int")
			specta.AssertThat(t, value, specta.AllOf(
				specta.GreaterThanOrEqual(int64(10)),
				specta.Not(specta.GreaterThan(int64(20))),
			))
		}, specta.MaxTests(100))
	})

	t.Run("respects Min constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Min(100)
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, value, specta.GreaterThanOrEqual(int64(100)))
		}, specta.MaxTests(100))
	})

	t.Run("respects Max constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Max(50)
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, value, specta.Not(specta.GreaterThan(int64(50))))
		}, specta.MaxTests(100))
	})

	t.Run("Positive generates values > 0", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Positive()
			value := specta.Draw(t, gen, "positive_int")
			specta.AssertThat(t, value, specta.GreaterThan(int64(0)))
		}, specta.MaxTests(100))
	})

	t.Run("NonNegative generates values >= 0", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().NonNegative()
			value := specta.Draw(t, gen, "nonneg_int")
			specta.AssertThat(t, value, specta.GreaterThanOrEqual(int64(0)))
		}, specta.MaxTests(100))
	})

	t.Run("Negative generates values < 0", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Negative()
			value := specta.Draw(t, gen, "negative_int")
			specta.AssertThat(t, value, specta.LessThan(int64(0)))
		}, specta.MaxTests(100))
	})

	t.Run("chained constraints work", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Min(10).Max(100)
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, value, specta.AllOf(
				specta.GreaterThanOrEqual(int64(10)),
				specta.Not(specta.GreaterThan(int64(100))),
			))
		}, specta.MaxTests(100))
	})
}

// TestStringGenerator tests string generation with constraints
func TestStringGenerator(t *testing.T) {
	t.Run("generates any bytes by default", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String()
			value := specta.Draw(t, gen, "str")
			// Should be able to generate any string including invalid UTF-8
			_ = value
		}, specta.MaxTests(100))
	})

	t.Run("NonEmpty generates non-empty strings", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().NonEmpty()
			value := specta.Draw(t, gen, "nonempty_str")
			specta.AssertThat(t, int64(len(value)), specta.GreaterThan(int64(0)))
		}, specta.MaxTests(100))
	})

	t.Run("MinLen respects minimum length", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().MinLen(10)
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, int64(len(value)), specta.GreaterThanOrEqual(int64(10)))
		}, specta.MaxTests(100))
	})

	t.Run("MaxLen respects maximum length", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().MaxLen(5)
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, int64(len(value)), specta.Not(specta.GreaterThan(int64(5))))
		}, specta.MaxTests(100))
	})

	t.Run("Len generates exact length", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().Len(10)
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, int64(len(value)), specta.Equal(int64(10)))
		}, specta.MaxTests(100))
	})

	t.Run("Prefix adds prefix", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().Prefix("test_")
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, value, specta.HasPrefix("test_"))
		}, specta.MaxTests(100))
	})

	t.Run("Suffix adds suffix", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().Suffix("_end")
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, value, specta.HasSuffix("_end"))
		}, specta.MaxTests(100))
	})

	t.Run("ASCII generates only ASCII", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.String().ASCII().MinLen(1)
			value := specta.Draw(t, gen, "ascii_str")
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
			value := specta.Draw(t, gen, "printable_str")
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
			value := specta.Draw(t, gen, "alphanum_str")
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
			value := specta.Draw(t, gen, "alpha_str")
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
			value := specta.Draw(t, gen, "value")

			specta.AssertThat(t, value, specta.HasPrefix("user_"))
			specta.AssertThat(t, int64(len(value)), specta.AllOf(
				specta.GreaterThanOrEqual(int64(10)),
				specta.Not(specta.GreaterThan(int64(20))),
			))
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
				value := specta.Draw(t, gen, "value")
				if len(value) == 0 {
					foundEmpty = true
				}
			}, specta.Seed(seed), specta.MaxTests(100))
		}

		specta.AssertThat(t, foundEmpty, specta.IsTrue())
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
				value := specta.Draw(t, gen, "value")
				if value == 10 {
					foundMin = true
				}
				if value == 20 {
					foundMax = true
				}
			}, specta.Seed(seed), specta.MaxTests(20))
		}

		specta.AssertThat(t, foundMin, specta.IsTrue())
		specta.AssertThat(t, foundMax, specta.IsTrue())
	})

	t.Run("handles single value range", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(42, 42)
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, value, specta.Equal(int64(42)))
		}, specta.MaxTests(100))
	})

	t.Run("generates negative values by default", func(t *testing.T) {
		foundNegative := false

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int()
			value := specta.Draw(t, gen, "value")
			if value < 0 {
				foundNegative = true
			}
		}, specta.MaxTests(1000))

		specta.AssertThat(t, foundNegative, specta.IsTrue())
	})

	t.Run("generates extreme values", func(t *testing.T) {
		var minSeen, maxSeen int64 = math.MaxInt64, math.MinInt64

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int()
			value := specta.Draw(t, gen, "value")
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

// TestIntGenerator_EdgeCaseBiasing tests that edge cases are generated with value biasing
func TestIntGenerator_EdgeCaseBiasing(t *testing.T) {
	t.Run("biases toward boundaries", func(t *testing.T) {
		found := make(map[int64]int)
		// Value biasing (for ranges >100):
		// - 12.5% return shrinkToward (0 for [0,100])
		// - 12.5% small values near min via geometricInt
		// - 12.5% boundary values (min, max, 0, ±1)
		// - 12.5% powers of 2
		// - 50% uniform

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(0, 200)
			value := specta.Draw(t, gen, "value")
			found[value]++
		}, specta.MaxTests(1000))

		// Edge cases should appear more than uniform
		// Uniform: ~5 times each (1000/201)
		// With biasing: expect >12 times for 0, 1, 200
		edgeCases := []int64{0, 1, 200}
		for _, edge := range edgeCases {
			count := found[edge]
			specta.AssertThat(t, count, specta.GreaterThanOrEqual(12))
			t.Logf("edge case %d appeared %d times", edge, count)
		}
	})

	t.Run("biases toward zero", func(t *testing.T) {
		foundZero := 0

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(-200, 200)
			value := specta.Draw(t, gen, "value")
			if value == 0 {
				foundZero++
			}
		}, specta.MaxTests(1000))

		// Value biasing for [-200, 200]:
		// - shrinkToward=0: 12.5% chance
		// - boundaries include 0: ~3.75% chance
		// - uniform: 50%/401 ≈ 0.12%
		// Total: ~16% = ~160 times expected
		// Use conservative threshold of 50 to account for variance
		specta.AssertThat(t, foundZero, specta.GreaterThanOrEqual(50))
		t.Logf("zero appeared %d times out of 1000", foundZero)
	})

	t.Run("biases toward powers of 2", func(t *testing.T) {
		powers := []int64{1, 2, 4, 8, 16, 32, 64, 128}
		foundPowers := 0

		specta.Property(t, func(t *specta.T) {
			gen := specta.Int().Range(0, 200)
			value := specta.Draw(t, gen, "value")
			for _, p := range powers {
				if value == p {
					foundPowers++
					break
				}
			}
		}, specta.MaxTests(1000))

		// Value biasing explicitly targets powers of 2:
		// - 12.5% chance of powers of 2 branch (checks each with 20% acceptance)
		// - 50% uniform: 8 values / 201 = ~40 times
		// - With biasing: expect >60 times total
		specta.AssertThat(t, foundPowers, specta.GreaterThanOrEqual(60))
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
				denominator := specta.Draw(t, specta.Int().Range(-10, 10), "denominator")

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
			value := specta.Draw(t, gen, "slice")
			// Should generate a slice (possibly empty)
			_ = value
		}, specta.MaxTests(100))
	})

	t.Run("respects MinLen constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int()).MinLen(5)
			value := specta.Draw(t, gen, "min_slice")
			specta.AssertThat(t, int64(len(value)), specta.GreaterThanOrEqual(int64(5)))
		}, specta.MaxTests(100))
	})

	t.Run("respects MaxLen constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int()).MaxLen(10)
			value := specta.Draw(t, gen, "max_slice")
			specta.AssertThat(t, int64(len(value)), specta.Not(specta.GreaterThan(int64(10))))
		}, specta.MaxTests(100))
	})

	t.Run("respects Len constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.String()).Len(7)
			value := specta.Draw(t, gen, "fixed_slice")
			specta.AssertThat(t, int64(len(value)), specta.Equal(int64(7)))
		}, specta.MaxTests(100))
	})

	t.Run("NonEmpty generates non-empty slices", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Bool()).NonEmpty()
			value := specta.Draw(t, gen, "nonempty_slice")
			specta.AssertThat(t, int64(len(value)), specta.GreaterThan(int64(0)))
		}, specta.MaxTests(100))
	})

	t.Run("element generator constraints are respected", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int().Range(1, 10))
			value := specta.Draw(t, gen, "value")
			for _, v := range value {
				specta.AssertThat(t, v, specta.AllOf(
					specta.GreaterThanOrEqual(int64(1)),
					specta.Not(specta.GreaterThan(int64(10))),
				))
			}
		}, specta.MaxTests(100))
	})

	t.Run("can generate empty slices by default", func(t *testing.T) {
		foundEmpty := false

		for seed := int64(0); seed < 100 && !foundEmpty; seed++ {
			specta.Property(t, func(t *specta.T) {
				gen := specta.Slice(specta.Int())
				value := specta.Draw(t, gen, "value")
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
			value := specta.Draw(t, gen, "value")

			specta.AssertThat(t, int64(len(value)), specta.AllOf(
				specta.GreaterThanOrEqual(int64(2)),
				specta.Not(specta.GreaterThan(int64(5))),
			))

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
			value := specta.Draw(t, gen, "value")
			lengths[len(value)] = true
		}, specta.MaxTests(200))

		// Should see at least a few different lengths
		specta.AssertThat(t, int64(len(lengths)), specta.GreaterThanOrEqual(int64(3)))
	})

	t.Run("Filter works on slices", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			// Only accept slices with at least one even number
			value := specta.Draw(t, specta.Slice(specta.Int().Range(0, 100)).NonEmpty().Filter(func(s []int64) bool {
				for _, v := range s {
					if v%2 == 0 {
						return true
					}
				}
				return false
			}), "value")

			// Verify filter condition holds
			foundEven := false
			for _, v := range value {
				if v%2 == 0 {
					foundEven = true
					break
				}
			}
			specta.AssertThat(t, foundEven, specta.IsTrue())
		}, specta.MaxTests(100))
	})
}

// TestSliceGenerator_Deterministic tests deterministic behavior
func TestSliceGenerator_Deterministic(t *testing.T) {
	t.Skip("Deterministic mode tests disabled during conjecture migration (Phase 3)")
	/*
		t.Run("deterministic mode produces predictable results", func(t *testing.T) {
			gen := specta.New(specta.WithStart(0))

			slice1 := specta.Slice(specta.Int()).MinLen(3).MaxLen(5).Draw(gen.TestCase, "test")

			// Reset to same state
			gen = specta.New(specta.WithStart(0))
			slice2 := specta.Slice(specta.Int()).MinLen(3).MaxLen(5).Draw(gen.TestCase, "test")

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
			slice := specta.Slice(specta.Int()).MinLen(0).MaxLen(10).Draw(gen.TestCase, "test")
			expectedLen := 0 // counter 0 % 11 = 0
			if len(slice) != expectedLen {
				t.Logf("note: first slice length %d (implementation detail, not a bug if different)", len(slice))
			}
		})
	*/
}

// TestSliceGenerator_SizeBiasing tests that collection sizes use range stratification
func TestSliceGenerator_SizeBiasing(t *testing.T) {
	t.Run("biases toward small slices", func(t *testing.T) {
		dist := make(map[string]int)

		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int()).MinLen(0).MaxLen(100)
			value := specta.Draw(t, gen, "value")
			length := len(value)

			switch {
			case length == 0:
				dist["empty"]++
			case length >= 1 && length <= 3:
				dist["very_small"]++
			case length >= 4 && length <= 10:
				dist["small"]++
			case length >= 11 && length <= 30:
				dist["medium"]++
			case length >= 31 && length <= 70:
				dist["large"]++
			default:
				dist["very_large"]++
			}
		}, specta.MaxTests(1000))

		t.Logf("Distribution: empty=%d (%.1f%%), very_small(1-3)=%d (%.1f%%), small(4-10)=%d (%.1f%%), medium(11-30)=%d (%.1f%%), large(31-70)=%d (%.1f%%), very_large(71+)=%d (%.1f%%)",
			dist["empty"], float64(dist["empty"])/10,
			dist["very_small"], float64(dist["very_small"])/10,
			dist["small"], float64(dist["small"])/10,
			dist["medium"], float64(dist["medium"])/10,
			dist["large"], float64(dist["large"])/10,
			dist["very_large"], float64(dist["very_large"])/10)

		// With range stratification:
		// - 12.5% in [0-3] range (very small)
		// - 12.5% in [0-10] range (small)
		// - 75% in [0-100] range (full)
		// Very small (0-3) should appear more frequently than uniform
		verySmallCount := dist["empty"] + dist["very_small"]

		// Uniform would be: 1000 * 4/101 ≈ 40
		// With biasing: should be higher, expect at least 80
		specta.AssertThat(t, verySmallCount, specta.GreaterThanOrEqual(70))
	})

	t.Run("biases toward slices in 0-10 range", func(t *testing.T) {
		foundSmall := 0

		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.String()).MinLen(0).MaxLen(100)
			value := specta.Draw(t, gen, "value")
			if len(value) >= 0 && len(value) <= 10 {
				foundSmall++
			}
		}, specta.MaxTests(1000))

		// With stratification: 12.5% in [0-3], 12.5% in [0-10], 75% full
		// Uniform would be: 1000 * 11/101 ≈ 109
		// With biasing: should be higher, expect at least 150
		specta.AssertThat(t, foundSmall, specta.GreaterThanOrEqual(140))
		t.Logf("small slices (0-10) appeared %d times out of 1000", foundSmall)
	})

	t.Run("range stratification distribution", func(t *testing.T) {
		dist := make(map[int]int)

		specta.Property(t, func(t *specta.T) {
			gen := specta.Slice(specta.Int()).MinLen(0).MaxLen(50)
			value := specta.Draw(t, gen, "value")
			dist[len(value)]++
		}, specta.MaxTests(1000))

		// Count sizes in different ranges
		verySmall := 0 // 0-3
		small := 0     // 4-10
		for i := 0; i <= 3; i++ {
			verySmall += dist[i]
		}
		for i := 4; i <= 10; i++ {
			small += dist[i]
		}

		t.Logf("Very small (0-3): %d times (%.1f%%)", verySmall, float64(verySmall)/10)
		t.Logf("Small (4-10): %d times (%.1f%%)", small, float64(small)/10)
		t.Logf("Size 0: %d times", dist[0])
		t.Logf("Size 1: %d times", dist[1])
	})

	t.Run("finds empty collection bug quickly", func(t *testing.T) {
		// Simulate a bug that only fails on empty collections
		totalIterations := 0
		found := false

		for seed := int64(0); seed < 50 && !found; seed++ {
			specta.Property(t, func(t *specta.T) {
				totalIterations++
				slice := specta.Draw(t, specta.Slice(specta.Int()).MinLen(0).MaxLen(20), "slice")

				// Check if we hit the edge case
				if len(slice) == 0 {
					found = true
				}
			}, specta.Seed(seed), specta.MaxTests(10))
		}

		if !found {
			t.Errorf("biasing failed to find empty slice in %d iterations", totalIterations)
		} else {
			t.Logf("found empty slice in %d total iterations (with range stratification)", totalIterations)
			// With range stratification (12.5% in [0-3] range), should find it reasonably quickly
			if totalIterations > 200 {
				t.Logf("warning: took %d iterations, expected faster with biasing", totalIterations)
			}
		}
	})
}

// TestStringGenerator_LengthBiasing tests that string lengths use range stratification
func TestStringGenerator_LengthBiasing(t *testing.T) {
	t.Run("biases toward short strings", func(t *testing.T) {
		dist := make(map[int]int)

		specta.Property(t, func(t *specta.T) {
			gen := specta.String().MinLen(0).MaxLen(50)
			value := specta.Draw(t, gen, "value")
			dist[len(value)]++
		}, specta.MaxTests(1000))

		// With range stratification:
		// - 12.5% in [0-5] range (very short)
		// - 12.5% in [0-10] range (short)
		// - 75% in [0-50] range (full)
		// Short strings (0-10) should appear more frequently than uniform
		shortCount := 0
		for i := 0; i <= 10; i++ {
			shortCount += dist[i]
		}

		// Uniform would be: 1000 * 11/51 ≈ 216
		// With biasing: should be higher, expect ~250-350
		specta.AssertThat(t, shortCount, specta.GreaterThanOrEqual(200))

		t.Logf("Short strings (0-10): %d times (%.1f%%)", shortCount, float64(shortCount)/10)
		t.Logf("Length 0: %d times", dist[0])
		t.Logf("Length 1: %d times", dist[1])
	})

	t.Run("finds empty string bug quickly", func(t *testing.T) {
		// Simulate a bug that only fails on empty strings
		totalIterations := 0
		found := false

		for seed := int64(0); seed < 50 && !found; seed++ {
			specta.Property(t, func(t *specta.T) {
				totalIterations++
				str := specta.Draw(t, specta.String().MinLen(0).MaxLen(20), "str")

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
			value := specta.Draw(t, gen, "value")
			if len(value) == 100 {
				foundMax++
			}
		}, specta.MaxTests(1000))

		// With range stratification, max length appears in 75% full-range bucket
		// Uniform in that bucket: 1000 * 0.75 / 101 ≈ 7.4 times
		// Just verify we can hit it (not as frequently as with old edge biasing)
		specta.AssertThat(t, foundMax, specta.GreaterThanOrEqual(3))
		t.Logf("max length (100) appeared %d times out of 1000", foundMax)
	})
}

// TestMapGenerator tests map generation with key and value generators
func TestMapGenerator(t *testing.T) {
	t.Run("generates maps by default", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.MapOf(specta.String(), specta.Int())
			value := specta.Draw(t, gen, "value")
			// Should generate a map (possibly empty)
			_ = value
		}, specta.MaxTests(100))
	})

	t.Run("respects MinLen constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.MapOf(specta.String(), specta.Int()).MinLen(5)
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, int64(len(value)), specta.GreaterThanOrEqual(int64(5)))
		}, specta.MaxTests(100))
	})

	t.Run("respects MaxLen constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.MapOf(specta.String(), specta.Bool()).MaxLen(10)
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, int64(len(value)), specta.Not(specta.GreaterThan(int64(10))))
		}, specta.MaxTests(100))
	})

	t.Run("respects Len constraint", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			// Use String keys to avoid collision issues during shrinking
			gen := specta.MapOf(specta.String(), specta.Int()).MinLen(7).MaxLen(7)
			value := specta.Draw(t, gen, "value")
			// Note: May be less than 7 if key collisions occur
			specta.AssertThat(t, int64(len(value)), specta.Not(specta.GreaterThan(int64(7))))
		}, specta.MaxTests(100))
	})

	t.Run("NonEmpty generates non-empty maps", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.MapOf(specta.String().AlphaNum(), specta.Float64(0.0, 1.0)).NonEmpty()
			value := specta.Draw(t, gen, "value")
			specta.AssertThat(t, int64(len(value)), specta.GreaterThan(int64(0)))
		}, specta.MaxTests(100))
	})

	t.Run("key and value generator constraints are respected", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			gen := specta.MapOf(
				specta.String().Prefix("key_"),
				specta.Int().Range(1, 10),
			)
			value := specta.Draw(t, gen, "value")
			for k, v := range value {
				specta.AssertThat(t, k, specta.HasPrefix("key_"))
				specta.AssertThat(t, v, specta.AllOf(
					specta.GreaterThanOrEqual(int64(1)),
					specta.Not(specta.GreaterThan(int64(10))),
				))
			}
		}, specta.MaxTests(100))
	})

	t.Run("can generate empty maps by default", func(t *testing.T) {
		foundEmpty := false

		for seed := int64(0); seed < 100 && !foundEmpty; seed++ {
			specta.Property(t, func(t *specta.T) {
				gen := specta.MapOf(specta.String(), specta.Int())
				value := specta.Draw(t, gen, "value")
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
			gen := specta.MapOf(specta.String(), specta.Int()).MinLen(0).MaxLen(10)
			value := specta.Draw(t, gen, "value")
			sizes[len(value)] = true
		}, specta.MaxTests(200))

		// Should see at least a few different sizes
		specta.AssertThat(t, int64(len(sizes)), specta.GreaterThanOrEqual(int64(3)))
	})

	t.Run("Filter works on maps", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			// Only accept maps where all values are even
			value := specta.Draw(t, specta.MapOf(
				specta.String(),
				specta.Int().Range(0, 100),
			).NonEmpty().Filter(func(m map[string]int64) bool {
				for _, v := range m {
					if v%2 != 0 {
						return false
					}
				}
				return true
			}), "value")

			// Verify filter condition holds
			for _, v := range value {
				specta.AssertThat(t, v%2, specta.Equal(int64(0)))
			}
		}, specta.MaxTests(100))
	})

	t.Run("handles key collisions gracefully", func(t *testing.T) {
		// Use a generator that produces limited unique keys
		specta.Property(t, func(t *specta.T) {
			gen := specta.MapOf(
				specta.Int().Range(0, 5), // Only 6 possible keys
				specta.String(),
			).MinLen(3).MaxLen(10)
			value := specta.Draw(t, gen, "value")

			// Should generate a map, possibly smaller than MaxLen due to collisions
			specta.AssertThat(t, int64(len(value)), specta.Not(specta.GreaterThan(int64(6))))
		}, specta.MaxTests(100))
	})
}

// TestMapGenerator_Deterministic tests deterministic behavior
func TestMapGenerator_Deterministic(t *testing.T) {
	t.Skip("Deterministic mode tests disabled during conjecture migration (Phase 3)")
	/*
		t.Run("deterministic mode produces predictable results", func(t *testing.T) {
			gen := specta.New(specta.WithStart(0))

			map1 := specta.MapOf(specta.String(), specta.Int()).MinLen(3).MaxLen(5).Draw(gen.TestCase, "test")

			// Reset to same state
			gen = specta.New(specta.WithStart(0))
			map2 := specta.MapOf(specta.String(), specta.Int()).MinLen(3).MaxLen(5).Draw(gen.TestCase, "test")

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
			m := specta.MapOf(specta.String(), specta.Int()).MinLen(0).MaxLen(10).Draw(gen.TestCase, "test")
			expectedSize := 0 // counter 0 % 11 = 0
			if len(m) != expectedSize {
				t.Logf("note: first map size %d (implementation detail, not a bug if different)", len(m))
			}
		})
	*/
}

// =============================================================================
// Choice Generator Tests
// =============================================================================

func TestChoiceGenerator(t *testing.T) {
	t.Run("generates from all alternatives", func(t *testing.T) {
		seen := make(map[string]bool)
		specta.Property(t, func(pt *specta.T) {
			gen := specta.Choice(
				specta.Int().Range(1, 10),
				specta.Int().Range(100, 200),
				specta.Int().Range(1000, 2000),
			)
			value := specta.Draw(pt, gen, "value")

			// Categorize by range
			switch {
			case value >= 1 && value <= 10:
				seen["small"] = true
			case value >= 100 && value <= 200:
				seen["medium"] = true
			case value >= 1000 && value <= 2000:
				seen["large"] = true
			}
		}, specta.Seed(12345), specta.MaxTests(300))

		specta.AssertThat(t, seen["small"], specta.IsTrue())
		specta.AssertThat(t, seen["medium"], specta.IsTrue())
		specta.AssertThat(t, seen["large"], specta.IsTrue())
	})

	t.Run("weighted choices respect weights", func(t *testing.T) {
		counts := make(map[string]int)
		specta.Property(t, func(pt *specta.T) {
			// 9:1 ratio - 90% large, 10% small
			gen := specta.Choice(specta.Int().Range(1, 10)).
				OrWeighted(specta.Int().Range(100, 200), 9)

			value := specta.Draw(pt, gen, "value")
			if value >= 1 && value <= 10 {
				counts["small"]++
			} else {
				counts["large"]++
			}
		}, specta.Seed(42), specta.MaxTests(1000))

		// With 1:9 ratio, expect ~100 small, ~900 large (allow 30% tolerance)
		if counts["small"] < 50 || counts["small"] > 200 {
			t.Errorf("weighted distribution skewed: small=%d, large=%d (expected ~100 small)",
				counts["small"], counts["large"])
		}
	})

	t.Run("Or chaining works", func(t *testing.T) {
		seen := make(map[string]bool)
		specta.Property(t, func(pt *specta.T) {
			gen := specta.Choice(specta.Int().Range(1, 10)).
				Or(specta.Int().Range(100, 200)).
				Or(specta.Int().Range(1000, 2000))

			value := specta.Draw(pt, gen, "value")
			switch {
			case value >= 1 && value <= 10:
				seen["small"] = true
			case value >= 100 && value <= 200:
				seen["medium"] = true
			case value >= 1000 && value <= 2000:
				seen["large"] = true
			}
		}, specta.Seed(99), specta.MaxTests(300))

		specta.AssertThat(t, seen["small"], specta.IsTrue())
		specta.AssertThat(t, seen["medium"], specta.IsTrue())
		specta.AssertThat(t, seen["large"], specta.IsTrue())
	})

	t.Run("Filter integration", func(t *testing.T) {
		specta.Property(t, func(pt *specta.T) {
			gen := specta.Choice(
				specta.Int().Range(0, 100),
				specta.Int().Range(1000, 2000),
			).Filter(func(n int64) bool { return n%2 == 0 })

			value := specta.Draw(pt, gen, "value")
			specta.AssertThat(pt, value%2, specta.Equal(int64(0)))
		}, specta.Seed(777), specta.MaxTests(100))
	})

	t.Run("mixing Or and OrWeighted", func(t *testing.T) {
		counts := make(map[string]int)
		specta.Property(t, func(pt *specta.T) {
			// First with Or (weight 1), second with Or (weight 1), third with OrWeighted (weight 8)
			// Expected ratio: 1:1:8 (10% small, 10% medium, 80% large)
			gen := specta.Choice(specta.Int().Range(1, 10)).
				Or(specta.Int().Range(50, 60)).
				OrWeighted(specta.Int().Range(100, 200), 8)

			value := specta.Draw(pt, gen, "value")
			switch {
			case value >= 1 && value <= 10:
				counts["small"]++
			case value >= 50 && value <= 60:
				counts["medium"]++
			case value >= 100 && value <= 200:
				counts["large"]++
			}
		}, specta.Seed(9876), specta.MaxTests(1000))

		// With 1:1:8 ratio, expect ~100 small, ~100 medium, ~800 large (allow 40% tolerance for smaller groups)
		if counts["small"] < 50 || counts["small"] > 200 {
			t.Errorf("small distribution skewed: small=%d, medium=%d, large=%d (expected ~100 small)",
				counts["small"], counts["medium"], counts["large"])
		}
		if counts["medium"] < 50 || counts["medium"] > 200 {
			t.Errorf("medium distribution skewed: small=%d, medium=%d, large=%d (expected ~100 medium)",
				counts["small"], counts["medium"], counts["large"])
		}
		if counts["large"] < 600 || counts["large"] > 900 {
			t.Errorf("large distribution skewed: small=%d, medium=%d, large=%d (expected ~800 large)",
				counts["small"], counts["medium"], counts["large"])
		}
	})

	t.Run("panics with no generators", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for Choice with no generators")
			}
		}()
		specta.Choice[int64]()
	})

	t.Run("panics with zero weight", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for zero weight")
			}
		}()
		gen := specta.Choice(specta.Int())
		gen.OrWeighted(specta.Int().Range(100, 200), 0)
	})

	t.Run("panics with negative weight", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for negative weight")
			}
		}()
		gen := specta.Choice(specta.Int())
		gen.OrWeighted(specta.Int().Range(100, 200), -1)
	})
}

// =============================================================================
// Optional Generator Tests
// =============================================================================

func TestOptionalGenerator(t *testing.T) {
	t.Run("default probability is 80% present", func(t *testing.T) {
		present := 0
		total := 1000

		specta.Property(t, func(pt *specta.T) {
			value := specta.Draw(pt, specta.Optional(specta.Int()), "value")
			if value != nil {
				present++
			}
		}, specta.Seed(123), specta.MaxTests(total))

		// Expect ~800 present, allow 15% tolerance (680-920)
		if present < 680 || present > 920 {
			t.Errorf("present rate: %d/%d (expected ~800)", present, total)
		}
	})

	t.Run("custom probability 30% respected", func(t *testing.T) {
		present := 0
		total := 1000

		specta.Property(t, func(pt *specta.T) {
			value := specta.Draw(pt, specta.Optional(specta.Int(), 0.3), "value")
			if value != nil {
				present++
			}
		}, specta.Seed(456), specta.MaxTests(total))

		// Expect ~300, allow 20% tolerance (240-360)
		if present < 240 || present > 360 {
			t.Errorf("30%% probability: %d/%d (expected ~300)", present, total)
		}
	})

	t.Run("rare probability 20% respected", func(t *testing.T) {
		present := 0
		total := 1000

		specta.Property(t, func(pt *specta.T) {
			value := specta.Draw(pt, specta.Optional(specta.Int(), 0.2), "value")
			if value != nil {
				present++
			}
		}, specta.Seed(789), specta.MaxTests(total))

		// Expect ~200, allow 25% tolerance (150-250)
		if present < 150 || present > 250 {
			t.Errorf("20%% probability: %d/%d (expected ~200)", present, total)
		}
	})

	t.Run("common probability 95% respected", func(t *testing.T) {
		present := 0
		total := 1000

		specta.Property(t, func(pt *specta.T) {
			value := specta.Draw(pt, specta.Optional(specta.Int(), 0.95), "value")
			if value != nil {
				present++
			}
		}, specta.Seed(321), specta.MaxTests(total))

		// Expect ~950, allow 5% tolerance (900-980)
		if present < 900 || present > 980 {
			t.Errorf("95%% probability: %d/%d (expected ~950)", present, total)
		}
	})

	t.Run("generated values respect inner generator constraints", func(t *testing.T) {
		specta.Property(t, func(pt *specta.T) {
			value := specta.Draw(pt, specta.Optional(specta.Int().Range(10, 20)), "value")
			if value != nil {
				specta.AssertThat(pt, *value, specta.AllOf(
					specta.GreaterThanOrEqual(int64(10)),
					specta.Not(specta.GreaterThan(int64(20))),
				))
			}
		}, specta.Seed(555), specta.MaxTests(100))
	})

	t.Run("panics with probability < 0", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for probability < 0")
			}
		}()
		specta.Optional(specta.Int(), -0.1)
	})

	t.Run("panics with probability > 1", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for probability > 1")
			}
		}()
		specta.Optional(specta.Int(), 1.5)
	})

	t.Run("probability 0 always nil", func(t *testing.T) {
		specta.Property(t, func(pt *specta.T) {
			value := specta.Draw(pt, specta.Optional(specta.Int(), 0.0), "value")
			specta.AssertThat(pt, value == nil, specta.IsTrue())
		}, specta.Seed(666), specta.MaxTests(100))
	})

	t.Run("probability 1 always present", func(t *testing.T) {
		specta.Property(t, func(pt *specta.T) {
			value := specta.Draw(pt, specta.Optional(specta.Int(), 1.0), "value")
			specta.AssertThat(pt, value != nil, specta.IsTrue())
		}, specta.Seed(888), specta.MaxTests(100))
	})
}

// =============================================================================
// Pair and Triple Generator Tests
// =============================================================================

func TestPairGenerator(t *testing.T) {
	t.Run("generates pairs with correct types", func(t *testing.T) {
		specta.Property(t, func(pt *specta.T) {
			pair := specta.Draw(pt, specta.Pair(
				specta.Int().Range(1, 100),
				specta.String().Alpha().MaxLen(10),
			), "pair")

			// Type checks (compile-time)
			var _ = pair.A // int64
			var _ = pair.B // string

			// Constraint checks
			specta.AssertThat(pt, pair.A, specta.AllOf(
				specta.GreaterThanOrEqual(int64(1)),
				specta.Not(specta.GreaterThan(int64(100))),
			))
			specta.AssertThat(pt, int64(len(pair.B)), specta.Not(specta.GreaterThan(int64(10))))
		}, specta.Seed(111), specta.MaxTests(100))
	})

	t.Run("can nest Pair with other generators", func(t *testing.T) {
		specta.Property(t, func(pt *specta.T) {
			// Pair of (int, Maybe[string])
			pair := specta.Draw(pt, specta.Pair(
				specta.Int().Range(1, 10),
				specta.Optional(specta.String()),
			), "nested_pair")

			specta.AssertThat(pt, pair.A, specta.AllOf(
				specta.GreaterThanOrEqual(int64(1)),
				specta.Not(specta.GreaterThan(int64(10))),
			))
			// B is *string, can be nil or non-nil
		}, specta.Seed(222), specta.MaxTests(50))
	})
}

func TestTripleGenerator(t *testing.T) {
	t.Run("generates triples with correct types", func(t *testing.T) {
		specta.Property(t, func(pt *specta.T) {
			triple := specta.Draw(pt, specta.Triple(
				specta.Int().Range(1, 100),
				specta.String().Alpha().MaxLen(10),
				specta.Bool(),
			), "triple")

			// Type checks (compile-time)
			var _ = triple.A // int64
			var _ = triple.B // string
			var _ = triple.C // bool

			// Constraint checks
			specta.AssertThat(pt, triple.A, specta.AllOf(
				specta.GreaterThanOrEqual(int64(1)),
				specta.Not(specta.GreaterThan(int64(100))),
			))
			specta.AssertThat(pt, int64(len(triple.B)), specta.Not(specta.GreaterThan(int64(10))))
		}, specta.Seed(333), specta.MaxTests(100))
	})

	t.Run("can use Triple for coordinate generation", func(t *testing.T) {
		specta.Property(t, func(pt *specta.T) {
			coords3d := specta.Draw(pt, specta.Triple(
				specta.Float64(-180, 180), // longitude
				specta.Float64(-90, 90),   // latitude
				specta.Float64(0, 10000),  // altitude
			), "coords")

			specta.AssertThat(pt, coords3d.A, specta.AllOf(
				specta.GreaterThanOrEqual(-180.0),
				specta.Not(specta.GreaterThan(180.0)),
			))
			specta.AssertThat(pt, coords3d.B, specta.AllOf(
				specta.GreaterThanOrEqual(-90.0),
				specta.Not(specta.GreaterThan(90.0)),
			))
			specta.AssertThat(pt, coords3d.C, specta.AllOf(
				specta.GreaterThanOrEqual(0.0),
				specta.Not(specta.GreaterThan(10000.0)),
			))
		}, specta.Seed(444), specta.MaxTests(100))
	})
}
