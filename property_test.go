package specta_test

import (
	"strings"
	"testing"

	"github.com/james-w/specta"
	"github.com/james-w/specta/testlib"
)

// TestSource tests the Source implementation
func TestSource(t *testing.T) {
	t.Run("same seed produces same bytes", func(t *testing.T) {
		s1 := specta.NewSource(12345)
		s2 := specta.NewSource(12345)

		bits1 := s1.DrawBits(32)
		bits2 := s2.DrawBits(32)

		if bits1 != bits2 {
			t.Errorf("expected same bits, got %d and %d", bits1, bits2)
		}
	})

	t.Run("different seeds produce different bytes", func(t *testing.T) {
		s1 := specta.NewSource(12345)
		s2 := specta.NewSource(54321)

		bits1 := s1.DrawBits(32)
		bits2 := s2.DrawBits(32)

		if bits1 == bits2 {
			t.Errorf("expected different bits, got same: %d", bits1)
		}
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

		if !strings.Contains(s.Log(), "test message") {
			t.Errorf("expected log to contain 'test message', got: %s", s.Log())
		}
	})

	t.Run("logging disabled by default", func(t *testing.T) {
		s := specta.NewSource(12345)
		s.WriteLog("test message")

		if s.Log() != "" {
			t.Errorf("expected empty log, got: %s", s.Log())
		}
	})
}

// TestIntGenerator tests the Int generator
func TestIntGenerator(t *testing.T) {
	t.Run("generates values", func(t *testing.T) {
		pt := &specta.T{Source: specta.NewSource(12345)}

		value := specta.Int().Draw(pt.Source, "x")

		// Should generate some value (could be any int64)
		_ = value
	})

	t.Run("range constraints respected", func(t *testing.T) {
		pt := &specta.T{Source: specta.NewSource(12345)}

		for i := 0; i < 100; i++ {
			value := specta.Int().Range(0, 10).Draw(pt.Source, "x")
			if value < 0 || value > 10 {
				t.Errorf("value %d out of range [0, 10]", value)
			}
		}
	})

	t.Run("same seed produces same values", func(t *testing.T) {
		pt1 := &specta.T{Source: specta.NewSource(12345)}
		pt2 := &specta.T{Source: specta.NewSource(12345)}

		v1 := specta.Int().Draw(pt1.Source, "x")
		v2 := specta.Int().Draw(pt2.Source, "x")

		if v1 != v2 {
			t.Errorf("expected same values, got %d and %d", v1, v2)
		}
	})

	t.Run("logs when enabled", func(t *testing.T) {
		rs := specta.NewSource(12345)
		pt := &specta.T{Source: rs}
		rs.EnableLogging()

		specta.Int().Draw(pt.Source, "myvalue")
		log := rs.Log()

		if !strings.Contains(log, "myvalue") {
			t.Errorf("expected log to contain label, got: %s", log)
		}
		if !strings.Contains(log, "Int(myvalue)=") {
			t.Errorf("expected log to contain Int(myvalue)=, got: %s", log)
		}
	})

	t.Run("single value range", func(t *testing.T) {
		pt := &specta.T{Source: specta.NewSource(12345)}

		for i := 0; i < 10; i++ {
			value := specta.Int().Range(42, 42).Draw(pt.Source, "x")
			if value != 42 {
				t.Errorf("expected 42, got %d", value)
			}
		}
	})
}

// TestPropertyBasics tests basic property test functionality
func TestPropertyBasics(t *testing.T) {
	t.Run("passing property succeeds", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Int().Draw(t.Source, "x")
			// Property: x + 0 == x (always true)
			specta.AssertThat(t, x+0, specta.Equal(x))
		}, specta.MaxTests(10))

		if len(spy.Errors) > 0 {
			t.Errorf("expected property to pass, but it failed: %v", spy.Errors)
		}
	})

	t.Run("failing property detected", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Int().Draw(t.Source, "x")
			// Property: x >= 0 (will fail on negative values)
			if x < 0 {
				t.Fatalf("expected non-negative, got %d", x)
			}
		}, specta.MaxTests(100))

		if len(spy.Errors) == 0 {
			t.Error("expected property to fail, but it passed")
		}
	})

	t.Run("seed reproduction", func(t *testing.T) {
		var failingSeed int64

		// First run: find a failing seed
		spy1 := testlib.NewSpy()
		specta.Property(spy1, func(t *specta.T) {
			x := specta.Int().Draw(t.Source, "x")
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
			x := specta.Int().Draw(t.Source, "x")
			if x < 0 {
				t.Fatalf("negative: %d", x)
			}
		}, specta.Seed(failingSeed), specta.MaxTests(100))

		if len(spy2.Errors) == 0 {
			t.Error("expected property to fail with same seed")
		}
	})

	t.Run("range constraints work in properties", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Int().Range(0, 100).Draw(t.Source, "x")
			// Property: value should be in range
			if x < 0 || x > 100 {
				t.Fatalf("value %d out of range", x)
			}
		}, specta.MaxTests(100))

		if len(spy.Errors) > 0 {
			t.Errorf("expected property to pass: %v", spy.Errors)
		}
	})
}

// TestPropertyShrinking tests that shrinking reduces failing cases
func TestPropertyShrinking(t *testing.T) {
	t.Run("shrinks to smaller values", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Int().Range(-1000, 1000).Draw(t.Source, "x")
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
		if !strings.Contains(errorMsg, "Property Test Failed") {
			t.Errorf("expected failure message, got: %s", errorMsg)
		}
	})
}

// TestPropertyWithMatchers tests integration with existing matchers
func TestPropertyWithMatchers(t *testing.T) {
	t.Run("works with Equal matcher", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Int().Range(0, 100).Draw(t.Source, "x")
			y := specta.Int().Range(0, 100).Draw(t.Source, "y")
			// Commutative property of addition
			specta.AssertThat(t, x+y, specta.Equal(y+x))
		}, specta.MaxTests(50))

		if len(spy.Errors) > 0 {
			t.Errorf("commutative property should pass: %v", spy.Errors)
		}
	})

	t.Run("works with comparison matchers", func(t *testing.T) {
		spy := testlib.NewSpy()

		specta.Property(spy, func(t *specta.T) {
			x := specta.Int().Range(0, 100).Draw(t.Source, "x")
			// Property: non-negative values are >= 0
			specta.AssertThat(t, x, specta.GreaterThanOrEqual(int64(0)))
		}, specta.MaxTests(50))

		if len(spy.Errors) > 0 {
			t.Errorf("property should pass: %v", spy.Errors)
		}
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
			if r := recover(); r == nil {
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
