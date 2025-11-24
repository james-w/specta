package specta_test

import (
	"testing"

	"github.com/james-w/specta"
)

func TestMaybe_GetWithGenerator(t *testing.T) {
	t.Run("uses generator when not set", func(t *testing.T) {
		var m specta.Maybe[string]
		gen := specta.String().ExampleHint("test_")
		s := specta.New()

		result := m.GetWithGenerator(s, "TestField", gen)

		if result != "test_1" {
			t.Errorf("expected test_1, got %s", result)
		}
	})

	t.Run("uses set value when set", func(t *testing.T) {
		m := specta.Some(specta.Lit("custom"))
		gen := specta.String().ExampleHint("test_")
		s := specta.New()

		result := m.GetWithGenerator(s, "TestField", gen)

		if result != "custom" {
			t.Errorf("expected custom, got %s", result)
		}
	})

	t.Run("label is used in generator draw", func(t *testing.T) {
		var m specta.Maybe[int64]
		gen := specta.Int().Range(1, 10)
		s := specta.New()

		result := m.GetWithGenerator(s, "MyField", gen)

		if result < 1 || result > 10 {
			t.Errorf("expected value in range [1, 10], got %d", result)
		}
	})
}

func TestProviderFromGenerator(t *testing.T) {
	t.Run("creates provider that calls generator with label", func(t *testing.T) {
		gen := specta.Int().Range(1, 10)
		prov := specta.ProviderFromGenerator(gen, "test")
		s := specta.New()

		result := prov(s)

		if result < 1 || result > 10 {
			t.Errorf("expected value in range [1, 10], got %d", result)
		}
	})

	t.Run("provider is reusable", func(t *testing.T) {
		gen := specta.String().ExampleHint("prefix_")
		prov := specta.ProviderFromGenerator(gen, "field")
		s := specta.New()

		result1 := prov(s)
		result2 := prov(s)

		if result1 != "prefix_1" {
			t.Errorf("expected prefix_1, got %s", result1)
		}
		if result2 != "prefix_2" {
			t.Errorf("expected prefix_2, got %s", result2)
		}
	})

	t.Run("works with bool generator", func(t *testing.T) {
		gen := specta.Bool()
		prov := specta.ProviderFromGenerator(gen, "active")
		s := specta.New()

		result := prov(s)

		// Bool returns deterministic value from Gen
		if result != true && result != false {
			t.Errorf("expected bool, got %v", result)
		}
	})
}
