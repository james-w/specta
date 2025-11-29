package specta_test

import (
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/testlib"
)

// TestPropertyIntegration_SimpleGenerators tests that Generators work with basic types
func TestPropertyIntegration_SimpleGenerators(t *testing.T) {
	t.Run("can use Generators in property test", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			// Generate various values from full type space
			id, err := specta.String().Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			str, err := specta.String().Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			_, err = specta.Int().Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}

			// Test a property that should ALWAYS hold
			// String concatenation length property
			combined := id + str
			expectedLen := len(id) + len(str)
			if len(combined) != expectedLen {
				t.Errorf("concatenation length wrong: got %d, expected %d", len(combined), expectedLen)
			}
		}, specta.MaxTests(50))

		if len(spy.Errors) > 0 {
			t.Errorf("property should pass: %v", spy.Errors)
		}
	})

	t.Run("Generators are deterministic with same seed", func(t *testing.T) {
		// Run property test twice with same seed
		var ids1, ids2 []string

		specta.Property(t, func(t *specta.T) {
			id, err := specta.String().Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			ids1 = append(ids1, id)
		}, specta.Seed(12345), specta.MaxTests(5))

		specta.Property(t, func(t *specta.T) {
			id, err := specta.String().Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			ids2 = append(ids2, id)
		}, specta.Seed(12345), specta.MaxTests(5))

		if len(ids1) != len(ids2) {
			t.Errorf("different number of IDs generated")
		}

		for i := range ids1 {
			if ids1[i] != ids2[i] {
				t.Errorf("ID %d differs: %s vs %s", i, ids1[i], ids2[i])
			}
		}
	})
}

// TestPropertyIntegration_FullSpaceExploration tests that Generators explores full type space
func TestPropertyIntegration_FullSpaceExploration(t *testing.T) {
	t.Run("finds empty strings", func(t *testing.T) {
		foundEmpty := false

		// Run many iterations to find an empty string
		// Probability of empty = 1/101 per call, so need enough attempts
		// Run up to 100 seeds × 100 tests = 10,000 attempts max
		for seed := int64(0); seed < 100 && !foundEmpty; seed++ {
			specta.Property(t, func(t *specta.T) {
				str, err := specta.String().Draw(t.Data)
				if err != nil {
					t.Fatalf("generator failed: %v", err)
				}
				if str == "" {
					foundEmpty = true
				}
			}, specta.Seed(seed), specta.MaxTests(100))
		}

		if !foundEmpty {
			t.Error("Generators should eventually generate empty strings")
		}
	})

	t.Run("finds negative integers", func(t *testing.T) {
		foundNegative := false

		specta.Property(t, func(t *specta.T) {
			n, err := specta.Int().Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			if n < 0 {
				foundNegative = true
			}
		}, specta.MaxTests(100))

		if !foundNegative {
			t.Error("Generators should generate negative integers")
		}
	})
}

// TestPropertyIntegration_RandomValues tests that values are random
func TestPropertyIntegration_RandomValues(t *testing.T) {
	t.Run("IDs are random and likely unique", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			// Generate multiple IDs - they should be different (extremely unlikely to collide)
			// Use MinLen to avoid empty strings (which would collide)
			id1, err := specta.String().MinLen(8).Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			id2, err := specta.String().MinLen(8).Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}

			// With sufficiently long random strings, collision is extremely unlikely
			if id1 == id2 {
				t.Errorf("got duplicate random IDs (extremely unlikely): %s", id1)
			}
		}, specta.MaxTests(50))

		if len(spy.Errors) > 0 {
			t.Errorf("property should pass: %v", spy.Errors)
		}
	})

	t.Run("values are different across iterations", func(t *testing.T) {
		var firstID string

		// First iteration
		specta.Property(t, func(t *specta.T) {
			if firstID == "" {
				id, err := specta.String().Draw(t.Data)
				if err != nil {
					t.Fatalf("generator failed: %v", err)
				}
				firstID = id
			}
		}, specta.MaxTests(1), specta.Seed(12345))

		// Second iteration with different seed should produce different ID
		var secondID string
		specta.Property(t, func(t *specta.T) {
			if secondID == "" {
				id, err := specta.String().Draw(t.Data)
				if err != nil {
					t.Fatalf("generator failed: %v", err)
				}
				secondID = id
			}
		}, specta.MaxTests(1), specta.Seed(54321))

		if firstID == secondID {
			t.Error("expected different IDs from different seeds")
		}
	})
}

// TestPropertyIntegration_PropertyExample shows a realistic property test
func TestPropertyIntegration_PropertyExample(t *testing.T) {
	t.Run("commutative property with random values", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			// Generate random integers using Primitives
			a, err := specta.Int().Range(0, 1000-1).Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			b, err := specta.Int().Range(0, 1000-1).Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}

			// Property: addition is commutative
			specta.AssertThat(t, a+b, specta.Equal(b+a))
		}, specta.MaxTests(100))

		if len(spy.Errors) > 0 {
			t.Errorf("commutative property should always hold: %v", spy.Errors)
		}
	})

	t.Run("string concatenation property", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			s1, err := specta.String().Prefix("prefix").Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			s2, err := specta.String().Prefix("prefix").Draw(t.Data)
			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}

			// Property: concatenation length equals sum of lengths
			combined := s1 + s2
			specta.AssertThat(t, len(combined), specta.Equal(len(s1)+len(s2)))
		}, specta.MaxTests(100))

		if len(spy.Errors) > 0 {
			t.Errorf("concatenation property should hold: %v", spy.Errors)
		}
	})
}
