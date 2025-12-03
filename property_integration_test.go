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
			// Generate random strings to test concatenation property
			id := specta.Draw(t, specta.String(), "id")
			str := specta.Draw(t, specta.String(), "str")

			// Test a property that should ALWAYS hold
			// String concatenation length property
			combined := id + str
			expectedLen := len(id) + len(str)
			specta.AssertThat(t, len(combined), specta.Equal(expectedLen))
		}, specta.MaxTests(50))

		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
	})

	t.Run("Generators are deterministic with same seed", func(t *testing.T) {
		// Run property test twice with same seed
		var ids1, ids2 []string

		specta.Property(t, func(t *specta.T) {
			id := specta.Draw(t, specta.String(), "id")
			ids1 = append(ids1, id)
		}, specta.Seed(12345), specta.MaxTests(5))

		specta.Property(t, func(t *specta.T) {
			id := specta.Draw(t, specta.String(), "id")
			ids2 = append(ids2, id)
		}, specta.Seed(12345), specta.MaxTests(5))

		specta.AssertThat(t, len(ids1), specta.Equal(len(ids2)))
		for i := range ids1 {
			specta.AssertThat(t, ids1[i], specta.Equal(ids2[i]))
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
				str := specta.Draw(t, specta.String(), "str")
				if str == "" {
					foundEmpty = true
				}
			}, specta.Seed(seed), specta.MaxTests(100))
		}

		specta.AssertThat(t, foundEmpty, specta.IsTrue())
	})

	t.Run("finds negative integers", func(t *testing.T) {
		foundNegative := false

		specta.Property(t, func(t *specta.T) {
			n := specta.Draw(t, specta.Int(), "n")
			if n < 0 {
				foundNegative = true
			}
		}, specta.MaxTests(100))

		specta.AssertThat(t, foundNegative, specta.IsTrue())
	})
}

// TestPropertyIntegration_RandomValues tests that values are random
func TestPropertyIntegration_RandomValues(t *testing.T) {
	t.Run("IDs are random and likely unique", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			// Generate multiple IDs - they should be different (extremely unlikely to collide)
			// Use MinLen to avoid empty strings (which would collide)
			id1 := specta.Draw(t, specta.String().MinLen(8), "id1")
			id2 := specta.Draw(t, specta.String().MinLen(8), "id2")

			// With sufficiently long random strings, collision is extremely unlikely
			specta.AssertThat(t, id1, specta.Not(specta.Equal(id2)))
		}, specta.MaxTests(50))

		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
	})

	t.Run("values are different across iterations", func(t *testing.T) {
		var firstID string

		// First iteration
		specta.Property(t, func(t *specta.T) {
			if firstID == "" {
				id := specta.Draw(t, specta.String(), "id")
				firstID = id
			}
		}, specta.MaxTests(1), specta.Seed(12345))

		// Second iteration with different seed should produce different ID
		var secondID string
		specta.Property(t, func(t *specta.T) {
			if secondID == "" {
				id := specta.Draw(t, specta.String(), "id")
				secondID = id
			}
		}, specta.MaxTests(1), specta.Seed(54321))

		specta.AssertThat(t, firstID, specta.Not(specta.Equal(secondID)))
	})
}

// TestPropertyIntegration_PropertyExample shows a realistic property test
func TestPropertyIntegration_PropertyExample(t *testing.T) {
	t.Run("commutative property with random values", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			// Generate random integers using Primitives
			a := specta.Draw(t, specta.Int().Range(0, 1000-1), "a")
			b := specta.Draw(t, specta.Int().Range(0, 1000-1), "b")

			// Property: addition is commutative
			specta.AssertThat(t, a+b, specta.Equal(b+a))
		}, specta.MaxTests(100))

		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
	})

	t.Run("string concatenation property", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			s1 := specta.Draw(t, specta.String().Prefix("prefix"), "s1")
			s2 := specta.Draw(t, specta.String().Prefix("prefix"), "s2")

			// Property: concatenation length equals sum of lengths
			combined := s1 + s2
			specta.AssertThat(t, len(combined), specta.Equal(len(s1)+len(s2)))
		}, specta.MaxTests(100))

		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
	})
}
