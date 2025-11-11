package showcase_test

import (
	"fmt"
	"strings"
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

		if addr.Street == "" {
			t.Error("expected non-empty Street")
		}
		if addr.City == "" {
			t.Error("expected non-empty City")
		}
		if addr.State == "" {
			t.Error("expected non-empty State")
		}
		if addr.ZipCode == "" {
			t.Error("expected non-empty ZipCode")
		}
		if addr.Country == "" {
			t.Error("expected non-empty Country")
		}
	})

	t.Run("custom values", func(t *testing.T) {
		addr := factory.Address().
			Street("123 Main St").
			City("Springfield").
			State("IL").
			ZipCode("62701").
			Country("USA").
			Build(p)

		if addr.Street != "123 Main St" {
			t.Errorf("expected Street='123 Main St', got %q", addr.Street)
		}
		if addr.City != "Springfield" {
			t.Errorf("expected City='Springfield', got %q", addr.City)
		}
		if addr.State != "IL" {
			t.Errorf("expected State='IL', got %q", addr.State)
		}
		if addr.ZipCode != "62701" {
			t.Errorf("expected ZipCode='62701', got %q", addr.ZipCode)
		}
		if addr.Country != "USA" {
			t.Errorf("expected Country='USA', got %q", addr.Country)
		}
	})

	t.Run("many with unique values", func(t *testing.T) {
		addresses := factory.Address().Many(3, p)

		if len(addresses) != 3 {
			t.Fatalf("expected 3 addresses, got %d", len(addresses))
		}

		// Each address should have unique values due to deterministic generation
		for i := range 2 {
			if addresses[i].Street == addresses[i+1].Street {
				t.Errorf("addresses[%d] and addresses[%d] have same Street", i, i+1)
			}
		}
	})
}

func TestUserFactory(t *testing.T) {
	p := specta.New()

	t.Run("default values with nested address", func(t *testing.T) {
		user := factory.User().Build(p)

		if user.ID == "" {
			t.Error("expected non-empty ID")
		}
		if user.Email == "" {
			t.Error("expected non-empty Email")
		}
		if user.FirstName == "" {
			t.Error("expected non-empty FirstName")
		}
		if user.LastName == "" {
			t.Error("expected non-empty LastName")
		}
		// Address should be auto-generated
		if user.Address.City == "" {
			t.Error("expected nested Address to be generated with non-empty City")
		}
		if user.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("custom active user", func(t *testing.T) {
		user := factory.User().
			Email("test@example.com").
			Active(true).
			FirstName("John").
			LastName("Doe").
			Build(p)

		if user.Email != "test@example.com" {
			t.Errorf("expected Email='test@example.com', got %q", user.Email)
		}
		if !user.Active {
			t.Error("expected Active=true")
		}
		if user.FirstName != "John" {
			t.Errorf("expected FirstName='John', got %q", user.FirstName)
		}
		if user.LastName != "Doe" {
			t.Errorf("expected LastName='Doe', got %q", user.LastName)
		}
	})

	t.Run("nested address from recipe", func(t *testing.T) {
		user := factory.User().
			AddressFromRecipe(
				factory.Address().City("Boston").State("MA"),
			).
			Build(p)

		if user.Address.City != "Boston" {
			t.Errorf("expected Address.City='Boston', got %q", user.Address.City)
		}
		if user.Address.State != "MA" {
			t.Errorf("expected Address.State='MA', got %q", user.Address.State)
		}
	})

	t.Run("literal address reuse", func(t *testing.T) {
		addr := factory.Address().City("NYC").Build(p)
		users := factory.User().Address(addr).Many(2, p)

		if len(users) != 2 {
			t.Fatalf("expected 2 users, got %d", len(users))
		}

		// Both users should have the same address (same object)
		if users[0].Address.City != "NYC" || users[1].Address.City != "NYC" {
			t.Error("expected both users to have NYC address")
		}
	})

	t.Run("address recipe generates unique instances", func(t *testing.T) {
		users := factory.User().
			AddressFromRecipe(factory.Address().State("CA")).
			Many(2, p)

		if len(users) != 2 {
			t.Fatalf("expected 2 users, got %d", len(users))
		}

		// Each user should have unique address (different cities)
		if users[0].Address.City == users[1].Address.City {
			t.Error("expected users to have different addresses when using FromRecipe")
		}
		// But both should have CA state
		if users[0].Address.State != "CA" || users[1].Address.State != "CA" {
			t.Error("expected both users to have CA state")
		}
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

		if product.Name != "Widget" {
			t.Errorf("expected Name='Widget', got %q", product.Name)
		}
		if product.Price != 19.99 {
			t.Errorf("expected Price=19.99, got %f", product.Price)
		}
		if !product.InStock {
			t.Error("expected InStock=true")
		}
		if product.ID == "" {
			t.Error("expected non-empty ID")
		}
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

		if order.User.Email != "customer@example.com" {
			t.Errorf("expected User.Email='customer@example.com', got %q", order.User.Email)
		}
		if order.Status != "pending" {
			t.Errorf("expected Status='pending', got %q", order.Status)
		}
		if order.Total != 100.50 {
			t.Errorf("expected Total=100.50, got %f", order.Total)
		}
		if order.ID == "" {
			t.Error("expected non-empty ID")
		}
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

		if len(order.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(order.Items))
		}
		if order.Items[0].Product.Name != "Item 1" {
			t.Errorf("expected first item name='Item 1', got %q", order.Items[0].Product.Name)
		}
		if order.Total != 35.00 {
			t.Errorf("expected Total=35.00, got %f", order.Total)
		}
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

		if post.Title != "My First Post" {
			t.Errorf("expected Title='My First Post', got %q", post.Title)
		}
		if post.Content != "Hello world!" {
			t.Errorf("expected Content='Hello world!', got %q", post.Content)
		}
		if !post.Published {
			t.Error("expected Published=true")
		}
		if !post.PublishedAt.Equal(now) {
			t.Error("expected PublishedAt to match provided time")
		}
		if post.Author.FirstName != "Alice" {
			t.Errorf("expected Author.FirstName='Alice', got %q", post.Author.FirstName)
		}
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

		if comment.Content != "Great post!" {
			t.Errorf("expected Content='Great post!', got %q", comment.Content)
		}
		if comment.Post.Title != "Test Post" {
			t.Errorf("expected Post.Title='Test Post', got %q", comment.Post.Title)
		}
		if comment.Author.FirstName != "Bob" {
			t.Errorf("expected Author.FirstName='Bob', got %q", comment.Author.FirstName)
		}
		if comment.ID == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("multiple comments on same post", func(t *testing.T) {
		post := factory.BlogPost().Title("Shared Post").Build(p)

		comments := factory.Comment().
			Post(post).
			Many(3, p)

		if len(comments) != 3 {
			t.Fatalf("expected 3 comments, got %d", len(comments))
		}

		// All comments should reference the same post
		for i, c := range comments {
			if c.Post.Title != "Shared Post" {
				t.Errorf("comment[%d] expected Post.Title='Shared Post', got %q", i, c.Post.Title)
			}
		}
	})
}

func TestDeterministicGeneration(t *testing.T) {
	t.Run("same primitives produce same values", func(t *testing.T) {
		p1 := specta.New()
		p2 := specta.New()

		user1 := factory.User().FirstName("Test").Build(p1)
		user2 := factory.User().FirstName("Test").Build(p2)

		if user1.ID != user2.ID {
			t.Errorf("expected same ID with same primitives, got %q and %q", user1.ID, user2.ID)
		}
		if user1.Email != user2.Email {
			t.Errorf("expected same Email with same primitives, got %q and %q", user1.Email, user2.Email)
		}
	})

	t.Run("sequential generation increments counter", func(t *testing.T) {
		p := specta.New()

		addr1 := factory.Address().Build(p)
		addr2 := factory.Address().Build(p)

		if addr1.Street == addr2.Street {
			t.Error("expected different Street values on sequential generation")
		}
		if addr1.City == addr2.City {
			t.Error("expected different City values on sequential generation")
		}
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
		if !order1.User.Active || !order2.User.Active {
			t.Error("expected both users to be active")
		}
		if order1.User.FirstName != "Composite" || order2.User.FirstName != "Composite" {
			t.Error("expected both users to have FirstName='Composite'")
		}

		// But they should have different IDs (unique instances)
		if order1.User.ID == order2.User.ID {
			t.Error("expected different user IDs (unique instances from recipe)")
		}
	})
}

func TestTimeHandling(t *testing.T) {
	p := specta.New()

	t.Run("default times are deterministic", func(t *testing.T) {
		user1 := factory.User().Build(p)
		// Reset primitives
		p = specta.New()
		user2 := factory.User().Build(p)

		if !user1.CreatedAt.Equal(user2.CreatedAt) {
			t.Error("expected deterministic CreatedAt times")
		}
	})

	t.Run("custom times can be set", func(t *testing.T) {
		customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		product := factory.Product().CreatedAt(customTime).Build(p)

		if !product.CreatedAt.Equal(customTime) {
			t.Errorf("expected CreatedAt=%v, got %v", customTime, product.CreatedAt)
		}
	})

	t.Run("sequential calls advance time", func(t *testing.T) {
		user1 := factory.User().Build(p)
		user2 := factory.User().Build(p)

		if !user2.CreatedAt.After(user1.CreatedAt) {
			t.Error("expected second user CreatedAt to be after first")
		}
	})
}

func TestBankAccountFactory(t *testing.T) {
	p := specta.New()

	t.Run("constructor-based factory with default values", func(t *testing.T) {
		account := factory.BankAccount().Build(p)

		name := account.GetName()
		balance := account.GetBalance()

		if name == "" {
			t.Error("expected non-empty Name")
		}
		// Default balance should be generated
		if balance == 0 {
			t.Error("expected non-zero balance from default provider")
		}
	})

	t.Run("constructor-based factory with custom values", func(t *testing.T) {
		account := factory.BankAccount().
			Name("Alice").
			Balance(1000).
			Build(p)

		if account.GetName() != "Alice" {
			t.Errorf("expected Name='Alice', got %q", account.GetName())
		}
		if account.GetBalance() != 1000 {
			t.Errorf("expected Balance=1000, got %d", account.GetBalance())
		}
	})

	t.Run("many generates unique instances", func(t *testing.T) {
		accounts := factory.BankAccount().Many(3, p)

		if len(accounts) != 3 {
			t.Fatalf("expected 3 accounts, got %d", len(accounts))
		}

		// Names should be unique
		names := make(map[string]bool)
		for i, acc := range accounts {
			name := acc.GetName()
			if names[name] {
				t.Errorf("duplicate name found: %q at index %d", name, i)
			}
			names[name] = true
		}

		if len(names) != 3 {
			t.Errorf("expected 3 unique names, got %d", len(names))
		}
	})
}

func TestEmailFactory(t *testing.T) {
	p := specta.New()

	t.Run("Build with valid email returns no error", func(t *testing.T) {
		email, err := factory.Email().Address("user@example.com").Build(p)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if email.GetAddress() != "user@example.com" {
			t.Errorf("expected Address='user@example.com', got %q", email.GetAddress())
		}
	})

	t.Run("Build with invalid email returns error", func(t *testing.T) {
		_, err := factory.Email().Address("notanemail").Build(p)

		if err == nil {
			t.Error("expected error for invalid email")
		}
	})

	t.Run("Build with default generates invalid email and returns error", func(t *testing.T) {
		// Default provider generates "address_" which is not a valid email
		_, err := factory.Email().Build(p)

		if err == nil {
			t.Error("expected error with default generation (address_ is not valid)")
		}
	})

	t.Run("Provider panics on error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected Provider to panic on invalid email")
			}
		}()

		provider := factory.Email().Address("invalid").Provider()
		provider(p) // Should panic
	})

	t.Run("Many panics on error", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected Many to panic on invalid email")
			} else {
				// Verify the panic message includes the item index
				msg := fmt.Sprint(r)
				if !strings.Contains(msg, "Many() failed on item 0") {
					t.Errorf("expected panic message to include item index, got: %v", r)
				}
			}
		}()

		factory.Email().Address("invalid").Many(3, p) // Should panic on first item
	})

	t.Run("Many succeeds with valid emails", func(t *testing.T) {
		emails := factory.Email().Address("test@example.com").Many(3, p)

		if len(emails) != 3 {
			t.Fatalf("expected 3 emails, got %d", len(emails))
		}
		for i, email := range emails {
			if email.GetAddress() != "test@example.com" {
				t.Errorf("email %d: expected Address='test@example.com', got %q", i, email.GetAddress())
			}
		}
	})
}

func TestIDGeneration(t *testing.T) {
	p := specta.New()

	t.Run("IDs are unique", func(t *testing.T) {
		users := factory.User().Many(5, p)

		seen := make(map[string]bool)
		for i, u := range users {
			if seen[u.ID] {
				t.Errorf("duplicate ID found: %q at index %d", u.ID, i)
			}
			seen[u.ID] = true
		}

		if len(seen) != 5 {
			t.Errorf("expected 5 unique IDs, got %d", len(seen))
		}
	})

	t.Run("ID format", func(t *testing.T) {
		user := factory.User().Build(p)

		if len(user.ID) == 0 {
			t.Error("expected non-empty ID")
		}
		// IDs should start with "id_"
		if len(user.ID) < 3 || user.ID[:3] != "id_" {
			t.Errorf("expected ID to start with 'id_', got %q", user.ID)
		}
	})
}
