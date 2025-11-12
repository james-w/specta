package specta_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/james-w/specta"
)

// testingSpy implements specta.TestingT and captures error messages for testing
type testingSpy struct {
	errors []string
}

func newTestingSpy() *testingSpy {
	return &testingSpy{errors: make([]string, 0)}
}

func (ts *testingSpy) Helper() {
	// No-op for spy
}

func (ts *testingSpy) Errorf(format string, args ...interface{}) {
	// Capture the formatted error message without failing the outer test
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}
	ts.errors = append(ts.errors, strings.TrimSpace(msg))
}

// TestAssertThat_ExpressionCapture verifies that AssertThat captures the correct
// argument (args[1] which is the actual value, not args[0] which is t)
func TestAssertThat_ExpressionCapture(t *testing.T) {
	t.Run("captures simple variable name", func(t *testing.T) {
		spy := newTestingSpy()
		value := 42

		// This will fail and should show "value" in the error message
		specta.AssertThat(spy, value, specta.Equal(99))

		if len(spy.errors) == 0 {
			t.Fatal("Expected AssertThat to call Errorf, but it didn't")
		}

		errorMsg := spy.errors[0]
		t.Logf("Captured error message: %s", errorMsg)

		// The error should contain "value", not "spy" or "t"
		if !strings.Contains(errorMsg, "value") {
			t.Errorf("Expected error message to contain 'value', got: %s", errorMsg)
		}

		// Should NOT contain references to the testing parameter
		if strings.Contains(errorMsg, "spy") || strings.HasPrefix(errorMsg, "t:") {
			t.Errorf("Error message incorrectly shows testing parameter instead of actual value: %s", errorMsg)
		}
	})

	t.Run("captures field access", func(t *testing.T) {
		spy := newTestingSpy()
		type User struct {
			Name string
		}
		user := User{Name: "Bob"}

		specta.AssertThat(spy, user.Name, specta.Equal("Alice"))

		if len(spy.errors) == 0 {
			t.Fatal("Expected AssertThat to call Errorf")
		}

		errorMsg := spy.errors[0]
		t.Logf("Captured error message: %s", errorMsg)

		// Should contain the field expression
		if !strings.Contains(errorMsg, "Name") && !strings.Contains(errorMsg, "user") {
			t.Errorf("Expected error message to reference the field, got: %s", errorMsg)
		}
	})

	t.Run("captures function call expression", func(t *testing.T) {
		spy := newTestingSpy()
		s := "hello"

		specta.AssertThat(spy, len(s), specta.GreaterThan(10))

		if len(spy.errors) == 0 {
			t.Fatal("Expected AssertThat to call Errorf")
		}

		errorMsg := spy.errors[0]
		t.Logf("Captured error message: %s", errorMsg)

		// Should contain reference to the len function call
		if !strings.Contains(errorMsg, "len") {
			t.Errorf("Expected error message to reference 'len', got: %s", errorMsg)
		}
	})

	t.Run("works correctly on success", func(t *testing.T) {
		spy := newTestingSpy()
		value := 42

		// This should succeed and not call Errorf
		specta.AssertThat(spy, value, specta.Equal(42))

		if len(spy.errors) > 0 {
			t.Errorf("Expected no errors for successful assertion, got: %v", spy.errors)
		}
	})
}
