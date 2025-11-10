package showcase_test

import (
	"strings"
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

func TestBankAccountMatcher(t *testing.T) {
	p := gomatchers.New()

	t.Run("matches via getter methods", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		matcher := factory.BankAccountMatches().
			Name(gomatchers.Equal("Alice")).
			Balance(gomatchers.Equal(1000))

		result := matcher.Matcher().Matches(account)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
			for _, detail := range result.Details {
				t.Errorf("  %s", detail)
			}
		}
	})

	t.Run("fails when getter value doesn't match", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		matcher := factory.BankAccountMatches().
			Name(gomatchers.Equal("Bob"))

		result := matcher.Matcher().Matches(account)
		if result.Matched {
			t.Error("Expected no match but got match")
		}
		if len(result.Details) == 0 || result.Details[0] == "" {
			t.Errorf("Expected error details, got: %v", result.Details)
		}
	})

	t.Run("partial matching - only checks specified getters", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		// Only check Name, ignore Balance
		matcher := factory.BankAccountMatches().
			Name(gomatchers.Equal("Alice"))

		result := matcher.Matcher().Matches(account)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("works with numeric matchers", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1500).Build(p)

		matcher := factory.BankAccountMatches().
			Balance(gomatchers.GreaterThan(1000))

		result := matcher.Matcher().Matches(account)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}

func TestEmailMatcher(t *testing.T) {
	t.Run("matches via getter methods", func(t *testing.T) {
		email, err := showcase.NewEmail("alice@example.com")
		if err != nil {
			t.Fatalf("Failed to create email: %v", err)
		}

		matcher := factory.EmailMatches().
			Address(gomatchers.Equal("alice@example.com"))

		result := matcher.Matcher().Matches(email)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
			for _, detail := range result.Details {
				t.Errorf("  %s", detail)
			}
		}
	})

	t.Run("fails when getter value doesn't match", func(t *testing.T) {
		email, err := showcase.NewEmail("alice@example.com")
		if err != nil {
			t.Fatalf("Failed to create email: %v", err)
		}

		matcher := factory.EmailMatches().
			Address(gomatchers.Equal("bob@example.com"))

		result := matcher.Matcher().Matches(email)
		if result.Matched {
			t.Error("Expected no match but got match")
		}
	})
}

func TestFieldExtractor(t *testing.T) {
	p := gomatchers.New()

	t.Run("matches computed value", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		// Match on double the balance
		matcher := gomatchers.Field("double balance",
			func(acc showcase.BankAccount) int {
				return acc.GetBalance() * 2
			},
			gomatchers.Equal(2000),
		)

		result := matcher.Matches(account)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("fails with clear error message", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		matcher := gomatchers.Field("double balance",
			func(acc showcase.BankAccount) int {
				return acc.GetBalance() * 2
			},
			gomatchers.Equal(3000),
		)

		result := matcher.Matches(account)
		if result.Matched {
			t.Error("Expected no match but got match")
		}
		// Error message should include field name
		if !strings.Contains(result.Message, "double balance") {
			t.Errorf("Expected error to mention 'double balance', got: %s", result.Message)
		}
	})

	t.Run("combines with generated matchers", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		// Use both generated matcher and Field extractor
		genMatcher := factory.BankAccountMatches().
			Name(gomatchers.Equal("Alice"))

		fieldMatcher := gomatchers.Field("double balance",
			func(acc showcase.BankAccount) int {
				return acc.GetBalance() * 2
			},
			gomatchers.GreaterThan(1500),
		)

		combined := gomatchers.AllOf(
			genMatcher.Matcher(),
			fieldMatcher,
		)

		result := combined.Matches(account)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
			for _, detail := range result.Details {
				t.Errorf("  %s", detail)
			}
		}
	})

	t.Run("multiple field extractors with AllOf", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1500).Build(p)

		matcher := gomatchers.AllOf(
			gomatchers.Field("name length",
				func(acc showcase.BankAccount) int {
					return len(acc.GetName())
				},
				gomatchers.Equal(5),
			),
			gomatchers.Field("balance >= 1000",
				func(acc showcase.BankAccount) bool {
					return acc.GetBalance() >= 1000
				},
				gomatchers.IsTrue(),
			),
			gomatchers.Field("double balance",
				func(acc showcase.BankAccount) int {
					return acc.GetBalance() * 2
				},
				gomatchers.GreaterThan(2000),
			),
		)

		result := matcher.Matches(account)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
			for _, detail := range result.Details {
				t.Errorf("  %s", detail)
			}
		}
	})

	t.Run("works with nested matchers", func(t *testing.T) {
		user := factory.User().
			FirstName("Alice").
			Email("alice@example.com").
			AddressFromRecipe(factory.Address().City("NYC")).
			Build(p)

		// Extract and match on nested Address
		matcher := gomatchers.Field("address city",
			func(u showcase.User) string {
				return u.Address.City
			},
			gomatchers.Equal("NYC"),
		)

		result := matcher.Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}
