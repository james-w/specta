package specta_test

import (
	"math"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/james-w/specta"
)

// TestPropertyPrimitives_Next tests random value generation
func TestPropertyPrimitives_Next(t *testing.T) {
	t.Run("generates random values", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		// Next() should return random uint64 values
		n1 := p.Next()
		n2 := p.Next()

		// Values should be different (extremely unlikely to be equal)
		if n1 == n2 {
			t.Log("warning: got same value twice (extremely unlikely but possible)")
		}
	})

	t.Run("deterministic with same seed", func(t *testing.T) {
		p1 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		p2 := specta.NewPropertyPrimitives(specta.NewSource(12345))

		if p1.Next() != p2.Next() {
			t.Error("Next() should be deterministic with same seed")
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
	t.Run("generates integers including negatives", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		// Should eventually generate both positive and negative
		var hasPositive, hasNegative bool
		for i := 0; i < 1000; i++ {
			value := p.Int()
			if value > 0 {
				hasPositive = true
			} else if value < 0 {
				hasNegative = true
			}
		}

		if !hasPositive {
			t.Error("never generated positive int")
		}
		if !hasNegative {
			t.Error("never generated negative int")
		}
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
	t.Run("generates floats from full range including special values", func(t *testing.T) {
		source := specta.NewSource(42)
		p := specta.NewPropertyPrimitives(source)

		var hasNaN, hasInf, hasNegInf, hasPositive bool
		for i := 0; i < 200; i++ {
			value := p.Float64()
			if math.IsNaN(value) {
				hasNaN = true
			} else if math.IsInf(value, 1) {
				hasInf = true
			} else if math.IsInf(value, -1) {
				hasNegInf = true
			} else if value > 0 {
				hasPositive = true
			}
		}

		// With our strategy distribution, we should see special values
		if !hasNaN {
			t.Error("never generated NaN")
		}
		if !hasInf {
			t.Error("never generated +Inf")
		}
		if !hasNegInf {
			t.Error("never generated -Inf")
		}
		// Should also get normal values
		if !hasPositive {
			t.Error("never generated positive float")
		}
	})
}

// TestPropertyPrimitives_String tests string generation
func TestPropertyPrimitives_String(t *testing.T) {
	t.Run("generates strings from full byte space including invalid UTF-8", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		var hasEmpty, hasShort, hasLong, hasInvalidUTF8 bool
		// Need more iterations to hit empty string (prob = 1/101 per call)
		for i := 0; i < 500; i++ {
			s := p.String()

			// Check variety of lengths
			if len(s) == 0 {
				hasEmpty = true
			} else if len(s) < 10 {
				hasShort = true
			} else if len(s) > 50 {
				hasLong = true
			}

			// Check if we get invalid UTF-8 (the whole point!)
			if !utf8.ValidString(s) {
				hasInvalidUTF8 = true
			}

			// Early exit if we've found everything
			if hasEmpty && hasShort && hasLong && hasInvalidUTF8 {
				break
			}
		}

		if !hasEmpty {
			t.Error("never generated empty string")
		}
		if !hasShort {
			t.Error("never generated short string")
		}
		if !hasInvalidUTF8 {
			t.Error("never generated invalid UTF-8 (should explore full byte space)")
		}
	})

	t.Run("deterministic with same seed", func(t *testing.T) {
		p1 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		p2 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		if p1.String() != p2.String() {
			t.Error("String() should be deterministic with same seed")
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
	t.Run("generates times from full range", func(t *testing.T) {
		source := specta.NewSource(42)
		p := specta.NewPropertyPrimitives(source)

		var hasZero bool
		for i := 0; i < 100; i++ {
			tm := p.Time()

			if tm.IsZero() {
				hasZero = true
			}
		}

		// With our strategy distribution, we should see edge cases
		if !hasZero {
			t.Error("never generated zero time")
		}
	})

	t.Run("deterministic with same seed", func(t *testing.T) {
		p1 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		p2 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		if !p1.Time().Equal(p2.Time()) {
			t.Error("Time() should be deterministic with same seed")
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
	t.Run("generates random IDs from full string space", func(t *testing.T) {
		source := specta.NewSource(12345)
		p := specta.NewPropertyPrimitives(source)

		id1 := p.ID()
		id2 := p.ID()

		// Should generate different random IDs
		if id1 == id2 {
			t.Error("expected different IDs")
		}

		// ID() should be same as String() - full space exploration
		// Should be deterministic with same seed
		p1 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		p2 := specta.NewPropertyPrimitives(specta.NewSource(12345))
		if p1.ID() != p2.ID() {
			t.Error("ID() should be deterministic with same seed")
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
