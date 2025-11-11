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

// ExampleAllOf demonstrates combining multiple matchers that all must pass.
func ExampleAllOf() {
	p := specta.New()
	user := factory.UserView().Name("Alice").Score(95).Active(true).Build(p)

	matcher := factory.UserViewMatches().
		Name(specta.HasPrefix("A")).
		Score(specta.GreaterThan(90)).
		Active(specta.IsTrue()).
		Matcher()

	result := matcher.Matches(user)
	fmt.Println(result.Matched)
	// Output:
	// true
}

// ExampleAllOf_failure demonstrates AllOf output when one matcher fails.
func ExampleAllOf_failure() {
	p := specta.New()
	user := factory.UserView().Name("Bob").Score(85).Active(true).Build(p)

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
	//   ✓ Active: true
	//   ~ ID: "id_1"
	//   ✗ Name: expected string to start with "A" but got "Bob"
	//   ✗ Score: expected value > 90 but got 85
	// }
}

// ExampleAnyOf demonstrates a matcher that passes if any condition matches.
func ExampleAnyOf() {
	matcher := specta.AnyOf(
		specta.Equal("admin"),
		specta.Equal("moderator"),
		specta.Equal("editor"),
	)

	result := matcher.Matches("editor")
	fmt.Println(result.Matched)
	// Output:
	// true
}

// ExampleNot demonstrates negating a matcher.
func ExampleNot() {
	matcher := specta.Not(specta.Contains("test"))

	result := matcher.Matches("production-db")
	fmt.Println(result.Matched)
	// Output:
	// true
}

// ExampleHasPrefix demonstrates string prefix matching.
func ExampleHasPrefix() {
	matcher := specta.HasPrefix("user_")
	result := matcher.Matches("user_12345")
	fmt.Println(result.Matched)
	// Output:
	// true
}

// ExampleHasSuffix demonstrates string suffix matching.
func ExampleHasSuffix() {
	matcher := specta.HasSuffix(".json")
	result := matcher.Matches("config.json")
	fmt.Println(result.Matched)
	// Output:
	// true
}

// ExampleLessThan demonstrates numeric less-than matching.
func ExampleLessThan() {
	matcher := specta.LessThan(100)
	result := matcher.Matches(50)
	fmt.Println(result.Matched)
	// Output:
	// true
}

// ExampleGreaterThanOrEqual demonstrates numeric greater-than-or-equal matching.
func ExampleGreaterThanOrEqual() {
	matcher := specta.GreaterThanOrEqual(18)
	result1 := matcher.Matches(18)
	result2 := matcher.Matches(25)
	fmt.Println(result1.Matched, result2.Matched)
	// Output:
	// true true
}

// ExampleFactory_buildingTestData demonstrates using factories to create test data.
func ExampleFactory_buildingTestData() {
	p := specta.New()

	// Build a single user with custom values
	user1 := factory.UserView().
		Name("Alice").
		Active(true).
		Build(p)

	// Build multiple users with unique generated values
	users := factory.UserView().Many(3, p)

	fmt.Printf("Single user: %s (Active: %v)\n", user1.Name, user1.Active)
	fmt.Printf("Generated %d users\n", len(users))
	fmt.Printf("First generated user ID: %s\n", users[0].ID)
	// Output:
	// Single user: Alice (Active: true)
	// Generated 3 users
	// First generated user ID: id_3
}

// ExampleFactory_partialMatching demonstrates matching only specific fields.
func ExampleFactory_partialMatching() {
	p := specta.New()
	user := factory.UserView().Name("Charlie").Score(75).Build(p)

	// Only check the name - don't care about other fields
	matcher := factory.UserViewMatches().Name(specta.Equal("Charlie")).Matcher()

	result := matcher.Matches(user)
	fmt.Println(result.Matched)
	// Output:
	// true
}

// ExampleFactory_nestedStructs demonstrates building nested structures.
func ExampleFactory_nestedStructs() {
	p := specta.New()

	// Build a parent with a custom child
	parent := factory.Parent().
		ChildFromRecipe(
			factory.UserView().Name("Nested User").Score(100),
		).
		Build(p)

	fmt.Printf("Parent's child name: %s, score: %d\n", parent.Child.Name, parent.Child.Score)
	// Output:
	// Parent's child name: Nested User, score: 100
}

// ExampleIsZero demonstrates matching zero values.
func ExampleIsZero() {
	var emptyString string
	var zero int

	matcher1 := specta.IsZero[string]()
	matcher2 := specta.IsZero[int]()

	result1 := matcher1.Matches(emptyString)
	result2 := matcher2.Matches(zero)

	fmt.Println(result1.Matched, result2.Matched)
	// Output:
	// true true
}

// ExampleIsTrue_IsFalse demonstrates boolean matchers.
func ExampleIsTrue_IsFalse() {
	trueVal := true
	falseVal := false

	matcherTrue := specta.IsTrue()
	matcherFalse := specta.IsFalse()

	result1 := matcherTrue.Matches(trueVal)
	result2 := matcherFalse.Matches(falseVal)

	fmt.Println(result1.Matched, result2.Matched)
	// Output:
	// true true
}

// ExampleComposition demonstrates composing matchers for complex validation.
func ExampleComposition() {
	p := specta.New()
	user := factory.UserView().
		Name("Administrator").
		Score(100).
		Active(true).
		Build(p)

	// Compose multiple requirements
	matcher := factory.UserViewMatches().
		Name(specta.AllOf(
			specta.HasPrefix("Admin"),
			specta.Contains("istrator"),
		)).
		Score(specta.GreaterThanOrEqual(100)).
		Active(specta.IsTrue()).
		Matcher()

	result := matcher.Matches(user)
	fmt.Println(result.Matched)
	// Output:
	// true
}
