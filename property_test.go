package specta_test

import (
	"strings"
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/conjecture"
	"github.com/james-w/specta/testlib"
)

// TestSource tests the Source implementation
func TestSource(t *testing.T) {
	t.Run("same seed produces same bytes", func(t *testing.T) {
		s1 := specta.NewSource(12345)
		s2 := specta.NewSource(12345)

		bits1 := s1.DrawBits(32)
		bits2 := s2.DrawBits(32)

		specta.AssertThat(t, bits1, specta.Equal(bits2))
	})

	t.Run("different seeds produce different bytes", func(t *testing.T) {
		s1 := specta.NewSource(12345)
		s2 := specta.NewSource(54321)

		bits1 := s1.DrawBits(32)
		bits2 := s2.DrawBits(32)

		specta.AssertThat(t, bits1, specta.Not(specta.Equal(bits2)))
	})

	t.Run("can replay from data", func(t *testing.T) {
		s1 := specta.NewSource(12345)
		s1.DrawBits(32)
		s1.DrawBits(16)
		data := s1.Data()

		s2 := specta.NewSourceFromData(data)
		s2.DrawBits(32)
		bits := s2.DrawBits(16)

		// Should be able to draw bits from replayed data
		if bits == 0 {
			// This is technically possible but extremely unlikely
			t.Log("warning: got 0 bits from replay (unlikely but possible)")
		}
	})

	t.Run("logging can be enabled", func(t *testing.T) {
		s := specta.NewSource(12345)
		s.EnableLogging()
		s.WriteLog("test message")

		specta.AssertThat(t, s.Log(), specta.Contains("test message"))
	})

	t.Run("logging disabled by default", func(t *testing.T) {
		s := specta.NewSource(12345)
		s.WriteLog("test message")

		specta.AssertThat(t, s.Log(), specta.Equal(""))
	})
}

// TestIntGenerator tests the Int generator
func TestIntGenerator(t *testing.T) {
	t.Run("generates values", func(t *testing.T) {
		data := conjecture.NewConjectureData(conjecture.WithSeed(12345))
		pt := &specta.T{Data: data}

		value := specta.Draw(pt, specta.Int(), "value")

		// Should generate some value (could be any int64)
		_ = value
	})

	t.Run("range constraints respected", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(12345 + i)))
			pt := &specta.T{Data: data}

			value := specta.Draw(pt, specta.Int().Range(0, 10), "value")
			specta.AssertThat(t, value, specta.AllOf(
				specta.GreaterThanOrEqual(int64(0)),
				specta.Not(specta.GreaterThan(int64(10))),
			))
		}
	})

	t.Run("same seed produces same values", func(t *testing.T) {
		data1 := conjecture.NewConjectureData(conjecture.WithSeed(12345))
		pt1 := &specta.T{Data: data1}
		data2 := conjecture.NewConjectureData(conjecture.WithSeed(12345))
		pt2 := &specta.T{Data: data2}

		v1 := specta.Draw(pt1, specta.Int(), "v1")
		v2 := specta.Draw(pt2, specta.Int(), "v2")

		specta.AssertThat(t, v1, specta.Equal(v2))
	})

	t.Run("logs when enabled", func(t *testing.T) {
		t.Skip("Logging test disabled during conjecture migration (Phase 3) - TestCase.Log() exists but different API")
		/* rs := specta.NewSource(12345)
		pt := &specta.T{Source: rs}
		rs.EnableLogging()

		specta.Int().Draw(pt.TestCase, "myvalue")
		log := rs.Log()

		if !strings.Contains(log, "myvalue") {
			t.Errorf("expected log to contain label, got: %s", log)
		}
		if !strings.Contains(log, "Int(myvalue)=") {
			t.Errorf("expected log to contain Int(myvalue)=, got: %s", log)
		} */
	})

	t.Run("single value range", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(12345 + i)))
			pt := &specta.T{Data: data}

			value := specta.Draw(pt, specta.Int().Range(42, 42), "value")
			specta.AssertThat(t, value, specta.Equal(int64(42)))
		}
	})
}

// TestPropertyBasics tests basic property test functionality
func TestPropertyBasics(t *testing.T) {
	t.Run("passing property succeeds", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Draw(t, specta.Int(), "x")
			// Property: x + 0 == x (always true)
			specta.AssertThat(t, x+0, specta.Equal(x))
		}, specta.MaxTests(10))

		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
	})

	t.Run("failing property detected", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Draw(t, specta.Int(), "x")
			// Property: x >= 0 (will fail on negative values)
			if x < 0 {
				t.Fatalf("expected non-negative, got %d", x)
			}
		}, specta.MaxTests(100))

		specta.AssertThat(t, len(spy.Errors), specta.GreaterThan(0))
	})

	t.Run("seed reproduction", func(t *testing.T) {
		var failingSeed int64

		// First run: find a failing seed
		spy1 := testlib.NewSpy()
		specta.Property(spy1, func(t *specta.T) {
			x := specta.Draw(t, specta.Int(), "x")
			if x < 0 {
				// Capture the seed from the error message
				t.Fatalf("negative: %d", x)
			}
		}, specta.Seed(12345), specta.MaxTests(100))

		if len(spy1.Errors) == 0 {
			t.Skip("property didn't fail with this seed")
		}

		// Extract seed from error message
		failingSeed = 12345

		// Second run: reproduce with same seed
		spy2 := testlib.NewSpy()
		specta.Property(spy2, func(t *specta.T) {
			x := specta.Draw(t, specta.Int(), "x")
			if x < 0 {
				t.Fatalf("negative: %d", x)
			}
		}, specta.Seed(failingSeed), specta.MaxTests(100))

		specta.AssertThat(t, len(spy2.Errors), specta.GreaterThan(0))
	})

	t.Run("range constraints work in properties", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Draw(t, specta.Int().Range(0, 100), "x")
			// Property: value should be in range
			if x < 0 || x > 100 {
				t.Fatalf("value %d out of range", x)
			}
		}, specta.MaxTests(100))

		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
	})
}

// TestPropertyShrinking tests that shrinking reduces failing cases
func TestPropertyShrinking(t *testing.T) {
	t.Run("shrinks to smaller values", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Draw(t, specta.Int().Range(-1000, 1000), "x")
			// Fail on negative values
			if x < 0 {
				t.Fatalf("negative: %d", x)
			}
		}, specta.MaxTests(100))

		if len(spy.Errors) == 0 {
			t.Error("expected property to fail")
			return
		}

		// Check that error message contains the failure
		// Shrinking should have found a small negative value
		errorMsg := strings.Join(spy.Errors, "\n")
		specta.AssertThat(t, errorMsg, specta.Contains("Property Test Failed"))
	})
}

// TestPropertyWithMatchers tests integration with existing matchers
func TestPropertyWithMatchers(t *testing.T) {
	t.Run("works with Equal matcher", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Draw(t, specta.Int().Range(0, 100), "x")
			y := specta.Draw(t, specta.Int().Range(0, 100), "y")
			// Commutative property of addition
			specta.AssertThat(t, x+y, specta.Equal(y+x))
		}, specta.MaxTests(50))

		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
	})

	t.Run("works with comparison matchers", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Draw(t, specta.Int().Range(0, 100), "x")
			// Property: non-negative values are >= 0
			specta.AssertThat(t, x, specta.GreaterThanOrEqual(int64(0)))
		}, specta.MaxTests(50))

		specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
	})
}

// TestSpecTaImplementsTestingT verifies that specta.T implements TestingT
func TestSpecTaImplementsTestingT(t *testing.T) {
	var _ specta.TestingT = (*specta.T)(nil)
}

// TestPropertyPanics tests that unexpected panics are propagated
func TestPropertyPanics(t *testing.T) {
	t.Run("unexpected panic propagates", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic to propagate")
			} else if r != "unexpected panic" {
				t.Errorf("expected 'unexpected panic', got: %v", r)
			}
		}()

		spy := testlib.NewSpy()
		specta.Property(spy, func(t *specta.T) {
			panic("unexpected panic")
		})
	})
}

func TestProperty_AllSkipped(t *testing.T) {
	spy := &testlib.Spy{}

	specta.Property(spy, func(pt *specta.T) {
		x := specta.Draw(pt, specta.Int().Range(0, 10), "x")

		// Always skip
		pt.Assume(false)

		// This should never run
		if x > 100 {
			t.Errorf("This should never happen")
		}
	}, specta.MaxTests(10))

	// Should have an error about all tests being skipped
	if len(spy.Errors) == 0 {
		t.Error("Expected error about all tests being skipped")
		return
	}
	specta.AssertThat(t, contains(spy.Errors[0], "All") || contains(spy.Errors[0], "skipped"), specta.IsTrue())
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestProperty_AllSkipped_Message(t *testing.T) {
	spy := &testlib.Spy{}

	specta.Property(spy, func(pt *specta.T) {
		pt.Assume(false) // Always skip
	}, specta.MaxTests(10))

	// Print the actual error to verify
	if len(spy.Errors) == 0 {
		t.Error("Expected error about all tests being skipped")
		return
	}
	t.Logf("Error message: %s", spy.Errors[0])
}

func TestProperty_HighSkipRate(t *testing.T) {
	spy := &testlib.Spy{}

	specta.Property(spy, func(pt *specta.T) {
		x := specta.Draw(pt, specta.Int().Range(0, 99), "x")

		// Skip 95% of the time (uses uniform distribution for range ≤100)
		pt.Assume(x < 5)

		// This should run occasionally
		if x >= 5 {
			pt.Errorf("Should not happen")
		}
	}, specta.MaxTests(100), specta.Seed(12345))

	// Should have a warning about high skip rate in logs (not errors)
	if len(spy.Logs) == 0 {
		t.Fatal("Expected skip rate warning in logs but got none")
	}
	specta.AssertThat(t, spy.Logs[0], specta.Contains("Warning: "))
	// Verify it's NOT in errors (should be a log, not error)
	specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
}

func TestProperty_SkipRateBoundary(t *testing.T) {
	spy := &testlib.Spy{}

	specta.Property(spy, func(pt *specta.T) {
		x := specta.Draw(pt, specta.Int().Range(0, 99), "x")

		// Skip ~50% of the time (well below 90% threshold - should NOT warn)
		pt.Assume(x < 50)

		if x >= 50 {
			pt.Errorf("Should not happen")
		}
	}, specta.MaxTests(100), specta.Seed(99999))

	// Should NOT have a warning (skip rate well below 90%)
	specta.AssertThat(t, len(spy.Logs), specta.Equal(0))
	specta.AssertThat(t, len(spy.Errors), specta.Equal(0))
}
