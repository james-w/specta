package specta

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// ==============================================================================
// Nil/Pointer Matchers
// ==============================================================================

// IsNil creates a matcher that checks if a pointer is nil.
func IsNil[T any]() Matcher[*T] {
	return MatcherFunc[*T](func(actual *T) MatchResult {
		if actual == nil {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected nil but got %#v", actual),
			Expected: nil,
			Actual:   actual,
		}
	})
}

// IsNotNil creates a matcher that checks if a pointer is not nil.
func IsNotNil[T any]() Matcher[*T] {
	return MatcherFunc[*T](func(actual *T) MatchResult {
		if actual != nil {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  "expected non-nil pointer but got nil",
			Expected: "non-nil",
			Actual:   nil,
		}
	})
}

// PointsTo creates a matcher that checks if a pointer points to a value matching the given matcher.
// Returns false if the pointer is nil.
func PointsTo[T any](matcher Matcher[T]) Matcher[*T] {
	return MatcherFunc[*T](func(actual *T) MatchResult {
		if actual == nil {
			return MatchResult{
				Matched:  false,
				Message:  "expected pointer to match but got nil",
				Expected: "non-nil pointer",
				Actual:   nil,
			}
		}
		result := matcher.Matches(*actual)
		if !result.Matched {
			result.Message = fmt.Sprintf("pointer value didn't match: %s", result.Message)
		}
		return result
	})
}

// ==============================================================================
// Error Matchers
// ==============================================================================

// IsError creates a matcher that checks if an error is non-nil.
func IsError() Matcher[error] {
	return MatcherFunc[error](func(actual error) MatchResult {
		if actual != nil {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  "expected an error but got nil",
			Expected: "non-nil error",
			Actual:   nil,
		}
	})
}

// NoErr creates a matcher that checks if an error is nil.
// This is a convenience matcher for the common case of asserting no error occurred.
func NoErr() Matcher[error] {
	return MatcherFunc[error](func(actual error) MatchResult {
		if actual == nil {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected no error but got: %v", actual),
			Expected: nil,
			Actual:   actual.Error(),
		}
	})
}

// ErrorContains creates a matcher that checks if an error's message contains the expected substring.
// Returns false if the error is nil.
func ErrorContains(substr string) Matcher[error] {
	return MatcherFunc[error](func(actual error) MatchResult {
		if actual == nil {
			return MatchResult{
				Matched:  false,
				Message:  fmt.Sprintf("expected error containing %q but got nil", substr),
				Expected: fmt.Sprintf("error containing %q", substr),
				Actual:   nil,
			}
		}
		errMsg := actual.Error()
		if contains(errMsg, substr) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected error containing %q but got %q", substr, errMsg),
			Expected: fmt.Sprintf("contains %q", substr),
			Actual:   errMsg,
		}
	})
}

// ErrorIs creates a matcher that checks if an error matches the target error using errors.Is.
// This checks the error chain, not just equality.
func ErrorIs(target error) Matcher[error] {
	return MatcherFunc[error](func(actual error) MatchResult {
		if errors.Is(actual, target) {
			return MatchResult{Matched: true}
		}
		if actual == nil {
			return MatchResult{
				Matched:  false,
				Message:  fmt.Sprintf("expected error matching %v but got nil", target),
				Expected: target,
				Actual:   nil,
			}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected error matching %v but got %v", target, actual),
			Expected: target,
			Actual:   actual,
		}
	})
}

// ErrorAs creates a matcher that checks if an error can be assigned to the target type using errors.As.
// The target parameter must be a pointer to an error type.
func ErrorAs[T error](target *T) Matcher[error] {
	return MatcherFunc[error](func(actual error) MatchResult {
		if errors.As(actual, target) {
			return MatchResult{Matched: true}
		}
		if actual == nil {
			return MatchResult{
				Matched:  false,
				Message:  fmt.Sprintf("expected error of type %T but got nil", target),
				Expected: fmt.Sprintf("error of type %T", target),
				Actual:   nil,
			}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected error of type %T but got %T: %v", target, actual, actual),
			Expected: fmt.Sprintf("error of type %T", target),
			Actual:   fmt.Sprintf("%T: %v", actual, actual),
		}
	})
}

// ==============================================================================
// Time Matchers
// ==============================================================================

// After creates a matcher that checks if a time is after the expected time.
func After(expected time.Time) Matcher[time.Time] {
	return MatcherFunc[time.Time](func(actual time.Time) MatchResult {
		if actual.After(expected) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected time after %v but got %v", expected, actual),
			Expected: fmt.Sprintf("after %v", expected),
			Actual:   actual,
		}
	})
}

// Before creates a matcher that checks if a time is before the expected time.
func Before(expected time.Time) Matcher[time.Time] {
	return MatcherFunc[time.Time](func(actual time.Time) MatchResult {
		if actual.Before(expected) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected time before %v but got %v", expected, actual),
			Expected: fmt.Sprintf("before %v", expected),
			Actual:   actual,
		}
	})
}

// Between creates a matcher that checks if a time is between start and end (inclusive).
func Between(start, end time.Time) Matcher[time.Time] {
	return MatcherFunc[time.Time](func(actual time.Time) MatchResult {
		if (actual.Equal(start) || actual.After(start)) && (actual.Equal(end) || actual.Before(end)) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected time between %v and %v but got %v", start, end, actual),
			Expected: fmt.Sprintf("between %v and %v", start, end),
			Actual:   actual,
		}
	})
}

// WithinDuration creates a matcher that checks if a time is within delta duration of the expected time.
// This is useful for fuzzy time comparisons where exact equality isn't practical.
func WithinDuration(expected time.Time, delta time.Duration) Matcher[time.Time] {
	return MatcherFunc[time.Time](func(actual time.Time) MatchResult {
		diff := actual.Sub(expected)
		if diff < 0 {
			diff = -diff
		}
		if diff <= delta {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected time within %v of %v but got %v (diff: %v)", delta, expected, actual, diff),
			Expected: fmt.Sprintf("within %v of %v", delta, expected),
			Actual:   fmt.Sprintf("%v (diff: %v)", actual, diff),
		}
	})
}

// ==============================================================================
// Regex Matchers
// ==============================================================================

// MatchesRegex creates a matcher that checks if a string matches the given regex pattern.
// The pattern is compiled internally; if compilation fails, the match will fail.
func MatchesRegex(pattern string) Matcher[string] {
	return MatcherFunc[string](func(actual string) MatchResult {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return MatchResult{
				Matched:  false,
				Message:  fmt.Sprintf("invalid regex pattern %q: %v", pattern, err),
				Expected: fmt.Sprintf("valid regex %q", pattern),
				Actual:   fmt.Sprintf("invalid: %v", err),
			}
		}
		if re.MatchString(actual) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected string matching regex %q but got %q", pattern, actual),
			Expected: fmt.Sprintf("matches /%s/", pattern),
			Actual:   actual,
		}
	})
}

// MatchesPattern creates a matcher that checks if a string matches the given compiled regex.
func MatchesPattern(regex *regexp.Regexp) Matcher[string] {
	return MatcherFunc[string](func(actual string) MatchResult {
		if regex.MatchString(actual) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected string matching pattern %s but got %q", regex.String(), actual),
			Expected: fmt.Sprintf("matches /%s/", regex.String()),
			Actual:   actual,
		}
	})
}

// Helper function for string contains check (used by ErrorContains)
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
