package showcase_test

import (
	"testing"

	"github.com/james-w/gomatchers"
	"github.com/james-w/gomatchers/showcase/factory"
)

// TestIntegrationExamples demonstrates using factories and matchers together
func TestIntegrationExamples(t *testing.T) {
	p := gomatchers.New()

	t.Run("verify factory output with matchers", func(t *testing.T) {
		// Build a user with factory
		user := factory.User().
			Email("alice@example.com").
			Active(true).
			Build(p)

		// Verify with matcher
		gomatchers.AssertThat(t, user,
			factory.UserMatches().
				Email(gomatchers.Contains("@example.com")).
				Active(gomatchers.IsTrue()).
				Matcher())
	})

	t.Run("partial matching with AsEqualMatcher", func(t *testing.T) {
		// Create a template that only checks specific fields
		template := factory.User().Active(true)
		matcher := template.AsEqualMatcher()

		// Build various users and check they match
		user1 := factory.User().Active(true).Email("alice@example.com").Build(p)
		user2 := factory.User().Active(true).Email("bob@example.com").Build(p)

		gomatchers.AssertThat(t, user1, matcher)
		gomatchers.AssertThat(t, user2, matcher)
	})

	t.Run("nested partial matching", func(t *testing.T) {
		// Template checks only Address.City
		template := factory.User().
			AddressFromRecipe(factory.Address().City("Boston"))
		matcher := template.AsEqualMatcher()

		// User in Boston matches, regardless of other address fields
		user := factory.User().
			FirstName("Alice").
			AddressFromRecipe(factory.Address().City("Boston").State("MA")).
			Build(p)

		gomatchers.AssertThat(t, user, matcher)
	})

	t.Run("template matching pattern", func(t *testing.T) {
		// Define a reusable template
		activeUserTemplate := factory.User().Active(true)
		matcher := activeUserTemplate.AsEqualMatcher()

		// Test that active users match
		activeUser1 := factory.User().Active(true).Email("alice@example.com").Build(p)
		activeUser2 := factory.User().Active(true).FirstName("Bob").Build(p)

		if !matcher.Matches(activeUser1).Matched {
			t.Error("Expected active user 1 to match")
		}
		if !matcher.Matches(activeUser2).Matched {
			t.Error("Expected active user 2 to match")
		}

		// Test that inactive user doesn't match
		inactiveUser := factory.User().Active(false).Build(p)
		if matcher.Matches(inactiveUser).Matched {
			t.Error("Expected inactive user not to match")
		}
	})

	t.Run("combining factory and explicit matchers", func(t *testing.T) {
		// Build an order with nested user
		order := factory.Order().
			UserFromRecipe(factory.User().Email("premium@example.com")).
			Total(250.0).
			Status("shipped").
			Build(p)

		// Use explicit matchers for flexible assertions
		gomatchers.AssertThat(t, order,
			factory.OrderMatches().
				Total(gomatchers.GreaterThan(200.0)).
				Status(gomatchers.Equal("shipped")).
				UserMatches(
					factory.UserMatches().
						Email(gomatchers.Contains("premium")),
				).
				Matcher())
	})
}

// TestAsEqualMatcherBasics demonstrates core AsEqualMatcher functionality
func TestAsEqualMatcherBasics(t *testing.T) {
	p := gomatchers.New()

	t.Run("partial matching - only checks set fields", func(t *testing.T) {
		template := factory.User().FirstName("Alice")
		matcher := template.AsEqualMatcher()

		// Matches any user with FirstName="Alice"
		user1 := factory.User().FirstName("Alice").Email("alice@example.com").Active(true).Build(p)
		user2 := factory.User().FirstName("Alice").Email("different@example.com").Active(false).Build(p)

		result1 := matcher.Matches(user1)
		if !result1.Matched {
			t.Errorf("Expected match for user1 but got: %s", result1.Message)
		}

		result2 := matcher.Matches(user2)
		if !result2.Matched {
			t.Errorf("Expected match for user2 but got: %s", result2.Message)
		}

		// Doesn't match different name
		user3 := factory.User().FirstName("Bob").Email("alice@example.com").Build(p)
		result3 := matcher.Matches(user3)
		if result3.Matched {
			t.Error("Expected mismatch for user3 with different name")
		}
	})

	t.Run("multiple fields - all must match", func(t *testing.T) {
		template := factory.User().FirstName("Alice").Active(true)
		matcher := template.AsEqualMatcher()

		user1 := factory.User().FirstName("Alice").Active(true).Email("any@example.com").Build(p)
		if !matcher.Matches(user1).Matched {
			t.Error("Expected match when both FirstName and Active match")
		}

		user2 := factory.User().FirstName("Alice").Active(false).Build(p)
		if matcher.Matches(user2).Matched {
			t.Error("Expected mismatch when Active doesn't match")
		}
	})

	t.Run("nested partial matching", func(t *testing.T) {
		template := factory.User().
			AddressFromRecipe(factory.Address().City("Boston"))
		matcher := template.AsEqualMatcher()

		user1 := factory.User().
			FirstName("Alice").
			AddressFromRecipe(factory.Address().City("Boston").State("MA").Street("123 Main")).
			Build(p)
		result1 := matcher.Matches(user1)
		if !result1.Matched {
			t.Errorf("Expected match for user1 but got: %s", result1.Message)
			for _, detail := range result1.Details {
				t.Errorf("  %s", detail)
			}
		}

		user2 := factory.User().
			FirstName("Bob").
			AddressFromRecipe(factory.Address().City("Boston").State("CA").ZipCode("90210")).
			Build(p)
		result2 := matcher.Matches(user2)
		if !result2.Matched {
			t.Errorf("Expected match for user2 but got: %s", result2.Message)
		}

		user3 := factory.User().
			FirstName("Charlie").
			AddressFromRecipe(factory.Address().City("NYC").State("NY")).
			Build(p)
		result3 := matcher.Matches(user3)
		if result3.Matched {
			t.Error("Expected mismatch for user3 with different city")
		}
	})

	t.Run("combined top-level and nested matching", func(t *testing.T) {
		template := factory.User().
			Active(true).
			AddressFromRecipe(factory.Address().City("Boston"))
		matcher := template.AsEqualMatcher()

		user1 := factory.User().
			Active(true).
			FirstName("Alice").
			AddressFromRecipe(factory.Address().City("Boston").State("MA")).
			Build(p)
		if !matcher.Matches(user1).Matched {
			t.Error("Expected match when both Active and City match")
		}

		user2 := factory.User().
			Active(false).
			AddressFromRecipe(factory.Address().City("Boston")).
			Build(p)
		if matcher.Matches(user2).Matched {
			t.Error("Expected mismatch when Active doesn't match")
		}

		user3 := factory.User().
			Active(true).
			AddressFromRecipe(factory.Address().City("NYC")).
			Build(p)
		if matcher.Matches(user3).Matched {
			t.Error("Expected mismatch when City doesn't match")
		}
	})

	t.Run("empty recipe matches everything", func(t *testing.T) {
		template := factory.User()
		matcher := template.AsEqualMatcher()

		user1 := factory.User().FirstName("Alice").Build(p)
		user2 := factory.User().FirstName("Bob").Active(true).Build(p)

		if !matcher.Matches(user1).Matched {
			t.Error("Expected empty recipe to match user1")
		}
		if !matcher.Matches(user2).Matched {
			t.Error("Expected empty recipe to match user2")
		}
	})
}
