package showcase_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/james-w/specta"
	"github.com/james-w/specta/showcase"
	"github.com/james-w/specta/showcase/factory"
)

func TestAddressFactory(t *testing.T) {
	p := specta.New()

	t.Run("default values", func(t *testing.T) {
		addr := factory.Address().Build(p)

		specta.AssertThat(t, addr.Street, specta.Not(specta.Equal("")))
		specta.AssertThat(t, addr.City, specta.Not(specta.Equal("")))
		specta.AssertThat(t, addr.State, specta.Not(specta.Equal("")))
		specta.AssertThat(t, addr.ZipCode, specta.Not(specta.Equal("")))
		specta.AssertThat(t, addr.Country, specta.Not(specta.Equal("")))
	})

	t.Run("custom values", func(t *testing.T) {
		addr := factory.Address().
			Street("123 Main St").
			City("Springfield").
			State("IL").
			ZipCode("62701").
			Country("USA").
			Build(p)

		specta.AssertThat(t, addr.Street, specta.Equal("123 Main St"))
		specta.AssertThat(t, addr.City, specta.Equal("Springfield"))
		specta.AssertThat(t, addr.State, specta.Equal("IL"))
		specta.AssertThat(t, addr.ZipCode, specta.Equal("62701"))
		specta.AssertThat(t, addr.Country, specta.Equal("USA"))
	})

	t.Run("many with unique values", func(t *testing.T) {
		addresses := factory.Address().Many(3, p)

		specta.AssertThat(t, addresses, specta.HasSize[showcase.Address](3))

		// Each address should have unique values due to deterministic generation
		for i := range 2 {
			specta.AssertThat(t, addresses[i].Street, specta.Not(specta.Equal(addresses[i+1].Street)))
		}
	})
}

func TestUserFactory(t *testing.T) {
	p := specta.New()

	t.Run("default values with nested address", func(t *testing.T) {
		user := factory.User().Build(p)

		specta.AssertThat(t, user.ID, specta.Not(specta.Equal("")))
		specta.AssertThat(t, user.Email, specta.Not(specta.Equal("")))
		specta.AssertThat(t, user.FirstName, specta.Not(specta.Equal("")))
		specta.AssertThat(t, user.LastName, specta.Not(specta.Equal("")))
		// Address should be auto-generated
		specta.AssertThat(t, user.Address.City, specta.Not(specta.Equal("")))
		specta.AssertThat(t, user.CreatedAt.IsZero(), specta.IsFalse())
	})

	t.Run("custom active user", func(t *testing.T) {
		user := factory.User().
			Email("test@example.com").
			Active(true).
			FirstName("John").
			LastName("Doe").
			Build(p)

		specta.AssertThat(t, user.Email, specta.Equal("test@example.com"))
		specta.AssertThat(t, user.Active, specta.IsTrue())
		specta.AssertThat(t, user.FirstName, specta.Equal("John"))
		specta.AssertThat(t, user.LastName, specta.Equal("Doe"))
	})

	t.Run("nested address from recipe", func(t *testing.T) {
		user := factory.User().
			AddressFromRecipe(
				factory.Address().City("Boston").State("MA"),
			).
			Build(p)

		specta.AssertThat(t, user.Address.City, specta.Equal("Boston"))
		specta.AssertThat(t, user.Address.State, specta.Equal("MA"))
	})

	t.Run("literal address reuse", func(t *testing.T) {
		addr := factory.Address().City("NYC").Build(p)
		users := factory.User().Address(addr).Many(2, p)

		specta.AssertThat(t, users, specta.HasSize[showcase.User](2))

		// Both users should have the same address (same object)
		specta.AssertThat(t, users[0].Address.City, specta.Equal("NYC"))
		specta.AssertThat(t, users[1].Address.City, specta.Equal("NYC"))
	})

	t.Run("address recipe generates unique instances", func(t *testing.T) {
		users := factory.User().
			AddressFromRecipe(factory.Address().State("CA")).
			Many(2, p)

		specta.AssertThat(t, users, specta.HasSize[showcase.User](2))

		// Each user should have unique address (different cities)
		specta.AssertThat(t, users[0].Address.City, specta.Not(specta.Equal(users[1].Address.City)))
		// But both should have CA state
		specta.AssertThat(t, users[0].Address.State, specta.Equal("CA"))
		specta.AssertThat(t, users[1].Address.State, specta.Equal("CA"))
	})
}

func TestProductFactory(t *testing.T) {
	p := specta.New()

	t.Run("product with price and stock", func(t *testing.T) {
		product := factory.Product().
			Name("Widget").
			Price(19.99).
			InStock(true).
			Build(p)

		specta.AssertThat(t, product.Name, specta.Equal("Widget"))
		specta.AssertThat(t, product.Price, specta.Equal(19.99))
		specta.AssertThat(t, product.InStock, specta.IsTrue())
		specta.AssertThat(t, product.ID, specta.Not(specta.Equal("")))
	})
}

func TestOrderFactory(t *testing.T) {
	p := specta.New()

	t.Run("order with nested user", func(t *testing.T) {
		order := factory.Order().
			UserFromRecipe(factory.User().Email("customer@example.com")).
			Status("pending").
			Total(100.50).
			Build(p)

		specta.AssertThat(t, order.User.Email, specta.Equal("customer@example.com"))
		specta.AssertThat(t, order.Status, specta.Equal("pending"))
		specta.AssertThat(t, order.Total, specta.Equal(100.50))
		specta.AssertThat(t, order.ID, specta.Not(specta.Equal("")))
	})

	t.Run("order with custom items slice", func(t *testing.T) {
		items := []showcase.OrderItem{
			{
				Product:  factory.Product().Name("Item 1").Price(10.00).Build(p),
				Quantity: 2,
				Price:    20.00,
			},
			{
				Product:  factory.Product().Name("Item 2").Price(15.00).Build(p),
				Quantity: 1,
				Price:    15.00,
			},
		}

		order := factory.Order().
			Items(items).
			Total(35.00).
			Build(p)

		specta.AssertThat(t, order.Items, specta.HasSize[showcase.OrderItem](2))
		specta.AssertThat(t, order.Items[0].Product.Name, specta.Equal("Item 1"))
		specta.AssertThat(t, order.Total, specta.Equal(35.00))
	})
}

func TestBlogPostFactory(t *testing.T) {
	p := specta.New()

	t.Run("published blog post", func(t *testing.T) {
		now := time.Now()
		post := factory.BlogPost().
			Title("My First Post").
			Content("Hello world!").
			Published(true).
			PublishedAt(now).
			AuthorFromRecipe(factory.User().FirstName("Alice")).
			Build(p)

		specta.AssertThat(t, post.Title, specta.Equal("My First Post"))
		specta.AssertThat(t, post.Content, specta.Equal("Hello world!"))
		specta.AssertThat(t, post.Published, specta.IsTrue())
		specta.AssertThat(t, post.PublishedAt.Equal(now), specta.IsTrue())
		specta.AssertThat(t, post.Author.FirstName, specta.Equal("Alice"))
	})
}

func TestCommentFactory(t *testing.T) {
	p := specta.New()

	t.Run("comment with nested post and author", func(t *testing.T) {
		comment := factory.Comment().
			Content("Great post!").
			PostFromRecipe(
				factory.BlogPost().Title("Test Post"),
			).
			AuthorFromRecipe(
				factory.User().FirstName("Bob"),
			).
			Build(p)

		specta.AssertThat(t, comment.Content, specta.Equal("Great post!"))
		specta.AssertThat(t, comment.Post.Title, specta.Equal("Test Post"))
		specta.AssertThat(t, comment.Author.FirstName, specta.Equal("Bob"))
		specta.AssertThat(t, comment.ID, specta.Not(specta.Equal("")))
	})

	t.Run("multiple comments on same post", func(t *testing.T) {
		post := factory.BlogPost().Title("Shared Post").Build(p)

		comments := factory.Comment().
			Post(post).
			Many(3, p)

		specta.AssertThat(t, comments, specta.HasSize[showcase.Comment](3))

		// All comments should reference the same post
		for _, c := range comments {
			specta.AssertThat(t, c.Post.Title, specta.Equal("Shared Post"))
		}
	})
}

func TestDeterministicGeneration(t *testing.T) {
	t.Run("same primitives produce same values", func(t *testing.T) {
		p1 := specta.New()
		p2 := specta.New()

		user1 := factory.User().FirstName("Test").Build(p1)
		user2 := factory.User().FirstName("Test").Build(p2)

		specta.AssertThat(t, user1.ID, specta.Equal(user2.ID))
		specta.AssertThat(t, user1.Email, specta.Equal(user2.Email))
	})

	t.Run("sequential generation increments counter", func(t *testing.T) {
		p := specta.New()

		addr1 := factory.Address().Build(p)
		addr2 := factory.Address().Build(p)

		specta.AssertThat(t, addr1.Street, specta.Not(specta.Equal(addr2.Street)))
		specta.AssertThat(t, addr1.City, specta.Not(specta.Equal(addr2.City)))
	})
}

func TestProviderComposition(t *testing.T) {
	p := specta.New()

	t.Run("provider creates unique instances", func(t *testing.T) {
		// Create a user recipe
		userRecipe := factory.User().Active(true).FirstName("Composite")

		// Use it multiple times - each should create a new instance
		order1 := factory.Order().UserFromRecipe(userRecipe).Build(p)
		order2 := factory.Order().UserFromRecipe(userRecipe).Build(p)

		// Both should have Active=true and FirstName="Composite"
		specta.AssertThat(t, order1.User.Active, specta.IsTrue())
		specta.AssertThat(t, order2.User.Active, specta.IsTrue())
		specta.AssertThat(t, order1.User.FirstName, specta.Equal("Composite"))
		specta.AssertThat(t, order2.User.FirstName, specta.Equal("Composite"))

		// But they should have different IDs (unique instances)
		specta.AssertThat(t, order1.User.ID, specta.Not(specta.Equal(order2.User.ID)))
	})
}

func TestTimeHandling(t *testing.T) {
	p := specta.New()

	t.Run("default times are deterministic", func(t *testing.T) {
		user1 := factory.User().Build(p)
		// Reset primitives
		p = specta.New()
		user2 := factory.User().Build(p)

		specta.AssertThat(t, user1.CreatedAt.Equal(user2.CreatedAt), specta.IsTrue())
	})

	t.Run("custom times can be set", func(t *testing.T) {
		customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		product := factory.Product().CreatedAt(customTime).Build(p)

		specta.AssertThat(t, product.CreatedAt.Equal(customTime), specta.IsTrue())
	})

	t.Run("sequential calls advance time", func(t *testing.T) {
		user1 := factory.User().Build(p)
		user2 := factory.User().Build(p)

		specta.AssertThat(t, user2.CreatedAt.After(user1.CreatedAt), specta.IsTrue())
	})
}

func TestBankAccountFactory(t *testing.T) {
	p := specta.New()

	t.Run("constructor-based factory with default values", func(t *testing.T) {
		account := factory.BankAccount().Build(p)

		name := account.GetName()
		balance := account.GetBalance()

		specta.AssertThat(t, name, specta.Not(specta.Equal("")))
		// Default balance should be generated
		specta.AssertThat(t, balance, specta.Not(specta.Equal(0)))
	})

	t.Run("constructor-based factory with custom values", func(t *testing.T) {
		account := factory.BankAccount().
			Name("Alice").
			Balance(1000).
			Build(p)

		specta.AssertThat(t, account.GetName(), specta.Equal("Alice"))
		specta.AssertThat(t, account.GetBalance(), specta.Equal(1000))
	})

	t.Run("many generates unique instances", func(t *testing.T) {
		accounts := factory.BankAccount().Many(3, p)

		specta.AssertThat(t, accounts, specta.HasSize[showcase.BankAccount](3))

		// Names should be unique
		names := make(map[string]bool)
		for _, acc := range accounts {
			name := acc.GetName()
			specta.AssertThat(t, names[name], specta.IsFalse())
			names[name] = true
		}

		specta.AssertThat(t, len(names), specta.Equal(3))
	})
}

func TestEmailFactory(t *testing.T) {
	p := specta.New()

	t.Run("Build with valid email returns no error", func(t *testing.T) {
		email, err := factory.Email().Address("user@example.com").Build(p)

		specta.AssertThat(t, err, specta.NoErr())
		specta.AssertThat(t, email.GetAddress(), specta.Equal("user@example.com"))
	})

	t.Run("Build with invalid email returns error", func(t *testing.T) {
		_, err := factory.Email().Address("notanemail").Build(p)

		specta.AssertThat(t, err, specta.IsError())
	})

	t.Run("Build with default pattern generates valid email", func(t *testing.T) {
		// Default provider uses email pattern which generates valid emails
		email, err := factory.Email().Build(p)

		specta.AssertThat(t, err, specta.NoErr())
		addr := email.GetAddress()
		specta.AssertThat(t, addr, specta.Contains("@"))
	})

	t.Run("Build returns error on invalid input", func(t *testing.T) {
		_, err := factory.Email().Address("invalid").Build(p)
		specta.AssertThat(t, err, specta.IsError())
	})

	t.Run("Many panics on error", func(t *testing.T) {
		defer func() {
			r := recover()
			specta.AssertThat(t, r, specta.Not(specta.Equal[any](nil)))
			// Verify the panic message includes the item index
			msg := fmt.Sprint(r)
			specta.AssertThat(t, msg, specta.Contains("Many() failed on item 0"))
		}()

		factory.Email().Address("invalid").Many(3, p) // Should panic on first item
	})

	t.Run("Many succeeds with valid emails", func(t *testing.T) {
		emails := factory.Email().Address("test@example.com").Many(3, p)

		specta.AssertThat(t, emails, specta.HasSize[showcase.Email](3))
		for _, email := range emails {
			specta.AssertThat(t, email.GetAddress(), specta.Equal("test@example.com"))
		}
	})
}

func TestIDGeneration(t *testing.T) {
	p := specta.New()

	t.Run("IDs are unique", func(t *testing.T) {
		users := factory.User().Many(5, p)

		seen := make(map[string]bool)
		for _, u := range users {
			specta.AssertThat(t, seen[u.ID], specta.IsFalse())
			seen[u.ID] = true
		}

		specta.AssertThat(t, len(seen), specta.Equal(5))
	})

	t.Run("ID format", func(t *testing.T) {
		user := factory.User().Build(p)

		specta.AssertThat(t, user.ID, specta.Not(specta.Equal("")))
		// IDs should start with "id_"
		specta.AssertThat(t, user.ID, specta.HasPrefix("id_"))
	})
}

// TestCustomUserRecipes demonstrates using custom recipe methods from user.go
func TestCustomUserRecipes(t *testing.T) {
	p := specta.New()

	t.Run("AdminUser", func(t *testing.T) {
		admin := factory.AdminUser().Build(p)

		specta.AssertThat(t, admin.FirstName, specta.Equal("Admin"))
		specta.AssertThat(t, admin.LastName, specta.Equal("User"))
		specta.AssertThat(t, admin.Email, specta.Equal("admin@example.com"))
		specta.AssertThat(t, admin.Active, specta.IsTrue())
	})

	t.Run("GuestUser", func(t *testing.T) {
		guest := factory.GuestUser().Build(p)

		specta.AssertThat(t, guest.FirstName, specta.Equal("Guest"))
		specta.AssertThat(t, guest.LastName, specta.Equal("User"))
		specta.AssertThat(t, guest.Active, specta.IsFalse())
	})

	t.Run("WithAdminRole", func(t *testing.T) {
		user := factory.User().WithAdminRole().Build(p)

		specta.AssertThat(t, user.FirstName, specta.Equal("Admin"))
		specta.AssertThat(t, user.Active, specta.IsTrue())
	})

	t.Run("WithTestEmail", func(t *testing.T) {
		user := factory.User().WithTestEmail("alice").Build(p)

		specta.AssertThat(t, user.Email, specta.Equal("alice@test.example.com"))
	})

	t.Run("composing custom methods", func(t *testing.T) {
		// Custom methods compose with generated methods
		user := factory.User().
			WithAdminRole().
			FirstName("SuperAdmin"). // Override the admin first name
			Build(p)

		specta.AssertThat(t, user.FirstName, specta.Equal("SuperAdmin"))
		specta.AssertThat(t, user.Email, specta.Equal("admin@example.com"))
	})
}

// TestCustomDefaults demonstrates that custom defaults from user_defaults.go are applied
func TestCustomDefaults(t *testing.T) {
	p := specta.New()

	t.Run("email uses custom default", func(t *testing.T) {
		user := factory.User().Build(p)

		// Custom default should generate user1@test.example.com format
		specta.AssertThat(t, user.Email, specta.Contains("@test.example.com"))
	})

	t.Run("active defaults to true", func(t *testing.T) {
		user := factory.User().Build(p)

		// Custom default sets Active to true
		specta.AssertThat(t, user.Active, specta.IsTrue())
	})
}
