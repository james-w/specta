package specta

import (
	"fmt"
	"strings"
	"time"
)

// T is a testing helper for property-based tests.
// It implements the TestingT interface and can be used with AssertThat and other matchers.
// Unlike testing.T, calling Fatalf does not immediately terminate the test process,
// but instead panics with a sentinel value to stop the current property iteration.
type T struct {
	failed bool
	errors []string
	Source Source // Exported so tests can construct T instances
}

// propertyFailure is a sentinel panic value used to stop property test iterations.
type propertyFailure struct{}

// skipTest is a sentinel panic value used to skip property test iterations.
// This is used by Assume() and by generators when Filter() exhausts retry attempts.
type skipTest struct{}

// Errorf records an error but allows the property test to continue.
// Multiple assertions can be checked before the property fails.
func (t *T) Errorf(format string, args ...any) {
	t.errors = append(t.errors, fmt.Sprintf(format, args...))
	t.failed = true
}

// Fatalf records an error and immediately stops this property test iteration.
func (t *T) Fatalf(format string, args ...any) {
	t.errors = append(t.errors, fmt.Sprintf(format, args...))
	t.failed = true
	panic(propertyFailure{})
}

// Helper marks the calling function as a test helper.
// This is a no-op for property testing but satisfies the TestingT interface.
func (t *T) Helper() {}

// DrawBits implements Source by delegating to T.Source.
func (t *T) DrawBits(n int) uint64 {
	return t.Source.DrawBits(n)
}

// WriteLog implements Source by delegating to T.Source.
func (t *T) WriteLog(msg string) {
	t.Source.WriteLog(msg)
}

// IsDeterministic implements Source by delegating to T.Source.
func (t *T) IsDeterministic() bool {
	return t.Source.IsDeterministic()
}

// StartInterval implements Source by delegating to T.Source.
func (t *T) StartInterval(label string) {
	t.Source.StartInterval(label)
}

// EndInterval implements Source by delegating to T.Source.
func (t *T) EndInterval() {
	t.Source.EndInterval()
}

// Assume skips the current property test iteration if the condition is false.
// This is useful for filtering generated values that don't meet preconditions.
// Property will track how many tests were skipped and warn if the skip rate is high.
//
// Example:
//
//	Property(t, func(t *T) {
//	    n := Int().Range(1, 100).Draw(t, "n")
//	    t.Assume(isPrime(n))  // Skip if not prime
//	    // Test with prime numbers only
//	})
func (t *T) Assume(condition bool) {
	if !condition {
		panic(skipTest{})
	}
}

// PropertyOption configures property testing behavior.
type PropertyOption func(*propertyConfig)

type propertyConfig struct {
	maxTests   int
	seed       int64
	maxShrinks int
}

// MaxTests sets the maximum number of test iterations to run.
// Default is 100.
func MaxTests(n int) PropertyOption {
	return func(c *propertyConfig) {
		c.maxTests = n
	}
}

// Seed sets the random seed for deterministic test generation.
// This is primarily used to reproduce failures.
// Default is based on the current time.
func Seed(seed int64) PropertyOption {
	return func(c *propertyConfig) {
		c.seed = seed
	}
}

// MaxShrinks sets the maximum number of shrinking attempts.
// Default is 1000.
func MaxShrinks(n int) PropertyOption {
	return func(c *propertyConfig) {
		c.maxShrinks = n
	}
}

// Property runs a property-based test.
// It executes the check function multiple times with randomly generated inputs.
// If a failure is detected, it attempts to shrink the input to a minimal failing case.
//
// Example:
//
//	Property(t, func(t *specta.T) {
//	    x := Int().Draw(t, "x")
//	    y := Int().Draw(t, "y")
//	    AssertThat(t, x+y, Equal(y+x))
//	})
func Property(t TestingT, check func(*T), opts ...PropertyOption) {
	// Only call Helper() if t is a real *testing.T
	if helper, ok := t.(interface{ Helper() }); ok {
		helper.Helper()
	}

	// Apply configuration
	cfg := propertyConfig{
		maxTests:   100,
		seed:       time.Now().UnixNano(),
		maxShrinks: 1000,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Run property tests
	var skipped int
	var tested int
	for i := 0; i < cfg.maxTests; i++ {
		pt := &T{
			Source: NewSource(cfg.seed + int64(i)),
		}

		failed, wasSkipped := runCheck(check, pt)

		if wasSkipped {
			skipped++
			continue
		}

		tested++

		if failed {
			// Property failed - shrink it using Hypothesis-style multi-pass shrinking
			// We know pt.Source is a *randomSource since we created it with NewSource
			rs := pt.Source.(*randomSource)

			// Create test function for the shrinker
			testFunc := func(data []byte) bool {
				shrinkT := &T{
					Source: NewSourceFromData(data),
				}
				failed, _ := runCheck(check, shrinkT)
				return failed
			}

			// Create shrinker with intervals from the original failing run
			shrinker := NewShrinker(rs.Data(), rs.Intervals(), testFunc)
			shrinker.maxCalls = cfg.maxShrinks
			shrunkData := shrinker.Shrink()

			// Report failure
			reportFailure(t, shrunkData, check, i+1, cfg.maxTests, cfg.seed+int64(i), tested, skipped)
			return
		}
	}

	// All tests passed
	// Warn if skip rate is high (>90%)
	if tested > 0 {
		skipRate := float64(skipped) / float64(skipped+tested) * 100
		if skipRate > 90 {
			t.Errorf("Warning: %.1f%% of property tests were skipped (%d/%d). Consider narrowing your generator or using Filter().",
				skipRate, skipped, skipped+tested)
		}
	}
}

// runCheck executes the property check function and returns whether it failed or was skipped.
// It recovers from propertyFailure and skipTest panics but re-panics on unexpected panics.
func runCheck(check func(*T), pt *T) (failed bool, skipped bool) {
	defer func() {
		if r := recover(); r != nil {
			// Check if it's our expected skip sentinel
			if _, ok := r.(skipTest); ok {
				skipped = true
				return
			}
			// Check if it's our expected failure sentinel
			if _, ok := r.(propertyFailure); !ok {
				// Unexpected panic - re-panic to propagate it
				panic(r)
			}
		}
		failed = pt.failed
	}()

	check(pt)
	return
}

// reportFailure creates a detailed error message and fails the test.
func reportFailure(t TestingT, shrunkData []byte, check func(*T), attempts, maxTests int, seed int64, tested, skipped int) {
	// Only call Helper() if t is a real *testing.T
	if helper, ok := t.(interface{ Helper() }); ok {
		helper.Helper()
	}

	// Replay with logging to show generated values
	finalT := &T{
		Source: NewSourceFromData(shrunkData),
	}
	// We know finalT.Source is a *randomSource since we created it with NewSourceFromData
	rs := finalT.Source.(*randomSource)
	rs.EnableLogging()
	runCheck(check, finalT)

	// Build error message
	var msg strings.Builder
	msg.WriteString("\n=== Property Test Failed ===\n")
	msg.WriteString(fmt.Sprintf("Seed: %d\n", seed))
	msg.WriteString(fmt.Sprintf("Attempts: %d/%d\n", attempts, maxTests))
	if skipped > 0 {
		msg.WriteString(fmt.Sprintf("Tested: %d, Skipped: %d (%.1f%% skip rate)\n",
			tested, skipped, float64(skipped)/float64(tested+skipped)*100))
	}

	// Show generated values
	if log := rs.Log(); log != "" {
		msg.WriteString("\nGenerated values:\n  ")
		msg.WriteString(strings.TrimSpace(log))
		msg.WriteString("\n")
	}

	// Show failure messages
	if len(finalT.errors) > 0 {
		msg.WriteString("\nFailure:\n")
		for _, err := range finalT.errors {
			// For multiline errors, indent each line
			lines := strings.Split(err, "\n")
			for _, line := range lines {
				if line != "" {
					msg.WriteString("  ")
					msg.WriteString(line)
				}
				msg.WriteString("\n")
			}
		}
	}

	// Show reproduction instructions
	msg.WriteString(fmt.Sprintf("\nReproduce: Property(t, check, Seed(%d))\n", seed))

	t.Errorf("%s", msg.String())
}
