package showcase_test

import (
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/showcase/factory"
)

// TestIntegrationExamples demonstrates using factories and matchers together
func TestIntegrationExamples(t *testing.T) {
	p := specta.New()

	t.Run("verify factory output with matchers", func(t *testing.T) {
		// Build a user with factory
		user := factory.User().
			Email("alice@example.com").
			Active(true).
			Build(p)

		// Verify with explicit matchers
		specta.AssertThat(t, user,
			factory.UserMatches().
				Email(specta.Contains("@example.com")).
				Active(specta.IsTrue()).
				Matcher())
	})

	t.Run("nested matching with explicit matchers", func(t *testing.T) {
		// Build user with specific address
		user := factory.User().
			FirstName("Alice").
			AddressFromRecipe(factory.Address().City("Boston").State("MA")).
			Build(p)

		// Verify nested fields - only check what matters
		specta.AssertThat(t, user,
			factory.UserMatches().
				FirstName(specta.Equal("Alice")).
				AddressMatches(
					factory.AddressMatches().
						City(specta.Equal("Boston")),
				).
				Matcher())
	})

	t.Run("flexible constraint matchers", func(t *testing.T) {
		// Build active users with different attributes
		activeUser1 := factory.User().Active(true).Email("alice@example.com").Build(p)
		activeUser2 := factory.User().Active(true).FirstName("Bob").Build(p)

		// Use explicit matcher to check only Active status
		matcher := factory.UserMatches().
			Active(specta.IsTrue()).
			Matcher()

		specta.AssertThat(t, activeUser1, matcher)
		specta.AssertThat(t, activeUser2, matcher)
	})

	t.Run("combining factory and explicit matchers", func(t *testing.T) {
		// Build an order with nested user
		order := factory.Order().
			UserFromRecipe(factory.User().Email("premium@example.com")).
			Total(250.0).
			Status("shipped").
			Build(p)

		// Use explicit matchers for flexible assertions
		specta.AssertThat(t, order,
			factory.OrderMatches().
				Total(specta.GreaterThan(200.0)).
				Status(specta.Equal("shipped")).
				UserMatches(
					factory.UserMatches().
						Email(specta.Contains("premium")),
				).
				Matcher())
	})
}

// TestAsEqualMatcher demonstrates when AsEqualMatcher is useful:
// when you want the matcher to verify the same fields that the recipe/generator
// would set, allowing you to create reusable templates.
func TestAsEqualMatcher(t *testing.T) {
	p := specta.New()

	t.Run("use case: reusable template matcher", func(t *testing.T) {
		// AsEqualMatcher shines when you want to match exactly what the recipe specifies.
		// This is useful for creating reusable test templates where the recipe defines
		// both the generation pattern AND the matching pattern.

		// Define a template: "active Boston users"
		activeBostonTemplate := factory.User().
			Active(true).
			AddressFromRecipe(factory.Address().City("Boston"))

		// Create matcher that checks exactly what the template specifies
		matcher := activeBostonTemplate.AsEqualMatcher()

		// Any user matching the template (active + Boston) should pass,
		// regardless of other fields
		user1 := factory.User().
			Active(true).
			FirstName("Alice").
			AddressFromRecipe(factory.Address().City("Boston").State("MA").Street("123 Main")).
			Build(p)

		user2 := factory.User().
			Active(true).
			Email("bob@example.com").
			AddressFromRecipe(factory.Address().City("Boston").ZipCode("02101")).
			Build(p)

		specta.AssertThat(t, user1, matcher)
		specta.AssertThat(t, user2, matcher)

		// Users not matching the template should fail
		inactiveBoston := factory.User().
			Active(false).
			AddressFromRecipe(factory.Address().City("Boston")).
			Build(p)

		activeNYC := factory.User().
			Active(true).
			AddressFromRecipe(factory.Address().City("NYC")).
			Build(p)

		result1 := matcher.Matches(inactiveBoston)
		if result1.Matched {
			t.Error("Expected mismatch for inactive Boston user")
		}

		result2 := matcher.Matches(activeNYC)
		if result2.Matched {
			t.Error("Expected mismatch for active NYC user")
		}
	})

	t.Run("comparison: AsEqualMatcher vs explicit matchers", func(t *testing.T) {
		// AsEqualMatcher automatically creates Equal matchers for set fields
		template := factory.User().FirstName("Alice").Active(true)
		asEqualMatcher := template.AsEqualMatcher()

		// Equivalent explicit matcher
		explicitMatcher := factory.UserMatches().
			FirstName(specta.Equal("Alice")).
			Active(specta.Equal(true)).
			Matcher()

		// Both should behave the same way
		user := factory.User().FirstName("Alice").Active(true).Email("alice@example.com").Build(p)

		specta.AssertThat(t, user, asEqualMatcher)
		specta.AssertThat(t, user, explicitMatcher)

		// Use explicit matchers when you need different constraints (not just equality)
		constraintMatcher := factory.UserMatches().
			FirstName(specta.Contains("Ali")).  // Substring match instead of exact
			Active(specta.IsTrue()).            // Boolean check
			Matcher()

		specta.AssertThat(t, user, constraintMatcher)
	})

	t.Run("empty recipe matches everything", func(t *testing.T) {
		// An empty recipe has no constraints, so matches any user
		emptyTemplate := factory.User()
		matcher := emptyTemplate.AsEqualMatcher()

		user1 := factory.User().FirstName("Alice").Build(p)
		user2 := factory.User().FirstName("Bob").Active(true).Build(p)

		specta.AssertThat(t, user1, matcher)
		specta.AssertThat(t, user2, matcher)
	})
}
