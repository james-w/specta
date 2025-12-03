package specta_test

import (
	"strings"
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/testlib"
)

// TestPropertyLogForwarding verifies that t.Logf() output from property tests
// is captured and included in the error output when a test fails.
func TestPropertyLogForwarding(t *testing.T) {
	t.Run("logs are forwarded on failure", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(pt *specta.T) {
			x := specta.Draw(pt, specta.Int().Range(0, 10), "x")

			// Log some debug info
			pt.Logf("Testing with x = %d", x)
			pt.Logf("This is a debug message")

			// Fail the test
			if x >= 5 {
				pt.Fatalf("x is too large: %d", x)
			}
		}, specta.Seed(42))

		// Verify that the test failed
		specta.AssertThat(t, len(spy.Errors), specta.GreaterThan(0))

		// Check that logs were forwarded in the error message
		errorMsg := spy.Errors[0]
		specta.AssertThat(t, errorMsg, specta.Contains("Logs:"))
		specta.AssertThat(t, errorMsg, specta.Contains("Testing with x"))
		specta.AssertThat(t, errorMsg, specta.Contains("This is a debug message"))
	})

	t.Run("logs are not shown on success", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(pt *specta.T) {
			x := specta.Draw(pt, specta.Int().Range(0, 10), "x")

			// Log something but don't fail
			pt.Logf("This log should not appear because test passes")

			// Always pass
			specta.AssertThat(pt, x, specta.GreaterThanOrEqual(int64(0)))
		}, specta.Seed(99), specta.MaxTests(10))

		// Verify no errors
		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
	})

	t.Run("logs are from shrunk iteration", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(pt *specta.T) {
			x := specta.Draw(pt, specta.Int(), "x")

			// Log the value being tested
			pt.Logf("Current x value: %d", x)

			// Fail on negative values
			if x < 0 {
				pt.Fatalf("negative value: %d", x)
			}
		}, specta.Seed(12345), specta.MaxTests(100))

		// Verify that the test failed
		specta.AssertThat(t, len(spy.Errors), specta.GreaterThan(0))

		// The error should contain logs from the shrunk (minimal) failing case
		errorMsg := spy.Errors[0]
		specta.AssertThat(t, errorMsg, specta.Contains("Logs:"))
		specta.AssertThat(t, errorMsg, specta.Contains("Current x value:"))

		// The shrunk value should be -1 or close to 0 (shrinking minimizes negative values)
		// We just verify that logs are present
	})

	t.Run("multiple log calls are all captured", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(pt *specta.T) {
			pt.Logf("Log 1")
			pt.Logf("Log 2")
			pt.Logf("Log 3")

			// Always fail immediately
			pt.Fatalf("forced failure")
		}, specta.MaxTests(1))

		// Verify all three logs appear
		errorMsg := spy.Errors[0]
		specta.AssertThat(t, strings.Count(errorMsg, "Log 1"), specta.Equal(1))
		specta.AssertThat(t, strings.Count(errorMsg, "Log 2"), specta.Equal(1))
		specta.AssertThat(t, strings.Count(errorMsg, "Log 3"), specta.Equal(1))
	})
}
