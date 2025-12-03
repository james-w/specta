package specta

import (
	"testing"
)

func TestIntervalTracking(t *testing.T) {
	t.Run("tracks intervals during generation", func(t *testing.T) {
		src := NewSource(42)

		// Simulate a generator drawing with labels
		src.StartInterval("x")
		src.DrawBits(8)
		src.DrawBits(8)
		src.EndInterval()

		src.StartInterval("y")
		src.DrawBits(16)
		src.EndInterval()

		intervals := src.Intervals()

		if len(intervals) != 2 {
			t.Fatalf("expected 2 intervals, got %d", len(intervals))
		}

		// First interval: "x" should cover 2 bytes (0-2)
		if intervals[0].Label != "x" {
			t.Errorf("interval 0: expected label 'x', got '%s'", intervals[0].Label)
		}
		if intervals[0].Start != 0 {
			t.Errorf("interval 0: expected start 0, got %d", intervals[0].Start)
		}
		if intervals[0].End != 2 {
			t.Errorf("interval 0: expected end 2, got %d", intervals[0].End)
		}

		// Second interval: "y" should cover 2 bytes (2-4)
		if intervals[1].Label != "y" {
			t.Errorf("interval 1: expected label 'y', got '%s'", intervals[1].Label)
		}
		if intervals[1].Start != 2 {
			t.Errorf("interval 1: expected start 2, got %d", intervals[1].Start)
		}
		if intervals[1].End != 4 {
			t.Errorf("interval 1: expected end 4, got %d", intervals[1].End)
		}
	})

	t.Run("does not track intervals during replay", func(t *testing.T) {
		// Create data from initial generation
		src := NewSource(42)
		src.StartInterval("x")
		src.DrawBits(8)
		src.EndInterval()
		data := src.Data()

		// Replay should not track intervals
		replaySrc := NewSourceFromData(data)
		replaySrc.StartInterval("x")
		replaySrc.DrawBits(8)
		replaySrc.EndInterval()

		intervals := replaySrc.Intervals()
		if len(intervals) != 0 {
			t.Errorf("replay source should not track intervals, got %d", len(intervals))
		}
	})

	t.Run("nested intervals not supported", func(t *testing.T) {
		// This documents current behavior - nested intervals overwrite
		// If we need nested intervals later, we'd need to change the implementation
		src := NewSource(42)

		src.StartInterval("outer")
		src.DrawBits(8)

		src.StartInterval("inner") // This overwrites currentInterval
		src.DrawBits(8)
		src.EndInterval()

		src.EndInterval() // This is now a no-op since currentInterval is nil

		intervals := src.Intervals()

		if len(intervals) != 1 {
			t.Fatalf("expected 1 interval (nested not supported), got %d", len(intervals))
		}

		if intervals[0].Label != "inner" {
			t.Errorf("expected 'inner' interval to win, got '%s'", intervals[0].Label)
		}
	})
}
