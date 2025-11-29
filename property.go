package specta

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/james-w/specta/conjecture"
)

// T is a testing helper for property-based tests.
// It implements the TestingT interface and can be used with AssertThat and other matchers.
// Unlike testing.T, calling Fatalf does not immediately terminate the test process,
// but instead panics with a sentinel value to stop the current property iteration.
type T struct {
	failed bool
	errors []string
	Data   *conjecture.ConjectureData // ConjectureData for drawing values
	Source Source                     // Deprecated: kept for backward compatibility, will panic if used

	// Track generated values for error reporting
	generatedValues      map[string]string // label -> formatted value
	generatedValuesOrder []string          // labels in draw order
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
		// Create ConjectureData for this iteration
		data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(cfg.seed + int64(i))))

		pt := &T{
			Data:            data,
			Source:          nil, // Don't set Source - it's deprecated
			generatedValues: make(map[string]string),
		}

		failed, wasSkipped := runCheck(check, pt)

		// Check if test overran (couldn't satisfy constraints)
		if data.Status() == conjecture.StatusOverrun {
			skipped++
			continue
		}

		if wasSkipped {
			skipped++
			continue
		}

		tested++

		if failed {
			// Property failed - attempt to shrink to minimal failing example
			data.Freeze()
			failingSeq := data.Sequence()

			// Shrink the failing example
			shrinkTest := func(d *conjecture.ConjectureData) bool {
				testT := &T{
					Data:            d,
					Source:          nil,
					generatedValues: make(map[string]string),
				}
				testFailed, testSkipped := runCheck(check, testT)

				// Check the data status - if it overran during replay, this shrink attempt is invalid
				if d.Status() == conjecture.StatusOverrun {
					return false
				}

				// Only consider it interesting if it failed (not skipped)
				if testFailed && !testSkipped {
					d.MarkInteresting("property failed")
					return true
				}
				return false
			}

			ctx := context.Background()
			shrinkResult := conjecture.Shrink(ctx, failingSeq, shrinkTest,
				conjecture.WithMaxShrinkCalls(cfg.maxShrinks))

			// Report with shrunk example
			reportFailure(t, shrinkResult.Sequence, check, i+1, cfg.maxTests, cfg.seed+int64(i), tested, skipped, shrinkResult.Calls)
			return
		}
	}

	// All tests passed
	// Warn if skip rate is high (>90%) or if no tests ran at all
	if tested == 0 {
		t.Errorf("All %d property test attempts were skipped - no tests actually ran! Check your generators and constraints.",
			skipped)
	} else if tested > 0 {
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
func reportFailure(t TestingT, shrunkSeq *conjecture.ChoiceSequence, check func(*T), attempts, maxTests int, seed int64, tested, skipped int, shrinkCalls int) {
	// Only call Helper() if t is a real *testing.T
	if helper, ok := t.(interface{ Helper() }); ok {
		helper.Helper()
	}

	// Replay the shrunk sequence to get the actual failure and capture generated values
	data := conjecture.ForReplay(shrunkSeq)
	finalT := &T{
		Data:            data,
		Source:          nil,
		generatedValues: make(map[string]string),
	}
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

	// Show generated values that were tracked during replay
	if len(finalT.generatedValues) > 0 {
		msg.WriteString("\nGenerated values:\n")
		// Use draw order for intuitive output
		for _, label := range finalT.generatedValuesOrder {
			msg.WriteString(fmt.Sprintf("  %s = %s\n", label, finalT.generatedValues[label]))
		}
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
