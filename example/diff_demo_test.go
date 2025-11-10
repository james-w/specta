package example_test

import (
	"testing"

	"github.com/james-w/gomatchers"
	"github.com/james-w/gomatchers/example"
	"github.com/james-w/gomatchers/example/factory"
)

// TestStructuredDiffDemo demonstrates the new structured diff output
func TestStructuredDiffDemo(t *testing.T) {
	t.Run("DeepEqual struct diff", func(t *testing.T) {
		expected := example.UserView{
			ID:     "user-123",
			Name:   "Alice",
			Active: true,
			Score:  100,
		}
		actual := example.UserView{
			ID:     "user-123",
			Name:   "Bob",
			Active: true,
			Score:  50,
		}

		matcher := gomatchers.DeepEqual(expected)
		result := matcher.Matches(actual)

		// This should fail and show structured diff
		t.Logf("DeepEqual structured diff output:\n%s", result.Message)
	})

	t.Run("Generated matcher struct diff", func(t *testing.T) {
		p := gomatchers.New()
		actual := factory.UserView().Build(p)

		// Create a matcher that will fail on some fields
		matcher := factory.UserViewMatches().
			Name(gomatchers.Equal("WrongName")).
			Score(gomatchers.GreaterThan(1000)).
			Matcher()

		result := matcher.Matches(actual)

		// This should fail and show structured diff
		t.Logf("Generated matcher structured diff output:\n%s", result.Message)
	})

	t.Run("All fields match", func(t *testing.T) {
		p := gomatchers.New()
		user := factory.UserView().Name("Alice").Score(100).Build(p)

		matcher := factory.UserViewMatches().
			Name(gomatchers.Equal("Alice")).
			Score(gomatchers.Equal(100)).
			Matcher()

		result := matcher.Matches(user)

		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}
