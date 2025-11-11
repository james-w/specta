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
