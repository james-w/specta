package spec

import (
	"fmt"

	"github.com/james-w/specta"
)

// Custom default providers for User type.
// This file demonstrates overriding generated defaults with custom logic.

func init() {
	// Override email default to use a more realistic test pattern
	UserDefaultEmail = func(p specta.Primitives) string {
		return fmt.Sprintf("user%d@test.example.com", p.Next())
	}

	// Override Active default to always be true (active by default)
	UserDefaultActive = func(p specta.Primitives) bool {
		return true
	}
}
