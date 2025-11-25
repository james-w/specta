package example_test

import (
	"fmt"

	"github.com/james-w/specta"
	"github.com/james-w/specta/example"
	"github.com/james-w/specta/example/factory"
	"github.com/james-w/specta/testlib"
)

// ExampleProperty_simpleFailure demonstrates the output when a property test fails.
// The output shows the seed for reproduction, which values were generated, and the failure message.
func ExampleProperty_simpleFailure() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		x := specta.Int().Range(-100, 100).Draw(t.Source, "x")
		// This property fails when x is negative
		specta.AssertThat(t, x, specta.GreaterThanOrEqual(int64(0)))
	}, specta.Seed(4), specta.MaxTests(100))

	// The property will fail and shrink will try to find the simplest failing case
	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}
	// Output:
	//
	// === Property Test Failed ===
	// Seed: 4
	// Attempts: 1/100
	//
	// Generated values:
	//   Int(x)=-1
	//
	// Failure:
	//   x: expected value >= 0 but got -1
	//
	// Reproduce: Property(t, check, Seed(4))
}

// ExampleProperty_shrinkingDemonstration demonstrates how shrinking finds minimal failing cases.
// When a property fails, specta tries to find the simplest input that still causes the failure.
// Notice how shrinking reduced the values from the range [-1000, 1000] down to 0.
func ExampleProperty_shrinkingDemonstration() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		x := specta.Int().Range(-1000, 1000).Draw(t.Source, "x")
		y := specta.Int().Range(-1000, 1000).Draw(t.Source, "y")

		// This property is false: it claims x*y == y*x only when both are positive
		// Shrinking should find small counterexamples
		if x > 0 && y > 0 {
			specta.AssertThat(t, x*y, specta.Equal(y*x))
		} else {
			t.Fatalf("Expected both values to be positive, but got x=%d, y=%d", x, y)
		}
	}, specta.Seed(1), specta.MaxTests(100))

	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}
	// Output:
	//
	// === Property Test Failed ===
	// Seed: 1
	// Attempts: 1/100
	//
	// Generated values:
	//   Int(x)=0 Int(y)=0
	//
	// Failure:
	//   Expected both values to be positive, but got x=0, y=0
	//
	// Reproduce: Property(t, check, Seed(1))
}

// ExampleProperty_structuredDiff demonstrates property failures with structured matcher output.
// When using generated types and matchers, you get full structured diffs in property failures.
func ExampleProperty_structuredDiff() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		// Generate a user with random score
		score := specta.Int().Range(0, 100).Draw(t.Source, "score")

		user := example.UserView{
			ID:     "user_123",
			Name:   "TestUser",
			Active: true,
			Score:  int(score),
		}

		// Property: all users should have score > 50
		// This will fail for users with low scores, showing structured diff
		matcher := factory.UserViewMatches().
			Score(specta.GreaterThan(50)).
			Active(specta.IsTrue()).
			Matcher()

		specta.AssertThat(t, user, matcher)
	}, specta.Seed(1), specta.MaxTests(100))

	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}
	// Output:
	//
	// === Property Test Failed ===
	// Seed: 1
	// Attempts: 1/100
	//
	// Generated values:
	//   Int(score)=0
	//
	// Failure:
	//
	//   user didn't match:
	//   UserView {
	//     ✓ Active: true
	//     ~ ID: "user_123"
	//     ~ Name: "TestUser"
	//     ✗ Score: expected value > 50 but got 0
	//   }
	//
	// Reproduce: Property(t, check, Seed(1))
}
