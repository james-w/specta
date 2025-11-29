// This file demonstrates property-based testing failure scenarios using intentionally
// buggy functions. Each Example shows:
// 1. A buggy implementation with a specific type of error
// 2. A property test that catches the bug
// 3. The actual error output with automatically tracked generated values
// 4. How shrinking finds minimal counterexamples
//
// The examples showcase specta's Draw() function which automatically tracks
// generated values and displays them in error messages - even if the user's
// error message doesn't mention them, or if the test panics.

package example_test

import (
	"fmt"
	"sort"

	"github.com/james-w/specta"
	"github.com/james-w/specta/example/factory"
	"github.com/james-w/specta/testlib"
)

// buggyAbs has a sign error - forgets to negate
func buggyAbs(x int64) int64 {
	if x < 0 {
		return x // BUG: should return -x
	}
	return x
}

// Example_absoluteValueNegative demonstrates catching a sign error in abs()
func Example_absoluteValueNegative() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		x := specta.Draw(t, specta.Int().Range(-100, -1), "x")

		// Property: abs(x) should always be non-negative
		specta.AssertThat(t, buggyAbs(x), specta.GreaterThanOrEqual(int64(0)))
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
	//   x = -1
	//
	// Failure:
	//   buggyAbs(x): expected value >= 0 but got -1
	//
	// Reproduce: Property(t, check, Seed(1))
}

// buggyMax returns wrong result when a == b
func buggyMax(a, b int64) int64 {
	if a > b {
		return a
	}
	if b > a {
		return b
	}
	return 0 // BUG: should return a (or b)
}

// Example_maximumEqualValues demonstrates failure on edge case
func Example_maximumEqualValues() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		x := specta.Draw(t, specta.Int().Range(0, 100), "x")

		// Property: max(x, x) should equal x
		specta.AssertThat(t, buggyMax(x, x), specta.Equal(x))
	}, specta.Seed(42), specta.MaxTests(100))

	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}

	// Output:
	//
	// === Property Test Failed ===
	// Seed: 43
	// Attempts: 2/100
	//
	// Generated values:
	//   x = 1
	//
	// Failure:
	//   buggyMax(x, x): expected 1 but got 0
	//
	// Reproduce: Property(t, check, Seed(43))
}

// buggySum starts at wrong initial value
func buggySum(xs []int64) int64 {
	sum := int64(1) // BUG: should start at 0
	for _, x := range xs {
		sum += x
	}
	return sum
}

// Example_sumEmptySlice demonstrates failure on empty input
func Example_sumEmptySlice() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		xs := specta.Draw(t,
			specta.Slice(specta.Int().Range(0, 10)).MinLen(0).MaxLen(0),
			"xs")

		// Property: sum of empty slice should be 0
		specta.AssertThat(t, buggySum(xs), specta.Equal(int64(0)))
	}, specta.Seed(1), specta.MaxTests(1))

	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}

	// Output:
	//
	// === Property Test Failed ===
	// Seed: 1
	// Attempts: 1/1
	//
	// Generated values:
	//   xs = []int64{}
	//
	// Failure:
	//   buggySum(xs): expected 0 but got 1
	//
	// Reproduce: Property(t, check, Seed(1))
}

// buggyIsSorted rejects equal adjacent elements
func buggyIsSorted(xs []int64) bool {
	for i := 1; i < len(xs); i++ {
		if xs[i] < xs[i-1] {
			return false
		}
		if xs[i] == xs[i-1] {
			return false // BUG: equal elements are fine in sorted arrays
		}
	}
	return true
}

// Example_sortedWithDuplicates demonstrates failure when checker rejects valid sorted arrays
func Example_sortedWithDuplicates() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		xs := specta.Draw(t,
			specta.Slice(specta.Int().Range(0, 100)).MinLen(1).MaxLen(10),
			"xs")

		sorted := make([]int64, len(xs))
		copy(sorted, xs)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

		// Property: sorted slice should be recognized as sorted
		specta.AssertThat(t, buggyIsSorted(sorted), specta.Equal(true))
	}, specta.Seed(17), specta.MaxTests(100))

	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}

	// Output:
	//
	// === Property Test Failed ===
	// Seed: 19
	// Attempts: 3/100
	//
	// Generated values:
	//   xs = []int64{
	//     0,
	//     0,
	//   }
	//
	// Failure:
	//   buggyIsSorted(sorted): expected true but got false
	//
	// Reproduce: Property(t, check, Seed(19))
}

// Example_multipleValues demonstrates tracking multiple generated values
func Example_multipleValues() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		a := specta.Draw(t, specta.Int().Range(0, 50), "a")
		b := specta.Draw(t, specta.Int().Range(0, 50), "b")
		c := specta.Draw(t, specta.Int().Range(0, 50), "c")

		// Property: a + b + c should equal sum([a, b, c])
		specta.AssertThat(t, buggySum([]int64{a, b, c}), specta.Equal(a+b+c))
	}, specta.Seed(7), specta.MaxTests(100))

	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}

	// Output:
	//
	// === Property Test Failed ===
	// Seed: 7
	// Attempts: 1/100
	//
	// Generated values:
	//   a = 0
	//   b = 0
	//   c = 0
	//
	// Failure:
	//   buggySum([]int64{a, b, c}): expected 0 but got 1
	//
	// Reproduce: Property(t, check, Seed(7))
}

// buggyContains only checks first element
func buggyContains(xs []int64, target int64) bool {
	if len(xs) > 0 && xs[0] == target {
		return true
	}
	return false // BUG: should check all elements
}

// Example_sliceContains demonstrates failure with slice search
func Example_sliceContains() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		// Generate value first
		value := specta.Draw(t, specta.Int().Range(0, 10), "value")

		// Generate a slice
		xs := specta.Draw(t,
			specta.Slice(specta.Int().Range(0, 10)).MinLen(1).MaxLen(5),
			"xs")

		// Append our value to ensure it's in there
		xs = append(xs, value)

		// Property: contains should find element that exists in the slice
		specta.AssertThat(t, buggyContains(xs, value), specta.Equal(true))
	}, specta.Seed(5), specta.MaxTests(100))

	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}

	// Output:
	//
	// === Property Test Failed ===
	// Seed: 5
	// Attempts: 1/100
	//
	// Generated values:
	//   value = 0
	//   xs = []int64{
	//     1,
	//   }
	//
	// Failure:
	//   buggyContains(xs, value): expected true but got false
	//
	// Reproduce: Property(t, check, Seed(5))
}

// Example_commutativeProperty demonstrates testing mathematical properties
func Example_commutativeProperty() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		x := specta.Draw(t, specta.Int().Range(1, 10), "x")
		y := specta.Draw(t, specta.Int().Range(1, 10), "y")

		// Property: max should be commutative (and also should work correctly!)
		result1 := buggyMax(x, y)
		result2 := buggyMax(y, x)

		// Also check that max(x,y) >= x and max(x,y) >= y
		specta.AssertThat(t, result1, specta.GreaterThanOrEqual(x))
		specta.AssertThat(t, result1, specta.GreaterThanOrEqual(y))

		// Check commutativity
		specta.AssertThat(t, result1, specta.Equal(result2))
	}, specta.Seed(10), specta.MaxTests(100))

	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}

	// Output:
	//
	// === Property Test Failed ===
	// Seed: 10
	// Attempts: 1/100
	//
	// Generated values:
	//   x = 1
	//   y = 1
	//
	// Failure:
	//   result1: expected value >= 1 but got 0
	//   result1: expected value >= 1 but got 0
	//
	// Reproduce: Property(t, check, Seed(10))
}

// Example_structFormatting demonstrates pretty-printed struct output in error messages
func Example_structFormatting() {
	spy := testlib.NewSpy()

	specta.Property(spy, func(t *specta.T) {
		parent := specta.Draw(t, factory.Parent().Gen(), "parent")

		// Deliberately fail to show struct formatting
		specta.AssertThat(t, parent.Child.Name, specta.Equal("deliberately-wrong"))
	}, specta.Seed(42), specta.MaxTests(1))

	if len(spy.Errors) > 0 {
		fmt.Println(spy.Errors[0])
	}

	// Output:
	//
	// === Property Test Failed ===
	// Seed: 42
	// Attempts: 1/1
	//
	// Generated values:
	//   parent = Parent{
	//     Child: UserView{
	//       ID: "a",
	//       Name: "a",
	//       Active: false,
	//       Score: 0,
	//     },
	//   }
	//
	// Failure:
	//   parent.Child.Name: expected "deliberately-wrong" but got "a"
	//
	// Reproduce: Property(t, check, Seed(42))
}
