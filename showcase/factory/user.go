package factory

// Custom recipe methods for User type.
// This file demonstrates extending generated recipes with domain-specific methods.

// WithAdminRole configures a user with admin privileges.
func (r UserRecipe) WithAdminRole() UserRecipe {
	return r.FirstName("Admin").LastName("User").Email("admin@example.com").Active(true)
}

// WithGuestRole configures a user as a guest with limited access.
func (r UserRecipe) WithGuestRole() UserRecipe {
	return r.FirstName("Guest").LastName("User").Active(false)
}

// WithTestEmail sets a test email with the given username.
func (r UserRecipe) WithTestEmail(username string) UserRecipe {
	return r.Email(username + "@test.example.com")
}

// AdminUser returns a recipe for an admin user.
// This is a convenience function for creating admin users in tests.
func AdminUser() UserRecipe {
	return User().WithAdminRole()
}

// GuestUser returns a recipe for a guest user.
// This is a convenience function for creating guest users in tests.
func GuestUser() UserRecipe {
	return User().WithGuestRole()
}
