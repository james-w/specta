package specta_test

import (
	"testing"

	"github.com/james-w/specta"
)

// ==============================================================================
// Slice Size Matchers Tests
// ==============================================================================

func TestHasSize(t *testing.T) {
	t.Run("matches slice with expected size", func(t *testing.T) {
		matcher := specta.HasSize[int](3)
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("HasSize(3) should match slice of size 3, but got: %s", result.Message)
		}
	})

	t.Run("rejects slice with different size", func(t *testing.T) {
		matcher := specta.HasSize[int](3)
		result := matcher.Matches([]int{1, 2})
		if result.Matched {
			t.Error("HasSize(3) should not match slice of size 2")
		}
		if result.Message == "" {
			t.Error("HasSize should provide error message on mismatch")
		}
	})

	t.Run("matches empty slice when size is 0", func(t *testing.T) {
		matcher := specta.HasSize[string](0)
		result := matcher.Matches([]string{})
		if !result.Matched {
			t.Errorf("HasSize(0) should match empty slice, but got: %s", result.Message)
		}
	})

	t.Run("matches nil slice when size is 0", func(t *testing.T) {
		matcher := specta.HasSize[int](0)
		var nilSlice []int
		result := matcher.Matches(nilSlice)
		if !result.Matched {
			t.Errorf("HasSize(0) should match nil slice, but got: %s", result.Message)
		}
	})
}

func TestIsEmpty(t *testing.T) {
	t.Run("matches empty slice", func(t *testing.T) {
		matcher := specta.IsEmpty[int]()
		result := matcher.Matches([]int{})
		if !result.Matched {
			t.Errorf("IsEmpty should match empty slice, but got: %s", result.Message)
		}
	})

	t.Run("matches nil slice", func(t *testing.T) {
		matcher := specta.IsEmpty[string]()
		var nilSlice []string
		result := matcher.Matches(nilSlice)
		if !result.Matched {
			t.Errorf("IsEmpty should match nil slice, but got: %s", result.Message)
		}
	})

	t.Run("rejects non-empty slice", func(t *testing.T) {
		matcher := specta.IsEmpty[int]()
		result := matcher.Matches([]int{1})
		if result.Matched {
			t.Error("IsEmpty should not match non-empty slice")
		}
		if result.Message == "" {
			t.Error("IsEmpty should provide error message on mismatch")
		}
	})
}

func TestIsNotEmpty(t *testing.T) {
	t.Run("matches non-empty slice", func(t *testing.T) {
		matcher := specta.IsNotEmpty[int]()
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("IsNotEmpty should match non-empty slice, but got: %s", result.Message)
		}
	})

	t.Run("matches single element slice", func(t *testing.T) {
		matcher := specta.IsNotEmpty[string]()
		result := matcher.Matches([]string{"hello"})
		if !result.Matched {
			t.Errorf("IsNotEmpty should match single element slice, but got: %s", result.Message)
		}
	})

	t.Run("rejects empty slice", func(t *testing.T) {
		matcher := specta.IsNotEmpty[int]()
		result := matcher.Matches([]int{})
		if result.Matched {
			t.Error("IsNotEmpty should not match empty slice")
		}
		if result.Message == "" {
			t.Error("IsNotEmpty should provide error message on mismatch")
		}
	})

	t.Run("rejects nil slice", func(t *testing.T) {
		matcher := specta.IsNotEmpty[int]()
		var nilSlice []int
		result := matcher.Matches(nilSlice)
		if result.Matched {
			t.Error("IsNotEmpty should not match nil slice")
		}
	})
}

// ==============================================================================
// Slice Content Matchers Tests
// ==============================================================================

func TestContainsElement(t *testing.T) {
	t.Run("matches when item is present", func(t *testing.T) {
		matcher := specta.ContainsElement(2)
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("ContainsElement(2) should match [1, 2, 3], but got: %s", result.Message)
		}
	})

	t.Run("matches when item is at start", func(t *testing.T) {
		matcher := specta.ContainsElement("hello")
		result := matcher.Matches([]string{"hello", "world"})
		if !result.Matched {
			t.Errorf("ContainsElement should match when item is at start, but got: %s", result.Message)
		}
	})

	t.Run("matches when item is at end", func(t *testing.T) {
		matcher := specta.ContainsElement(42)
		result := matcher.Matches([]int{1, 2, 42})
		if !result.Matched {
			t.Errorf("ContainsElement should match when item is at end, but got: %s", result.Message)
		}
	})

	t.Run("rejects when item is not present", func(t *testing.T) {
		matcher := specta.ContainsElement(99)
		result := matcher.Matches([]int{1, 2, 3})
		if result.Matched {
			t.Error("ContainsElement(99) should not match [1, 2, 3]")
		}
		if result.Message == "" {
			t.Error("ContainsElement should provide error message on mismatch")
		}
	})

	t.Run("rejects empty slice", func(t *testing.T) {
		matcher := specta.ContainsElement("test")
		result := matcher.Matches([]string{})
		if result.Matched {
			t.Error("ContainsElement should not match empty slice")
		}
	})
}

func TestContainsAllElements(t *testing.T) {
	t.Run("matches when all items are present", func(t *testing.T) {
		matcher := specta.ContainsAllElements(1, 2, 3)
		result := matcher.Matches([]int{1, 2, 3, 4, 5})
		if !result.Matched {
			t.Errorf("ContainsAllElements should match when all items present, but got: %s", result.Message)
		}
	})

	t.Run("matches when all items present in different order", func(t *testing.T) {
		matcher := specta.ContainsAllElements("a", "b", "c")
		result := matcher.Matches([]string{"c", "a", "b"})
		if !result.Matched {
			t.Errorf("ContainsAllElements should match regardless of order, but got: %s", result.Message)
		}
	})

	t.Run("matches with duplicates in actual", func(t *testing.T) {
		matcher := specta.ContainsAllElements(1, 2)
		result := matcher.Matches([]int{1, 1, 2, 2})
		if !result.Matched {
			t.Errorf("ContainsAllElements should handle duplicates, but got: %s", result.Message)
		}
	})

	t.Run("rejects when some items are missing", func(t *testing.T) {
		matcher := specta.ContainsAllElements(1, 2, 3, 4)
		result := matcher.Matches([]int{1, 2})
		if result.Matched {
			t.Error("ContainsAllElements should not match when items are missing")
		}
		if result.Message == "" {
			t.Error("ContainsAllElements should provide error message on mismatch")
		}
	})

	t.Run("matches empty items list", func(t *testing.T) {
		matcher := specta.ContainsAllElements[int]()
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("ContainsAllElements with no items should always match, but got: %s", result.Message)
		}
	})
}

func TestContainsAnyElement(t *testing.T) {
	t.Run("matches when one item is present", func(t *testing.T) {
		matcher := specta.ContainsAnyElement(2, 99, 100)
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("ContainsAnyElement should match when one item present, but got: %s", result.Message)
		}
	})

	t.Run("matches when multiple items are present", func(t *testing.T) {
		matcher := specta.ContainsAnyElement("a", "b", "c")
		result := matcher.Matches([]string{"a", "b"})
		if !result.Matched {
			t.Errorf("ContainsAnyElement should match when multiple items present, but got: %s", result.Message)
		}
	})

	t.Run("matches when all items are present", func(t *testing.T) {
		matcher := specta.ContainsAnyElement(1, 2, 3)
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("ContainsAnyElement should match when all items present, but got: %s", result.Message)
		}
	})

	t.Run("rejects when no items are present", func(t *testing.T) {
		matcher := specta.ContainsAnyElement(99, 100, 101)
		result := matcher.Matches([]int{1, 2, 3})
		if result.Matched {
			t.Error("ContainsAnyElement should not match when no items are present")
		}
		if result.Message == "" {
			t.Error("ContainsAnyElement should provide error message on mismatch")
		}
	})

	t.Run("rejects empty slice", func(t *testing.T) {
		matcher := specta.ContainsAnyElement(1, 2, 3)
		result := matcher.Matches([]int{})
		if result.Matched {
			t.Error("ContainsAnyElement should not match empty slice")
		}
	})
}

// ==============================================================================
// Slice Predicate Matchers Tests
// ==============================================================================

func TestEvery(t *testing.T) {
	t.Run("matches when all elements satisfy matcher", func(t *testing.T) {
		matcher := specta.Every(specta.GreaterThan(0))
		result := matcher.Matches([]int{1, 2, 3, 4})
		if !result.Matched {
			t.Errorf("Every should match when all elements satisfy matcher, but got: %s", result.Message)
		}
	})

	t.Run("matches empty slice", func(t *testing.T) {
		matcher := specta.Every(specta.Equal(42))
		result := matcher.Matches([]int{})
		if !result.Matched {
			t.Errorf("Every should match empty slice (vacuous truth), but got: %s", result.Message)
		}
	})

	t.Run("rejects when one element fails", func(t *testing.T) {
		matcher := specta.Every(specta.GreaterThan(5))
		result := matcher.Matches([]int{6, 7, 3, 9})
		if result.Matched {
			t.Error("Every should not match when one element fails")
		}
		if result.Message == "" {
			t.Error("Every should provide error message on mismatch")
		}
	})

	t.Run("provides index of failing element", func(t *testing.T) {
		matcher := specta.Every(specta.Equal("test"))
		result := matcher.Matches([]string{"test", "test", "fail"})
		if result.Matched {
			t.Error("Every should not match when element fails")
		}
		// Check that the error message mentions the index
		if result.Message == "" || result.Details == nil {
			t.Error("Every should provide details about which element failed")
		}
	})

	t.Run("works with string matchers", func(t *testing.T) {
		matcher := specta.Every(specta.HasPrefix("hello"))
		result := matcher.Matches([]string{"hello world", "hello there"})
		if !result.Matched {
			t.Errorf("Every should work with string matchers, but got: %s", result.Message)
		}
	})
}

func TestAny(t *testing.T) {
	t.Run("matches when one element satisfies matcher", func(t *testing.T) {
		matcher := specta.Any(specta.Equal(42))
		result := matcher.Matches([]int{1, 2, 42, 3})
		if !result.Matched {
			t.Errorf("Any should match when one element satisfies matcher, but got: %s", result.Message)
		}
	})

	t.Run("matches when all elements satisfy matcher", func(t *testing.T) {
		matcher := specta.Any(specta.GreaterThan(0))
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("Any should match when all elements satisfy matcher, but got: %s", result.Message)
		}
	})

	t.Run("matches when first element satisfies", func(t *testing.T) {
		matcher := specta.Any(specta.Equal("first"))
		result := matcher.Matches([]string{"first", "second", "third"})
		if !result.Matched {
			t.Errorf("Any should match when first element satisfies, but got: %s", result.Message)
		}
	})

	t.Run("rejects when no elements satisfy matcher", func(t *testing.T) {
		matcher := specta.Any(specta.Equal(99))
		result := matcher.Matches([]int{1, 2, 3})
		if result.Matched {
			t.Error("Any should not match when no elements satisfy")
		}
		if result.Message == "" {
			t.Error("Any should provide error message on mismatch")
		}
	})

	t.Run("rejects empty slice", func(t *testing.T) {
		matcher := specta.Any(specta.Equal(42))
		result := matcher.Matches([]int{})
		if result.Matched {
			t.Error("Any should not match empty slice")
		}
		if result.Message == "" {
			t.Error("Any should provide error message for empty slice")
		}
	})

	t.Run("works with string matchers", func(t *testing.T) {
		matcher := specta.Any(specta.Contains("@"))
		result := matcher.Matches([]string{"hello", "user@example.com", "world"})
		if !result.Matched {
			t.Errorf("Any should work with string matchers, but got: %s", result.Message)
		}
	})
}

func TestNone(t *testing.T) {
	t.Run("matches when no elements satisfy matcher", func(t *testing.T) {
		matcher := specta.None(specta.Equal(99))
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("None should match when no elements satisfy matcher, but got: %s", result.Message)
		}
	})

	t.Run("matches empty slice", func(t *testing.T) {
		matcher := specta.None(specta.Equal(42))
		result := matcher.Matches([]int{})
		if !result.Matched {
			t.Errorf("None should match empty slice (vacuous truth), but got: %s", result.Message)
		}
	})

	t.Run("rejects when one element satisfies matcher", func(t *testing.T) {
		matcher := specta.None(specta.Equal(2))
		result := matcher.Matches([]int{1, 2, 3})
		if result.Matched {
			t.Error("None should not match when one element satisfies")
		}
		if result.Message == "" {
			t.Error("None should provide error message on mismatch")
		}
	})

	t.Run("provides index of matching element", func(t *testing.T) {
		matcher := specta.None(specta.GreaterThan(5))
		result := matcher.Matches([]int{1, 2, 10, 3})
		if result.Matched {
			t.Error("None should not match when element satisfies")
		}
		// Check that the error message mentions the index
		if result.Message == "" || result.Details == nil {
			t.Error("None should provide details about which element matched")
		}
	})

	t.Run("works with string matchers", func(t *testing.T) {
		matcher := specta.None(specta.Contains("@"))
		result := matcher.Matches([]string{"hello", "world", "test"})
		if !result.Matched {
			t.Errorf("None should work with string matchers, but got: %s", result.Message)
		}
	})
}

// ==============================================================================
// Map Matchers Tests
// ==============================================================================

func TestHasKey(t *testing.T) {
	t.Run("matches when key is present", func(t *testing.T) {
		matcher := specta.HasKey[string, int]("age")
		result := matcher.Matches(map[string]int{"age": 30, "score": 100})
		if !result.Matched {
			t.Errorf("HasKey should match when key is present, but got: %s", result.Message)
		}
	})

	t.Run("matches with zero value", func(t *testing.T) {
		matcher := specta.HasKey[string, int]("count")
		result := matcher.Matches(map[string]int{"count": 0})
		if !result.Matched {
			t.Errorf("HasKey should match even with zero value, but got: %s", result.Message)
		}
	})

	t.Run("rejects when key is not present", func(t *testing.T) {
		matcher := specta.HasKey[string, int]("missing")
		result := matcher.Matches(map[string]int{"age": 30})
		if result.Matched {
			t.Error("HasKey should not match when key is not present")
		}
		if result.Message == "" {
			t.Error("HasKey should provide error message on mismatch")
		}
	})

	t.Run("rejects empty map", func(t *testing.T) {
		matcher := specta.HasKey[string, int]("key")
		result := matcher.Matches(map[string]int{})
		if result.Matched {
			t.Error("HasKey should not match empty map")
		}
	})

	t.Run("rejects nil map", func(t *testing.T) {
		matcher := specta.HasKey[string, int]("key")
		var nilMap map[string]int
		result := matcher.Matches(nilMap)
		if result.Matched {
			t.Error("HasKey should not match nil map")
		}
	})
}

func TestHasValue(t *testing.T) {
	t.Run("matches when value is present", func(t *testing.T) {
		matcher := specta.HasValue[string, int](30)
		result := matcher.Matches(map[string]int{"age": 30, "score": 100})
		if !result.Matched {
			t.Errorf("HasValue should match when value is present, but got: %s", result.Message)
		}
	})

	t.Run("matches zero value", func(t *testing.T) {
		matcher := specta.HasValue[string, int](0)
		result := matcher.Matches(map[string]int{"count": 0})
		if !result.Matched {
			t.Errorf("HasValue should match zero value, but got: %s", result.Message)
		}
	})

	t.Run("matches when multiple keys have same value", func(t *testing.T) {
		matcher := specta.HasValue[string, int](100)
		result := matcher.Matches(map[string]int{"score1": 100, "score2": 100})
		if !result.Matched {
			t.Errorf("HasValue should match when value appears multiple times, but got: %s", result.Message)
		}
	})

	t.Run("rejects when value is not present", func(t *testing.T) {
		matcher := specta.HasValue[string, int](999)
		result := matcher.Matches(map[string]int{"age": 30, "score": 100})
		if result.Matched {
			t.Error("HasValue should not match when value is not present")
		}
		if result.Message == "" {
			t.Error("HasValue should provide error message on mismatch")
		}
	})

	t.Run("rejects empty map", func(t *testing.T) {
		matcher := specta.HasValue[string, int](42)
		result := matcher.Matches(map[string]int{})
		if result.Matched {
			t.Error("HasValue should not match empty map")
		}
	})
}

func TestHasEntry(t *testing.T) {
	t.Run("matches when key-value pair is present", func(t *testing.T) {
		matcher := specta.HasEntry[string, int]("age", 30)
		result := matcher.Matches(map[string]int{"age": 30, "score": 100})
		if !result.Matched {
			t.Errorf("HasEntry should match when key-value pair is present, but got: %s", result.Message)
		}
	})

	t.Run("matches with zero value", func(t *testing.T) {
		matcher := specta.HasEntry[string, int]("count", 0)
		result := matcher.Matches(map[string]int{"count": 0})
		if !result.Matched {
			t.Errorf("HasEntry should match zero value, but got: %s", result.Message)
		}
	})

	t.Run("rejects when key is missing", func(t *testing.T) {
		matcher := specta.HasEntry[string, int]("missing", 30)
		result := matcher.Matches(map[string]int{"age": 30})
		if result.Matched {
			t.Error("HasEntry should not match when key is missing")
		}
		if result.Message == "" {
			t.Error("HasEntry should provide error message when key is missing")
		}
	})

	t.Run("rejects when value is different", func(t *testing.T) {
		matcher := specta.HasEntry[string, int]("age", 25)
		result := matcher.Matches(map[string]int{"age": 30})
		if result.Matched {
			t.Error("HasEntry should not match when value is different")
		}
		if result.Message == "" {
			t.Error("HasEntry should provide error message when value is different")
		}
	})

	t.Run("rejects empty map", func(t *testing.T) {
		matcher := specta.HasEntry[string, int]("key", 42)
		result := matcher.Matches(map[string]int{})
		if result.Matched {
			t.Error("HasEntry should not match empty map")
		}
	})
}

func TestMapHasSize(t *testing.T) {
	t.Run("matches map with expected size", func(t *testing.T) {
		matcher := specta.MapHasSize[string, int](2)
		result := matcher.Matches(map[string]int{"a": 1, "b": 2})
		if !result.Matched {
			t.Errorf("MapHasSize should match map with expected size, but got: %s", result.Message)
		}
	})

	t.Run("matches empty map when size is 0", func(t *testing.T) {
		matcher := specta.MapHasSize[string, int](0)
		result := matcher.Matches(map[string]int{})
		if !result.Matched {
			t.Errorf("MapHasSize(0) should match empty map, but got: %s", result.Message)
		}
	})

	t.Run("matches nil map when size is 0", func(t *testing.T) {
		matcher := specta.MapHasSize[string, int](0)
		var nilMap map[string]int
		result := matcher.Matches(nilMap)
		if !result.Matched {
			t.Errorf("MapHasSize(0) should match nil map, but got: %s", result.Message)
		}
	})

	t.Run("rejects map with different size", func(t *testing.T) {
		matcher := specta.MapHasSize[string, int](3)
		result := matcher.Matches(map[string]int{"a": 1, "b": 2})
		if result.Matched {
			t.Error("MapHasSize should not match map with different size")
		}
		if result.Message == "" {
			t.Error("MapHasSize should provide error message on mismatch")
		}
	})

	t.Run("works with different key/value types", func(t *testing.T) {
		matcher := specta.MapHasSize[int, string](1)
		result := matcher.Matches(map[int]string{42: "answer"})
		if !result.Matched {
			t.Errorf("MapHasSize should work with different types, but got: %s", result.Message)
		}
	})
}

// ==============================================================================
// Integration Tests - Combining Matchers
// ==============================================================================

func TestCollectionMatchersWithCombinators(t *testing.T) {
	t.Run("AllOf with multiple collection matchers", func(t *testing.T) {
		matcher := specta.AllOf(
			specta.HasSize[int](3),
			specta.ContainsElement(2),
			specta.IsNotEmpty[int](),
		)
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("AllOf with collection matchers should match, but got: %s", result.Message)
		}
	})

	t.Run("AnyOf with collection matchers", func(t *testing.T) {
		matcher := specta.AnyOf(
			specta.IsEmpty[int](),
			specta.HasSize[int](3),
		)
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("AnyOf should match when one condition is met, but got: %s", result.Message)
		}
	})

	t.Run("Not with collection matcher", func(t *testing.T) {
		matcher := specta.Not(specta.IsEmpty[int]())
		result := matcher.Matches([]int{1, 2, 3})
		if !result.Matched {
			t.Errorf("Not(IsEmpty) should match non-empty slice, but got: %s", result.Message)
		}
	})

	t.Run("Every with combined matcher", func(t *testing.T) {
		matcher := specta.Every(
			specta.AllOf(
				specta.GreaterThan(0),
				specta.LessThan(10),
			),
		)
		result := matcher.Matches([]int{1, 5, 9})
		if !result.Matched {
			t.Errorf("Every with AllOf should match, but got: %s", result.Message)
		}
	})
}
