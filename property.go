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
	Source *Source // Exported so tests can construct T instances
}

// propertyFailure is a sentinel panic value used to stop property test iterations.
type propertyFailure struct{}

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

// Primitives returns a Primitives implementation backed by this property test's Source.
// This is a convenience method for using existing factories in property tests.
//
// Example:
//
//	Property(t, func(t *T) {
//	    user := UserRecipe{}.Build(t.Primitives())
//	    // Test properties of user...
//	})
func (t *T) Primitives(opts ...PrimitivesOption) Primitives {
	return NewPropertyPrimitives(t.Source, opts...)
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
	for i := 0; i < cfg.maxTests; i++ {
		pt := &T{
			Source: NewSource(cfg.seed + int64(i)),
		}

		failed := runCheck(check, pt)

		if failed {
			// Property failed - shrink it
			shrunkData := shrink(pt.Source.Data(), func(data []byte) bool {
				shrinkT := &T{
					Source: NewSourceFromData(data),
				}
				return runCheck(check, shrinkT)
			}, cfg.maxShrinks)

			// Report failure
			reportFailure(t, shrunkData, check, i+1, cfg.maxTests, cfg.seed+int64(i))
			return
		}
	}

	// All tests passed - TestingT doesn't have Logf, so we skip logging success
	// Users can enable verbose mode if they want to see this
}

// runCheck executes the property check function and returns whether it failed.
// It recovers from propertyFailure panics but re-panics on unexpected panics.
func runCheck(check func(*T), pt *T) (failed bool) {
	defer func() {
		if r := recover(); r != nil {
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

// shrink attempts to find a minimal failing test case by trying progressively smaller byte streams.
// It uses a greedy approach with two strategies:
// 1. Binary search on byte stream length
// 2. Zero out individual bytes
func shrink(original []byte, test func([]byte) bool, maxTries int) []byte {
	if len(original) == 0 {
		return original
	}

	best := make([]byte, len(original))
	copy(best, original)
	tries := 0

	// Strategy 1: Binary search on length
	// Try progressively shorter prefixes of the byte stream
	for length := len(best) / 2; length > 0; length /= 2 {
		if tries >= maxTries {
			break
		}
		candidate := best[:length]
		if test(candidate) {
			// Still fails with shorter stream - keep it
			newBest := make([]byte, length)
			copy(newBest, candidate)
			best = newBest
			tries++
			continue
		}
		tries++
	}

	// Strategy 2: Zero out bytes
	// Try replacing individual bytes with 0x00
	for i := len(best) - 1; i >= 0; i-- {
		if tries >= maxTries {
			break
		}
		if best[i] == 0 {
			continue // Already zero
		}

		candidate := make([]byte, len(best))
		copy(candidate, best)
		candidate[i] = 0

		if test(candidate) {
			// Still fails with this byte zeroed - keep it
			best = candidate
		}
		tries++
	}

	return best
}

// reportFailure creates a detailed error message and fails the test.
func reportFailure(t TestingT, shrunkData []byte, check func(*T), attempts, maxTests int, seed int64) {
	// Only call Helper() if t is a real *testing.T
	if helper, ok := t.(interface{ Helper() }); ok {
		helper.Helper()
	}

	// Replay with logging to show generated values
	finalT := &T{
		Source: NewSourceFromData(shrunkData),
	}
	finalT.Source.EnableLogging()
	runCheck(check, finalT)

	// Build error message
	var msg strings.Builder
	msg.WriteString("\n=== Property Test Failed ===\n")
	msg.WriteString(fmt.Sprintf("Seed: %d\n", seed))
	msg.WriteString(fmt.Sprintf("Attempts: %d/%d\n", attempts, maxTests))

	// Show generated values
	if log := finalT.Source.Log(); log != "" {
		msg.WriteString("\nGenerated values:\n  ")
		msg.WriteString(strings.TrimSpace(log))
		msg.WriteString("\n")
	}

	// Show failure messages
	if len(finalT.errors) > 0 {
		msg.WriteString("\nFailure:\n")
		for _, err := range finalT.errors {
			msg.WriteString("  ")
			msg.WriteString(err)
			msg.WriteString("\n")
		}
	}

	// Show reproduction instructions
	msg.WriteString(fmt.Sprintf("\nReproduce: Property(t, check, Seed(%d))\n", seed))

	t.Errorf("%s", msg.String())
}
