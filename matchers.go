package gomatchers

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Matcher defines an interface for matching values of type T.
type Matcher[T any] interface {
	Matches(actual T) MatchResult
}

// MatchResult represents the result of a match operation.
type MatchResult struct {
	Matched bool     // Whether the match succeeded
	Message string   // Human-readable description of the failure
	Details []string // Additional details about what didn't match
}

// MatcherFunc is a function that implements the Matcher interface.
type MatcherFunc[T any] func(T) MatchResult

func (f MatcherFunc[T]) Matches(actual T) MatchResult {
	return f(actual)
}

// AssertThat checks if actual matches the given matcher, failing the test if not.
func AssertThat[T any](t *testing.T, actual T, matcher Matcher[T]) {
	t.Helper()
	result := matcher.Matches(actual)
	if !result.Matched {
		t.Errorf("Assertion failed: %s", result.Message)
		for _, detail := range result.Details {
			t.Errorf("  %s", detail)
		}
	}
}

// ==============================================================================
// Basic Value Matchers
// ==============================================================================

// Equal creates a matcher that checks for exact equality.
func Equal[T comparable](expected T) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		if actual == expected {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected %v but got %v", expected, actual),
		}
	})
}

// DeepEqual creates a matcher that uses reflect.DeepEqual for comparison.
// This works with any type, including slices, maps, and other non-comparable types.
func DeepEqual[T any](expected T) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		if reflect.DeepEqual(actual, expected) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected %v but got %v", expected, actual),
		}
	})
}

// Is is an alias for Equal for more readable assertions.
func Is[T comparable](expected T) Matcher[T] {
	return Equal(expected)
}

// Not negates a matcher.
func Not[T any](matcher Matcher[T]) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		result := matcher.Matches(actual)
		if result.Matched {
			return MatchResult{
				Matched: false,
				Message: fmt.Sprintf("expected not to match, but it did"),
			}
		}
		return MatchResult{Matched: true}
	})
}

// IsZero checks if a value is the zero value for its type.
func IsZero[T any]() Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		var zero T
		if reflect.DeepEqual(actual, zero) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected zero value but got %v", actual),
		}
	})
}

// ==============================================================================
// Numeric Matchers
// ==============================================================================

// GreaterThan creates a matcher for numeric types.
func GreaterThan[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}](threshold T) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		if actual > threshold {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected value > %v but got %v", threshold, actual),
		}
	})
}

// LessThan creates a matcher for numeric types.
func LessThan[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}](threshold T) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		if actual < threshold {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected value < %v but got %v", threshold, actual),
		}
	})
}

// GreaterThanOrEqual creates a matcher for numeric types.
func GreaterThanOrEqual[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}](threshold T) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		if actual >= threshold {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected value >= %v but got %v", threshold, actual),
		}
	})
}

// ==============================================================================
// String Matchers
// ==============================================================================

// Contains checks if a string contains a substring.
func Contains(substr string) Matcher[string] {
	return MatcherFunc[string](func(actual string) MatchResult {
		if strings.Contains(actual, substr) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected string to contain %q but got %q", substr, actual),
		}
	})
}

// HasPrefix checks if a string has a given prefix.
func HasPrefix(prefix string) Matcher[string] {
	return MatcherFunc[string](func(actual string) MatchResult {
		if strings.HasPrefix(actual, prefix) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected string to start with %q but got %q", prefix, actual),
		}
	})
}

// HasSuffix checks if a string has a given suffix.
func HasSuffix(suffix string) Matcher[string] {
	return MatcherFunc[string](func(actual string) MatchResult {
		if strings.HasSuffix(actual, suffix) {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected string to end with %q but got %q", suffix, actual),
		}
	})
}

// ==============================================================================
// Boolean Matchers
// ==============================================================================

// IsTrue checks if a boolean is true.
func IsTrue() Matcher[bool] {
	return MatcherFunc[bool](func(actual bool) MatchResult {
		if actual {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: "expected true but got false",
		}
	})
}

// IsFalse checks if a boolean is false.
func IsFalse() Matcher[bool] {
	return MatcherFunc[bool](func(actual bool) MatchResult {
		if !actual {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched: false,
			Message: "expected false but got true",
		}
	})
}

// ==============================================================================
// Combinators
// ==============================================================================

// AllOf creates a matcher that requires all sub-matchers to match.
func AllOf[T any](matchers ...Matcher[T]) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		var failures []string
		for i, matcher := range matchers {
			result := matcher.Matches(actual)
			if !result.Matched {
				failures = append(failures, fmt.Sprintf("matcher %d: %s", i, result.Message))
				failures = append(failures, result.Details...)
			}
		}
		if len(failures) > 0 {
			return MatchResult{
				Matched: false,
				Message: "not all matchers succeeded",
				Details: failures,
			}
		}
		return MatchResult{Matched: true}
	})
}

// AnyOf creates a matcher that requires at least one sub-matcher to match.
func AnyOf[T any](matchers ...Matcher[T]) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		for _, matcher := range matchers {
			result := matcher.Matches(actual)
			if result.Matched {
				return MatchResult{Matched: true}
			}
		}
		return MatchResult{
			Matched: false,
			Message: "none of the matchers succeeded",
		}
	})
}

// ==============================================================================
// Field Extractors
// ==============================================================================

// Field creates a matcher that extracts a value from T using an extractor function,
// then applies a matcher to the extracted value. This is useful for:
// - Matching on computed/derived values
// - Matching on values from getter methods
// - Matching on unexported fields (via getters)
//
// Example:
//
//	Field("balance * 2", func(acc BankAccount) int {
//	    return acc.GetBalance() * 2
//	}, Equal(2000))
func Field[T any, V any](name string, extractor func(T) V, matcher Matcher[V]) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		value := extractor(actual)
		result := matcher.Matches(value)
		if !result.Matched {
			return MatchResult{
				Matched: false,
				Message: fmt.Sprintf("%s: %s", name, result.Message),
				Details: result.Details,
			}
		}
		return MatchResult{Matched: true}
	})
}
