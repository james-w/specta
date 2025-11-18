package specta_test

import (
	"strings"
	"testing"

	"github.com/james-w/specta"
)

// TestAssumeSkipsIteration tests that Assume skips test iterations
func TestAssumeSkipsIteration(t *testing.T) {
	var testExecuted int
	var totalAttempts int

	specta.Property(t, func(t *specta.T) {
		totalAttempts++
		n := specta.Int().Range(1, 10).Draw(t, "n")
		t.Assume(n > 5) // Skip if n <= 5
		testExecuted++
		// This code only runs for n > 5
	}, specta.MaxTests(100))

	// Should have executed about 50% of tests (n > 5)
	if testExecuted < 30 || testExecuted > 70 {
		t.Errorf("expected ~50%% execution, got %d/%d", testExecuted, totalAttempts)
	}
}

// TestFilterWithHighPassRate tests Filter with a predicate that passes frequently
func TestFilterWithHighPassRate(t *testing.T) {
	specta.Property(t, func(t *specta.T) {
		// Filter for even numbers - should pass ~50% of the time
		n := specta.Int().Range(1, 100).Filter(func(x int64) bool {
			return x%2 == 0
		}).Draw(t, "n")

		// Verify it's even
		if n%2 != 0 {
			t.Errorf("expected even number but got %d", n)
		}
	}, specta.MaxTests(50))
}

// TestFilterWithLowPassRate tests Filter with a predicate that rarely passes
func TestFilterWithLowPassRate(t *testing.T) {
	var skipCount int
	originalErrorf := func(format string, args ...interface{}) {
		// Count skip warnings
		if strings.Contains(format, "skipped") {
			skipCount++
		}
	}
	_ = originalErrorf // Use if needed

	// This filter is very selective - only primes in a small range
	// Most iterations should be skipped after filter exhaustion
	specta.Property(t, func(t *specta.T) {
		n := specta.Int().Range(1, 100).Filter(func(x int64) bool {
			return isPrime(int(x))
		}).Draw(t, "n")

		// If we get here, n should be prime
		if !isPrime(int(n)) {
			t.Errorf("expected prime but got %d", n)
		}
	}, specta.MaxTests(20))

	// Filter should eventually find primes, but may skip some iterations
	// No assertion needed - just making sure it doesn't hang
}

// TestStringFilter tests Filter on string generator
func TestStringFilter(t *testing.T) {
	specta.Property(t, func(t *specta.T) {
		s := specta.String().AlphaNum().MinLen(5).MaxLen(10).Filter(func(s string) bool {
			return strings.Contains(s, "a") || strings.Contains(s, "A")
		}).Draw(t, "s")

		// Should contain 'a' or 'A'
		if !strings.Contains(s, "a") && !strings.Contains(s, "A") {
			t.Errorf("expected string to contain 'a' or 'A', got %q", s)
		}

		// Should be alphanumeric
		for _, r := range s {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
				t.Errorf("expected alphanumeric string, got %q", s)
			}
		}
	}, specta.MaxTests(30))
}

// TestAssumeWithMultipleConditions tests multiple Assume calls
func TestAssumeWithMultipleConditions(t *testing.T) {
	var executedCount int

	specta.Property(t, func(t *specta.T) {
		x := specta.Int().Range(1, 100).Draw(t, "x")
		y := specta.Int().Range(1, 100).Draw(t, "y")

		t.Assume(x > 50)   // Skip if x <= 50
		t.Assume(y < 50)   // Skip if y >= 50
		t.Assume(x+y > 60) // Skip if sum too small

		executedCount++

		// If we get here, all assumptions hold
		if x <= 50 || y >= 50 || x+y <= 60 {
			t.Errorf("assumptions violated: x=%d, y=%d", x, y)
		}
	}, specta.MaxTests(100))

	// Should have executed some tests (maybe 10-20%)
	if executedCount == 0 {
		t.Error("no tests executed - assumptions too restrictive")
	}
}

// TestFilterComposition tests chaining Filter with other constraints
func TestFilterComposition(t *testing.T) {
	specta.Property(t, func(t *specta.T) {
		n := specta.Int().Range(10, 50).Filter(func(x int64) bool {
			return x%3 == 0 // divisible by 3
		}).Draw(t, "n")

		if n < 10 || n > 50 {
			t.Errorf("expected n in [10, 50], got %d", n)
		}
		if n%3 != 0 {
			t.Errorf("expected n divisible by 3, got %d", n)
		}
	}, specta.MaxTests(30))
}

// Helper function to check if a number is prime
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := 3; i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}
