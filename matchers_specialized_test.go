package specta_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/james-w/specta"
)

// ==============================================================================
// Nil/Pointer Matchers Tests
// ==============================================================================

func TestIsNil(t *testing.T) {
	t.Run("matches nil pointer", func(t *testing.T) {
		var ptr *int
		matcher := specta.IsNil[int]()
		result := matcher.Matches(ptr)
		if !result.Matched {
			t.Errorf("IsNil should match nil pointer, but got: %s", result.Message)
		}
	})

	t.Run("rejects non-nil pointer", func(t *testing.T) {
		value := 42
		matcher := specta.IsNil[int]()
		result := matcher.Matches(&value)
		if result.Matched {
			t.Error("IsNil should not match non-nil pointer")
		}
		if result.Message == "" {
			t.Error("IsNil should provide error message on mismatch")
		}
	})

	t.Run("works with different types", func(t *testing.T) {
		var strPtr *string
		matcher := specta.IsNil[string]()
		result := matcher.Matches(strPtr)
		if !result.Matched {
			t.Errorf("IsNil should match nil string pointer, but got: %s", result.Message)
		}
	})
}

func TestIsNotNil(t *testing.T) {
	t.Run("matches non-nil pointer", func(t *testing.T) {
		value := 42
		matcher := specta.IsNotNil[int]()
		result := matcher.Matches(&value)
		if !result.Matched {
			t.Errorf("IsNotNil should match non-nil pointer, but got: %s", result.Message)
		}
	})

	t.Run("rejects nil pointer", func(t *testing.T) {
		var ptr *int
		matcher := specta.IsNotNil[int]()
		result := matcher.Matches(ptr)
		if result.Matched {
			t.Error("IsNotNil should not match nil pointer")
		}
		if result.Message == "" {
			t.Error("IsNotNil should provide error message on mismatch")
		}
	})

	t.Run("works with struct pointers", func(t *testing.T) {
		type Person struct{ Name string }
		person := &Person{Name: "Alice"}
		matcher := specta.IsNotNil[Person]()
		result := matcher.Matches(person)
		if !result.Matched {
			t.Errorf("IsNotNil should match non-nil struct pointer, but got: %s", result.Message)
		}
	})
}

func TestPointsTo(t *testing.T) {
	t.Run("matches when value satisfies matcher", func(t *testing.T) {
		value := 42
		matcher := specta.PointsTo(specta.Equal(42))
		result := matcher.Matches(&value)
		if !result.Matched {
			t.Errorf("PointsTo should match when value is 42, but got: %s", result.Message)
		}
	})

	t.Run("rejects when value doesn't satisfy matcher", func(t *testing.T) {
		value := 99
		matcher := specta.PointsTo(specta.Equal(42))
		result := matcher.Matches(&value)
		if result.Matched {
			t.Error("PointsTo should not match when value is not 42")
		}
		if result.Message == "" {
			t.Error("PointsTo should provide error message on mismatch")
		}
	})

	t.Run("rejects nil pointer", func(t *testing.T) {
		var ptr *int
		matcher := specta.PointsTo(specta.Equal(42))
		result := matcher.Matches(ptr)
		if result.Matched {
			t.Error("PointsTo should not match nil pointer")
		}
		if result.Message == "" {
			t.Error("PointsTo should provide error message for nil pointer")
		}
	})

	t.Run("works with complex matchers", func(t *testing.T) {
		value := 50
		matcher := specta.PointsTo(specta.GreaterThan(40))
		result := matcher.Matches(&value)
		if !result.Matched {
			t.Errorf("PointsTo should match with GreaterThan matcher, but got: %s", result.Message)
		}
	})

	t.Run("works with string pointers", func(t *testing.T) {
		value := "hello world"
		matcher := specta.PointsTo(specta.Contains("world"))
		result := matcher.Matches(&value)
		if !result.Matched {
			t.Errorf("PointsTo should match string with Contains matcher, but got: %s", result.Message)
		}
	})
}

// ==============================================================================
// Error Matchers Tests
// ==============================================================================

// CustomError is a test error type for ErrorAs tests
type CustomError struct {
	Code    int
	Message string
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("error %d: %s", e.Code, e.Message)
}

func TestIsError(t *testing.T) {
	t.Run("matches non-nil error", func(t *testing.T) {
		err := errors.New("something went wrong")
		matcher := specta.IsError()
		result := matcher.Matches(err)
		if !result.Matched {
			t.Errorf("IsError should match non-nil error, but got: %s", result.Message)
		}
	})

	t.Run("rejects nil error", func(t *testing.T) {
		var err error
		matcher := specta.IsError()
		result := matcher.Matches(err)
		if result.Matched {
			t.Error("IsError should not match nil error")
		}
		if result.Message == "" {
			t.Error("IsError should provide error message on mismatch")
		}
	})

	t.Run("matches wrapped errors", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
		matcher := specta.IsError()
		result := matcher.Matches(wrappedErr)
		if !result.Matched {
			t.Errorf("IsError should match wrapped error, but got: %s", result.Message)
		}
	})
}

func TestNoErr(t *testing.T) {
	t.Run("matches nil error", func(t *testing.T) {
		var err error
		matcher := specta.NoErr()
		result := matcher.Matches(err)
		if !result.Matched {
			t.Errorf("NoErr should match nil error, but got: %s", result.Message)
		}
	})

	t.Run("rejects non-nil error", func(t *testing.T) {
		err := errors.New("something went wrong")
		matcher := specta.NoErr()
		result := matcher.Matches(err)
		if result.Matched {
			t.Error("NoErr should not match non-nil error")
		}
		if result.Message == "" {
			t.Error("NoErr should provide error message on mismatch")
		}
	})

	t.Run("provides error message in output", func(t *testing.T) {
		err := errors.New("database connection failed")
		matcher := specta.NoErr()
		result := matcher.Matches(err)
		if !result.Matched && result.Actual != "database connection failed" {
			t.Errorf("NoErr should include actual error message in output, got: %s", result.Message)
		}
	})
}

func TestErrorContains(t *testing.T) {
	t.Run("matches when error contains substring", func(t *testing.T) {
		err := errors.New("database connection failed")
		matcher := specta.ErrorContains("connection")
		result := matcher.Matches(err)
		if !result.Matched {
			t.Errorf("ErrorContains should match when substring is present, but got: %s", result.Message)
		}
	})

	t.Run("matches with exact error message", func(t *testing.T) {
		err := errors.New("error")
		matcher := specta.ErrorContains("error")
		result := matcher.Matches(err)
		if !result.Matched {
			t.Errorf("ErrorContains should match exact error message, but got: %s", result.Message)
		}
	})

	t.Run("rejects when substring not found", func(t *testing.T) {
		err := errors.New("database error")
		matcher := specta.ErrorContains("network")
		result := matcher.Matches(err)
		if result.Matched {
			t.Error("ErrorContains should not match when substring is absent")
		}
		if result.Message == "" {
			t.Error("ErrorContains should provide error message on mismatch")
		}
	})

	t.Run("rejects nil error", func(t *testing.T) {
		var err error
		matcher := specta.ErrorContains("something")
		result := matcher.Matches(err)
		if result.Matched {
			t.Error("ErrorContains should not match nil error")
		}
	})

	t.Run("case sensitive matching", func(t *testing.T) {
		err := errors.New("Database Error")
		matcher := specta.ErrorContains("database")
		result := matcher.Matches(err)
		if result.Matched {
			t.Error("ErrorContains should be case sensitive")
		}
	})

	t.Run("matches empty substring", func(t *testing.T) {
		err := errors.New("any error")
		matcher := specta.ErrorContains("")
		result := matcher.Matches(err)
		if !result.Matched {
			t.Errorf("ErrorContains should match empty substring, but got: %s", result.Message)
		}
	})
}

func TestErrorIs(t *testing.T) {
	targetErr := errors.New("target error")

	t.Run("matches identical error", func(t *testing.T) {
		matcher := specta.ErrorIs(targetErr)
		result := matcher.Matches(targetErr)
		if !result.Matched {
			t.Errorf("ErrorIs should match identical error, but got: %s", result.Message)
		}
	})

	t.Run("matches wrapped error", func(t *testing.T) {
		wrappedErr := fmt.Errorf("wrapped: %w", targetErr)
		matcher := specta.ErrorIs(targetErr)
		result := matcher.Matches(wrappedErr)
		if !result.Matched {
			t.Errorf("ErrorIs should match wrapped error, but got: %s", result.Message)
		}
	})

	t.Run("rejects different error", func(t *testing.T) {
		differentErr := errors.New("different error")
		matcher := specta.ErrorIs(targetErr)
		result := matcher.Matches(differentErr)
		if result.Matched {
			t.Error("ErrorIs should not match different error")
		}
	})

	t.Run("rejects nil error", func(t *testing.T) {
		var err error
		matcher := specta.ErrorIs(targetErr)
		result := matcher.Matches(err)
		if result.Matched {
			t.Error("ErrorIs should not match nil error")
		}
	})
}

func TestErrorAs(t *testing.T) {
	t.Run("matches when error can be assigned to target type", func(t *testing.T) {
		err := &CustomError{Code: 404, Message: "not found"}
		var target *CustomError
		matcher := specta.ErrorAs(&target)
		result := matcher.Matches(err)
		if !result.Matched {
			t.Errorf("ErrorAs should match when error is assignable to target, but got: %s", result.Message)
		}
		if target.Code != 404 {
			t.Errorf("ErrorAs should populate target, expected Code=404 but got %d", target.Code)
		}
	})

	t.Run("matches wrapped error of target type", func(t *testing.T) {
		customErr := &CustomError{Code: 500, Message: "server error"}
		wrappedErr := fmt.Errorf("wrapped: %w", customErr)
		var target *CustomError
		matcher := specta.ErrorAs(&target)
		result := matcher.Matches(wrappedErr)
		if !result.Matched {
			t.Errorf("ErrorAs should match wrapped error of target type, but got: %s", result.Message)
		}
	})

	t.Run("rejects error of different type", func(t *testing.T) {
		err := errors.New("regular error")
		var target *CustomError
		matcher := specta.ErrorAs(&target)
		result := matcher.Matches(err)
		if result.Matched {
			t.Error("ErrorAs should not match error of different type")
		}
	})

	t.Run("rejects nil error", func(t *testing.T) {
		var err error
		var target *CustomError
		matcher := specta.ErrorAs(&target)
		result := matcher.Matches(err)
		if result.Matched {
			t.Error("ErrorAs should not match nil error")
		}
	})
}

// ==============================================================================
// Time Matchers Tests
// ==============================================================================

func TestAfter(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	t.Run("matches time after expected", func(t *testing.T) {
		matcher := specta.After(now)
		result := matcher.Matches(future)
		if !result.Matched {
			t.Errorf("After should match future time, but got: %s", result.Message)
		}
	})

	t.Run("rejects time before expected", func(t *testing.T) {
		matcher := specta.After(now)
		result := matcher.Matches(past)
		if result.Matched {
			t.Error("After should not match past time")
		}
		if result.Message == "" {
			t.Error("After should provide error message on mismatch")
		}
	})

	t.Run("rejects equal time", func(t *testing.T) {
		matcher := specta.After(now)
		result := matcher.Matches(now)
		if result.Matched {
			t.Error("After should not match equal time")
		}
	})
}

func TestBefore(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	t.Run("matches time before expected", func(t *testing.T) {
		matcher := specta.Before(now)
		result := matcher.Matches(past)
		if !result.Matched {
			t.Errorf("Before should match past time, but got: %s", result.Message)
		}
	})

	t.Run("rejects time after expected", func(t *testing.T) {
		matcher := specta.Before(now)
		result := matcher.Matches(future)
		if result.Matched {
			t.Error("Before should not match future time")
		}
		if result.Message == "" {
			t.Error("Before should provide error message on mismatch")
		}
	})

	t.Run("rejects equal time", func(t *testing.T) {
		matcher := specta.Before(now)
		result := matcher.Matches(now)
		if result.Matched {
			t.Error("Before should not match equal time")
		}
	})
}

func TestBetween(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	middle := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	before := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("matches time in middle of range", func(t *testing.T) {
		matcher := specta.Between(start, end)
		result := matcher.Matches(middle)
		if !result.Matched {
			t.Errorf("Between should match time in middle of range, but got: %s", result.Message)
		}
	})

	t.Run("matches time at start boundary", func(t *testing.T) {
		matcher := specta.Between(start, end)
		result := matcher.Matches(start)
		if !result.Matched {
			t.Errorf("Between should match time at start boundary, but got: %s", result.Message)
		}
	})

	t.Run("matches time at end boundary", func(t *testing.T) {
		matcher := specta.Between(start, end)
		result := matcher.Matches(end)
		if !result.Matched {
			t.Errorf("Between should match time at end boundary, but got: %s", result.Message)
		}
	})

	t.Run("rejects time before range", func(t *testing.T) {
		matcher := specta.Between(start, end)
		result := matcher.Matches(before)
		if result.Matched {
			t.Error("Between should not match time before range")
		}
		if result.Message == "" {
			t.Error("Between should provide error message on mismatch")
		}
	})

	t.Run("rejects time after range", func(t *testing.T) {
		matcher := specta.Between(start, end)
		result := matcher.Matches(after)
		if result.Matched {
			t.Error("Between should not match time after range")
		}
	})
}

func TestWithinDuration(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("matches time within delta", func(t *testing.T) {
		closeTime := baseTime.Add(30 * time.Second)
		matcher := specta.WithinDuration(baseTime, 1*time.Minute)
		result := matcher.Matches(closeTime)
		if !result.Matched {
			t.Errorf("WithinDuration should match time within delta, but got: %s", result.Message)
		}
	})

	t.Run("matches time exactly at delta", func(t *testing.T) {
		exactTime := baseTime.Add(1 * time.Minute)
		matcher := specta.WithinDuration(baseTime, 1*time.Minute)
		result := matcher.Matches(exactTime)
		if !result.Matched {
			t.Errorf("WithinDuration should match time at exact delta, but got: %s", result.Message)
		}
	})

	t.Run("matches time within negative delta", func(t *testing.T) {
		pastTime := baseTime.Add(-30 * time.Second)
		matcher := specta.WithinDuration(baseTime, 1*time.Minute)
		result := matcher.Matches(pastTime)
		if !result.Matched {
			t.Errorf("WithinDuration should match time within negative delta, but got: %s", result.Message)
		}
	})

	t.Run("matches equal time with zero delta", func(t *testing.T) {
		matcher := specta.WithinDuration(baseTime, 0)
		result := matcher.Matches(baseTime)
		if !result.Matched {
			t.Errorf("WithinDuration should match equal time with zero delta, but got: %s", result.Message)
		}
	})

	t.Run("rejects time outside delta", func(t *testing.T) {
		farTime := baseTime.Add(2 * time.Minute)
		matcher := specta.WithinDuration(baseTime, 1*time.Minute)
		result := matcher.Matches(farTime)
		if result.Matched {
			t.Error("WithinDuration should not match time outside delta")
		}
		if result.Message == "" {
			t.Error("WithinDuration should provide error message on mismatch")
		}
	})

	t.Run("works with large durations", func(t *testing.T) {
		futureTime := baseTime.Add(6 * time.Hour)
		matcher := specta.WithinDuration(baseTime, 12*time.Hour)
		result := matcher.Matches(futureTime)
		if !result.Matched {
			t.Errorf("WithinDuration should work with large durations, but got: %s", result.Message)
		}
	})
}

// ==============================================================================
// Regex Matchers Tests
// ==============================================================================

func TestMatchesRegex(t *testing.T) {
	t.Run("matches string with valid regex", func(t *testing.T) {
		matcher := specta.MatchesRegex(`^\d{3}-\d{4}$`)
		result := matcher.Matches("123-4567")
		if !result.Matched {
			t.Errorf("MatchesRegex should match valid pattern, but got: %s", result.Message)
		}
	})

	t.Run("rejects string not matching pattern", func(t *testing.T) {
		matcher := specta.MatchesRegex(`^\d{3}-\d{4}$`)
		result := matcher.Matches("abc-defg")
		if result.Matched {
			t.Error("MatchesRegex should not match invalid pattern")
		}
		if result.Message == "" {
			t.Error("MatchesRegex should provide error message on mismatch")
		}
	})

	t.Run("works with simple patterns", func(t *testing.T) {
		matcher := specta.MatchesRegex("hello")
		result := matcher.Matches("hello world")
		if !result.Matched {
			t.Errorf("MatchesRegex should match simple pattern, but got: %s", result.Message)
		}
	})

	t.Run("works with email pattern", func(t *testing.T) {
		matcher := specta.MatchesRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		result := matcher.Matches("user@example.com")
		if !result.Matched {
			t.Errorf("MatchesRegex should match email pattern, but got: %s", result.Message)
		}
	})

	t.Run("rejects invalid regex pattern", func(t *testing.T) {
		matcher := specta.MatchesRegex(`[invalid(`)
		result := matcher.Matches("anything")
		if result.Matched {
			t.Error("MatchesRegex should fail on invalid regex pattern")
		}
		if result.Message == "" {
			t.Error("MatchesRegex should provide error message for invalid pattern")
		}
	})

	t.Run("handles case sensitivity", func(t *testing.T) {
		matcher := specta.MatchesRegex(`^HELLO$`)
		result := matcher.Matches("hello")
		if result.Matched {
			t.Error("MatchesRegex should be case sensitive by default")
		}
	})

	t.Run("works with case insensitive flag", func(t *testing.T) {
		matcher := specta.MatchesRegex(`(?i)^HELLO$`)
		result := matcher.Matches("hello")
		if !result.Matched {
			t.Errorf("MatchesRegex should support case insensitive flag, but got: %s", result.Message)
		}
	})
}

func TestMatchesPattern(t *testing.T) {
	t.Run("matches string with compiled regex", func(t *testing.T) {
		pattern := regexp.MustCompile(`^\d{3}-\d{4}$`)
		matcher := specta.MatchesPattern(pattern)
		result := matcher.Matches("123-4567")
		if !result.Matched {
			t.Errorf("MatchesPattern should match valid pattern, but got: %s", result.Message)
		}
	})

	t.Run("rejects string not matching pattern", func(t *testing.T) {
		pattern := regexp.MustCompile(`^\d{3}-\d{4}$`)
		matcher := specta.MatchesPattern(pattern)
		result := matcher.Matches("abc-defg")
		if result.Matched {
			t.Error("MatchesPattern should not match invalid pattern")
		}
		if result.Message == "" {
			t.Error("MatchesPattern should provide error message on mismatch")
		}
	})

	t.Run("works with complex patterns", func(t *testing.T) {
		pattern := regexp.MustCompile(`^[A-Z][a-z]+\s[A-Z][a-z]+$`)
		matcher := specta.MatchesPattern(pattern)
		result := matcher.Matches("John Smith")
		if !result.Matched {
			t.Errorf("MatchesPattern should match name pattern, but got: %s", result.Message)
		}
	})

	t.Run("reuses compiled pattern efficiently", func(t *testing.T) {
		pattern := regexp.MustCompile(`\d+`)
		matcher := specta.MatchesPattern(pattern)

		// Test multiple matches
		if !matcher.Matches("123").Matched {
			t.Error("MatchesPattern should match first string")
		}
		if !matcher.Matches("456").Matched {
			t.Error("MatchesPattern should match second string")
		}
		if matcher.Matches("abc").Matched {
			t.Error("MatchesPattern should not match non-numeric string")
		}
	})
}
