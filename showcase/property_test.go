package showcase_test

import (
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/showcase/factory"
)

// TestPropertyBasicGenerators demonstrates using basic generators with AssertThat
func TestPropertyBasicGenerators(t *testing.T) {
	t.Run("string concatenation length property", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			s1 := specta.Draw(t, specta.String(), "s1")
			s2 := specta.Draw(t, specta.String(), "s2")

			combined := s1 + s2
			expectedLen := len(s1) + len(s2)

			specta.AssertThat(t, len(combined), specta.Equal(expectedLen))
		}, specta.MaxTests(50))
	})

	t.Run("integer addition is commutative", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			x := specta.Draw(t, specta.Int().Range(-1000, 1000), "x")
			y := specta.Draw(t, specta.Int().Range(-1000, 1000), "y")

			specta.AssertThat(t, x+y, specta.Equal(y+x))
		}, specta.MaxTests(50))
	})

	t.Run("boolean negation is involutive", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			b := specta.Draw(t, specta.Bool(), "b")

			// Applying NOT twice should return original value
			specta.AssertThat(t, !(!b), specta.Equal(b))
		}, specta.MaxTests(50))
	})

	t.Run("strings with AlphaNum charset contain only letters and digits", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			s := specta.Draw(t, specta.String().AlphaNum().MinLen(1).MaxLen(50), "s")

			// Every character should be alphanumeric
			for _, ch := range s {
				isAlphaNum := (ch >= 'a' && ch <= 'z') ||
					(ch >= 'A' && ch <= 'Z') ||
					(ch >= '0' && ch <= '9')
				specta.AssertThat(t, isAlphaNum, specta.IsTrue())
			}
		}, specta.MaxTests(50))
	})
}

// TestPropertySliceOperations demonstrates slice property tests
func TestPropertySliceOperations(t *testing.T) {
	t.Run("appending to slice increases length by one", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			xs := specta.Draw(t, specta.Slice(specta.Int()).MaxLen(20), "xs")
			newElem := specta.Draw(t, specta.Int(), "newElem")

			originalLen := len(xs)
			xs = append(xs, newElem)

			specta.AssertThat(t, len(xs), specta.Equal(originalLen+1))
			specta.AssertThat(t, xs[len(xs)-1], specta.Equal(newElem))
		}, specta.MaxTests(50))
	})

	t.Run("reversing a slice twice returns original", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			xs := specta.Draw(t, specta.Slice(specta.Int().Range(0, 100)).MaxLen(20), "xs")

			// Make a copy of original
			original := make([]int64, len(xs))
			copy(original, xs)

			// Reverse twice
			reverse(xs)
			reverse(xs)

			// Should match original
			specta.AssertThat(t, len(xs), specta.Equal(len(original)))
			for i := range xs {
				specta.AssertThat(t, xs[i], specta.Equal(original[i]))
			}
		}, specta.MaxTests(50))
	})

	t.Run("slice length is within specified bounds", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			xs := specta.Draw(t, specta.Slice(specta.Int()).MinLen(5).MaxLen(10), "xs")

			specta.AssertThat(t, len(xs), specta.GreaterThanOrEqual(5))
			specta.AssertThat(t, len(xs), specta.Not(specta.GreaterThan(10)))
		}, specta.MaxTests(50))
	})
}

// TestPropertyUserFactory demonstrates property tests with factory-generated Users
func TestPropertyUserFactory(t *testing.T) {
	t.Run("using factory recipes as generators with Draw", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			// You can use factory recipes as generators with .Gen()
			user := specta.Draw(t, factory.User().Gen(), "user")

			// Generated user should have non-empty ID
			specta.AssertThat(t, user.ID, specta.Not(specta.Equal("")))
		}, specta.MaxTests(50))
	})

	t.Run("configured recipe as generator with Draw", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			// You can configure the recipe before converting to Gen
			activeUser := specta.Draw(t, factory.User().Active(true).Gen(), "activeUser")

			specta.AssertThat(t, activeUser.Active, specta.IsTrue())
			specta.AssertThat(t, activeUser.ID, specta.Not(specta.Equal("")))
		}, specta.MaxTests(50))
	})

	t.Run("nested recipe as generator", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			city := specta.Draw(t, specta.String().AlphaNum().MinLen(3).MaxLen(10), "city")

			// Use configured nested recipe as generator
			user := specta.Draw(t,
				factory.User().
					AddressFromRecipe(factory.Address().City(city)).
					Gen(),
				"user")

			specta.AssertThat(t, user.Address.City, specta.Equal(city))
		}, specta.MaxTests(50))
	})

	t.Run("generated users always have valid IDs", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			user := factory.User().Build(p)

			// ID should be non-empty
			specta.AssertThat(t, user.ID, specta.Not(specta.Equal("")))
		}, specta.MaxTests(50))
	})

	t.Run("custom email always contains specified domain", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			domain := specta.Draw(t, specta.String().AlphaNum().MinLen(3).MaxLen(10), "domain")
			email := "user@" + domain + ".com"

			user := factory.User().Email(email).Build(p)

			specta.AssertThat(t, user.Email, specta.Contains("@"+domain+".com"))
		}, specta.MaxTests(50))
	})

	t.Run("active flag is preserved through factory", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			isActive := specta.Draw(t, specta.Bool(), "isActive")

			user := factory.User().Active(isActive).Build(p)

			specta.AssertThat(t, user.Active, specta.Equal(isActive))
		}, specta.MaxTests(50))
	})

	t.Run("many users have unique IDs", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			count := specta.Draw(t, specta.Int().Range(2, 10), "count")

			users := factory.User().Many(int(count), p)

			// Collect IDs
			ids := make(map[string]bool)
			for _, user := range users {
				specta.AssertThat(t, ids[user.ID], specta.IsFalse())
				ids[user.ID] = true
			}

			// Should have count unique IDs
			specta.AssertThat(t, len(ids), specta.Equal(int(count)))
		}, specta.MaxTests(20))
	})
}

// TestPropertyMatchers demonstrates property tests with matchers
func TestPropertyMatchers(t *testing.T) {
	t.Run("Equal matcher is reflexive", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			x := specta.Draw(t, specta.Int().Range(-100, 100), "x")

			// Every value equals itself
			specta.AssertThat(t, x, specta.Equal(x))
		}, specta.MaxTests(50))
	})

	t.Run("Contains matcher works with any substring", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			base := specta.Draw(t, specta.String().AlphaNum().MinLen(5).MaxLen(20), "base")
			prefix := specta.Draw(t, specta.String().AlphaNum().MinLen(1).MaxLen(5), "prefix")
			suffix := specta.Draw(t, specta.String().AlphaNum().MinLen(1).MaxLen(5), "suffix")

			fullString := prefix + base + suffix

			// Full string should contain base
			specta.AssertThat(t, fullString, specta.Contains(base))
			specta.AssertThat(t, fullString, specta.Contains(prefix))
			specta.AssertThat(t, fullString, specta.Contains(suffix))
		}, specta.MaxTests(50))
	})

	t.Run("GreaterThan and LessThan are transitive", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			a := specta.Draw(t, specta.Int().Range(0, 50), "a")
			b := specta.Draw(t, specta.Int().Range(51, 100), "b")
			c := specta.Draw(t, specta.Int().Range(101, 150), "c")

			// a < b < c, so a < c
			specta.AssertThat(t, a, specta.LessThan(b))
			specta.AssertThat(t, b, specta.LessThan(c))
			specta.AssertThat(t, a, specta.LessThan(c))
		}, specta.MaxTests(50))
	})

	t.Run("Not matcher inverts result", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			x := specta.Draw(t, specta.Int().Range(1, 100), "x")
			y := specta.Draw(t, specta.Int().Range(1, 100), "y")

			// Exactly one of these should be true: x == y or x != y
			if x == y {
				specta.AssertThat(t, x, specta.Equal(y))
				specta.AssertThat(t, x, specta.Not(specta.Not(specta.Equal(y))))
			} else {
				specta.AssertThat(t, x, specta.Not(specta.Equal(y)))
			}
		}, specta.MaxTests(50))
	})
}

// TestPropertyNestedStructures demonstrates property tests with nested structures
func TestPropertyNestedStructures(t *testing.T) {
	t.Run("user with address has valid nested data", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			city := specta.Draw(t, specta.String().AlphaNum().MinLen(3).MaxLen(15), "city")

			user := factory.User().
				AddressFromRecipe(factory.Address().City(city)).
				Build(p)

			// Nested address should have the specified city
			specta.AssertThat(t, user.Address.City, specta.Equal(city))

			// User should match the partial matcher
			matcher := factory.UserMatches().
				AddressMatches(
					factory.AddressMatches().City(specta.Equal(city)),
				).Matcher()

			specta.AssertThat(t, user, matcher)
		}, specta.MaxTests(50))
	})

	t.Run("order total is always non-negative", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			total := specta.Draw(t, specta.Float64(0, 10000), "total")

			order := factory.Order().Total(total).Build(p)

			specta.AssertThat(t, order.Total, specta.GreaterThanOrEqual(0.0))
			specta.AssertThat(t, order.Total, specta.Equal(total))
		}, specta.MaxTests(50))
	})

	t.Run("order with generated user has consistent nested structure", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			firstName := specta.Draw(t, specta.String().AlphaNum().MinLen(3).MaxLen(10), "firstName")

			order := factory.Order().
				UserFromRecipe(factory.User().FirstName(firstName)).
				Build(p)

			// Nested user should have the specified first name
			specta.AssertThat(t, order.User.FirstName, specta.Equal(firstName))

			// Order should match the nested partial matcher
			matcher := factory.OrderMatches().
				UserMatches(
					factory.UserMatches().FirstName(specta.Equal(firstName)),
				).Matcher()

			specta.AssertThat(t, order, matcher)
		}, specta.MaxTests(50))
	})
}

// TestPropertyBankAccount demonstrates property tests with constructor-based types
func TestPropertyBankAccount(t *testing.T) {
	t.Run("bank account preserves name and balance", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			name := specta.Draw(t, specta.String().AlphaNum().MinLen(3).MaxLen(20), "name")
			balance := specta.Draw(t, specta.Int().Range(0, 1000000), "balance")

			account := factory.BankAccount().
				Name(name).
				Balance(int(balance)).
				Build(p)

			specta.AssertThat(t, account.GetName(), specta.Equal(name))
			specta.AssertThat(t, account.GetBalance(), specta.Equal(int(balance)))
		}, specta.MaxTests(50))
	})

	t.Run("multiple accounts with same balance are equal via matcher", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			balance := specta.Draw(t, specta.Int().Range(100, 1000), "balance")

			account1 := factory.BankAccount().Balance(int(balance)).Build(p)
			account2 := factory.BankAccount().Balance(int(balance)).Build(p)

			// Both should match a matcher checking only balance
			balanceMatcher := factory.BankAccountMatches().
				Balance(specta.Equal(int(balance))).
				Matcher()

			specta.AssertThat(t, account1, balanceMatcher)
			specta.AssertThat(t, account2, balanceMatcher)
		}, specta.MaxTests(50))
	})
}

// TestPropertyCustomMatchers demonstrates property tests with custom matchers
func TestPropertyCustomMatchers(t *testing.T) {
	t.Run("admin users always match IsAdmin", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			admin := factory.AdminUser().Build(p)

			specta.AssertThat(t, admin, factory.IsAdmin())
			specta.AssertThat(t, admin.FirstName, specta.Equal("Admin"))
		}, specta.MaxTests(50))
	})

	t.Run("active users always match IsActive", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			user := factory.User().Active(true).Build(p)

			specta.AssertThat(t, user, factory.IsActive())
			specta.AssertThat(t, user.Active, specta.IsTrue())
		}, specta.MaxTests(50))
	})

	t.Run("test emails always match HasTestEmail", func(t *testing.T) {
		specta.Property(t, func(t *specta.T) {
			p := specta.New()
			username := specta.Draw(t, specta.String().AlphaNum().MinLen(3).MaxLen(10), "username")

			user := factory.User().WithTestEmail(username).Build(p)

			specta.AssertThat(t, user, factory.HasTestEmail())
			specta.AssertThat(t, user.Email, specta.Contains("@test.example.com"))
		}, specta.MaxTests(50))
	})
}

// Helper function for reversing a slice
func reverse(xs []int64) {
	for i, j := 0, len(xs)-1; i < j; i, j = i+1, j-1 {
		xs[i], xs[j] = xs[j], xs[i]
	}
}
