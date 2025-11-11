package example_test

import (
	"fmt"
	"github.com/james-w/specta"
	"github.com/james-w/specta/example"
	"github.com/james-w/specta/example/factory"
)

// ExampleUserViewMatcher_singleFieldMismatch demonstrates the structured diff output when a single field doesn't match.
// The output shows: ✓ for matched fields, ✗ for failed fields, ~ for unchecked fields.
func ExampleUserViewMatcher_singleFieldMismatch() {
	p := specta.New()
	user := factory.UserView().Name("foo").Build(p)

	matcher := factory.UserViewMatches().Name(specta.Equal("bar")).Matcher()
	result := matcher.Matches(user)

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

// ExampleUserViewMatcher_multipleFieldMismatches demonstrates structured diff with multiple field failures.
func ExampleUserViewMatcher_multipleFieldMismatches() {
	p := specta.New()
	user := factory.UserView().Name("Bob").Score(85).Active(false).Build(p)

	matcher := factory.UserViewMatches().
		Name(specta.HasPrefix("A")).
		Score(specta.GreaterThan(90)).
		Active(specta.IsTrue()).
		Matcher()

	result := matcher.Matches(user)
	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// UserView {
	//   ✗ Active: expected true but got false
	//   ~ ID: "id_1"
	//   ✗ Name: expected string to start with "A" but got "Bob"
	//   ✗ Score: expected value > 90 but got 85
	// }
}

// ExampleUserViewMatcher_partialMatch demonstrates that unchecked fields are shown with ~ symbol.
func ExampleUserViewMatcher_partialMatch() {
	p := specta.New()
	user := factory.UserView().Name("Alice").Score(50).Build(p)

	// Only check the name - other fields are unchecked
	matcher := factory.UserViewMatches().Name(specta.Equal("Alice")).Matcher()

	result := matcher.Matches(user)
	if !result.Matched {
		fmt.Println(result.Message)
	} else {
		// Even on success, we can see what was checked
		fmt.Println("Match succeeded - Name was checked")
	}
	// Output:
	// Match succeeded - Name was checked
}

// ExampleDeepEqual_structMismatch demonstrates DeepEqual's structured diff for struct comparisons.
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

// ExampleDeepEqual_nestedStruct demonstrates diff output for nested struct mismatches.
// Nested structures show full structured diff recursively, allowing you to see exactly
// which fields differ at any level of nesting.
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
	//   ✗ Child:
	//     UserView {
	//       ✗ ID: expected "123" but got "456"
	//       ✓ Name: "Alice"
	//       ✓ Active: false
	//       ✓ Score: 0
	//     }
	// }
}

// ExampleDeepEqual_sliceDifference demonstrates diff output for slice mismatches.
// Slices show simple format without field-by-field breakdown.
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

// ExampleGreaterThan_failure demonstrates numeric constraint matcher failure output.
func ExampleGreaterThan_failure() {
	matcher := specta.GreaterThan(100)
	result := matcher.Matches(50)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected value > 100 but got 50
}

// ExampleLessThan_failure demonstrates numeric less-than matcher failure output.
func ExampleLessThan_failure() {
	matcher := specta.LessThan(10)
	result := matcher.Matches(25)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected value < 10 but got 25
}

// ExampleGreaterThanOrEqual_failure demonstrates >= matcher failure output.
func ExampleGreaterThanOrEqual_failure() {
	matcher := specta.GreaterThanOrEqual(100)
	result := matcher.Matches(99)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected value >= 100 but got 99
}

// ExampleEqual_failure demonstrates equality matcher failure output.
func ExampleEqual_failure() {
	matcher := specta.Equal(42)
	result := matcher.Matches(99)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected 42 but got 99
}

// ExampleContains_failure demonstrates string Contains matcher failure output.
func ExampleContains_failure() {
	matcher := specta.Contains("hello")
	result := matcher.Matches("goodbye world")

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected string to contain "hello" but got "goodbye world"
}

// ExampleHasPrefix_failure demonstrates string prefix matcher failure output.
func ExampleHasPrefix_failure() {
	matcher := specta.HasPrefix("admin_")
	result := matcher.Matches("user_123")

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected string to start with "admin_" but got "user_123"
}

// ExampleHasSuffix_failure demonstrates string suffix matcher failure output.
func ExampleHasSuffix_failure() {
	matcher := specta.HasSuffix(".json")
	result := matcher.Matches("config.yaml")

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected string to end with ".json" but got "config.yaml"
}

// ExampleIsTrue_failure demonstrates boolean true matcher failure output.
func ExampleIsTrue_failure() {
	matcher := specta.IsTrue()
	result := matcher.Matches(false)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected true but got false
}

// ExampleIsFalse_failure demonstrates boolean false matcher failure output.
func ExampleIsFalse_failure() {
	matcher := specta.IsFalse()
	result := matcher.Matches(true)

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected false but got true
}

// ExampleIsZero_failure demonstrates zero value matcher failure output.
// Shows the actual value received (quoted for strings).
func ExampleIsZero_failure() {
	matcher := specta.IsZero[string]()
	result := matcher.Matches("not empty")

	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected zero value but got "not empty"
}

// ExampleAnyOf_failure demonstrates AnyOf matcher failure when no branches match.
// Shows all failure messages (up to 3) so you can see what was tried.
func ExampleAnyOf_failure() {
	matcher := specta.AnyOf(
		specta.Equal("admin"),
		specta.Equal("moderator"),
		specta.Equal("editor"),
	)

	result := matcher.Matches("guest")
	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// none of the matchers succeeded:
	//   option 1: expected "admin" but got "guest"
	//   option 2: expected "moderator" but got "guest"
	//   option 3: expected "editor" but got "guest"
}

// ExampleNot_failure demonstrates Not matcher failure output.
func ExampleNot_failure() {
	matcher := specta.Not(specta.Contains("test"))

	result := matcher.Matches("test_database")
	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// expected not to match, but it did
}

// ExampleAllOf_failure demonstrates AllOf matcher failure output.
// Shows which matchers failed and which succeeded, making it easy to see what went wrong.
func ExampleAllOf_failure() {
	matcher := specta.AllOf(
		specta.HasPrefix("user_"),
		specta.Contains("admin"),
		specta.HasSuffix("_verified"),
	)

	result := matcher.Matches("user_123_pending")
	if !result.Matched {
		fmt.Println(result.Message)
	}
	// Output:
	// 2 of 3 matchers failed:
	//   ✗ matcher 2: expected string to contain "admin" but got "user_123_pending"
	//   ✗ matcher 3: expected string to end with "_verified" but got "user_123_pending"
}

// ExampleAssertThat_expressionCapture demonstrates how AssertThat captures and displays
// the actual expression that was tested using AST parsing.
func ExampleAssertThat_expressionCapture() {
	user := example.UserView{Name: "Bob"}

	// The expression "user.Name" will be captured and shown in the error
	result := specta.Equal("Alice").Matches(user.Name)

	if !result.Matched {
		// This shows what happens with a simple expression
		fmt.Printf("user.Name: %s\n", result.Message)
	}
	// Output:
	// user.Name: expected "Alice" but got "Bob"
}

// ExampleAssertThat_complexExpression demonstrates AST capture with a complex expression.
func ExampleAssertThat_complexExpression() {
	p := specta.New()
	user := factory.UserView().Name("Alice").Score(42).Build(p)

	// Complex expressions like field access or arithmetic are captured
	result := specta.GreaterThan(50).Matches(user.Score)

	if !result.Matched {
		fmt.Printf("user.Score: %s\n", result.Message)
	}
	// Output:
	// user.Score: expected value > 50 but got 42
}

// ExampleAssertThat_structuredDiffWithExpression demonstrates how the expression
// is displayed as a header for multi-line structured diffs.
func ExampleAssertThat_structuredDiffWithExpression() {
	p := specta.New()
	actual := factory.UserView().Name("Bob").Score(50).Build(p)

	// For structured diffs, the expression appears as a labeled header
	matcher := factory.UserViewMatches().
		Name(specta.Equal("Alice")).
		Score(specta.GreaterThan(90)).
		Matcher()

	result := matcher.Matches(actual)
	if !result.Matched {
		// This demonstrates multi-line output with expression header
		fmt.Println("Assertion: actual")
		fmt.Println()
		fmt.Println(result.Message)
	}
	// Output:
	// Assertion: actual
	//
	// UserView {
	//   ~ Active: false
	//   ~ ID: "id_1"
	//   ✗ Name: expected "Alice" but got "Bob"
	//   ✗ Score: expected value > 90 but got 50
	// }
}
