package showcase_test

import (
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/showcase"
	"github.com/james-w/specta/showcase/factory"
)

func TestAccountNestedRecipeTracking(t *testing.T) {
	p := specta.New()

	// Test 1: Build an account with a nested user recipe
	account := factory.Account().
		UserFromRecipe(factory.User().FirstName("Alice").LastName("Smith")).
		Status("active").
		Build(p)

	// Verify the account was built correctly
	specta.AssertThat(t, account.GetUser().FirstName, specta.Equal("Alice"))
	specta.AssertThat(t, account.GetUser().LastName, specta.Equal("Smith"))
	specta.AssertThat(t, account.GetStatus(), specta.Equal("active"))
}

func TestAccountPartialMatchingWithNestedRecipe(t *testing.T) {
	p := specta.New()

	// Build an account with specific user fields
	account := factory.Account().
		UserFromRecipe(factory.User().FirstName("Alice")).
		Build(p)

	// Test 2: AsEqualMatcher should use nested recipe for partial matching
	matcher := factory.Account().
		UserFromRecipe(factory.User().FirstName("Alice")).
		AsEqualMatcher()

	// The matcher should pass because the FirstName matches
	specta.AssertThat(t, account, matcher)

	// Test 3: Matcher should fail if FirstName doesn't match
	matcherBob := factory.Account().
		UserFromRecipe(factory.User().FirstName("Bob")).
		AsEqualMatcher()

	specta.AssertThat(t, account, specta.Not(matcherBob))
}

func TestAccountFromRecipeCreatesUniqueInstances(t *testing.T) {
	p := specta.New()

	// Test 4: UserFromRecipe should create unique user instances for each account
	accounts := factory.Account().
		UserFromRecipe(factory.User()).
		Status("active").
		Many(5, p)

	specta.AssertThat(t, accounts, specta.HasSize[showcase.Account](5))

	// Verify all accounts have different user IDs (unique instances)
	userIDs := make(map[string]bool)
	for _, account := range accounts {
		id := account.GetUser().ID
		specta.AssertThat(t, userIDs[id], specta.IsFalse())
		userIDs[id] = true
	}
}

func TestAccountDirectUserSetClearsNestedRecipe(t *testing.T) {
	p := specta.New()

	// Test 5: Setting User directly should clear the nested recipe
	user := factory.User().FirstName("Charlie").Build(p)

	account := factory.Account().
		UserFromRecipe(factory.User().FirstName("Alice")).
		User(user). // This should clear the nested recipe
		Build(p)

	// Verify the account uses the directly set user, not the recipe
	specta.AssertThat(t, account.GetUser().FirstName, specta.Equal("Charlie"))
}

func TestDeeplyNestedRecipePartialMatching(t *testing.T) {
	p := specta.New()

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
	specta.AssertThat(t, account.GetUser().FirstName, specta.Equal("Alice"))
	specta.AssertThat(t, account.GetUser().Address.City, specta.Equal("NYC"))
	specta.AssertThat(t, account.GetUser().Address.State, specta.Equal("NY"))

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

	specta.AssertThat(t, account, matcher)

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

	specta.AssertThat(t, accountWithDifferentCity, specta.Not(matcher))
}
