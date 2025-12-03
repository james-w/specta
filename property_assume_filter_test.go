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
		n := specta.Draw(t, specta.Int().Range(1, 10), "n")
		t.Assume(n > 5) // Skip if n <= 5
		testExecuted++
		// This code only runs for n > 5
	}, specta.MaxTests(100))

	// Should have executed about 50% of tests (n > 5)
	specta.AssertThat(t, testExecuted, specta.AllOf(
		specta.GreaterThanOrEqual(30),
		specta.Not(specta.GreaterThan(70)),
	))
}

// TestFilterWithHighPassRate tests Filter with a predicate that passes frequently
func TestFilterWithHighPassRate(t *testing.T) {
	specta.Property(t, func(t *specta.T) {
		// Filter for even numbers - should pass ~50% of the time
		n := specta.Draw(t, specta.Int().Range(1, 100).Filter(func(x int64) bool {
			return x%2 == 0
		}), "n")

		// Verify it's even
		specta.AssertThat(t, n%2, specta.Equal(int64(0)))
	}, specta.MaxTests(50))
}

// TestFilterWithLowPassRate tests Filter with a predicate that rarely passes
func TestFilterWithLowPassRate(t *testing.T) {
	// This filter is very selective - only primes in a small range
	// Most iterations should be skipped after filter exhaustion
	specta.Property(t, func(t *specta.T) {
		n := specta.Draw(t, specta.Int().Range(1, 100).Filter(func(x int64) bool {
			return isPrime(int(x))
		}), "n")

		// If we get here, n should be prime
		specta.AssertThat(t, isPrime(int(n)), specta.IsTrue())
	}, specta.MaxTests(20))

	// Filter should eventually find primes, but may skip some iterations
	// No assertion needed - just making sure it doesn't hang
}

// TestStringFilter tests Filter on string generator
func TestStringFilter(t *testing.T) {
	specta.Property(t, func(t *specta.T) {
		s := specta.Draw(t, specta.String().AlphaNum().MinLen(5).MaxLen(10).Filter(func(s string) bool {
			return strings.Contains(s, "a") || strings.Contains(s, "A")
		}), "s")

		// Should contain 'a' or 'A'
		specta.AssertThat(t, strings.Contains(s, "a") || strings.Contains(s, "A"), specta.IsTrue())

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
		x := specta.Draw(t, specta.Int().Range(1, 100), "x")
		y := specta.Draw(t, specta.Int().Range(1, 100), "y")

		t.Assume(x > 50)   // Skip if x <= 50
		t.Assume(y < 50)   // Skip if y >= 50
		t.Assume(x+y > 60) // Skip if sum too small

		executedCount++

		// If we get here, all assumptions hold
		specta.AssertThat(t, x, specta.GreaterThan(int64(50)))
		specta.AssertThat(t, y, specta.LessThan(int64(50)))
		specta.AssertThat(t, x+y, specta.GreaterThan(int64(60)))
	}, specta.MaxTests(100))

	// Should have executed some tests (maybe 10-20%)
	specta.AssertThat(t, executedCount, specta.GreaterThan(0))
}

// TestFilterComposition tests chaining Filter with other constraints
func TestFilterComposition(t *testing.T) {
	specta.Property(t, func(t *specta.T) {
		n := specta.Draw(t, specta.Int().Range(10, 50).Filter(func(x int64) bool {
			return x%3 == 0 // divisible by 3
		}), "n")

		specta.AssertThat(t, n, specta.AllOf(
			specta.GreaterThanOrEqual(int64(10)),
			specta.Not(specta.GreaterThan(int64(50))),
		))
		specta.AssertThat(t, n%3, specta.Equal(int64(0)))
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
