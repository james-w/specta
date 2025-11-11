package specta_test

import (
	"testing"
	"time"

	"github.com/james-w/specta"
)

func TestGen_Next(t *testing.T) {
	g := specta.New()

	first := g.Next()
	second := g.Next()
	third := g.Next()

	if first != 1 || second != 2 || third != 3 {
		t.Errorf("Expected sequential values 1, 2, 3 but got %d, %d, %d", first, second, third)
	}
}

func TestGen_WithStart(t *testing.T) {
	g := specta.New(specta.WithStart(100))

	first := g.Next()
	if first != 101 {
		t.Errorf("Expected first value to be 101, got %d", first)
	}
}

func TestGen_Bool(t *testing.T) {
	g := specta.New()

	// Currently always returns false - test actual behavior
	result := g.Bool()
	if result != false {
		t.Errorf("Expected Bool() to return false, got %v", result)
	}
}

func TestGen_Int(t *testing.T) {
	g := specta.New()

	first := g.Int()
	second := g.Int()

	if first >= second {
		t.Errorf("Expected deterministic increasing values, got %d then %d", first, second)
	}

	if first != 1 || second != 2 {
		t.Errorf("Expected 1, 2 but got %d, %d", first, second)
	}
}

func TestGen_IntN(t *testing.T) {
	g := specta.New()

	// Test with valid max
	for range 10 {
		val := g.IntN(5)
		if val < 0 || val >= 5 {
			t.Errorf("IntN(5) returned %d, expected 0 <= val < 5", val)
		}
	}

	// Test with zero max
	zeroResult := g.IntN(0)
	if zeroResult != 0 {
		t.Errorf("IntN(0) returned %d, expected 0", zeroResult)
	}

	// Test with negative max
	negResult := g.IntN(-5)
	if negResult != 0 {
		t.Errorf("IntN(-5) returned %d, expected 0", negResult)
	}
}

func TestGen_Int64(t *testing.T) {
	g := specta.New()

	first := g.Int64()
	second := g.Int64()

	if first != 1 || second != 2 {
		t.Errorf("Expected 1, 2 but got %d, %d", first, second)
	}
}

func TestGen_Uint64(t *testing.T) {
	g := specta.New()

	first := g.Uint64()
	second := g.Uint64()

	if first != 1 || second != 2 {
		t.Errorf("Expected 1, 2 but got %d, %d", first, second)
	}
}

func TestGen_Float64(t *testing.T) {
	g := specta.New()

	first := g.Float64()
	second := g.Float64()

	// Float64 adds 0.123 to the counter value
	if first != 1.123 || second != 2.123 {
		t.Errorf("Expected 1.123, 2.123 but got %f, %f", first, second)
	}
}

func TestGen_String(t *testing.T) {
	g := specta.New()

	first := g.String()
	second := g.String()

	if first != "str_1" || second != "str_2" {
		t.Errorf("Expected 'str_1', 'str_2' but got '%s', '%s'", first, second)
	}
}

func TestGen_StringWith(t *testing.T) {
	g := specta.New()

	first := g.StringWith("user_")
	second := g.StringWith("user_")

	if first != "user_1" || second != "user_2" {
		t.Errorf("Expected 'user_1', 'user_2' but got '%s', '%s'", first, second)
	}
}

func TestGen_WithPrefix(t *testing.T) {
	g := specta.New(specta.WithPrefix("test_"))

	str := g.String()
	id := g.ID()

	if str != "test_str_1" {
		t.Errorf("Expected 'test_str_1', got '%s'", str)
	}
	if id != "test_id_2" {
		t.Errorf("Expected 'test_id_2', got '%s'", id)
	}
}

func TestGen_BytesN(t *testing.T) {
	g := specta.New()

	// Test with valid length
	bytes := g.BytesN(3)
	if len(bytes) != 3 {
		t.Errorf("Expected 3 bytes, got %d", len(bytes))
	}

	// Bytes should be deterministic based on sequence
	if bytes[0] != 1 || bytes[1] != 2 || bytes[2] != 3 {
		t.Errorf("Expected bytes [1, 2, 3], got %v", bytes)
	}

	// Test with zero length
	emptyBytes := g.BytesN(0)
	if emptyBytes != nil {
		t.Errorf("Expected nil for BytesN(0), got %v", emptyBytes)
	}

	// Test with negative length
	negBytes := g.BytesN(-5)
	if negBytes != nil {
		t.Errorf("Expected nil for BytesN(-5), got %v", negBytes)
	}
}

func TestGen_Time(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	g := specta.New(specta.WithBaseTime(baseTime), specta.WithTimeStep(time.Hour))

	first := g.Time()
	second := g.Time()

	expectedFirst := baseTime.Add(1 * time.Hour)
	expectedSecond := baseTime.Add(2 * time.Hour)

	if !first.Equal(expectedFirst) {
		t.Errorf("Expected %v, got %v", expectedFirst, first)
	}
	if !second.Equal(expectedSecond) {
		t.Errorf("Expected %v, got %v", expectedSecond, second)
	}
}

func TestGen_TimeAtOffset(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	g := specta.New(specta.WithBaseTime(baseTime))

	offset := 5 * time.Hour
	result := g.TimeAtOffset(offset)
	expected := baseTime.Add(offset)

	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// TimeAtOffset should not affect counter
	nextVal := g.Next()
	if nextVal != 1 {
		t.Errorf("Expected counter to be at 1, got %d", nextVal)
	}
}

func TestGen_Duration(t *testing.T) {
	g := specta.New(specta.WithTimeStep(time.Minute))

	first := g.Duration()
	second := g.Duration()

	if first != 1*time.Minute || second != 2*time.Minute {
		t.Errorf("Expected 1m, 2m but got %v, %v", first, second)
	}
}

func TestGen_ID(t *testing.T) {
	g := specta.New()

	first := g.ID()
	second := g.ID()

	if first != "id_1" || second != "id_2" {
		t.Errorf("Expected 'id_1', 'id_2' but got '%s', '%s'", first, second)
	}
}

func TestGen_UUID(t *testing.T) {
	g := specta.New()

	first := g.UUID()
	second := g.UUID()

	// UUIDs should be valid
	if first.String() == "" || second.String() == "" {
		t.Error("Generated UUIDs should not be empty")
	}

	// UUIDs should be deterministic
	g2 := specta.New()
	firstAgain := g2.UUID()

	if first != firstAgain {
		t.Errorf("Expected deterministic UUID generation, got %s and %s", first, firstAgain)
	}

	// Different sequence numbers should produce different UUIDs
	if first == second {
		t.Error("Sequential UUIDs should be different")
	}
}

func TestGen_Deterministic(t *testing.T) {
	// Test that two generators produce identical sequences
	g1 := specta.New()
	g2 := specta.New()

	for range 10 {
		if g1.Next() != g2.Next() {
			t.Error("Expected identical sequences from separate generators")
		}
	}
}

func TestGen_Options(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	g := specta.New(
		specta.WithStart(50),
		specta.WithBaseTime(baseTime),
		specta.WithTimeStep(time.Hour),
		specta.WithPrefix("custom_"),
	)

	// Test start
	if g.Next() != 51 {
		t.Errorf("Expected Next() to be 51 with WithStart(50)")
	}

	// Test prefix
	id := g.ID()
	if id != "custom_id_52" {
		t.Errorf("Expected 'custom_id_52', got '%s'", id)
	}

	// Test time options
	tm := g.Time()
	expected := baseTime.Add(53 * time.Hour)
	if !tm.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, tm)
	}
}
