package specta_test

import (
	"testing"

	"github.com/james-w/specta"
)

// TestIntGenerator_Validation tests that IntGenerator validates constraints
func TestIntGenerator_Validation(t *testing.T) {
	t.Run("Min after conflicting Max panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when Min > Max")
			}
		}()
		specta.Int().Max(5).Min(10)
	})

	t.Run("Max after conflicting Min panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when Max < Min")
			}
		}()
		specta.Int().Min(10).Max(5)
	})

	t.Run("Positive after Negative panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when Positive after Negative")
			}
		}()
		specta.Int().Negative().Positive()
	})

	t.Run("Negative after Positive panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when Negative after Positive")
			}
		}()
		specta.Int().Positive().Negative()
	})

	t.Run("Range with min > max panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when Range(min > max)")
			}
		}()
		specta.Int().Range(10, 5)
	})

	t.Run("valid constraints don't panic", func(t *testing.T) {
		// These should all work fine
		_ = specta.Int().Min(0).Max(100)
		_ = specta.Int().Max(100).Min(0)
		_ = specta.Int().Range(0, 100)
		_ = specta.Int().Positive()
		_ = specta.Int().Negative()
		_ = specta.Int().NonNegative()
	})
}

// TestStringGenerator_Validation tests that StringGenerator validates constraints
func TestStringGenerator_Validation(t *testing.T) {
	t.Run("MinLen after conflicting MaxLen panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when MinLen > MaxLen")
			}
		}()
		specta.String().MaxLen(5).MinLen(10)
	})

	t.Run("MaxLen after conflicting MinLen panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when MaxLen < MinLen")
			}
		}()
		specta.String().MinLen(10).MaxLen(5)
	})

	t.Run("Prefix exceeding MaxLen panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when Prefix length > MaxLen")
			}
		}()
		specta.String().MaxLen(5).Prefix("this is too long")
	})

	t.Run("Suffix exceeding MaxLen panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when Suffix length > MaxLen")
			}
		}()
		specta.String().MaxLen(5).Suffix("too long")
	})

	t.Run("Prefix+Suffix exceeding MaxLen panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when Prefix + Suffix > MaxLen")
			}
		}()
		specta.String().MaxLen(10).Prefix("foo").Suffix("barbazqux") // 3 + 9 = 12 > 10
	})

	t.Run("MaxLen after Prefix+Suffix panics if too small", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when MaxLen < existing Prefix + Suffix")
			}
		}()
		specta.String().Prefix("foo").Suffix("bar").MaxLen(5)
	})

	t.Run("valid constraints don't panic", func(t *testing.T) {
		// These should all work fine
		_ = specta.String().MinLen(0).MaxLen(100)
		_ = specta.String().MaxLen(100).MinLen(0)
		_ = specta.String().Len(50)
		_ = specta.String().Prefix("foo").Suffix("bar").MaxLen(10)
		_ = specta.String().Prefix("foo").Suffix("bar").MinLen(6)
		_ = specta.String().NonEmpty()
		_ = specta.String().Printable()
	})
}
