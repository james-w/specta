package example_test

import (
	"fmt"
	"github.com/james-w/specta"
	"github.com/james-w/specta/example"
	"github.com/james-w/specta/example/factory"
)

// ExampleUserViewMatcher_fieldMismatch demonstrates the diff output when a single field doesn't match.
func ExampleUserViewMatcher_fieldMismatch() {
	p := specta.New()
	child := factory.UserView().Name("foo").Build(p)

	matcher := factory.UserViewMatches().Name(specta.Equal("bar")).Matcher()
	result := matcher.Matches(child)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// UserView {
	//   ~ Active: false
	//   ~ ID: "id_1"
	//   ✗ Name: expected "bar" but got "foo"
	//   ~ Score: 2
	// }
}

// ExampleDeepEqual_structMismatch demonstrates the diff output for DeepEqual with multiple field differences.
func ExampleDeepEqual_structMismatch() {
	expected := example.UserView{
		ID:     "123",
		Name:   "Alice",
		Active: true,
		Score:  100,
	}

	actual := example.UserView{
		ID:     "123",
		Name:   "Bob",
		Active: false,
		Score:  85,
	}

	matcher := specta.DeepEqual(expected)
	result := matcher.Matches(actual)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// UserView {
	//   ✓ ID: "123"
	//   ✗ Name: expected "Alice" but got "Bob"
	//   ✗ Active: expected true but got false
	//   ✗ Score: expected 100 but got 85
	// }
}

// ExampleDeepEqual_nestedStruct demonstrates the diff output for nested struct mismatches.
func ExampleDeepEqual_nestedStruct() {
	expected := example.Parent{
		Child: example.UserView{
			ID:   "123",
			Name: "Alice",
		},
	}

	actual := example.Parent{
		Child: example.UserView{
			ID:   "456",
			Name: "Alice",
		},
	}

	matcher := specta.DeepEqual(expected)
	result := matcher.Matches(actual)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// Parent {
	//   ✗ Child: expected UserView{...} but got UserView{...}
	// }
}

// ExampleDeepEqual_sliceDifference demonstrates the diff output for slice mismatches.
func ExampleDeepEqual_sliceDifference() {
	expected := []string{"apple", "banana", "cherry"}
	actual := []string{"apple", "blueberry", "cherry"}

	matcher := specta.DeepEqual(expected)
	result := matcher.Matches(actual)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected []string{"apple", "banana", "cherry"} but got []string{"apple", "blueberry", "cherry"}
}

// ExampleGreaterThan_failure demonstrates the output when a GreaterThan matcher fails.
func ExampleGreaterThan_failure() {
	matcher := specta.GreaterThan(100)
	result := matcher.Matches(50)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected value > 100 but got 50
}

// ExampleContains_failure demonstrates the output when a Contains matcher fails.
func ExampleContains_failure() {
	matcher := specta.Contains("hello")
	result := matcher.Matches("goodbye world")

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected string to contain "hello" but got "goodbye world"
}

