package showcase_test

import (
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

		specta.AssertThat(t, user, matcher)
	})

	t.Run("fails when field doesn't match", func(t *testing.T) {
		user := showcase.User{Email: "alice@example.com"}

		matcher := factory.UserMatches().
			Email(specta.Equal("bob@example.com"))

		specta.AssertThat(t, user, specta.Not(matcher))
	})

	t.Run("only checks specified fields", func(t *testing.T) {
		user := showcase.User{
			Email:     "alice@example.com",
			FirstName: "Alice",
		}

		// Only check email, ignore other fields
		matcher := factory.UserMatches().
			Email(specta.Equal("alice@example.com"))

		specta.AssertThat(t, user, matcher)
	})

	t.Run("works with string matchers", func(t *testing.T) {
		user := showcase.User{Email: "alice@example.com"}

		matcher := factory.UserMatches().
			Email(specta.Contains("@example.com"))

		specta.AssertThat(t, user, matcher)
	})

	t.Run("works with nested matchers", func(t *testing.T) {
		user := showcase.User{
			FirstName: "Alice",
			Address: showcase.Address{
				City: "Boston",
			},
		}

		matcher := factory.UserMatches().
			Address(
				factory.AddressMatches().City(specta.Equal("Boston")),
			)

		specta.AssertThat(t, user, matcher)
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

		specta.AssertThat(t, addr, matcher)
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

		specta.AssertThat(t, addr, matcher)
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

		specta.AssertThat(t, product, matcher)
	})

	t.Run("fails with wrong price", func(t *testing.T) {
		product := showcase.Product{Price: 5.0}

		matcher := factory.ProductMatches().
			Price(specta.GreaterThan(10.0))

		specta.AssertThat(t, product, specta.Not(matcher))
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
			User(
				factory.UserMatches().
					Email(specta.Equal("customer@example.com")),
			).
			Status(specta.Equal("pending")).
			Total(specta.GreaterThan(100.0))

		specta.AssertThat(t, order, matcher)
	})

	t.Run("provides detailed error on nested mismatch", func(t *testing.T) {
		order := showcase.Order{
			User: showcase.User{
				Email: "wrong@example.com",
			},
		}

		matcher := factory.OrderMatches().
			User(
				factory.UserMatches().
					Email(specta.Equal("right@example.com")),
			)

		specta.AssertThat(t, order, specta.Not(matcher))
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
				Email(specta.Contains("@example.com")),
			factory.UserMatches().
				FirstName(specta.Equal("Alice")),
		)

		specta.AssertThat(t, user, matcher)
	})

	t.Run("AnyOf with matchers", func(t *testing.T) {
		user := showcase.User{Email: "alice@example.com"}

		matcher := specta.AnyOf(
			factory.UserMatches().
				Email(specta.Equal("bob@example.com")),
			factory.UserMatches().
				Email(specta.Equal("alice@example.com")),
		)

		specta.AssertThat(t, user, matcher)
	})
}

func TestNotMatcher(t *testing.T) {
	t.Run("negates matcher result", func(t *testing.T) {
		user := showcase.User{Active: false}

		// Not(IsTrue()) should match when Active is false
		matcher := factory.UserMatches().
			Active(specta.Not(specta.IsTrue()))

		specta.AssertThat(t, user, matcher)
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

		specta.AssertThat(t, product, matcher)
	})
}

func TestBankAccountMatcher(t *testing.T) {
	p := specta.New()

	t.Run("matches via getter methods", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		matcher := factory.BankAccountMatches().
			Name(specta.Equal("Alice")).
			Balance(specta.Equal(1000))

		specta.AssertThat(t, account, matcher)
	})

	t.Run("fails when getter value doesn't match", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		matcher := factory.BankAccountMatches().
			Name(specta.Equal("Bob"))

		specta.AssertThat(t, account, specta.Not(matcher))
	})

	t.Run("partial matching - only checks specified getters", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		// Only check Name, ignore Balance
		matcher := factory.BankAccountMatches().
			Name(specta.Equal("Alice"))

		specta.AssertThat(t, account, matcher)
	})

	t.Run("works with numeric matchers", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1500).Build(p)

		matcher := factory.BankAccountMatches().
			Balance(specta.GreaterThan(1000))

		specta.AssertThat(t, account, matcher)
	})
}

func TestEmailMatcher(t *testing.T) {
	t.Run("matches via getter methods", func(t *testing.T) {
		email, err := showcase.NewEmail("alice@example.com")
		specta.AssertThat(t, err, specta.NoErr())

		matcher := factory.EmailMatches().
			Address(specta.Equal("alice@example.com"))

		specta.AssertThat(t, email, matcher)
	})

	t.Run("fails when getter value doesn't match", func(t *testing.T) {
		email, err := showcase.NewEmail("alice@example.com")
		specta.AssertThat(t, err, specta.NoErr())

		addressIsBob := factory.EmailMatches().
			Address(specta.Equal("bob@example.com"))

		specta.AssertThat(t, email, specta.Not(addressIsBob))
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

		specta.AssertThat(t, account, matcher)
	})

	t.Run("fails with clear error message", func(t *testing.T) {
		account := factory.BankAccount().Name("Alice").Balance(1000).Build(p)

		doubleBalanceIs3000 := specta.Field("double balance",
			func(acc showcase.BankAccount) int {
				return acc.GetBalance() * 2
			},
			specta.Equal(3000),
		)

		result := doubleBalanceIs3000.Matches(account)
		specta.AssertThat(t, result.Matched, specta.IsFalse())
		// Error message should include field name
		specta.AssertThat(t, result.Message, specta.Contains("double balance"))
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
			genMatcher,
			fieldMatcher,
		)

		specta.AssertThat(t, account, combined)
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

		specta.AssertThat(t, account, matcher)
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

		specta.AssertThat(t, user, matcher)
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

		specta.AssertThat(t, account1, matcher)
		specta.AssertThat(t, account2, matcher)
	})

	t.Run("BankAccount matcher checks multiple fields", func(t *testing.T) {
		// Create a matcher that checks both Name and Balance
		namedBobWith500 := factory.BankAccount().
			Name("Bob").
			Balance(500).
			AsEqualMatcher()

		// Should match account with both fields
		account := factory.BankAccount().Name("Bob").Balance(500).Build(p)
		specta.AssertThat(t, account, namedBobWith500)

		// Should fail if Name is different
		wrongName := factory.BankAccount().Name("Alice").Balance(500).Build(p)
		specta.AssertThat(t, wrongName, specta.Not(namedBobWith500))

		// Should fail if Balance is different
		wrongBalance := factory.BankAccount().Name("Bob").Balance(1000).Build(p)
		specta.AssertThat(t, wrongBalance, specta.Not(namedBobWith500))
	})

	t.Run("Email recipe converts to partial matcher", func(t *testing.T) {
		// Create a matcher from a recipe
		addressIsTest := factory.Email().Address("test@example.com").AsEqualMatcher()

		// This should match any email with the same address
		email, err := factory.Email().Address("test@example.com").Build(p)
		specta.AssertThat(t, err, specta.NoErr())
		specta.AssertThat(t, email, addressIsTest)

		// Should fail for different address
		otherEmail, err := factory.Email().Address("other@example.com").Build(p)
		specta.AssertThat(t, err, specta.NoErr())

		specta.AssertThat(t, otherEmail, specta.Not(addressIsTest))
	})

	t.Run("Empty recipe creates matcher that matches anything", func(t *testing.T) {
		// A recipe with no set fields should match any instance
		matcher := factory.BankAccount().AsEqualMatcher()

		account1 := factory.BankAccount().Name("Alice").Balance(1000).Build(p)
		account2 := factory.BankAccount().Name("Bob").Balance(2000).Build(p)

		specta.AssertThat(t, account1, matcher)
		specta.AssertThat(t, account2, matcher)
	})

	t.Run("Matcher provides structured diff on failure", func(t *testing.T) {
		matcher := factory.BankAccount().
			Name("Alice").
			Balance(1000).
			AsEqualMatcher()

		account := factory.BankAccount().Name("Bob").Balance(500).Build(p)

		result := matcher.Matches(account)
		specta.AssertThat(t, result.Matched, specta.IsFalse())

		// Message should include field names and values
		specta.AssertThat(t, result.Message, specta.Contains("Name"))
		specta.AssertThat(t, result.Message, specta.Contains("Balance"))
	})
}

// TestCustomUserMatchers demonstrates using custom matchers from user_matcher.go
func TestCustomUserMatchers(t *testing.T) {
	p := specta.New()

	t.Run("IsAdmin matcher", func(t *testing.T) {
		admin := factory.AdminUser().Build(p)
		guest := factory.GuestUser().Build(p)

		// Should match admin user
		specta.AssertThat(t, admin, factory.IsAdmin())

		// Should not match guest user
		specta.AssertThat(t, guest, specta.Not(factory.IsAdmin()))
	})

	t.Run("IsActive matcher", func(t *testing.T) {
		activeUser := factory.User().Active(true).Build(p)
		inactiveUser := factory.User().Active(false).Build(p)

		// Should match active user
		specta.AssertThat(t, activeUser, factory.IsActive())

		// Should not match inactive user
		specta.AssertThat(t, inactiveUser, specta.Not(factory.IsActive()))
	})

	t.Run("IsGuest matcher", func(t *testing.T) {
		guest := factory.GuestUser().Build(p)
		admin := factory.AdminUser().Build(p)

		// Should match guest user
		specta.AssertThat(t, guest, factory.IsGuest())

		// Should not match admin user
		specta.AssertThat(t, admin, specta.Not(factory.IsGuest()))
	})

	t.Run("HasTestEmail matcher", func(t *testing.T) {
		testUser := factory.User().WithTestEmail("alice").Build(p)
		prodUser := factory.User().Email("user@production.com").Build(p)

		// Should match test email
		specta.AssertThat(t, testUser, factory.HasTestEmail())

		// Should not match production email - check error message includes expected domain
		result := factory.HasTestEmail().Matches(prodUser)
		specta.AssertThat(t, result.Matched, specta.IsFalse())
		specta.AssertThat(t, result.Message, specta.Contains("@test.example.com"))
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

		specta.AssertThat(t, user, matcher)
	})

	t.Run("using with AssertThat", func(t *testing.T) {
		admin := factory.AdminUser().Build(p)

		// Custom matchers work with AssertThat
		specta.AssertThat(t, admin, factory.IsAdmin())
		specta.AssertThat(t, admin, factory.IsActive())
	})
}
