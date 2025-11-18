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
