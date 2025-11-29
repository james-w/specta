package conjecture

import (
	"testing"
)

func TestConjectureData_DrawInteger(t *testing.T) {
	t.Run("draws within range", func(t *testing.T) {
		d := NewConjectureData(WithSeed(1))
		val, err := d.DrawInteger(IntegerParams{Min: 10, Max: 20})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val < 10 || val > 20 {
			t.Errorf("value %d outside range [10, 20]", val)
		}
	})

	t.Run("replay produces same value", func(t *testing.T) {
		d := NewConjectureData(WithSeed(42))
		val1, _ := d.DrawInteger(IntegerParams{Min: 0, Max: 100})

		seq := d.Sequence()
		d2 := ForReplay(seq)
		val2, _ := d2.DrawInteger(IntegerParams{Min: 0, Max: 100})

		if val1 != val2 {
			t.Errorf("replay produced different value: %d vs %d", val1, val2)
		}
	})
}

func TestConjectureData_DrawString(t *testing.T) {
	t.Run("respects min and max size", func(t *testing.T) {
		d := NewConjectureData(WithSeed(1))
		str, err := d.DrawString(StringParams{
			MinSize:   5,
			MaxSize:   10,
			Intervals: DefaultASCIIPrintable(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(str) < 5 || len(str) > 10 {
			t.Errorf("string length %d outside range [5, 10]", len(str))
		}
	})
}

func TestConjectureData_MarkInteresting(t *testing.T) {
	d := NewConjectureData()

	if d.Status() != StatusValid {
		t.Errorf("initial status should be Valid, got %v", d.Status())
	}

	d.MarkInteresting("test reason")

	if d.Status() != StatusInteresting {
		t.Errorf("after MarkInteresting, status should be Interesting, got %v", d.Status())
	}
}

func TestConjectureData_Spans(t *testing.T) {
	d := NewConjectureData()

	spanID := d.StartSpan("test-span")
	d.DrawInteger(IntegerParams{Min: 0, Max: 10})
	d.EndSpan(spanID, false)

	spans := d.Sequence().Spans()
	if len(spans) != 1 {
		t.Errorf("expected 1 span, got %d", len(spans))
	}

	if spans[0].Label != "test-span" {
		t.Errorf("expected span label 'test-span', got %q", spans[0].Label)
	}
}
