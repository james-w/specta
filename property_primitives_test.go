package specta_test

import (
	"strings"
	"testing"
	"time"

	"github.com/james-w/specta"
)

// TestPropertyPrimitives_Next tests the counter functionality
func TestPropertyPrimitives_Next(t *testing.T) {
	t.Run("increments counter", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		n1 := p.Next()
		n2 := p.Next()
		n3 := p.Next()

		if n1 != 1 || n2 != 2 || n3 != 3 {
			t.Errorf("expected sequential counters 1,2,3 got %d,%d,%d", n1, n2, n3)
		}
	})
}

// TestPropertyPrimitives_Determinism tests that same seed produces same values
func TestPropertyPrimitives_Determinism(t *testing.T) {
	t.Run("same seed produces same random values", func(t *testing.T) {
		p1 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		p2 := specta.NewPropertyPrimitives(specta.NewSource(12345))

		if p1.Bool() != p2.Bool() {
			t.Error("Bool() not deterministic")
		}

		if p1.Int() != p2.Int() {
			t.Error("Int() not deterministic")
		}

		if p1.Float64() != p2.Float64() {
			t.Error("Float64() not deterministic")
		}
	})

	t.Run("different seeds produce different values", func(t *testing.T) {
		p1 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		p2 := specta.NewPropertyPrimitives(specta.NewSource(54321))

		// These could theoretically be equal, but it's extremely unlikely
		if p1.Int64() == p2.Int64() {
			t.Log("warning: same value from different seeds (unlikely but possible)")
		}
	})
}

// TestPropertyPrimitives_Bool tests boolean generation
func TestPropertyPrimitives_Bool(t *testing.T) {
	t.Run("generates booleans", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		// Generate many booleans and check we get both true and false
		var trueCount, falseCount int
		for i := 0; i < 100; i++ {
			if p.Bool() {
				trueCount++
			} else {
				falseCount++
			}
		}

		if trueCount == 0 {
			t.Error("never generated true")
		}
		if falseCount == 0 {
			t.Error("never generated false")
		}
	})
}

// TestPropertyPrimitives_Int tests integer generation
func TestPropertyPrimitives_Int(t *testing.T) {
	t.Run("generates integers", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		value := p.Int()
		_ = value // Just check it doesn't panic
	})

	t.Run("IntN respects max", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		for i := 0; i < 100; i++ {
			value := p.IntN(10)
			if value < 0 || value >= 10 {
				t.Errorf("IntN(10) returned %d, expected [0, 10)", value)
			}
		}
	})

	t.Run("IntN with zero or negative max", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		if p.IntN(0) != 0 {
			t.Error("IntN(0) should return 0")
		}
		if p.IntN(-5) != 0 {
			t.Error("IntN(-5) should return 0")
		}
	})
}

// TestPropertyPrimitives_Float64 tests float generation
func TestPropertyPrimitives_Float64(t *testing.T) {
	t.Run("generates floats in range [0, 1)", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		for i := 0; i < 100; i++ {
			value := p.Float64()
			if value < 0.0 || value >= 1.0 {
				t.Errorf("Float64() returned %f, expected [0.0, 1.0)", value)
			}
		}
	})
}

// TestPropertyPrimitives_String tests string generation
func TestPropertyPrimitives_String(t *testing.T) {
	t.Run("generates strings with counter", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		s1 := p.String()
		s2 := p.String()

		if s1 == s2 {
			t.Error("expected different strings")
		}

		if !strings.Contains(s1, "str_") {
			t.Errorf("expected string to contain 'str_', got %s", s1)
		}
	})

	t.Run("StringWith uses custom prefix", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		s := p.StringWith("user")

		if !strings.HasPrefix(s, "user_") {
			t.Errorf("expected string to start with 'user_', got %s", s)
		}
	})

	t.Run("respects WithPropertyPrefix option", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source, specta.WithPropertyPrefix("test"))

		s := p.String()

		if !strings.HasPrefix(s, "test_") {
			t.Errorf("expected string to start with 'test_', got %s", s)
		}
	})
}

// TestPropertyPrimitives_BytesN tests byte generation
func TestPropertyPrimitives_BytesN(t *testing.T) {
	t.Run("generates correct number of bytes", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		bytes := p.BytesN(16)

		if len(bytes) != 16 {
			t.Errorf("expected 16 bytes, got %d", len(bytes))
		}
	})

	t.Run("handles zero bytes", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		bytes := p.BytesN(0)

		if len(bytes) != 0 {
			t.Errorf("expected 0 bytes, got %d", len(bytes))
		}
	})

	t.Run("same seed produces same bytes", func(t *testing.T) {
		p1 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		p2 := specta.NewPropertyPrimitives(specta.NewSource(12345))

		b1 := p1.BytesN(16)
		b2 := p2.BytesN(16)

		if len(b1) != len(b2) {
			t.Error("different lengths")
		}

		for i := range b1 {
			if b1[i] != b2[i] {
				t.Errorf("bytes differ at index %d: %d != %d", i, b1[i], b2[i])
			}
		}
	})
}

// TestPropertyPrimitives_Time tests time generation
func TestPropertyPrimitives_Time(t *testing.T) {
	t.Run("generates incrementing times", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		t1 := p.Time()
		t2 := p.Time()
		t3 := p.Time()

		if !t2.After(t1) || !t3.After(t2) {
			t.Error("expected incrementing times")
		}
	})

	t.Run("TimeAtOffset works", func(t *testing.T) {
		baseTime, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source, specta.WithPropertyBaseTime(baseTime))

		t1 := p.TimeAtOffset(0)
		t2 := p.TimeAtOffset(24 * time.Hour)

		expected, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")
		if !t2.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, t2)
		}

		if !t1.Before(t2) {
			t.Error("expected t1 before t2")
		}
	})
}

// TestPropertyPrimitives_ID tests ID generation
func TestPropertyPrimitives_ID(t *testing.T) {
	t.Run("generates unique IDs", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		id1 := p.ID()
		id2 := p.ID()

		if id1 == id2 {
			t.Error("expected different IDs")
		}

		if !strings.HasPrefix(id1, "id_") {
			t.Errorf("expected ID to start with 'id_', got %s", id1)
		}
	})
}

// TestPropertyPrimitives_UUID tests UUID generation
func TestPropertyPrimitives_UUID(t *testing.T) {
	t.Run("generates valid UUIDs", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		uuid1 := p.UUID()
		uuid2 := p.UUID()

		if uuid1 == uuid2 {
			t.Error("expected different UUIDs")
		}

		// Check version 4 (random) marker
		if uuid1.Version() != 4 {
			t.Errorf("expected UUID version 4, got %d", uuid1.Version())
		}
	})

	t.Run("same seed produces same UUIDs", func(t *testing.T) {
		p1 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		p2 := specta.NewPropertyPrimitives(specta.NewSource(12345))

		if p1.UUID() != p2.UUID() {
			t.Error("UUID() not deterministic")
		}
	})
}

// TestPropertyPrimitives_ImplementsPrimitives verifies interface compliance
func TestPropertyPrimitives_ImplementsPrimitives(t *testing.T) {
	source := specta.NewSource(12345)
	var _ specta.Primitives = specta.NewPropertyPrimitives(source)
}
