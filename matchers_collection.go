package specta

import "fmt"

// ==============================================================================
// Slice Size Matchers
// ==============================================================================

// HasSize creates a matcher that checks if a slice has exactly the expected size.
func HasSize[T any](expectedSize int) Matcher[[]T] {
	return MatcherFunc[[]T](func(actual []T) MatchResult {
		actualSize := len(actual)
		if actualSize == expectedSize {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected slice of size %d but got size %d", expectedSize, actualSize),
			Expected: expectedSize,
			Actual:   actualSize,
		}
	})
}

// IsEmpty creates a matcher that checks if a slice is empty.
func IsEmpty[T any]() Matcher[[]T] {
	return MatcherFunc[[]T](func(actual []T) MatchResult {
		if len(actual) == 0 {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected empty slice but got size %d", len(actual)),
			Expected: 0,
			Actual:   len(actual),
		}
	})
}

// IsNotEmpty creates a matcher that checks if a slice has at least one element.
func IsNotEmpty[T any]() Matcher[[]T] {
	return MatcherFunc[[]T](func(actual []T) MatchResult {
		if len(actual) > 0 {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  "expected non-empty slice but got empty slice",
			Expected: "> 0",
			Actual:   0,
		}
	})
}

// ==============================================================================
// Slice Content Matchers
// ==============================================================================

// ContainsElement creates a matcher that checks if a slice contains the expected item.
func ContainsElement[T comparable](item T) Matcher[[]T] {
	return MatcherFunc[[]T](func(actual []T) MatchResult {
		for _, v := range actual {
			if v == item {
				return MatchResult{Matched: true}
			}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected slice to contain %#v but it was not found (size: %d)", item, len(actual)),
			Expected: fmt.Sprintf("contains %#v", item),
			Actual:   actual,
		}
	})
}

// ContainsAllElements creates a matcher that checks if a slice contains all expected items.
func ContainsAllElements[T comparable](items ...T) Matcher[[]T] {
	return MatcherFunc[[]T](func(actual []T) MatchResult {
		// Create a set of actual items for efficient lookup
		actualSet := make(map[T]bool)
		for _, v := range actual {
			actualSet[v] = true
		}

		// Check each expected item
		var missing []T
		for _, item := range items {
			if !actualSet[item] {
				missing = append(missing, item)
			}
		}

		if len(missing) == 0 {
			return MatchResult{Matched: true}
		}

		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected slice to contain all of %#v but missing %#v", items, missing),
			Expected: fmt.Sprintf("contains all of %#v", items),
			Actual:   actual,
		}
	})
}

// ContainsAnyElement creates a matcher that checks if a slice contains at least one of the expected items.
func ContainsAnyElement[T comparable](items ...T) Matcher[[]T] {
	return MatcherFunc[[]T](func(actual []T) MatchResult {
		// Create a set of actual items for efficient lookup
		actualSet := make(map[T]bool)
		for _, v := range actual {
			actualSet[v] = true
		}

		// Check if any expected item exists
		for _, item := range items {
			if actualSet[item] {
				return MatchResult{Matched: true}
			}
		}

		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected slice to contain at least one of %#v but none were found", items),
			Expected: fmt.Sprintf("contains any of %#v", items),
			Actual:   actual,
		}
	})
}

// ==============================================================================
// Slice Predicate Matchers
// ==============================================================================

// Every creates a matcher that checks if all elements in a slice match the given matcher.
func Every[T any](matcher Matcher[T]) Matcher[[]T] {
	return MatcherFunc[[]T](func(actual []T) MatchResult {
		for i, v := range actual {
			result := matcher.Matches(v)
			if !result.Matched {
				return MatchResult{
					Matched: false,
					Message: fmt.Sprintf("expected all elements to match, but element at index %d failed: %s", i, result.Message),
					Details: []string{fmt.Sprintf("element[%d] = %#v", i, v)},
				}
			}
		}
		return MatchResult{Matched: true}
	})
}

// Any creates a matcher that checks if at least one element in a slice matches the given matcher.
func Any[T any](matcher Matcher[T]) Matcher[[]T] {
	return MatcherFunc[[]T](func(actual []T) MatchResult {
		if len(actual) == 0 {
			return MatchResult{
				Matched: false,
				Message: "expected at least one element to match, but slice is empty",
			}
		}

		for _, v := range actual {
			result := matcher.Matches(v)
			if result.Matched {
				return MatchResult{Matched: true}
			}
		}

		return MatchResult{
			Matched: false,
			Message: fmt.Sprintf("expected at least one element to match, but none of the %d elements matched", len(actual)),
		}
	})
}

// None creates a matcher that checks if no elements in a slice match the given matcher.
func None[T any](matcher Matcher[T]) Matcher[[]T] {
	return MatcherFunc[[]T](func(actual []T) MatchResult {
		for i, v := range actual {
			result := matcher.Matches(v)
			if result.Matched {
				return MatchResult{
					Matched: false,
					Message: fmt.Sprintf("expected no elements to match, but element at index %d matched", i),
					Details: []string{fmt.Sprintf("element[%d] = %#v", i, v)},
				}
			}
		}
		return MatchResult{Matched: true}
	})
}

// ==============================================================================
// Map Matchers
// ==============================================================================

// HasKey creates a matcher that checks if a map contains the specified key.
func HasKey[K comparable, V any](key K) Matcher[map[K]V] {
	return MatcherFunc[map[K]V](func(actual map[K]V) MatchResult {
		if _, exists := actual[key]; exists {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected map to have key %#v but it was not found (map size: %d)", key, len(actual)),
			Expected: fmt.Sprintf("has key %#v", key),
			Actual:   actual,
		}
	})
}

// HasValue creates a matcher that checks if a map contains the specified value.
func HasValue[K comparable, V comparable](value V) Matcher[map[K]V] {
	return MatcherFunc[map[K]V](func(actual map[K]V) MatchResult {
		for _, v := range actual {
			if v == value {
				return MatchResult{Matched: true}
			}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected map to have value %#v but it was not found (map size: %d)", value, len(actual)),
			Expected: fmt.Sprintf("has value %#v", value),
			Actual:   actual,
		}
	})
}

// HasEntry creates a matcher that checks if a map contains the specified key-value pair.
func HasEntry[K comparable, V comparable](key K, value V) Matcher[map[K]V] {
	return MatcherFunc[map[K]V](func(actual map[K]V) MatchResult {
		actualValue, exists := actual[key]
		if !exists {
			return MatchResult{
				Matched:  false,
				Message:  fmt.Sprintf("expected map to have key %#v but it was not found", key),
				Expected: fmt.Sprintf("has entry [%#v]=%#v", key, value),
				Actual:   actual,
			}
		}
		if actualValue == value {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected map[%#v] to be %#v but got %#v", key, value, actualValue),
			Expected: value,
			Actual:   actualValue,
		}
	})
}

// MapHasSize creates a matcher that checks if a map has exactly the expected size.
func MapHasSize[K comparable, V any](expectedSize int) Matcher[map[K]V] {
	return MatcherFunc[map[K]V](func(actual map[K]V) MatchResult {
		actualSize := len(actual)
		if actualSize == expectedSize {
			return MatchResult{Matched: true}
		}
		return MatchResult{
			Matched:  false,
			Message:  fmt.Sprintf("expected map of size %d but got size %d", expectedSize, actualSize),
			Expected: expectedSize,
			Actual:   actualSize,
		}
	})
}
