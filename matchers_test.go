package gomatchers_test

import (
	"testing"

	"github.com/james-w/gomatchers"
)

func TestEqual(t *testing.T) {
	t.Run("matches equal values", func(t *testing.T) {
		matcher := gomatchers.Equal(42)
		result := matcher.Matches(42)
		if !result.Matched {
			t.Error("Expected match but got mismatch")
		}
	})

	t.Run("rejects unequal values", func(t *testing.T) {
		matcher := gomatchers.Equal(42)
		result := matcher.Matches(99)
		if result.Matched {
			t.Error("Expected mismatch but got match")
		}
		if result.Message == "" {
			t.Error("Expected error message")
		}
	})

	t.Run("works with strings", func(t *testing.T) {
		matcher := gomatchers.Equal("hello")
		if !matcher.Matches("hello").Matched {
			t.Error("Expected match")
		}
		if matcher.Matches("world").Matched {
			t.Error("Expected mismatch")
		}
	})
}

func TestIs(t *testing.T) {
	t.Run("is alias for Equal", func(t *testing.T) {
		matcher := gomatchers.Is(true)
		if !matcher.Matches(true).Matched {
			t.Error("Expected match")
		}
		if matcher.Matches(false).Matched {
			t.Error("Expected mismatch")
		}
	})
}

func TestNot(t *testing.T) {
	t.Run("negates matcher", func(t *testing.T) {
		matcher := gomatchers.Not(gomatchers.Equal(42))
		if !matcher.Matches(99).Matched {
			t.Error("Expected match for non-42")
		}
		if matcher.Matches(42).Matched {
			t.Error("Expected mismatch for 42")
		}
	})
}

func TestIsZero(t *testing.T) {
	t.Run("matches zero int", func(t *testing.T) {
		matcher := gomatchers.IsZero[int]()
		if !matcher.Matches(0).Matched {
			t.Error("Expected match for 0")
		}
		if matcher.Matches(42).Matched {
			t.Error("Expected mismatch for 42")
		}
	})

	t.Run("matches zero string", func(t *testing.T) {
		matcher := gomatchers.IsZero[string]()
		if !matcher.Matches("").Matched {
			t.Error("Expected match for empty string")
		}
		if matcher.Matches("hello").Matched {
			t.Error("Expected mismatch for non-empty string")
		}
	})
}

func TestGreaterThan(t *testing.T) {
	t.Run("matches greater values", func(t *testing.T) {
		matcher := gomatchers.GreaterThan(10)
		if !matcher.Matches(20).Matched {
			t.Error("Expected 20 > 10")
		}
		if matcher.Matches(5).Matched {
			t.Error("Expected 5 not > 10")
		}
		if matcher.Matches(10).Matched {
			t.Error("Expected 10 not > 10")
		}
	})

	t.Run("works with floats", func(t *testing.T) {
		matcher := gomatchers.GreaterThan(3.14)
		if !matcher.Matches(3.15).Matched {
			t.Error("Expected 3.15 > 3.14")
		}
		if matcher.Matches(3.13).Matched {
			t.Error("Expected 3.13 not > 3.14")
		}
	})
}

func TestLessThan(t *testing.T) {
	t.Run("matches lesser values", func(t *testing.T) {
		matcher := gomatchers.LessThan(10)
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
		matcher := gomatchers.GreaterThanOrEqual(10)
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

func TestContains(t *testing.T) {
	t.Run("matches substring", func(t *testing.T) {
		matcher := gomatchers.Contains("world")
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
		matcher := gomatchers.HasPrefix("hello")
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
		matcher := gomatchers.HasSuffix("world")
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
		matcher := gomatchers.IsTrue()
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
		matcher := gomatchers.IsFalse()
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
		matcher := gomatchers.AllOf(
			gomatchers.GreaterThan(10),
			gomatchers.LessThan(20),
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
		matcher := gomatchers.AllOf(
			gomatchers.Equal(42),
			gomatchers.GreaterThan(50),
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
		matcher := gomatchers.AnyOf(
			gomatchers.Equal(10),
			gomatchers.Equal(20),
			gomatchers.Equal(30),
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
		gomatchers.AssertThat(t, 42, gomatchers.Equal(42))
		gomatchers.AssertThat(t, "hello", gomatchers.Contains("ell"))
		gomatchers.AssertThat(t, true, gomatchers.IsTrue())
	})
}

func TestMatcherComposition(t *testing.T) {
	t.Run("complex matcher combinations", func(t *testing.T) {
		// Value must be > 10 AND (< 20 OR > 30)
		matcher := gomatchers.AllOf(
			gomatchers.GreaterThan(10),
			gomatchers.AnyOf(
				gomatchers.LessThan(20),
				gomatchers.GreaterThan(30),
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
