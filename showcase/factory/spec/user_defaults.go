package spec

import (
	"fmt"

	"github.com/james-w/specta"
)

// Custom default providers for User type.
// This file demonstrates overriding generated defaults with custom logic.

func init() {
	// Override email generator to use a more realistic test pattern
	UserEmailGenerator = specta.GeneratorFromProvider(func(s specta.Source) string {
		counter := s.DrawBits(32)
		return fmt.Sprintf("user%d@test.example.com", counter)
	})

	// Override Active generator to always be true (active by default)
	UserActiveGenerator = specta.GeneratorFromProvider(func(s specta.Source) bool {
		return true
	})
}
