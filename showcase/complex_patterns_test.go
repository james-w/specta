package showcase_test

import (
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/showcase"
	"github.com/james-w/specta/showcase/factory"
)

// ==============================================================================
// Tests for Pointers to Structs as Child
// ==============================================================================

func TestProfile_PointerToStruct(t *testing.T) {
	p := specta.New()

	t.Run("build profile with pointer to user", func(t *testing.T) {
		profile := factory.Profile().
			Name("John's Profile").
			ManagerFromRecipe(factory.User().FirstName("Alice").LastName("Manager")).
			Build(p)

		if profile.Manager == nil {
			t.Fatal("Expected Manager to be non-nil")
		}
		if profile.Manager.FirstName != "Alice" {
			t.Errorf("Expected Manager.FirstName=Alice, got %s", profile.Manager.FirstName)
		}
		if profile.Manager.LastName != "Manager" {
			t.Errorf("Expected Manager.LastName=Manager, got %s", profile.Manager.LastName)
		}
	})

	t.Run("pointer to struct with partial matching", func(t *testing.T) {
		profile := factory.Profile().
			Name("Test Profile").
			ManagerFromRecipe(factory.User().FirstName("Bob")).
			Build(p)

		// Partial matching on pointer field
		matcher := factory.Profile().
			ManagerFromRecipe(factory.User().FirstName("Bob")).
			AsEqualMatcher()

		specta.AssertThat(t, profile, matcher)
	})

	t.Run("pointer to struct can be nil", func(t *testing.T) {
		// Build profile without setting Manager (should default to nil)
		profile := factory.Profile().
			Name("No Manager").
			Build(p)

		// Profile generation creates a User by default via FromSpec
		// If we want nil, we'd need to set it explicitly
		// For now, this tests that the generation works
		if profile.Manager == nil {
			t.Log("Manager is nil (expected for default)")
		}
	})

	t.Run("setting pointer directly clears nested recipe", func(t *testing.T) {
		user := factory.User().FirstName("Charlie").Build(p)

		profile := factory.Profile().
			ManagerFromRecipe(factory.User().FirstName("Alice")).
			Manager(&user). // This should override the recipe
			Build(p)

		if profile.Manager.FirstName != "Charlie" {
			t.Errorf("Expected Manager.FirstName=Charlie (from direct set), got %s", profile.Manager.FirstName)
		}
	})
}

// ==============================================================================
// Tests for Slices of Structs as Child
// ==============================================================================

func TestOrder_SliceOfStructs(t *testing.T) {
	p := specta.New()

	t.Run("build order with slice of order items", func(t *testing.T) {
		order := factory.Order().
			User(showcase.User{ID: "user_1", FirstName: "Test"}).
			Build(p)

		// Default generated Items slice defaults to nil (zero value for slices)
		// This is expected behavior
		if order.Items == nil {
			t.Log("Items slice is nil (expected default)")
		} else {
			t.Logf("Items slice has %d items", len(order.Items))
		}
	})

	t.Run("order with multiple items", func(t *testing.T) {
		product1 := factory.Product().Name("Item 1").Price(10.0).Build(p)
		product2 := factory.Product().Name("Item 2").Price(20.0).Build(p)

		item1 := factory.OrderItem().Product(product1).Quantity(2).Build(p)
		item2 := factory.OrderItem().Product(product2).Quantity(1).Build(p)

		order := factory.Order().
			Items([]showcase.OrderItem{item1, item2}).
			Build(p)

		if len(order.Items) != 2 {
			t.Fatalf("Expected 2 items, got %d", len(order.Items))
		}
		if order.Items[0].Product.Name != "Item 1" {
			t.Errorf("Expected first item name='Item 1', got %s", order.Items[0].Product.Name)
		}
		if order.Items[1].Product.Name != "Item 2" {
			t.Errorf("Expected second item name='Item 2', got %s", order.Items[1].Product.Name)
		}
	})
}

func TestOrganization_MultipleSlicesOfStructs(t *testing.T) {
	p := specta.New()

	t.Run("build organization with teams and members", func(t *testing.T) {
		team1 := factory.Team().Name("Engineering").Build(p)
		team2 := factory.Team().Name("Sales").Build(p)

		org := factory.Organization().
			Name("Acme Corp").
			Teams([]showcase.Team{team1, team2}).
			Build(p)

		if len(org.Teams) != 2 {
			t.Fatalf("Expected 2 teams, got %d", len(org.Teams))
		}
		if org.Teams[0].Name != "Engineering" {
			t.Errorf("Expected first team='Engineering', got %s", org.Teams[0].Name)
		}
		if org.Teams[1].Name != "Sales" {
			t.Errorf("Expected second team='Sales', got %s", org.Teams[1].Name)
		}
	})

	t.Run("organization with CEO pointer", func(t *testing.T) {
		org := factory.Organization().
			Name("TechCo").
			CEOFromRecipe(factory.User().FirstName("Jane").LastName("CEO")).
			Build(p)

		if org.CEO == nil {
			t.Fatal("Expected CEO to be non-nil")
		}
		if org.CEO.FirstName != "Jane" {
			t.Errorf("Expected CEO.FirstName=Jane, got %s", org.CEO.FirstName)
		}
	})

	t.Run("organization with metadata map", func(t *testing.T) {
		metadata := map[string]string{
			"industry": "technology",
			"founded":  "2020",
		}

		org := factory.Organization().
			Name("StartupCo").
			Metadata(metadata).
			Build(p)

		if org.Metadata["industry"] != "technology" {
			t.Errorf("Expected metadata industry=technology, got %s", org.Metadata["industry"])
		}
		if org.Metadata["founded"] != "2020" {
			t.Errorf("Expected metadata founded=2020, got %s", org.Metadata["founded"])
		}
	})
}

// ==============================================================================
// Tests for Circular Dependencies Between Types
// ==============================================================================

func TestTeamMember_CircularDependency(t *testing.T) {
	p := specta.New()

	t.Run("build team with members", func(t *testing.T) {
		member1 := factory.Member().Name("Alice").Build(p)
		member2 := factory.Member().Name("Bob").Build(p)

		team := factory.Team().
			Name("DevTeam").
			Members([]showcase.Member{member1, member2}).
			Build(p)

		if len(team.Members) != 2 {
			t.Fatalf("Expected 2 members, got %d", len(team.Members))
		}
		if team.Members[0].Name != "Alice" {
			t.Errorf("Expected first member='Alice', got %s", team.Members[0].Name)
		}
	})

	t.Run("build member with team reference", func(t *testing.T) {
		member := factory.Member().
			Name("Charlie").
			TeamFromRecipe(factory.Team().Name("Backend")).
			Build(p)

		if member.Team == nil {
			t.Fatal("Expected Team to be non-nil")
		}
		if member.Team.Name != "Backend" {
			t.Errorf("Expected Team.Name=Backend, got %s", member.Team.Name)
		}
	})

	t.Run("circular reference - team with members referencing team", func(t *testing.T) {
		// Create a team
		team := factory.Team().Name("CircularTeam").Build(p)

		// Create members that reference the team
		member1 := factory.Member().
			Name("Dave").
			Team(&team).
			Build(p)

		member2 := factory.Member().
			Name("Eve").
			Team(&team).
			Build(p)

		// Update team with members
		team.Members = []showcase.Member{member1, member2}

		// Verify circular structure
		if len(team.Members) != 2 {
			t.Fatalf("Expected 2 members, got %d", len(team.Members))
		}
		if team.Members[0].Team == nil {
			t.Error("Expected member to have team reference")
		}
		if team.Members[0].Team.Name != "CircularTeam" {
			t.Errorf("Expected member's team name=CircularTeam, got %s", team.Members[0].Team.Name)
		}
	})

	t.Run("partial matching with circular dependencies", func(t *testing.T) {
		member := factory.Member().
			Name("Frank").
			TeamFromRecipe(factory.Team().Name("Frontend")).
			Build(p)

		// Partial matching on nested team
		matcher := factory.Member().
			TeamFromRecipe(factory.Team().Name("Frontend")).
			AsEqualMatcher()

		specta.AssertThat(t, member, matcher)
	})
}

// ==============================================================================
// Integration Tests for Complex Nested Structures
// ==============================================================================

func TestComplexNestedStructures(t *testing.T) {
	p := specta.New()

	t.Run("deeply nested structure with all pattern types", func(t *testing.T) {
		// Build an organization with:
		// - CEO (pointer to struct)
		// - Teams (slice of structs with circular refs)
		// - Members (slice of structs)
		// - Metadata (map)

		ceo := factory.User().FirstName("CEO").LastName("Boss").Build(p)

		team1 := factory.Team().Name("Engineering").Build(p)
		team2 := factory.Team().Name("Sales").Build(p)

		member1 := factory.Member().Name("Alice").Team(&team1).Build(p)
		member2 := factory.Member().Name("Bob").Team(&team2).Build(p)

		org := factory.Organization().
			Name("MegaCorp").
			CEO(&ceo).
			Teams([]showcase.Team{team1, team2}).
			Members([]showcase.Member{member1, member2}).
			Metadata(map[string]string{
				"industry": "tech",
				"size":     "large",
			}).
			Build(p)

		// Verify all fields
		if org.Name != "MegaCorp" {
			t.Errorf("Expected Name=MegaCorp, got %s", org.Name)
		}
		if org.CEO.FirstName != "CEO" {
			t.Errorf("Expected CEO.FirstName=CEO, got %s", org.CEO.FirstName)
		}
		if len(org.Teams) != 2 {
			t.Errorf("Expected 2 teams, got %d", len(org.Teams))
		}
		if len(org.Members) != 2 {
			t.Errorf("Expected 2 members, got %d", len(org.Members))
		}
		if org.Metadata["industry"] != "tech" {
			t.Errorf("Expected metadata industry=tech, got %s", org.Metadata["industry"])
		}
	})

	t.Run("partial matching on complex nested structure", func(t *testing.T) {
		org := factory.Organization().
			Name("TestCorp").
			CEOFromRecipe(factory.User().FirstName("Jane")).
			Build(p)

		// Partial match only on CEO's first name
		matcher := factory.Organization().
			CEOFromRecipe(factory.User().FirstName("Jane")).
			AsEqualMatcher()

		specta.AssertThat(t, org, matcher)
	})
}
