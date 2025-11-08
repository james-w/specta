package showcase_test

import (
	"testing"
	"time"

	"github.com/james-w/gomatchers"
	"github.com/james-w/gomatchers/showcase"
	"github.com/james-w/gomatchers/showcase/factory"
)

func TestUserMatcher(t *testing.T) {
	t.Run("matches all specified fields", func(t *testing.T) {
		user := showcase.User{
			Email:     "alice@example.com",
			FirstName: "Alice",
			Active:    true,
		}

		matcher := factory.UserMatches().
			Email(gomatchers.Equal("alice@example.com")).
			FirstName(gomatchers.Equal("Alice")).
			Active(gomatchers.IsTrue())

		result := matcher.Matcher().Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
			for _, detail := range result.Details {
				t.Errorf("  %s", detail)
			}
		}
	})

	t.Run("fails when field doesn't match", func(t *testing.T) {
		user := showcase.User{Email: "alice@example.com"}

		matcher := factory.UserMatches().
			Email(gomatchers.Equal("bob@example.com"))

		result := matcher.Matcher().Matches(user)
		if result.Matched {
			t.Error("Expected mismatch but got match")
		}
		if len(result.Details) == 0 {
			t.Error("Expected failure details")
		}
	})

	t.Run("only checks specified fields", func(t *testing.T) {
		user := showcase.User{
			Email:     "alice@example.com",
			FirstName: "Alice",
		}

		// Only check email, ignore other fields
		matcher := factory.UserMatches().
			Email(gomatchers.Equal("alice@example.com"))

		result := matcher.Matcher().Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("works with string matchers", func(t *testing.T) {
		user := showcase.User{Email: "alice@example.com"}

		matcher := factory.UserMatches().
			Email(gomatchers.Contains("@example.com"))

		result := matcher.Matcher().Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("works with nested matchers", func(t *testing.T) {
		user := showcase.User{
			FirstName: "Alice",
			Address: showcase.Address{
				City: "Boston",
			},
		}

		matcher := factory.UserMatches().
			AddressMatches(
				factory.AddressMatches().City(gomatchers.Equal("Boston")),
			)

		result := matcher.Matcher().Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
			for _, detail := range result.Details {
				t.Errorf("  %s", detail)
			}
		}
	})
}

func TestAddressMatcher(t *testing.T) {
	t.Run("matches address fields", func(t *testing.T) {
		addr := showcase.Address{
			Street: "123 Main St",
			City:   "Springfield",
			State:  "IL",
		}

		matcher := factory.AddressMatches().
			Street(gomatchers.Equal("123 Main St")).
			City(gomatchers.Equal("Springfield")).
			State(gomatchers.Equal("IL"))

		result := matcher.Matcher().Matches(addr)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("works with partial matching", func(t *testing.T) {
		addr := showcase.Address{
			City:    "Boston",
			Street:  "Any Street",
			ZipCode: "02101",
		}

		// Only check city
		matcher := factory.AddressMatches().
			City(gomatchers.Equal("Boston"))

		result := matcher.Matcher().Matches(addr)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}

func TestProductMatcher(t *testing.T) {
	t.Run("matches with numeric matchers", func(t *testing.T) {
		product := showcase.Product{
			Name:    "Widget",
			Price:   19.99,
			InStock: true,
		}

		matcher := factory.ProductMatches().
			Name(gomatchers.Equal("Widget")).
			Price(gomatchers.GreaterThan(10.0)).
			InStock(gomatchers.IsTrue())

		result := matcher.Matcher().Matches(product)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("fails with wrong price", func(t *testing.T) {
		product := showcase.Product{Price: 5.0}

		matcher := factory.ProductMatches().
			Price(gomatchers.GreaterThan(10.0))

		result := matcher.Matcher().Matches(product)
		if result.Matched {
			t.Error("Expected mismatch for price < 10")
		}
	})
}

func TestOrderMatcher(t *testing.T) {
	t.Run("matches nested user", func(t *testing.T) {
		order := showcase.Order{
			User: showcase.User{
				Email: "customer@example.com",
			},
			Status: "pending",
			Total:  100.50,
		}

		matcher := factory.OrderMatches().
			UserMatches(
				factory.UserMatches().
					Email(gomatchers.Equal("customer@example.com")),
			).
			Status(gomatchers.Equal("pending")).
			Total(gomatchers.GreaterThan(100.0))

		result := matcher.Matcher().Matches(order)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
			for _, detail := range result.Details {
				t.Errorf("  %s", detail)
			}
		}
	})

	t.Run("provides detailed error on nested mismatch", func(t *testing.T) {
		order := showcase.Order{
			User: showcase.User{
				Email: "wrong@example.com",
			},
		}

		matcher := factory.OrderMatches().
			UserMatches(
				factory.UserMatches().
					Email(gomatchers.Equal("right@example.com")),
			)

		result := matcher.Matcher().Matches(order)
		if result.Matched {
			t.Error("Expected mismatch")
		}
		// Should have details about the nested field mismatch
		if len(result.Details) == 0 {
			t.Error("Expected detailed error message")
		}
	})
}

func TestMatcherCombinations(t *testing.T) {
	t.Run("AllOf with matchers", func(t *testing.T) {
		user := showcase.User{
			Email:     "alice@example.com",
			FirstName: "Alice",
		}

		matcher := gomatchers.AllOf(
			factory.UserMatches().
				Email(gomatchers.Contains("@example.com")).
				Matcher(),
			factory.UserMatches().
				FirstName(gomatchers.Equal("Alice")).
				Matcher(),
		)

		result := matcher.Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("AnyOf with matchers", func(t *testing.T) {
		user := showcase.User{Email: "alice@example.com"}

		matcher := gomatchers.AnyOf(
			factory.UserMatches().
				Email(gomatchers.Equal("bob@example.com")).
				Matcher(),
			factory.UserMatches().
				Email(gomatchers.Equal("alice@example.com")).
				Matcher(),
		)

		result := matcher.Matches(user)
		if !result.Matched {
			t.Error("Expected match with AnyOf")
		}
	})
}

func TestNotMatcher(t *testing.T) {
	t.Run("negates matcher result", func(t *testing.T) {
		user := showcase.User{Active: false}

		// Not(IsTrue()) should match when Active is false
		matcher := factory.UserMatches().
			Active(gomatchers.Not(gomatchers.IsTrue()))

		result := matcher.Matcher().Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}

func TestTimeMatcher(t *testing.T) {
	t.Run("matches time fields", func(t *testing.T) {
		now := time.Now()
		product := showcase.Product{
			Name:      "Widget",
			CreatedAt: now,
		}

		matcher := factory.ProductMatches().
			CreatedAt(gomatchers.Equal(now))

		result := matcher.Matcher().Matches(product)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}
