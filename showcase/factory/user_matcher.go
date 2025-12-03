package factory

import (
	"strings"

	"github.com/james-w/specta"
	"github.com/james-w/specta/showcase"
)

// Custom matchers for User type.
// This file demonstrates extending generated matchers with domain-specific assertions.

// IsAdmin matches users with admin privileges.
func IsAdmin() specta.Matcher[showcase.User] {
	return UserMatches().
		FirstName(specta.Equal("Admin")).
		LastName(specta.Equal("User")).
		Email(specta.Equal("admin@example.com")).
		Active(specta.Equal(true))
}

// IsActive matches users where Active is true.
func IsActive() specta.Matcher[showcase.User] {
	return UserMatches().Active(specta.Equal(true))
}

// IsGuest matches users with guest privileges.
func IsGuest() specta.Matcher[showcase.User] {
	return UserMatches().
		FirstName(specta.Equal("Guest")).
		LastName(specta.Equal("User")).
		Active(specta.Equal(false))
}

// HasTestEmail matches users with @test.example.com emails.
func HasTestEmail() specta.Matcher[showcase.User] {
	return specta.MatcherFunc[showcase.User](func(actual showcase.User) specta.MatchResult {
		if strings.HasSuffix(actual.Email, "@test.example.com") {
			return specta.MatchResult{Matched: true}
		}
		return specta.MatchResult{
			Matched: false,
			Message: "expected email to end with @test.example.com, got: " + actual.Email,
		}
	})
}
