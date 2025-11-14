package showcase_test

import (
	"strings"
	"testing"
	"time"

	"github.com/james-w/specta"
	"github.com/james-w/specta/showcase"
	"github.com/james-w/specta/showcase/factory"
)

func TestUserMatcher(t *testing.T) {
	t.Run("matches all specified fields", func(t *testing.T) {
		user := showcase.User{
			Email:     "alice@example.com",
			FirstName: "Alice",
			Active:    true,
		}

		matcher := factory.UserMatches().
			Email(specta.Equal("alice@example.com")).
			FirstName(specta.Equal("Alice")).
			Active(specta.IsTrue())

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
			Email(specta.Equal("bob@example.com"))

		result := matcher.Matcher().Matches(user)
		if result.Matched {
			t.Error("Expected mismatch but got match")
		}
	})

	t.Run("only checks specified fields", func(t *testing.T) {
		user := showcase.User{
			Email:     "alice@example.com",
			FirstName: "Alice",
		}

		// Only check email, ignore other fields
		matcher := factory.UserMatches().
			Email(specta.Equal("alice@example.com"))

		result := matcher.Matcher().Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("works with string matchers", func(t *testing.T) {
		user := showcase.User{Email: "alice@example.com"}

		matcher := factory.UserMatches().
			Email(specta.Contains("@example.com"))

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
				factory.AddressMatches().City(specta.Equal("Boston")),
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
			Street(specta.Equal("123 Main St")).
			City(specta.Equal("Springfield")).
			State(specta.Equal("IL"))

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
			City(specta.Equal("Boston"))

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
			Name(specta.Equal("Widget")).
			Price(specta.GreaterThan(10.0)).
			InStock(specta.IsTrue())

		result := matcher.Matcher().Matches(product)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("fails with wrong price", func(t *testing.T) {
		product := showcase.Product{Price: 5.0}

		matcher := factory.ProductMatches().
			Price(specta.GreaterThan(10.0))

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
					Email(specta.Equal("customer@example.com")),
			).
			Status(specta.Equal("pending")).
			Total(specta.GreaterThan(100.0))

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
					Email(specta.Equal("right@example.com")),
			)

		result := matcher.Matcher().Matches(order)
		if result.Matched {
			t.Error("Expected mismatch")
		}
	})
}

func TestMatcherCombinations(t *testing.T) {
	t.Run("AllOf with matchers", func(t *testing.T) {
		user := showcase.User{
			Email:     "alice@example.com",
			FirstName: "Alice",
		}

		matcher := specta.AllOf(
			factory.UserMatches().
				Email(specta.Contains("@example.com")).
				Matcher(),
			factory.UserMatches().
				FirstName(specta.Equal("Alice")).
				Matcher(),
		)

		result := matcher.Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("AnyOf with matchers", func(t *testing.T) {
		user := showcase.User{Email: "alice@example.com"}

		matcher := specta.AnyOf(
			factory.UserMatches().
				Email(specta.Equal("bob@example.com")).
				Matcher(),
			factory.UserMatches().
				Email(specta.Equal("alice@example.com")).
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
			Active(specta.Not(specta.IsTrue()))

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
			CreatedAt(specta.Equal(now))

		result := matcher.Matcher().Matches(product)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}

func TestBankAccountMatcher(t *testing.T) {
	p := specta.New()

	t.Run("matches via getter methods", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		matcher := factory.BankAccountMatches().
			Name(specta.Equal("Alice")).
			Balance(specta.Equal(1000))

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
			Name(specta.Equal("Bob"))

		result := matcher.Matcher().Matches(account)
		if result.Matched {
			t.Error("Expected no match but got match")
		}
	})

	t.Run("partial matching - only checks specified getters", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		// Only check Name, ignore Balance
		matcher := factory.BankAccountMatches().
			Name(specta.Equal("Alice"))

		result := matcher.Matcher().Matches(account)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("works with numeric matchers", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1500).Build(p)

		matcher := factory.BankAccountMatches().
			Balance(specta.GreaterThan(1000))

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
			Address(specta.Equal("alice@example.com"))

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
			Address(specta.Equal("bob@example.com"))

		result := matcher.Matcher().Matches(email)
		if result.Matched {
			t.Error("Expected no match but got match")
		}
	})
}

func TestFieldExtractor(t *testing.T) {
	p := specta.New()

	t.Run("matches computed value", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		// Match on double the balance
		matcher := specta.Field("double balance",
			func(acc showcase.BankAccount) int {
				return acc.GetBalance() * 2
			},
			specta.Equal(2000),
		)

		result := matcher.Matches(account)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("fails with clear error message", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		matcher := specta.Field("double balance",
			func(acc showcase.BankAccount) int {
				return acc.GetBalance() * 2
			},
			specta.Equal(3000),
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
			Name(specta.Equal("Alice"))

		fieldMatcher := specta.Field("double balance",
			func(acc showcase.BankAccount) int {
				return acc.GetBalance() * 2
			},
			specta.GreaterThan(1500),
		)

		combined := specta.AllOf(
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

		matcher := specta.AllOf(
			specta.Field("name length",
				func(acc showcase.BankAccount) int {
					return len(acc.GetName())
				},
				specta.Equal(5),
			),
			specta.Field("balance >= 1000",
				func(acc showcase.BankAccount) bool {
					return acc.GetBalance() >= 1000
				},
				specta.IsTrue(),
			),
			specta.Field("double balance",
				func(acc showcase.BankAccount) int {
					return acc.GetBalance() * 2
				},
				specta.GreaterThan(2000),
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
		matcher := specta.Field("address city",
			func(u showcase.User) string {
				return u.Address.City
			},
			specta.Equal("NYC"),
		)

		result := matcher.Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}

func TestConstructorTypeAsEqualMatcher(t *testing.T) {
	p := specta.New()

	t.Run("BankAccount recipe converts to partial matcher", func(t *testing.T) {
		// Create a matcher from a recipe - only checks Name
		matcher := factory.BankAccount().Name("Alice").AsEqualMatcher()

		// This should match any account with Name="Alice", regardless of Balance
		account1 := factory.BankAccount().Name("Alice").Balance(1000).Build(p)
		account2 := factory.BankAccount().Name("Alice").Balance(2000).Build(p)

		result1 := matcher.Matches(account1)
		if !result1.Matched {
			t.Errorf("Expected match for account1 but got: %s", result1.Message)
		}

		result2 := matcher.Matches(account2)
		if !result2.Matched {
			t.Errorf("Expected match for account2 but got: %s", result2.Message)
		}
	})

	t.Run("BankAccount matcher checks multiple fields", func(t *testing.T) {
		// Create a matcher that checks both Name and Balance
		matcher := factory.BankAccount().
			Name("Bob").
			Balance(500).
			AsEqualMatcher()

		// Should match account with both fields
		account := factory.BankAccount().Name("Bob").Balance(500).Build(p)
		result := matcher.Matches(account)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}

		// Should fail if Name is different
		wrongName := factory.BankAccount().Name("Alice").Balance(500).Build(p)
		result = matcher.Matches(wrongName)
		if result.Matched {
			t.Error("Expected no match when Name differs")
		}

		// Should fail if Balance is different
		wrongBalance := factory.BankAccount().Name("Bob").Balance(1000).Build(p)
		result = matcher.Matches(wrongBalance)
		if result.Matched {
			t.Error("Expected no match when Balance differs")
		}
	})

	t.Run("Email recipe converts to partial matcher", func(t *testing.T) {
		// Create a matcher from a recipe
		matcher := factory.Email().Address("test@example.com").AsEqualMatcher()

		// This should match any email with the same address
		email, err := factory.Email().Address("test@example.com").Build(p)
		if err != nil {
			t.Fatalf("Failed to create email: %v", err)
		}

		result := matcher.Matches(email)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}

		// Should fail for different address
		otherEmail, err := factory.Email().Address("other@example.com").Build(p)
		if err != nil {
			t.Fatalf("Failed to create email: %v", err)
		}

		result = matcher.Matches(otherEmail)
		if result.Matched {
			t.Error("Expected no match when Address differs")
		}
	})

	t.Run("Empty recipe creates matcher that matches anything", func(t *testing.T) {
		// A recipe with no set fields should match any instance
		matcher := factory.BankAccount().AsEqualMatcher()

		account1 := factory.BankAccount().Name("Alice").Balance(1000).Build(p)
		account2 := factory.BankAccount().Name("Bob").Balance(2000).Build(p)

		result1 := matcher.Matches(account1)
		if !result1.Matched {
			t.Errorf("Expected match for account1 but got: %s", result1.Message)
		}

		result2 := matcher.Matches(account2)
		if !result2.Matched {
			t.Errorf("Expected match for account2 but got: %s", result2.Message)
		}
	})

	t.Run("Matcher provides structured diff on failure", func(t *testing.T) {
		matcher := factory.BankAccount().
			Name("Alice").
			Balance(1000).
			AsEqualMatcher()

		account := factory.BankAccount().Name("Bob").Balance(500).Build(p)

		result := matcher.Matches(account)
		if result.Matched {
			t.Error("Expected no match")
		}

		// Message should include field names and values
		if !strings.Contains(result.Message, "Name") {
			t.Errorf("Expected error message to mention 'Name', got: %s", result.Message)
		}
		if !strings.Contains(result.Message, "Balance") {
			t.Errorf("Expected error message to mention 'Balance', got: %s", result.Message)
		}
	})
}

// TestCustomUserMatchers demonstrates using custom matchers from user_matcher.go
func TestCustomUserMatchers(t *testing.T) {
	p := specta.New()

	t.Run("IsAdmin matcher", func(t *testing.T) {
		admin := factory.AdminUser().Build(p)
		guest := factory.GuestUser().Build(p)

		// Should match admin user
		result := factory.IsAdmin().Matches(admin)
		if !result.Matched {
			t.Errorf("expected IsAdmin to match admin user, got: %s", result.Message)
		}

		// Should not match guest user
		result = factory.IsAdmin().Matches(guest)
		if result.Matched {
			t.Error("expected IsAdmin to not match guest user")
		}
	})

	t.Run("IsActive matcher", func(t *testing.T) {
		activeUser := factory.User().Active(true).Build(p)
		inactiveUser := factory.User().Active(false).Build(p)

		// Should match active user
		result := factory.IsActive().Matches(activeUser)
		if !result.Matched {
			t.Errorf("expected IsActive to match active user, got: %s", result.Message)
		}

		// Should not match inactive user
		result = factory.IsActive().Matches(inactiveUser)
		if result.Matched {
			t.Error("expected IsActive to not match inactive user")
		}
	})

	t.Run("IsGuest matcher", func(t *testing.T) {
		guest := factory.GuestUser().Build(p)
		admin := factory.AdminUser().Build(p)

		// Should match guest user
		result := factory.IsGuest().Matches(guest)
		if !result.Matched {
			t.Errorf("expected IsGuest to match guest user, got: %s", result.Message)
		}

		// Should not match admin user
		result = factory.IsGuest().Matches(admin)
		if result.Matched {
			t.Error("expected IsGuest to not match admin user")
		}
	})

	t.Run("HasTestEmail matcher", func(t *testing.T) {
		testUser := factory.User().WithTestEmail("alice").Build(p)
		prodUser := factory.User().Email("user@production.com").Build(p)

		// Should match test email
		result := factory.HasTestEmail().Matches(testUser)
		if !result.Matched {
			t.Errorf("expected HasTestEmail to match test user, got: %s", result.Message)
		}

		// Should not match production email
		result = factory.HasTestEmail().Matches(prodUser)
		if result.Matched {
			t.Error("expected HasTestEmail to not match production email")
		}
		if !strings.Contains(result.Message, "@test.example.com") {
			t.Errorf("expected error message to mention @test.example.com, got: %s", result.Message)
		}
	})

	t.Run("composing custom matchers with AllOf", func(t *testing.T) {
		user := factory.User().
			Active(true).
			WithTestEmail("admin").
			Build(p)

		// Combine custom matchers
		matcher := specta.AllOf(
			factory.IsActive(),
			factory.HasTestEmail(),
		)

		result := matcher.Matches(user)
		if !result.Matched {
			t.Errorf("expected user to match both IsActive and HasTestEmail, got: %s", result.Message)
		}
	})

	t.Run("using with AssertThat", func(t *testing.T) {
		admin := factory.AdminUser().Build(p)

		// Custom matchers work with AssertThat
		specta.AssertThat(t, admin, factory.IsAdmin())
		specta.AssertThat(t, admin, factory.IsActive())
	})
}
