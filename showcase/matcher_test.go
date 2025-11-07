package showcase_test

import (
	"testing"
	"time"

	"github.com/james-w/gomatchers"
	"github.com/james-w/gomatchers/showcase/factory"
)

func TestUserMatcher(t *testing.T) {
	p := gomatchers.New()

	t.Run("matches all specified fields", func(t *testing.T) {
		user := factory.User().
			Email("alice@example.com").
			FirstName("Alice").
			Active(true).
			Build(p)

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
		user := factory.User().Email("alice@example.com").Build(p)

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
		user := factory.User().
			Email("alice@example.com").
			FirstName("Alice").
			Build(p)

		// Only check email, ignore other fields
		matcher := factory.UserMatches().
			Email(gomatchers.Equal("alice@example.com"))

		result := matcher.Matcher().Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("works with string matchers", func(t *testing.T) {
		user := factory.User().Email("alice@example.com").Build(p)

		matcher := factory.UserMatches().
			Email(gomatchers.Contains("@example.com"))

		result := matcher.Matcher().Matches(user)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})

	t.Run("works with nested matchers", func(t *testing.T) {
		user := factory.User().
			AddressFromRecipe(factory.Address().City("Boston")).
			Build(p)

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
	p := gomatchers.New()

	t.Run("matches address fields", func(t *testing.T) {
		addr := factory.Address().
			Street("123 Main St").
			City("Springfield").
			State("IL").
			Build(p)

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
		addr := factory.Address().City("Boston").Build(p)

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
	p := gomatchers.New()

	t.Run("matches with numeric matchers", func(t *testing.T) {
		product := factory.Product().
			Name("Widget").
			Price(19.99).
			InStock(true).
			Build(p)

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
		product := factory.Product().Price(5.0).Build(p)

		matcher := factory.ProductMatches().
			Price(gomatchers.GreaterThan(10.0))

		result := matcher.Matcher().Matches(product)
		if result.Matched {
			t.Error("Expected mismatch for price < 10")
		}
	})
}

func TestOrderMatcher(t *testing.T) {
	p := gomatchers.New()

	t.Run("matches nested user", func(t *testing.T) {
		order := factory.Order().
			UserFromRecipe(
				factory.User().Email("customer@example.com"),
			).
			Status("pending").
			Total(100.50).
			Build(p)

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
		order := factory.Order().
			UserFromRecipe(
				factory.User().Email("wrong@example.com"),
			).
			Build(p)

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

func TestBlogPostMatcher(t *testing.T) {
	p := gomatchers.New()

	t.Run("matches blog post with author", func(t *testing.T) {
		post := factory.BlogPost().
			Title("My First Post").
			Published(true).
			AuthorFromRecipe(
				factory.User().FirstName("Alice"),
			).
			Build(p)

		matcher := factory.BlogPostMatches().
			Title(gomatchers.Equal("My First Post")).
			Published(gomatchers.IsTrue()).
			AuthorMatches(
				factory.UserMatches().
					FirstName(gomatchers.Equal("Alice")),
			)

		result := matcher.Matcher().Matches(post)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}

func TestCommentMatcher(t *testing.T) {
	p := gomatchers.New()

	t.Run("matches deeply nested structures", func(t *testing.T) {
		comment := factory.Comment().
			Content("Great post!").
			PostFromRecipe(
				factory.BlogPost().Title("Test Post"),
			).
			AuthorFromRecipe(
				factory.User().FirstName("Bob"),
			).
			Build(p)

		matcher := factory.CommentMatches().
			Content(gomatchers.Equal("Great post!")).
			PostMatches(
				factory.BlogPostMatches().
					Title(gomatchers.Equal("Test Post")),
			).
			AuthorMatches(
				factory.UserMatches().
					FirstName(gomatchers.Equal("Bob")),
			)

		result := matcher.Matcher().Matches(comment)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
			for _, detail := range result.Details {
				t.Errorf("  %s", detail)
			}
		}
	})
}

func TestMatcherWithAssertThat(t *testing.T) {
	p := gomatchers.New()

	t.Run("AssertThat integration", func(t *testing.T) {
		user := factory.User().
			Email("test@example.com").
			Active(true).
			Build(p)

		// This should not fail the test
		gomatchers.AssertThat(t, user,
			factory.UserMatches().
				Email(gomatchers.Contains("@example.com")).
				Active(gomatchers.IsTrue()).
				Matcher())
	})
}

func TestMatcherCombinations(t *testing.T) {
	p := gomatchers.New()

	t.Run("AllOf with matchers", func(t *testing.T) {
		user := factory.User().
			Email("alice@example.com").
			FirstName("Alice").
			Build(p)

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
		user := factory.User().Email("alice@example.com").Build(p)

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

func TestTimeMatcher(t *testing.T) {
	p := gomatchers.New()

	t.Run("matches time fields", func(t *testing.T) {
		now := time.Now()
		product := factory.Product().CreatedAt(now).Build(p)

		matcher := factory.ProductMatches().
			CreatedAt(gomatchers.Equal(now))

		result := matcher.Matcher().Matches(product)
		if !result.Matched {
			t.Errorf("Expected match but got: %s", result.Message)
		}
	})
}

func TestMatcherReadability(t *testing.T) {
	p := gomatchers.New()

	// This test demonstrates the readability of the matcher API
	t.Run("readable fluent API", func(t *testing.T) {
		order := factory.Order().
			Status("shipped").
			Total(250.00).
			UserFromRecipe(
				factory.User().
					Email("premium@example.com").
					Active(true),
			).
			Build(p)

		// The matcher reads naturally
		gomatchers.AssertThat(t, order,
			factory.OrderMatches().
				Status(gomatchers.Equal("shipped")).
				Total(gomatchers.GreaterThan(200.0)).
				UserMatches(
					factory.UserMatches().
						Email(gomatchers.Contains("premium")).
						Active(gomatchers.IsTrue()),
				).
				Matcher())
	})
}
