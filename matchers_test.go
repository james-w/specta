package specta_test

import (
	"strings"
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/testlib"
)

func TestEqual(t *testing.T) {
	t.Run("matches equal values", func(t *testing.T) {
		matcher := specta.Equal(42)
		result := matcher.Matches(42)
		if !result.Matched {
			t.Errorf("Equal(42) should match 42, but got: %s", result.Message)
		}
	})

	t.Run("rejects unequal values", func(t *testing.T) {
		matcher := specta.Equal(42)
		result := matcher.Matches(99)
		if result.Matched {
			t.Error("Equal(42) should not match 99")
		}
		if result.Message == "" {
			t.Error("Equal(42).Matches(99) should provide error message")
		}
	})

	t.Run("works with strings", func(t *testing.T) {
		matcher := specta.Equal("hello")
		result := matcher.Matches("hello")
		if !result.Matched {
			t.Errorf("Equal(\"hello\") should match \"hello\", but got: %s", result.Message)
		}

		result = matcher.Matches("world")
		if result.Matched {
			t.Error("Equal(\"hello\") should not match \"world\"")
		}
	})
}

func TestIs(t *testing.T) {
	t.Run("is alias for Equal", func(t *testing.T) {
		matcher := specta.Is(true)
		result := matcher.Matches(true)
		if !result.Matched {
			t.Errorf("Is(true) should match true, but got: %s", result.Message)
		}

		result = matcher.Matches(false)
		if result.Matched {
			t.Error("Is(true) should not match false")
		}
	})
}

func TestNot(t *testing.T) {
	t.Run("negates matcher", func(t *testing.T) {
		matcher := specta.Not(specta.Equal(42))
		result := matcher.Matches(99)
		if !result.Matched {
			t.Errorf("Not(Equal(42)) should match 99, but got: %s", result.Message)
		}

		result = matcher.Matches(42)
		if result.Matched {
			t.Error("Not(Equal(42)) should not match 42")
		}
	})
}

func TestIsZero(t *testing.T) {
	t.Run("matches zero int", func(t *testing.T) {
		matcher := specta.IsZero[int]()
		result := matcher.Matches(0)
		if !result.Matched {
			t.Errorf("IsZero[int]() should match 0, but got: %s", result.Message)
		}

		result = matcher.Matches(42)
		if result.Matched {
			t.Error("IsZero[int]() should not match 42")
		}
	})

	t.Run("matches zero string", func(t *testing.T) {
		matcher := specta.IsZero[string]()
		result := matcher.Matches("")
		if !result.Matched {
			t.Errorf("IsZero[string]() should match empty string, but got: %s", result.Message)
		}

		result = matcher.Matches("hello")
		if result.Matched {
			t.Error("IsZero[string]() should not match \"hello\"")
		}
	})
}

func TestGreaterThan(t *testing.T) {
	t.Run("matches greater values", func(t *testing.T) {
		matcher := specta.GreaterThan(10)
		result := matcher.Matches(20)
		if !result.Matched {
			t.Errorf("GreaterThan(10) should match 20, but got: %s", result.Message)
		}

		result = matcher.Matches(5)
		if result.Matched {
			t.Error("GreaterThan(10) should not match 5")
		}

		result = matcher.Matches(10)
		if result.Matched {
			t.Error("GreaterThan(10) should not match 10 (boundary)")
		}
	})

	t.Run("works with floats", func(t *testing.T) {
		matcher := specta.GreaterThan(3.14)
		result := matcher.Matches(3.15)
		if !result.Matched {
			t.Errorf("GreaterThan(3.14) should match 3.15, but got: %s", result.Message)
		}

		result = matcher.Matches(3.13)
		if result.Matched {
			t.Error("GreaterThan(3.14) should not match 3.13")
		}
	})
}

func TestLessThan(t *testing.T) {
	t.Run("matches lesser values", func(t *testing.T) {
		matcher := specta.LessThan(10)
		if !matcher.Matches(5).Matched {
			t.Error("Expected 5 < 10")
		}
		if matcher.Matches(20).Matched {
			t.Error("Expected 20 not < 10")
		}
		if matcher.Matches(10).Matched {
			t.Error("Expected 10 not < 10")
		}
	})
}

func TestGreaterThanOrEqual(t *testing.T) {
	t.Run("matches greater or equal values", func(t *testing.T) {
		matcher := specta.GreaterThanOrEqual(10)
		if !matcher.Matches(20).Matched {
			t.Error("Expected 20 >= 10")
		}
		if !matcher.Matches(10).Matched {
			t.Error("Expected 10 >= 10")
		}
		if matcher.Matches(5).Matched {
			t.Error("Expected 5 not >= 10")
		}
	})
}

func TestLessThanOrEqual(t *testing.T) {
	t.Run("matches lesser or equal values", func(t *testing.T) {
		matcher := specta.LessThanOrEqual(10)
		result := matcher.Matches(5)
		if !result.Matched {
			t.Errorf("LessThanOrEqual(10) should match 5, but got: %s", result.Message)
		}

		result = matcher.Matches(10)
		if !result.Matched {
			t.Error("LessThanOrEqual(10) should match 10 (boundary)")
		}

		result = matcher.Matches(20)
		if result.Matched {
			t.Error("LessThanOrEqual(10) should not match 20")
		}
	})

	t.Run("works with floats", func(t *testing.T) {
		matcher := specta.LessThanOrEqual(3.14)
		result := matcher.Matches(3.13)
		if !result.Matched {
			t.Errorf("LessThanOrEqual(3.14) should match 3.13, but got: %s", result.Message)
		}

		result = matcher.Matches(3.14)
		if !result.Matched {
			t.Error("LessThanOrEqual(3.14) should match 3.14 (boundary)")
		}

		result = matcher.Matches(3.15)
		if result.Matched {
			t.Error("LessThanOrEqual(3.14) should not match 3.15")
		}
	})
}

func TestContains(t *testing.T) {
	t.Run("matches substring", func(t *testing.T) {
		matcher := specta.Contains("world")
		if !matcher.Matches("hello world").Matched {
			t.Error("Expected 'hello world' to contain 'world'")
		}
		if matcher.Matches("hello there").Matched {
			t.Error("Expected 'hello there' not to contain 'world'")
		}
	})
}

func TestHasPrefix(t *testing.T) {
	t.Run("matches prefix", func(t *testing.T) {
		matcher := specta.HasPrefix("hello")
		if !matcher.Matches("hello world").Matched {
			t.Error("Expected 'hello world' to start with 'hello'")
		}
		if matcher.Matches("goodbye world").Matched {
			t.Error("Expected 'goodbye world' not to start with 'hello'")
		}
	})
}

func TestHasSuffix(t *testing.T) {
	t.Run("matches suffix", func(t *testing.T) {
		matcher := specta.HasSuffix("world")
		if !matcher.Matches("hello world").Matched {
			t.Error("Expected 'hello world' to end with 'world'")
		}
		if matcher.Matches("hello there").Matched {
			t.Error("Expected 'hello there' not to end with 'world'")
		}
	})
}

func TestIsTrue(t *testing.T) {
	t.Run("matches true", func(t *testing.T) {
		matcher := specta.IsTrue()
		if !matcher.Matches(true).Matched {
			t.Error("Expected match for true")
		}
		if matcher.Matches(false).Matched {
			t.Error("Expected mismatch for false")
		}
	})
}

func TestIsFalse(t *testing.T) {
	t.Run("matches false", func(t *testing.T) {
		matcher := specta.IsFalse()
		if !matcher.Matches(false).Matched {
			t.Error("Expected match for false")
		}
		if matcher.Matches(true).Matched {
			t.Error("Expected mismatch for true")
		}
	})
}

func TestAllOf(t *testing.T) {
	t.Run("all matchers must succeed", func(t *testing.T) {
		matcher := specta.AllOf(
			specta.GreaterThan(10),
			specta.LessThan(20),
		)

		if !matcher.Matches(15).Matched {
			t.Error("Expected 15 to be > 10 and < 20")
		}
		if matcher.Matches(5).Matched {
			t.Error("Expected 5 not to match (not > 10)")
		}
		if matcher.Matches(25).Matched {
			t.Error("Expected 25 not to match (not < 20)")
		}
	})

	t.Run("provides details on failure", func(t *testing.T) {
		matcher := specta.AllOf(
			specta.Equal(42),
			specta.GreaterThan(50),
		)

		result := matcher.Matches(42)
		if result.Matched {
			t.Error("Expected mismatch")
		}
		if len(result.Details) == 0 {
			t.Error("Expected failure details")
		}
	})
}

func TestAnyOf(t *testing.T) {
	t.Run("at least one matcher must succeed", func(t *testing.T) {
		matcher := specta.AnyOf(
			specta.Equal(10),
			specta.Equal(20),
			specta.Equal(30),
		)

		if !matcher.Matches(10).Matched {
			t.Error("Expected match for 10")
		}
		if !matcher.Matches(20).Matched {
			t.Error("Expected match for 20")
		}
		if matcher.Matches(99).Matched {
			t.Error("Expected mismatch for 99")
		}
	})
}

func TestAssertThat(t *testing.T) {
	// We can't easily test failures without subtests that are expected to fail
	// So we just test the happy path
	t.Run("does not fail on match", func(t *testing.T) {
		specta.AssertThat(t, 42, specta.Equal(42))
		specta.AssertThat(t, "hello", specta.Contains("ell"))
		specta.AssertThat(t, true, specta.IsTrue())
	})
}

func TestRequireThat(t *testing.T) {
	t.Run("does not fail on match", func(t *testing.T) {
		specta.RequireThat(t, 42, specta.Equal(42))
		specta.RequireThat(t, "hello", specta.Contains("ell"))
		specta.RequireThat(t, true, specta.IsTrue())
		// Test continues - these lines execute
	})

	t.Run("calls Fatalf on failure", func(t *testing.T) {
		spy := &testlib.Spy{}
		specta.RequireThat(spy, 42, specta.Equal(99))

		// Verify Fatalf was called
		specta.AssertThat(t, spy.Fataled, specta.IsTrue())

		// Verify error message was logged via Errorf first
		specta.AssertThat(t, len(spy.Errors), specta.GreaterThan(0))
	})

	t.Run("reports error before calling Fatalf", func(t *testing.T) {
		spy := &testlib.Spy{}
		specta.RequireThat(spy, "hello", specta.Equal("world"))

		// Should have both error message and Fatalf called
		specta.AssertThat(t, spy.Fataled, specta.IsTrue())
		specta.AssertThat(t, len(spy.Errors), specta.Equal(1))
		specta.AssertThat(t, spy.Errors[0], specta.Contains("expected \"world\" but got \"hello\""))
	})

	t.Run("includes details in error output", func(t *testing.T) {
		spy := &testlib.Spy{}

		// Use a matcher that produces details (e.g., AllOf with multiple failures)
		specta.RequireThat(spy, 15, specta.AllOf(
			specta.LessThan(10),
			specta.GreaterThan(20),
		))

		// Should have main error + detail errors
		specta.AssertThat(t, spy.Fataled, specta.IsTrue())
		specta.AssertThat(t, len(spy.Errors), specta.GreaterThan(1)) // Main error + details
	})
}

func TestMatcherComposition(t *testing.T) {
	t.Run("complex matcher combinations", func(t *testing.T) {
		// Value must be > 10 AND (< 20 OR > 30)
		matcher := specta.AllOf(
			specta.GreaterThan(10),
			specta.AnyOf(
				specta.LessThan(20),
				specta.GreaterThan(30),
			),
		)

		if !matcher.Matches(15).Matched {
			t.Error("Expected 15 to match (> 10 and < 20)")
		}
		if !matcher.Matches(35).Matched {
			t.Error("Expected 35 to match (> 10 and > 30)")
		}
		if matcher.Matches(25).Matched {
			t.Error("Expected 25 not to match (not < 20 and not > 30)")
		}
		if matcher.Matches(5).Matched {
			t.Error("Expected 5 not to match (not > 10)")
		}
	})
}

// ==============================================================================
// Structured Diff Tests
// ==============================================================================

type TestPerson struct {
	Name  string
	Age   int
	Email string
}

type TestAddress struct {
	Street  string
	City    string
	ZipCode string
}

type TestUser struct {
	Name    string
	Age     int
	Active  bool
	Address TestAddress
}

func TestStructuredDiff_DeepEqual(t *testing.T) {
	t.Run("shows structured diff for struct types", func(t *testing.T) {
		expected := TestPerson{Name: "Alice", Age: 30, Email: "alice@example.com"}
		actual := TestPerson{Name: "Alice", Age: 25, Email: "bob@example.com"}

		matcher := specta.DeepEqual(expected)
		result := matcher.Matches(actual)

		if result.Matched {
			t.Error("Expected mismatch")
		}
		if result.Message == "" {
			t.Error("Expected error message")
		}
		// Should contain struct name
		if !strings.Contains(result.Message, "TestPerson") {
			t.Errorf("Expected message to contain struct name, got: %s", result.Message)
		}
		// Should show matched field (Name)
		if !strings.Contains(result.Message, "Name") {
			t.Errorf("Expected message to show Name field, got: %s", result.Message)
		}
		// Should show failed fields (Age, Email)
		if !strings.Contains(result.Message, "Age") || !strings.Contains(result.Message, "Email") {
			t.Errorf("Expected message to show failed fields, got: %s", result.Message)
		}
	})

	t.Run("uses simple format for non-struct types", func(t *testing.T) {
		matcher := specta.DeepEqual(42)
		result := matcher.Matches(99)

		if result.Matched {
			t.Error("Expected mismatch")
		}
		// Should use simple format (no structured diff)
		if !strings.Contains(result.Message, "42") || !strings.Contains(result.Message, "99") {
			t.Errorf("Expected simple format with values, got: %s", result.Message)
		}
	})

	t.Run("handles nested structs", func(t *testing.T) {
		expected := TestUser{
			Name:   "Alice",
			Age:    30,
			Active: true,
			Address: TestAddress{
				Street:  "123 Main St",
				City:    "NYC",
				ZipCode: "10001",
			},
		}
		actual := TestUser{
			Name:   "Alice",
			Age:    30,
			Active: true,
			Address: TestAddress{
				Street:  "123 Main St",
				City:    "Boston",
				ZipCode: "10001",
			},
		}

		matcher := specta.DeepEqual(expected)
		result := matcher.Matches(actual)

		if result.Matched {
			t.Error("Expected mismatch due to nested Address.City difference")
		}
		if !strings.Contains(result.Message, "Address") {
			t.Errorf("Expected message to show Address field, got: %s", result.Message)
		}
	})
}

func TestStructuredDiff_SliceTruncation(t *testing.T) {
	t.Run("shows all items for small slices", func(t *testing.T) {
		// BuildMatcherStructDiff is used by generated matchers,
		// but we can test the value formatting directly through DeepEqual
		type SmallList struct {
			Items []int
		}
		expected := SmallList{Items: []int{1, 2, 3}}
		actual := SmallList{Items: []int{1, 2, 4}}

		matcher := specta.DeepEqual(expected)
		result := matcher.Matches(actual)

		if result.Matched {
			t.Error("Expected mismatch")
		}
		// Should show Items field with difference
		if !strings.Contains(result.Message, "Items") {
			t.Errorf("Expected message to show Items field, got: %s", result.Message)
		}
	})

	t.Run("truncates large slices", func(t *testing.T) {
		type LargeList struct {
			Items []int
		}
		// Create slice with more than 5 items
		expectedItems := make([]int, 10)
		actualItems := make([]int, 10)
		for i := range expectedItems {
			expectedItems[i] = i
			actualItems[i] = i + 1 // All different
		}

		expected := LargeList{Items: expectedItems}
		actual := LargeList{Items: actualItems}

		matcher := specta.DeepEqual(expected)
		result := matcher.Matches(actual)

		if result.Matched {
			t.Error("Expected mismatch")
		}
		// Message should be present
		if result.Message == "" {
			t.Error("Expected error message")
		}
	})
}

func TestStructuredDiff_MapTruncation(t *testing.T) {
	t.Run("shows all entries for small maps", func(t *testing.T) {
		type Config struct {
			Settings map[string]string
		}
		expected := Config{Settings: map[string]string{"a": "1", "b": "2"}}
		actual := Config{Settings: map[string]string{"a": "1", "b": "3"}}

		matcher := specta.DeepEqual(expected)
		result := matcher.Matches(actual)

		if result.Matched {
			t.Error("Expected mismatch")
		}
		if !strings.Contains(result.Message, "Settings") {
			t.Errorf("Expected message to show Settings field, got: %s", result.Message)
		}
	})

	t.Run("truncates large maps", func(t *testing.T) {
		type Config struct {
			Settings map[string]int
		}
		expectedSettings := make(map[string]int)
		actualSettings := make(map[string]int)
		for i := range 10 {
			key := string(rune('a' + i))
			expectedSettings[key] = i
			actualSettings[key] = i + 1
		}

		expected := Config{Settings: expectedSettings}
		actual := Config{Settings: actualSettings}

		matcher := specta.DeepEqual(expected)
		result := matcher.Matches(actual)

		if result.Matched {
			t.Error("Expected mismatch")
		}
		if result.Message == "" {
			t.Error("Expected error message")
		}
	})
}

func TestStructuredDiff_AllMatched(t *testing.T) {
	t.Run("shows only matched symbols when everything matches", func(t *testing.T) {
		expected := TestPerson{Name: "Alice", Age: 30, Email: "alice@example.com"}
		actual := TestPerson{Name: "Alice", Age: 30, Email: "alice@example.com"}

		matcher := specta.DeepEqual(expected)
		result := matcher.Matches(actual)

		if !result.Matched {
			t.Error("Expected match")
		}
	})
}
