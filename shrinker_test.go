package specta

import (
	"bytes"
	"testing"
)

func TestShrinkBuffer(t *testing.T) {
	t.Run("DeleteRange removes bytes", func(t *testing.T) {
		sb := NewShrinkBuffer([]byte{1, 2, 3, 4, 5})

		if !sb.DeleteRange(1, 3) {
			t.Fatal("DeleteRange failed")
		}

		expected := []byte{1, 4, 5}
		if !bytes.Equal(sb.Data(), expected) {
			t.Errorf("expected %v, got %v", expected, sb.Data())
		}
	})

	t.Run("DeleteRange validates indices", func(t *testing.T) {
		sb := NewShrinkBuffer([]byte{1, 2, 3})

		if sb.DeleteRange(-1, 2) {
			t.Error("should reject negative start")
		}
		if sb.DeleteRange(0, 10) {
			t.Error("should reject end > length")
		}
		if sb.DeleteRange(2, 1) {
			t.Error("should reject start >= end")
		}
	})

	t.Run("MinimizeByte reduces value", func(t *testing.T) {
		sb := NewShrinkBuffer([]byte{10, 20, 30})

		if !sb.MinimizeByte(1, 5) {
			t.Fatal("MinimizeByte failed")
		}

		expected := []byte{10, 15, 30}
		if !bytes.Equal(sb.Data(), expected) {
			t.Errorf("expected %v, got %v", expected, sb.Data())
		}
	})

	t.Run("MinimizeByte prevents negative values", func(t *testing.T) {
		sb := NewShrinkBuffer([]byte{5})

		if sb.MinimizeByte(0, 10) {
			t.Error("should prevent value going negative")
		}

		// Value should be unchanged
		if sb.Data()[0] != 5 {
			t.Errorf("value should be unchanged, got %d", sb.Data()[0])
		}
	})

	t.Run("SetByte changes value", func(t *testing.T) {
		sb := NewShrinkBuffer([]byte{1, 2, 3})

		if !sb.SetByte(1, 99) {
			t.Fatal("SetByte failed")
		}

		expected := []byte{1, 99, 3}
		if !bytes.Equal(sb.Data(), expected) {
			t.Errorf("expected %v, got %v", expected, sb.Data())
		}
	})

	t.Run("Swap exchanges bytes", func(t *testing.T) {
		sb := NewShrinkBuffer([]byte{1, 2, 3})

		if !sb.Swap(0, 2) {
			t.Fatal("Swap failed")
		}

		expected := []byte{3, 2, 1}
		if !bytes.Equal(sb.Data(), expected) {
			t.Errorf("expected %v, got %v", expected, sb.Data())
		}
	})
}

func TestShrinker(t *testing.T) {
	t.Run("accepts smaller failing examples", func(t *testing.T) {
		// Test function that fails if first byte >= 10
		testFunc := func(data []byte) bool {
			if len(data) == 0 {
				return false
			}
			return data[0] >= 10
		}

		shrinker := NewShrinker([]byte{20, 30, 40}, nil, testFunc)

		// Try a shorter example that still fails
		if !shrinker.test([]byte{15}) {
			t.Error("should accept shorter failing example")
		}

		if !bytes.Equal(shrinker.Current(), []byte{15}) {
			t.Errorf("current should be updated to %v, got %v", []byte{15}, shrinker.Current())
		}
	})

	t.Run("rejects passing examples", func(t *testing.T) {
		testFunc := func(data []byte) bool {
			if len(data) == 0 {
				return false
			}
			return data[0] >= 10
		}

		shrinker := NewShrinker([]byte{20}, nil, testFunc)

		// Try an example that passes (should be rejected)
		if shrinker.test([]byte{5}) {
			t.Error("should reject passing example")
		}

		if !bytes.Equal(shrinker.Current(), []byte{20}) {
			t.Error("current should be unchanged")
		}
	})

	t.Run("prefers lexicographically smaller for same length", func(t *testing.T) {
		testFunc := func(data []byte) bool {
			// Always fails
			return true
		}

		shrinker := NewShrinker([]byte{10, 20, 30}, nil, testFunc)

		// Same length, but lexicographically smaller
		if !shrinker.test([]byte{5, 20, 30}) {
			t.Error("should accept lexicographically smaller")
		}

		if !bytes.Equal(shrinker.Current(), []byte{5, 20, 30}) {
			t.Errorf("expected %v, got %v", []byte{5, 20, 30}, shrinker.Current())
		}

		// Try one that's larger at first byte (should be rejected)
		if shrinker.test([]byte{6, 10, 10}) {
			t.Error("should reject lexicographically larger")
		}
	})

	t.Run("tracks test calls", func(t *testing.T) {
		testFunc := func(data []byte) bool {
			return true
		}

		shrinker := NewShrinker([]byte{1, 2, 3}, nil, testFunc)

		if shrinker.Calls() != 0 {
			t.Errorf("initial calls should be 0, got %d", shrinker.Calls())
		}

		shrinker.test([]byte{1, 2})
		shrinker.test([]byte{1})

		if shrinker.Calls() != 2 {
			t.Errorf("expected 2 calls, got %d", shrinker.Calls())
		}
	})

	t.Run("respects max calls limit", func(t *testing.T) {
		testFunc := func(data []byte) bool {
			return true
		}

		shrinker := NewShrinker([]byte{1, 2, 3}, nil, testFunc)
		shrinker.maxCalls = 5

		// Make 5 calls
		for i := 0; i < 5; i++ {
			shrinker.test([]byte{byte(i)})
		}

		// 6th call should be rejected
		if shrinker.test([]byte{99}) {
			t.Error("should reject call beyond maxCalls")
		}

		if shrinker.Calls() != 5 {
			t.Errorf("expected 5 calls, got %d", shrinker.Calls())
		}
	})
}

func TestAdaptiveBlockDeletion(t *testing.T) {
	t.Run("deletes unnecessary bytes", func(t *testing.T) {
		// Test fails if the first byte is 'X' (0x58)
		// The rest of the bytes are irrelevant
		testFunc := func(data []byte) bool {
			if len(data) == 0 {
				return false
			}
			return data[0] == 'X'
		}

		// Start with: X + 20 irrelevant bytes
		initial := make([]byte, 21)
		initial[0] = 'X'
		for i := 1; i < 21; i++ {
			initial[i] = byte(i)
		}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should shrink to just [X]
		expected := []byte{'X'}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		t.Logf("Shrunk from %d bytes to %d bytes in %d test calls",
			len(initial), len(result), shrinker.Calls())
	})

	t.Run("finds minimal failing subsequence", func(t *testing.T) {
		// Test fails if it contains the sequence [5, 6, 7]
		testFunc := func(data []byte) bool {
			if len(data) < 3 {
				return false
			}
			for i := 0; i <= len(data)-3; i++ {
				if data[i] == 5 && data[i+1] == 6 && data[i+2] == 7 {
					return true
				}
			}
			return false
		}

		// Start with: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
		initial := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should shrink to [5, 6, 7]
		expected := []byte{5, 6, 7}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		t.Logf("Shrunk from %d bytes to %d bytes in %d test calls",
			len(initial), len(result), shrinker.Calls())
	})

	t.Run("handles case where nothing can be deleted", func(t *testing.T) {
		// Test fails if all bytes sum to >= 10
		testFunc := func(data []byte) bool {
			sum := 0
			for _, b := range data {
				sum += int(b)
			}
			return sum >= 10
		}

		// [5, 5] - can't delete either byte without making sum < 10
		initial := []byte{5, 5}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should remain unchanged (but might reorder or minimize values later)
		if len(result) != 2 {
			t.Errorf("expected length 2, got %d", len(result))
		}

		t.Logf("Unable to delete any bytes, result: %v after %d calls",
			result, shrinker.Calls())
	})

	t.Run("uses binary search for efficiency", func(t *testing.T) {
		// Test fails if first byte is 'X'
		testFunc := func(data []byte) bool {
			if len(data) == 0 {
				return false
			}
			return data[0] == 'X'
		}

		// Start with X + 100 irrelevant bytes
		initial := make([]byte, 101)
		initial[0] = 'X'

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should shrink to [X]
		if !bytes.Equal(result, []byte{'X'}) {
			t.Errorf("expected [X], got %v", result)
		}

		// With adaptive delta debugging, this should be much fewer than 100 calls
		// (binary search is O(log n) instead of O(n))
		t.Logf("Shrunk %d bytes with only %d test calls (efficient!)",
			len(initial), shrinker.Calls())

		// Sanity check: should be way less than linear scan
		if shrinker.Calls() > 50 {
			t.Logf("Warning: expected < 50 calls with binary search, got %d", shrinker.Calls())
		}
	})

	t.Run("handles empty result", func(t *testing.T) {
		// Test always fails (even on empty input)
		testFunc := func(data []byte) bool {
			return true
		}

		initial := []byte{1, 2, 3, 4, 5}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should shrink to empty
		if len(result) != 0 {
			t.Errorf("expected empty, got %v", result)
		}
	})

	t.Run("minimizes deletion size", func(t *testing.T) {
		// Test fails if it contains byte value 99
		testFunc := func(data []byte) bool {
			for _, b := range data {
				if b == 99 {
					return true
				}
			}
			return false
		}

		// Put 99 in the middle of a large block
		initial := make([]byte, 20)
		for i := range initial {
			initial[i] = byte(i)
		}
		initial[10] = 99

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should shrink to just [99]
		expected := []byte{99}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		t.Logf("Found minimal failing byte in %d calls", shrinker.Calls())
	})
}

func TestByteMinimization(t *testing.T) {
	t.Run("minimizes individual bytes", func(t *testing.T) {
		// Test fails if any byte >= 10
		testFunc := func(data []byte) bool {
			for _, b := range data {
				if b >= 10 {
					return true
				}
			}
			return false
		}

		// Start with high values
		initial := []byte{100, 50, 200, 30}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Block deletion will remove unnecessary bytes, leaving just one
		// That byte should be minimized to 10 (the minimum value that still fails)
		expected := []byte{10}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		t.Logf("Minimized bytes from %v to %v in %d calls",
			initial, result, shrinker.Calls())
	})

	t.Run("minimizes duplicates efficiently", func(t *testing.T) {
		// Test fails if length == 3 AND all bytes are the same AND >= 10
		// This forces all bytes to be kept (can't delete any)
		testFunc := func(data []byte) bool {
			if len(data) != 3 {
				return false
			}
			// All bytes must be the same
			first := data[0]
			for _, b := range data {
				if b != first {
					return false
				}
			}
			// And that value must be >= 10
			return first >= 10
		}

		// Start with run of identical high values: [50, 50, 50]
		initial := []byte{50, 50, 50}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should minimize all three bytes together to [10, 10, 10]
		// Can't delete any because test requires exactly 3 bytes
		expected := []byte{10, 10, 10}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		t.Logf("Minimized duplicates from %v to %v in %d calls",
			initial, result, shrinker.Calls())
	})

	t.Run("handles bytes that cannot be minimized", func(t *testing.T) {
		// Test fails if first byte is exactly 42
		testFunc := func(data []byte) bool {
			if len(data) == 0 {
				return false
			}
			return data[0] == 42
		}

		initial := []byte{42, 99, 100}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// First byte must stay 42, others should be deleted
		expected := []byte{42}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("combines deletion and minimization", func(t *testing.T) {
		// Test fails if it contains byte value >= 50
		testFunc := func(data []byte) bool {
			for _, b := range data {
				if b >= 50 {
					return true
				}
			}
			return false
		}

		// Mix of high and low values
		initial := []byte{1, 2, 100, 4, 5, 200, 7}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should delete unnecessary bytes and minimize the high one
		// Result should be single byte with value 50
		expected := []byte{50}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		t.Logf("Optimized from %v to %v in %d calls",
			initial, result, shrinker.Calls())
	})

	t.Run("minimizes all bytes to zero when possible", func(t *testing.T) {
		// Test always fails
		testFunc := func(data []byte) bool {
			return len(data) > 0
		}

		initial := []byte{100, 200, 50}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should minimize to single zero byte
		expected := []byte{0}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})
}

func TestStructuralShrinking(t *testing.T) {
	t.Run("sorts intervals when order doesn't matter", func(t *testing.T) {
		// Test fails if it contains bytes [5, 3] in any order
		testFunc := func(data []byte) bool {
			if len(data) < 2 {
				return false
			}
			// Check for [5, 3] or [3, 5]
			for i := 0; i < len(data)-1; i++ {
				if (data[i] == 5 && data[i+1] == 3) || (data[i] == 3 && data[i+1] == 5) {
					return true
				}
			}
			return false
		}

		// Start with the larger ordering: [5, 3]
		initial := []byte{5, 3}

		// Create intervals for the two elements
		intervals := []Interval{
			{Start: 0, End: 1, Label: "a"},
			{Start: 1, End: 2, Label: "b"},
		}

		shrinker := NewShrinker(initial, intervals, testFunc)
		result := shrinker.Shrink()

		// Should sort to [3, 5] which is lexicographically smaller
		expected := []byte{3, 5}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		t.Logf("Sorted from %v to %v in %d calls", initial, result, shrinker.Calls())
	})

	t.Run("handles case without intervals", func(t *testing.T) {
		// Without interval information, structural passes should be skipped gracefully
		testFunc := func(data []byte) bool {
			return len(data) > 0
		}

		initial := []byte{5, 3, 7}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should still shrink using other passes (block deletion, byte minimization)
		// Just verify it doesn't crash
		if len(result) == 0 {
			t.Error("should not shrink to empty when test requires len > 0")
		}

		t.Logf("Shrunk to %v without intervals", result)
	})

	t.Run("lexicographic comparison helper", func(t *testing.T) {
		tests := []struct {
			a, b     []byte
			expected bool
		}{
			{[]byte{1}, []byte{2}, true},       // 1 < 2
			{[]byte{2}, []byte{1}, false},      // 2 > 1
			{[]byte{1, 2}, []byte{1, 3}, true}, // [1,2] < [1,3]
			{[]byte{1, 3}, []byte{1, 2}, false},
			{[]byte{1}, []byte{1, 2}, true},     // shorter is smaller when prefixes match
			{[]byte{1, 2}, []byte{1}, false},    // longer is not smaller
			{[]byte{1, 2}, []byte{1, 2}, false}, // equal is not smaller
		}

		for _, tt := range tests {
			result := isLexicographicallySmaller(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("isLexicographicallySmaller(%v, %v) = %v, expected %v",
					tt.a, tt.b, result, tt.expected)
			}
		}
	})
}

func TestPassOrchestration(t *testing.T) {
	t.Run("multiple rounds improve shrinking", func(t *testing.T) {
		// Test fails if it contains value 50 somewhere
		testFunc := func(data []byte) bool {
			for _, b := range data {
				if b == 50 {
					return true
				}
			}
			return false
		}

		// Start with: [100, 50, 200, 75, 50, 90]
		// Multiple bytes = 50, plus many distractors
		initial := []byte{100, 50, 200, 75, 50, 90}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should shrink to just [50]
		// - Round 1: Delete unnecessary bytes (might not get all)
		// - Round 1: Minimize values
		// - Round 2: Delete more now that values are smaller
		// - Eventually converges to [50]
		expected := []byte{50}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		t.Logf("Multi-round shrinking: %v → %v in %d calls",
			initial, result, shrinker.Calls())
	})

	t.Run("respects fixed point", func(t *testing.T) {
		// Test that requires specific structure that can't be shrunk further
		testFunc := func(data []byte) bool {
			// Fails if length >= 2 AND both bytes >= 10
			if len(data) < 2 {
				return false
			}
			return data[0] >= 10 && data[1] >= 10
		}

		initial := []byte{100, 100, 50, 75}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should minimize to [10, 10] (minimum values) and delete extra bytes
		expected := []byte{10, 10}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		// Should not run too many rounds (should detect fixed point)
		if shrinker.Calls() > 100 {
			t.Logf("Warning: used %d calls, might not be detecting fixed point efficiently",
				shrinker.Calls())
		}

		t.Logf("Reached fixed point at %v in %d calls", result, shrinker.Calls())
	})

	t.Run("handles complex shrinking scenario", func(t *testing.T) {
		// Test fails if sum of bytes >= 100 AND length >= 3
		testFunc := func(data []byte) bool {
			if len(data) < 3 {
				return false
			}
			sum := 0
			for _, b := range data {
				sum += int(b)
			}
			return sum >= 100
		}

		// Start with varied high values
		initial := []byte{50, 40, 30, 20, 10, 5}

		shrinker := NewShrinker(initial, nil, testFunc)
		result := shrinker.Shrink()

		// Should find minimal failing case:
		// - Need at least 3 bytes
		// - Need sum >= 100
		// - Minimum is [34, 33, 33] (sum = 100) or similar
		if len(result) < 3 {
			t.Errorf("expected at least 3 bytes, got %d", len(result))
		}

		sum := 0
		for _, b := range result {
			sum += int(b)
		}

		if sum < 100 {
			t.Errorf("expected sum >= 100, got %d", sum)
		}

		t.Logf("Complex shrink: %v → %v (len=%d, sum=%d) in %d calls",
			initial, result, len(result), sum, shrinker.Calls())
	})
}
