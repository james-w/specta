package specta_test

import (
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/testlib"
)

// TestPropertyIntegration_SimplePrimitives tests that PropertyPrimitives works with basic types
func TestPropertyIntegration_SimplePrimitives(t *testing.T) {
	t.Run("can use Primitives in property test", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			p := t.Primitives()

			// Generate various primitives
			id := p.ID()
			str := p.String()
			num := p.Int()

			// Properties should hold
			if id == "" {
				t.Errorf("ID should not be empty")
			}
			if str == "" {
				t.Errorf("String should not be empty")
			}
			// num can be any value, just check it was generated
			_ = num
		}, specta.MaxTests(50))

		if len(spy.Errors) > 0 {
			t.Errorf("property should pass: %v", spy.Errors)
		}
	})

	t.Run("Primitives are deterministic with same seed", func(t *testing.T) {
		// Run property test twice with same seed
		var ids1, ids2 []string

		specta.Property(t, func(t *specta.T) {
			p := t.Primitives()
			ids1 = append(ids1, p.ID())
		}, specta.Seed(12345), specta.MaxTests(5))

		specta.Property(t, func(t *specta.T) {
			p := t.Primitives()
			ids2 = append(ids2, p.ID())
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

// TestPropertyIntegration_WithOptions tests that Primitives options work
func TestPropertyIntegration_WithOptions(t *testing.T) {
	t.Run("WithPropertyPrefix works", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			p := t.Primitives(specta.WithPropertyPrefix("test"))
			str := p.String()

			if str[0:5] != "test_" {
				t.Errorf("expected string to start with 'test_', got %s", str)
			}
		}, specta.MaxTests(10))

		if len(spy.Errors) > 0 {
			t.Errorf("property should pass: %v", spy.Errors)
		}
	})
}

// TestPropertyIntegration_UniqueValues tests that counter-based values are unique within an iteration
func TestPropertyIntegration_UniqueValues(t *testing.T) {
	t.Run("IDs are unique within a single property iteration", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			p := t.Primitives()

			// Within a single iteration, IDs should be unique
			seenIDs := make(map[string]bool)
			for i := 0; i < 10; i++ {
				id := p.ID()
				if seenIDs[id] {
					t.Errorf("duplicate ID generated in same iteration: %s", id)
				}
				seenIDs[id] = true
			}
		}, specta.MaxTests(50))

		if len(spy.Errors) > 0 {
			t.Errorf("property should pass: %v", spy.Errors)
		}
	})
}

// TestPropertyIntegration_PropertyExample shows a realistic property test
func TestPropertyIntegration_PropertyExample(t *testing.T) {
	t.Run("commutative property with random values", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			p := t.Primitives()

			// Generate random integers using Primitives
			a := p.IntN(1000)
			b := p.IntN(1000)

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
			p := t.Primitives()

			s1 := p.StringWith("prefix")
			s2 := p.StringWith("prefix")

			// Property: concatenation length equals sum of lengths
			combined := s1 + s2
			specta.AssertThat(t, len(combined), specta.Equal(len(s1)+len(s2)))
		}, specta.MaxTests(100))

		if len(spy.Errors) > 0 {
			t.Errorf("concatenation property should hold: %v", spy.Errors)
		}
	})
}
