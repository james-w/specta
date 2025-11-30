// Package testlib provides testing utilities for specta.
package testlib

import (
	"fmt"
	"strings"
)

// Spy is a test helper that implements specta.TestingT interface
// to capture error messages and log messages without failing the outer test.
// Useful for testing matchers and assertions.
type Spy struct {
	Errors []string
	Logs   []string
}

// NewSpy creates a new testing spy.
func NewSpy() *Spy {
	return &Spy{
		Errors: make([]string, 0),
		Logs:   make([]string, 0),
	}
}

// Helper marks the calling function as a test helper function.
func (s *Spy) Helper() {
	// No-op for spy
}

// Errorf captures the formatted error message without failing the test.
func (s *Spy) Errorf(format string, args ...interface{}) {
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}
	s.Errors = append(s.Errors, strings.TrimSpace(msg))
}

// Logf captures the formatted log message.
func (s *Spy) Logf(format string, args ...interface{}) {
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}
	s.Logs = append(s.Logs, strings.TrimSpace(msg))
}
