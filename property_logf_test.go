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
	t.Run("logs are forwarded from shrunk iteration", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(pt *specta.T) {
			x := specta.Draw(pt, specta.Int().Range(0, 10), "x")

			// Log the value being tested
			pt.Logf("Testing with x = %d", x)
			pt.Logf("This is a debug message")

			// Fail the test for x >= 5
			if x >= 5 {
				pt.Fatalf("x is too large: %d", x)
			}
		}, specta.Seed(42))

		// Verify that the test failed
		specta.AssertThat(t, len(spy.Errors), specta.GreaterThan(0))

		// Verify logs were forwarded
		specta.AssertThat(t, len(spy.Logs), specta.GreaterThan(0))

		// Logs should show x = 5, the minimal failing value after shrinking
		// (not a larger value that might have been the original failure)
		allLogs := strings.Join(spy.Logs, "\n")
		specta.AssertThat(t, allLogs, specta.Contains("Testing with x = 5"))
		specta.AssertThat(t, allLogs, specta.Contains("This is a debug message"))
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

		// Verify no errors and no logs (logs only forwarded on failure)
		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
		specta.AssertThat(t, len(spy.Logs), specta.Equal(0))
	})

	t.Run("logs show shrunk values not original failures", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(pt *specta.T) {
			x := specta.Draw(pt, specta.Int().Range(-1000, 1000), "x")

			// Log the value being tested
			pt.Logf("Current x value: %d", x)

			// Fail on negative values
			if x < 0 {
				pt.Fatalf("negative value: %d", x)
			}
		}, specta.Seed(12345), specta.MaxTests(100))

		// Verify that the test failed
		specta.AssertThat(t, len(spy.Errors), specta.GreaterThan(0))

		// Logs should show x = -1, the minimal negative value after shrinking
		// (not a larger negative number like -500 that might have been the original failure)
		specta.AssertThat(t, len(spy.Logs), specta.GreaterThan(0))
		allLogs := strings.Join(spy.Logs, "\n")
		specta.AssertThat(t, allLogs, specta.Contains("Current x value: -1"))
	})

	t.Run("multiple log calls are all captured in order", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(pt *specta.T) {
			pt.Logf("Log 1")
			pt.Logf("Log 2")
			pt.Logf("Log 3")

			// Always fail immediately
			pt.Fatalf("forced failure")
		}, specta.MaxTests(1))

		// Verify all three logs appear in correct order
		specta.AssertThat(t, spy.Logs, specta.DeepEqual([]string{"Log 1", "Log 2", "Log 3"}))
	})
}
