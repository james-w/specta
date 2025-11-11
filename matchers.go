package specta

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/rliebz/ghost/ghostlib"
)

// Matcher defines an interface for matching values of type T.
type Matcher[T any] interface {
	Matches(actual T) MatchResult
}

// MatchResult represents the result of a match operation.
type MatchResult struct {
	Matched  bool     // Whether the match succeeded
	Message  string   // Human-readable description of the failure
	Details  []string // Additional details about what didn't match
	Expected any      // Expected value (optional, for better error messages)
	Actual   any      // Actual value (optional, for better error messages)
	Path     string   // Field path for nested failures (e.g., "User.Address.City")
}

// MatcherFunc is a function that implements the Matcher interface.
type MatcherFunc[T any] func(T) MatchResult

func (f MatcherFunc[T]) Matches(actual T) MatchResult {
	return f(actual)
}

// AssertThat checks if actual matches the given matcher, failing the test if not.
func AssertThat[T any](t *testing.T, actual T, matcher Matcher[T]) {
	t.Helper()

	// Capture the expression from AST for better error messages
	args := ghostlib.ArgsFromAST(actual)
	expr := ""
	if len(args) > 0 {
		expr = args[0]
	}

	result := matcher.Matches(actual)
	if !result.Matched {
		var msg string
		if strings.Contains(result.Message, "\n") {
			// Multi-line: add "didn't match:" header
			if expr != "" {
				msg = fmt.Sprintf("%s didn't match:\n%s", expr, result.Message)
			} else {
				msg = result.Message
			}
			t.Errorf("\n%s", msg)
		} else {
			// Single-line: prepend with colon
			if expr != "" {
				msg = fmt.Sprintf("%s: %s", expr, result.Message)
			} else {
				msg = result.Message
			}
			t.Errorf("%s", msg)
		}
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
			Matched:  false,
			Message:  fmt.Sprintf("expected %#v but got %#v", expected, actual),
			Expected: expected,
			Actual:   actual,
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

		// Check if both values are structs - use structured diff
		expectedVal := reflect.ValueOf(expected)
		actualVal := reflect.ValueOf(actual)
		if expectedVal.Kind() == reflect.Struct && actualVal.Kind() == reflect.Struct {
			return MatchResult{
				Matched:  false,
				Message:  buildReflectionStructDiff(expected, actual),
				Expected: expected,
				Actual:   actual,
			}
		}

		// For non-structs, use simple format
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected %#v but got %#v", expected, actual),
			Expected: expected,
			Actual:   actual,
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
			Message: fmt.Sprintf("expected zero value but got %#v", actual),
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
			Matched:  false,
			Message:  fmt.Sprintf("expected value > %v but got %v", threshold, actual),
			Expected: fmt.Sprintf("> %v", threshold),
			Actual:   actual,
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
			Matched:  false,
			Message:  fmt.Sprintf("expected value < %v but got %v", threshold, actual),
			Expected: fmt.Sprintf("< %v", threshold),
			Actual:   actual,
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
			Matched:  false,
			Message:  fmt.Sprintf("expected value >= %v but got %v", threshold, actual),
			Expected: fmt.Sprintf(">= %v", threshold),
			Actual:   actual,
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
		var successes []string
		for i, matcher := range matchers {
			result := matcher.Matches(actual)
			if !result.Matched {
				failures = append(failures, fmt.Sprintf("matcher %d: %s", i+1, result.Message))
			} else {
				successes = append(successes, fmt.Sprintf("matcher %d: %s", i+1, result.Message))
			}
		}
		if len(failures) > 0 {
			msg := fmt.Sprintf("%d of %d matchers failed:\n", len(failures), len(matchers))
			for _, f := range failures {
				msg += "  ✗ " + f + "\n"
			}
			msg = strings.TrimSuffix(msg, "\n")
			return MatchResult{
				Matched: false,
				Message: msg,
				Details: failures,
			}
		}
		return MatchResult{Matched: true}
	})
}

// AnyOf creates a matcher that requires at least one sub-matcher to match.
func AnyOf[T any](matchers ...Matcher[T]) Matcher[T] {
	return MatcherFunc[T](func(actual T) MatchResult {
		var failures []string
		for i, matcher := range matchers {
			result := matcher.Matches(actual)
			if result.Matched {
				return MatchResult{Matched: true}
			}
			failures = append(failures, fmt.Sprintf("option %d: %s", i+1, result.Message))
		}
		// Show up to 3 failure messages
		msg := "none of the matchers succeeded:\n"
		limit := 3
		if len(failures) < limit {
			limit = len(failures)
		}
		for i := 0; i < limit; i++ {
			msg += "  " + failures[i] + "\n"
		}
		if len(failures) > limit {
			msg += fmt.Sprintf("  ... and %d more", len(failures)-limit)
		} else {
			msg = strings.TrimSuffix(msg, "\n")
		}
		return MatchResult{
			Matched: false,
			Message: msg,
			Details: failures,
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
