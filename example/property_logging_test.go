// This file demonstrates using pt.Logf() in property tests for debugging.
// Logs are forwarded to testing.T.Logf() during the final replay of a failing test,
// with correct line attribution to user code (not framework code).
//
// Key behaviors:
// 1. Logs only appear for failing tests (no spam from successful tests)
// 2. Logs show values from the shrunk counterexample (minimal failing case)
// 3. File:line attribution points to user code where pt.Logf() was called

package example_test

import (
	"fmt"

	"github.com/james-w/specta"
	"github.com/james-w/specta/testlib"
)

// buggySqrt incorrectly handles negative inputs
func buggySqrt(x int64) int64 {
	if x < 0 {
		return x // BUG: should return error or abs, but returns negative
	}
	// Simplified: just return the input (not real sqrt, but demonstrates the point)
	return x
}

// Example_propertyLogging demonstrates using pt.Logf() for debugging property tests.
// The logs help understand what values caused the failure, especially when shrinking
// produces minimal counterexamples.
func Example_propertyLogging() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(pt *specta.T) {
		x := specta.Draw(pt, specta.Int().Range(-100, -1), "x")

		// Log the value being tested - this will appear in error output
		pt.Logf("Testing sqrt(%d)", x)

		// Property: sqrt result should always be non-negative
		result := buggySqrt(x)

		pt.Logf("Got result: %d", result)

		specta.AssertThat(pt, result, specta.GreaterThanOrEqual(int64(0)))
	}, specta.Seed(42), specta.MaxTests(100))

	// Print the logs that were forwarded from the shrunk failure
	if len(spy.Logs) > 0 {
		fmt.Println("Logs from failing test:")
		for _, log := range spy.Logs {
			fmt.Printf("  %s\n", log)
		}
		fmt.Println()
	}

	// Print the error
	if len(spy.Errors) > 0 {
		fmt.Println("Error:")
		fmt.Println(spy.Errors[0])
	}

	// Output:
	// Logs from failing test:
	//   Testing sqrt(-1)
	//   Got result: -1
	//
	// Error:
	// === Property Test Failed ===
	// Seed: 42
	// Attempts: 1/100
	//
	// Generated values:
	//   x = -1
	//
	// Failure:
	//   result: expected value >= 0 but got -1
	//
	// Reproduce: Property(t, check, Seed(42))
}

// Example_propertyLoggingMultipleValues demonstrates using pt.Logf() with multiple
// log statements to trace execution flow during property test failures.
func Example_propertyLoggingMultipleValues() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(pt *specta.T) {
		x := specta.Draw(pt, specta.Int().Range(-10, 10), "x")

		pt.Logf("Starting test with x = %d", x)

		// Fail on negative values
		if x < 0 {
			pt.Logf("Found negative value, failing test")
			pt.Fatalf("x should be non-negative, got: %d", x)
		}

		pt.Logf("Test passed for x = %d", x)
	}, specta.Seed(999), specta.MaxTests(100))

	// Print all logs from the shrunk failing case
	if len(spy.Logs) > 0 {
		fmt.Println("Execution trace from shrunk failure:")
		for _, log := range spy.Logs {
			fmt.Printf("  %s\n", log)
		}
	}

	// Output:
	// Execution trace from shrunk failure:
	//   Starting test with x = -1
	//   Found negative value, failing test
}
