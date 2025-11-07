package showcase_test

import (
	"testing"

	"github.com/james-w/gomatchers"
	"github.com/james-w/gomatchers/showcase"
	"github.com/james-w/gomatchers/showcase/factory"
)

func TestAsEqualMatcher(t *testing.T) {
	p := gomatchers.New()

	t.Run("partial matching - only checks set fields", func(t *testing.T) {
		// Recipe only sets FirstName
		template := factory.User().FirstName("Alice")
		matcher := template.AsEqualMatcher()

		// Should match any user with FirstName="Alice", regardless of other fields
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

		// Should NOT match if FirstName is different
		user3 := factory.User().FirstName("Bob").Email("alice@example.com").Build(p)
		result3 := matcher.Matches(user3)
		if result3.Matched {
			t.Error("Expected mismatch for user3 with different name")
		}
	})

	t.Run("multiple fields - all must match", func(t *testing.T) {
		template := factory.User().FirstName("Alice").Active(true)
		matcher := template.AsEqualMatcher()

		// Matches: both FirstName and Active match
		user1 := factory.User().FirstName("Alice").Active(true).Email("any@example.com").Build(p)
		if !matcher.Matches(user1).Matched {
			t.Error("Expected match when both FirstName and Active match")
		}

		// Doesn't match: FirstName matches but Active doesn't
		user2 := factory.User().FirstName("Alice").Active(false).Build(p)
		if matcher.Matches(user2).Matched {
			t.Error("Expected mismatch when Active doesn't match")
		}

		// Doesn't match: Active matches but FirstName doesn't
		user3 := factory.User().FirstName("Bob").Active(true).Build(p)
		if matcher.Matches(user3).Matched {
			t.Error("Expected mismatch when FirstName doesn't match")
		}
	})

	t.Run("nested partial matching - only checks specified nested fields", func(t *testing.T) {
		// Template only checks Address.City, ignoring other Address fields
		template := factory.User().
			AddressFromRecipe(factory.Address().City("Boston"))
		matcher := template.AsEqualMatcher()

		// Should match any user with Address.City="Boston", regardless of other Address fields
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
			for _, detail := range result2.Details {
				t.Errorf("  %s", detail)
			}
		}

		// Should NOT match if Address.City is different
		user3 := factory.User().
			FirstName("Charlie").
			AddressFromRecipe(factory.Address().City("NYC").State("NY")).
			Build(p)
		result3 := matcher.Matches(user3)
		if result3.Matched {
			t.Error("Expected mismatch for user3 with different city")
		}
	})

	t.Run("nested partial matching with multiple nested fields", func(t *testing.T) {
		// Template checks both Address.City and Address.State
		template := factory.User().
			AddressFromRecipe(factory.Address().City("Boston").State("MA"))
		matcher := template.AsEqualMatcher()

		// Matches: both City and State match
		user1 := factory.User().
			AddressFromRecipe(factory.Address().City("Boston").State("MA").Street("Any Street")).
			Build(p)
		if !matcher.Matches(user1).Matched {
			t.Error("Expected match when both City and State match")
		}

		// Doesn't match: City matches but State doesn't
		user2 := factory.User().
			AddressFromRecipe(factory.Address().City("Boston").State("CA")).
			Build(p)
		if matcher.Matches(user2).Matched {
			t.Error("Expected mismatch when State doesn't match")
		}
	})

	t.Run("nested partial matching combined with top-level fields", func(t *testing.T) {
		// Template checks User.Active AND Address.City
		template := factory.User().
			Active(true).
			AddressFromRecipe(factory.Address().City("Boston"))
		matcher := template.AsEqualMatcher()

		// Matches: both Active and Address.City match
		user1 := factory.User().
			Active(true).
			FirstName("Alice").
			AddressFromRecipe(factory.Address().City("Boston").State("MA")).
			Build(p)
		if !matcher.Matches(user1).Matched {
			t.Error("Expected match when both Active and City match")
		}

		// Doesn't match: City matches but Active doesn't
		user2 := factory.User().
			Active(false).
			AddressFromRecipe(factory.Address().City("Boston")).
			Build(p)
		if matcher.Matches(user2).Matched {
			t.Error("Expected mismatch when Active doesn't match")
		}

		// Doesn't match: Active matches but City doesn't
		user3 := factory.User().
			Active(true).
			AddressFromRecipe(factory.Address().City("NYC")).
			Build(p)
		if matcher.Matches(user3).Matched {
			t.Error("Expected mismatch when City doesn't match")
		}
	})

	t.Run("empty recipe matches everything", func(t *testing.T) {
		// Recipe with no fields set
		template := factory.User()
		matcher := template.AsEqualMatcher()

		// Should match any user since no fields are being checked
		user1 := factory.User().FirstName("Alice").Build(p)
		user2 := factory.User().FirstName("Bob").Active(true).Build(p)

		if !matcher.Matches(user1).Matched {
			t.Error("Expected empty recipe to match user1")
		}
		if !matcher.Matches(user2).Matched {
			t.Error("Expected empty recipe to match user2")
		}
	})

	t.Run("use case: template matching in loop", func(t *testing.T) {
		// Define a template for active users
		activeUserTemplate := factory.User().Active(true)
		matcher := activeUserTemplate.AsEqualMatcher()

		// Create various users
		users := []struct {
			user          showcase.User
			shouldMatch   bool
			description   string
		}{
			{factory.User().Active(true).FirstName("Alice").Build(p), true, "active user"},
			{factory.User().Active(true).Email("bob@example.com").Build(p), true, "another active user"},
			{factory.User().Active(false).FirstName("Charlie").Build(p), false, "inactive user"},
		}

		for _, tc := range users {
			result := matcher.Matches(tc.user)
			if result.Matched != tc.shouldMatch {
				t.Errorf("%s: expected match=%v but got match=%v", tc.description, tc.shouldMatch, result.Matched)
			}
		}
	})

	t.Run("use case: verify factory produces expected template", func(t *testing.T) {
		// Define what we expect from the factory
		expected := factory.User().FirstName("TestUser").Active(false)

		// Build from factory
		actual := expected.Build(p)

		// Verify it matches the template
		gomatchers.AssertThat(t, actual, expected.AsEqualMatcher())
	})

	t.Run("with AssertThat for concise tests", func(t *testing.T) {
		user := factory.User().
			Email("test@example.com").
			Active(true).
			Build(p)

		// Concise partial verification - only check Email
		template := factory.User().Email("test@example.com")
		gomatchers.AssertThat(t, user, template.AsEqualMatcher())
	})
}

func TestAsEqualMatcherWithAddress(t *testing.T) {
	p := gomatchers.New()

	t.Run("address partial matching", func(t *testing.T) {
		// Only check city
		template := factory.Address().City("Boston")
		matcher := template.AsEqualMatcher()

		addr1 := factory.Address().City("Boston").State("MA").Street("123 Main").Build(p)
		addr2 := factory.Address().City("Boston").State("CA").Street("456 Oak").Build(p)
		addr3 := factory.Address().City("NYC").State("NY").Build(p)

		if !matcher.Matches(addr1).Matched {
			t.Error("Expected match for addr1")
		}
		if !matcher.Matches(addr2).Matched {
			t.Error("Expected match for addr2")
		}
		if matcher.Matches(addr3).Matched {
			t.Error("Expected mismatch for addr3")
		}
	})

	t.Run("multiple address fields", func(t *testing.T) {
		template := factory.Address().City("Boston").State("MA")
		matcher := template.AsEqualMatcher()

		addr1 := factory.Address().City("Boston").State("MA").Build(p)
		addr2 := factory.Address().City("Boston").State("CA").Build(p)

		if !matcher.Matches(addr1).Matched {
			t.Error("Expected match when both City and State match")
		}
		if matcher.Matches(addr2).Matched {
			t.Error("Expected mismatch when State doesn't match")
		}
	})
}

func TestAsEqualMatcherWithProduct(t *testing.T) {
	p := gomatchers.New()

	t.Run("numeric field matching", func(t *testing.T) {
		// Template with specific price
		template := factory.Product().Price(19.99)
		matcher := template.AsEqualMatcher()

		product1 := factory.Product().Price(19.99).Name("Widget").Build(p)
		product2 := factory.Product().Price(29.99).Name("Widget").Build(p)

		if !matcher.Matches(product1).Matched {
			t.Error("Expected match for product with correct price")
		}
		if matcher.Matches(product2).Matched {
			t.Error("Expected mismatch for product with wrong price")
		}
	})

	t.Run("boolean field matching", func(t *testing.T) {
		template := factory.Product().InStock(true)
		matcher := template.AsEqualMatcher()

		product1 := factory.Product().InStock(true).Build(p)
		product2 := factory.Product().InStock(false).Build(p)

		if !matcher.Matches(product1).Matched {
			t.Error("Expected match for in-stock product")
		}
		if matcher.Matches(product2).Matched {
			t.Error("Expected mismatch for out-of-stock product")
		}
	})
}

func TestAsEqualMatcherComposition(t *testing.T) {
	p := gomatchers.New()

	t.Run("progressive refinement", func(t *testing.T) {
		// Start with broad template
		activeUsers := factory.User().Active(true)

		// Can narrow down later
		activeCompanyUsers := factory.User().Active(true).Email("@company.com")

		user1 := factory.User().Active(true).Email("@company.com").Build(p)
		user2 := factory.User().Active(true).Email("@example.com").Build(p)

		// Both match the broad template
		if !activeUsers.AsEqualMatcher().Matches(user1).Matched {
			t.Error("Expected user1 to match activeUsers")
		}
		if !activeUsers.AsEqualMatcher().Matches(user2).Matched {
			t.Error("Expected user2 to match activeUsers")
		}

		// Only user1 matches the narrow template
		if !activeCompanyUsers.AsEqualMatcher().Matches(user1).Matched {
			t.Error("Expected user1 to match activeCompanyUsers")
		}
		if activeCompanyUsers.AsEqualMatcher().Matches(user2).Matched {
			t.Error("Expected user2 NOT to match activeCompanyUsers")
		}
	})
}

func TestAsEqualMatcherDetailedErrors(t *testing.T) {
	p := gomatchers.New()

	t.Run("provides clear error messages", func(t *testing.T) {
		template := factory.User().FirstName("Alice").Email("alice@example.com")
		matcher := template.AsEqualMatcher()

		user := factory.User().FirstName("Bob").Email("bob@example.com").Build(p)

		result := matcher.Matches(user)
		if result.Matched {
			t.Error("Expected mismatch")
		}

		// Should have details about which fields didn't match
		if len(result.Details) == 0 {
			t.Error("Expected detailed error messages")
		}
		// Message should mention both FirstName and Email failures
		hasFirstNameError := false
		hasEmailError := false
		for _, detail := range result.Details {
			if detail == "FirstName: expected Alice but got Bob" {
				hasFirstNameError = true
			}
			if detail == "Email: expected alice@example.com but got bob@example.com" {
				hasEmailError = true
			}
		}
		if !hasFirstNameError {
			t.Error("Expected error detail about FirstName mismatch")
		}
		if !hasEmailError {
			t.Error("Expected error detail about Email mismatch")
		}
	})
}
