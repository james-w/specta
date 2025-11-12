package showcase_test

import (
	"testing"

	testgen "github.com/james-w/specta"
	"github.com/james-w/specta/showcase/factory"
)

func TestAccountNestedRecipeTracking(t *testing.T) {
	p := testgen.New()

	// Test 1: Build an account with a nested user recipe
	account := factory.Account().
		UserFromRecipe(factory.User().FirstName("Alice").LastName("Smith")).
		Status("active").
		Build(p)

	// Verify the account was built correctly
	if account.GetUser().FirstName != "Alice" {
		t.Errorf("Expected FirstName=Alice, got %s", account.GetUser().FirstName)
	}
	if account.GetUser().LastName != "Smith" {
		t.Errorf("Expected LastName=Smith, got %s", account.GetUser().LastName)
	}
	if account.GetStatus() != "active" {
		t.Errorf("Expected Status=active, got %s", account.GetStatus())
	}
}

func TestAccountPartialMatchingWithNestedRecipe(t *testing.T) {
	p := testgen.New()

	// Build an account with specific user fields
	account := factory.Account().
		UserFromRecipe(factory.User().FirstName("Alice")).
		Build(p)

	// Test 2: AsEqualMatcher should use nested recipe for partial matching
	matcher := factory.Account().
		UserFromRecipe(factory.User().FirstName("Alice")).
		AsEqualMatcher()

	// The matcher should pass because the FirstName matches
	result := matcher.Matches(account)
	if !result.Matched {
		t.Errorf("Expected match to pass, but got failure: %s", result.Message)
	}

	// Test 3: Matcher should fail if FirstName doesn't match
	matcherBob := factory.Account().
		UserFromRecipe(factory.User().FirstName("Bob")).
		AsEqualMatcher()

	result = matcherBob.Matches(account)
	if result.Matched {
		t.Errorf("Expected match to fail for different FirstName, but it passed")
	}
}

func TestAccountFromRecipeCreatesUniqueInstances(t *testing.T) {
	p := testgen.New()

	// Test 4: UserFromRecipe should create unique user instances for each account
	accounts := factory.Account().
		UserFromRecipe(factory.User()).
		Status("active").
		Many(5, p)

	if len(accounts) != 5 {
		t.Errorf("Expected 5 accounts, got %d", len(accounts))
	}

	// Verify all accounts have different user IDs (unique instances)
	userIDs := make(map[string]bool)
	for i, account := range accounts {
		id := account.GetUser().ID
		if userIDs[id] {
			t.Errorf("Account %d has duplicate user ID: %s", i, id)
		}
		userIDs[id] = true
	}
}

func TestAccountDirectUserSetClearsNestedRecipe(t *testing.T) {
	p := testgen.New()

	// Test 5: Setting User directly should clear the nested recipe
	user := factory.User().FirstName("Charlie").Build(p)

	account := factory.Account().
		UserFromRecipe(factory.User().FirstName("Alice")).
		User(user). // This should clear the nested recipe
		Build(p)

	// Verify the account uses the directly set user, not the recipe
	if account.GetUser().FirstName != "Charlie" {
		t.Errorf("Expected FirstName=Charlie, got %s", account.GetUser().FirstName)
	}
}

func TestDeeplyNestedRecipePartialMatching(t *testing.T) {
	p := testgen.New()

	// Test 6: Deeply nested recipes (Account → User → Address)
	// Build an account with deeply nested recipe specifications
	account := factory.Account().
		UserFromRecipe(
			factory.User().
				FirstName("Alice").
				AddressFromRecipe(
					factory.Address().City("NYC").State("NY"),
				),
		).
		Status("active").
		Build(p)

	// Verify the deeply nested structure was built correctly
	if account.GetUser().FirstName != "Alice" {
		t.Errorf("Expected FirstName=Alice, got %s", account.GetUser().FirstName)
	}
	if account.GetUser().Address.City != "NYC" {
		t.Errorf("Expected City=NYC, got %s", account.GetUser().Address.City)
	}
	if account.GetUser().Address.State != "NY" {
		t.Errorf("Expected State=NY, got %s", account.GetUser().Address.State)
	}

	// Test recursive partial matching with deeply nested recipes
	// This should only check: FirstName=Alice, Address.City=NYC, Address.State=NY
	matcher := factory.Account().
		UserFromRecipe(
			factory.User().
				FirstName("Alice").
				AddressFromRecipe(
					factory.Address().City("NYC").State("NY"),
				),
		).
		AsEqualMatcher()

	result := matcher.Matches(account)
	if !result.Matched {
		t.Errorf("Expected deeply nested partial match to pass, but got failure: %s", result.Message)
	}

	// Test that deeply nested partial matching fails appropriately
	accountWithDifferentCity := factory.Account().
		UserFromRecipe(
			factory.User().
				FirstName("Alice").
				AddressFromRecipe(
					factory.Address().City("LA"), // Different city
				),
		).
		Build(p)

	result = matcher.Matches(accountWithDifferentCity)
	if result.Matched {
		t.Errorf("Expected deeply nested partial match to fail for different city, but it passed")
	}
}
