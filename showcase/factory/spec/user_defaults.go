package spec

import (
	"fmt"

	"github.com/james-w/specta"
)

// Custom default providers for User type.
// This file demonstrates overriding generated defaults with custom logic.

func init() {
	// Override email generator to use a more realistic test pattern
	UserEmailGenerator = specta.Build("UserEmail", func(d specta.DataSource) (string, error) {
		counter := d.DrawBits(32)
		return fmt.Sprintf("user%d@test.example.com", counter), nil
	})

	// Override Active generator to always be true (active by default)
	UserActiveGenerator = specta.Just(true)
}
