package conjecture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T wraps testing.T with property-based testing capabilities.
type T struct {
	*testing.T
	engine *Engine
}

// NewT creates a property testing wrapper for a standard Go test.
func NewT(t *testing.T, opts ...SettingsOption) *T {
	// Add a file-based database by default in test mode
	defaultOpts := []SettingsOption{
		WithDatabase(NewFileDatabase(".hypothesis")),
	}
	opts = append(defaultOpts, opts...)

	return &T{
		T:      t,
		engine: NewEngine(opts...),
	}
}

// Property runs a property test and fails the test if a counterexample is found.
func (t *T) Property(name string, prop Property) {
	t.Helper()

	ctx := context.Background()
	result := t.engine.Run(ctx, name, prop)

	if !result.Success {
		t.Errorf("Property %q failed:\n%s\n\nFailing example:\n%s\n\nStats: %+v",
			name,
			result.FailureReason,
			formatSequence(result.FailingExample),
			result.Stats,
		)
	}
}

// Check runs a simple property with a generator and assertion.
func Check[V any](t *T, name string, gen Gen[V], check func(V) error) {
	t.Helper()

	prop := func(d *ConjectureData) error {
		v, err := gen.Draw(d)
		if err != nil {
			return nil // Generation issue, not property failure
		}
		return check(v)
	}

	t.Property(name, prop)
}

// Check2 runs a property with two generators.
func Check2[A, B any](t *T, name string, ga Gen[A], gb Gen[B], check func(A, B) error) {
	t.Helper()

	prop := func(d *ConjectureData) error {
		a, err := ga.Draw(d)
		if err != nil {
			return nil
		}
		b, err := gb.Draw(d)
		if err != nil {
			return nil
		}
		return check(a, b)
	}

	t.Property(name, prop)
}

// Check3 runs a property with three generators.
func Check3[A, B, C any](t *T, name string, ga Gen[A], gb Gen[B], gc Gen[C], check func(A, B, C) error) {
	t.Helper()

	prop := func(d *ConjectureData) error {
		a, err := ga.Draw(d)
		if err != nil {
			return nil
		}
		b, err := gb.Draw(d)
		if err != nil {
			return nil
		}
		c, err := gc.Draw(d)
		if err != nil {
			return nil
		}
		return check(a, b, c)
	}

	t.Property(name, prop)
}

// Given creates a property builder for fluent test definition.
// Usage:
//
//	t.Given(Integer(0, 100)).
//	    Check(func(n int64) error {
//	        if n*n < 0 { return errors.New("overflow") }
//	        return nil
//	    })
func (t *T) Given(g Gen[any]) *TestBuilder {
	return &TestBuilder{
		t:   t,
		gen: g,
	}
}

// TestBuilder allows fluent property test construction.
type TestBuilder struct {
	t     *T
	gen   Gen[any]
	name  string
	setup func(any) error
}

// Named sets the test name.
func (tb *TestBuilder) Named(name string) *TestBuilder {
	tb.name = name
	return tb
}

// Where adds a precondition.
func (tb *TestBuilder) Where(pred func(any) bool) *TestBuilder {
	tb.setup = func(v any) error {
		if !pred(v) {
			return fmt.Errorf("precondition failed")
		}
		return nil
	}
	return tb
}

// Then runs the property check.
func (tb *TestBuilder) Then(check func(any) error) {
	tb.t.Helper()

	name := tb.name
	if name == "" {
		name = tb.gen.String()
	}

	prop := func(d *ConjectureData) error {
		v, err := tb.gen.Draw(d)
		if err != nil {
			return nil
		}
		if tb.setup != nil {
			if err := tb.setup(v); err != nil {
				return d.Assume(false)
			}
		}
		return check(v)
	}

	tb.t.Property(name, prop)
}

// formatSequence creates a human-readable representation of a choice sequence.
func formatSequence(seq *ChoiceSequence) string {
	if seq == nil {
		return "<nil>"
	}

	var sb strings.Builder
	for i := 0; i < seq.Len(); i++ {
		c := seq.Get(i)
		fmt.Fprintf(&sb, "[%d] %s: %v\n", i, c.Type, c.Value)
	}
	return sb.String()
}

// =============================================================================
// File-based Database
// =============================================================================

// FileDatabase persists examples to disk.
type FileDatabase struct {
	dir string
}

// NewFileDatabase creates a file-based database in the given directory.
func NewFileDatabase(dir string) *FileDatabase {
	return &FileDatabase{dir: dir}
}

func (db *FileDatabase) path(key string) string {
	return filepath.Join(db.dir, key)
}

func (db *FileDatabase) Save(key string, seq *ChoiceSequence) {
	dir := db.path(key)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	// Serialize the sequence
	data := serializeSequence(seq)
	hash := hashSequence(data)

	filename := filepath.Join(dir, hash)
	_ = os.WriteFile(filename, data, 0644) // Best effort - ignore errors
}

func (db *FileDatabase) Load(key string) []*ChoiceSequence {
	dir := db.path(key)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var result []*ChoiceSequence
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		seq, err := deserializeSequence(data)
		if err != nil {
			continue
		}
		result = append(result, seq)
	}
	return result
}

func (db *FileDatabase) Delete(key string) {
	_ = os.RemoveAll(db.path(key)) // Best effort - ignore errors
}

// Simple serialization for choice sequences.
// In production, you'd want a more robust format.
func serializeSequence(seq *ChoiceSequence) []byte {
	var buf []byte

	// Write number of choices
	n := uint32(seq.Len())
	buf = append(buf, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))

	for i := 0; i < seq.Len(); i++ {
		c := seq.Get(i)
		buf = append(buf, byte(c.Type))

		switch c.Type {
		case ChoiceBoolean:
			if c.Value.(bool) {
				buf = append(buf, 1)
			} else {
				buf = append(buf, 0)
			}
		case ChoiceInteger:
			v := c.Value.(int64)
			buf = appendInt64(buf, v)
		case ChoiceFloat:
			v := c.Value.(float64)
			buf = appendFloat64(buf, v)
		case ChoiceString:
			v := c.Value.(string)
			buf = appendString(buf, v)
		case ChoiceBytes:
			v := c.Value.([]byte)
			buf = appendBytes(buf, v)
		}
	}

	return buf
}

func deserializeSequence(data []byte) (*ChoiceSequence, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}

	n := int(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]))
	data = data[4:]

	seq := NewChoiceSequence()

	for i := 0; i < n; i++ {
		if len(data) < 1 {
			return nil, fmt.Errorf("unexpected end of data")
		}

		choiceType := ChoiceType(data[0])
		data = data[1:]

		var c Choice
		c.Type = choiceType

		switch choiceType {
		case ChoiceBoolean:
			if len(data) < 1 {
				return nil, fmt.Errorf("unexpected end of data")
			}
			c.Value = data[0] != 0
			data = data[1:]

		case ChoiceInteger:
			if len(data) < 8 {
				return nil, fmt.Errorf("unexpected end of data")
			}
			c.Value = readInt64(data)
			data = data[8:]

		case ChoiceFloat:
			if len(data) < 8 {
				return nil, fmt.Errorf("unexpected end of data")
			}
			c.Value = readFloat64(data)
			data = data[8:]

		case ChoiceString:
			s, rest, err := readString(data)
			if err != nil {
				return nil, err
			}
			c.Value = s
			data = rest

		case ChoiceBytes:
			b, rest, err := readBytes(data)
			if err != nil {
				return nil, err
			}
			c.Value = b
			data = rest
		}

		seq.Append(c)
	}

	return seq, nil
}

func appendInt64(buf []byte, v int64) []byte {
	return append(buf,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func readInt64(data []byte) int64 {
	return int64(data[0])<<56 | int64(data[1])<<48 | int64(data[2])<<40 | int64(data[3])<<32 |
		int64(data[4])<<24 | int64(data[5])<<16 | int64(data[6])<<8 | int64(data[7])
}

func appendFloat64(buf []byte, v float64) []byte {
	bits := math.Float64bits(v)
	return append(buf,
		byte(bits>>56), byte(bits>>48), byte(bits>>40), byte(bits>>32),
		byte(bits>>24), byte(bits>>16), byte(bits>>8), byte(bits))
}

func readFloat64(data []byte) float64 {
	bits := uint64(data[0])<<56 | uint64(data[1])<<48 | uint64(data[2])<<40 | uint64(data[3])<<32 |
		uint64(data[4])<<24 | uint64(data[5])<<16 | uint64(data[6])<<8 | uint64(data[7])
	return math.Float64frombits(bits)
}

func appendString(buf []byte, s string) []byte {
	n := uint32(len(s))
	buf = append(buf, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	return append(buf, s...)
}

func readString(data []byte) (string, []byte, error) {
	if len(data) < 4 {
		return "", nil, fmt.Errorf("unexpected end of data")
	}
	n := int(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]))
	data = data[4:]
	if len(data) < n {
		return "", nil, fmt.Errorf("unexpected end of data")
	}
	return string(data[:n]), data[n:], nil
}

func appendBytes(buf []byte, b []byte) []byte {
	n := uint32(len(b))
	buf = append(buf, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	return append(buf, b...)
}

func readBytes(data []byte) ([]byte, []byte, error) {
	if len(data) < 4 {
		return nil, nil, fmt.Errorf("unexpected end of data")
	}
	n := int(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]))
	data = data[4:]
	if len(data) < n {
		return nil, nil, fmt.Errorf("unexpected end of data")
	}
	result := make([]byte, n)
	copy(result, data[:n])
	return result, data[n:], nil
}

func hashSequence(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:8])
}

// =============================================================================
// Quick Test Functions (for simple cases)
// =============================================================================

// Quick runs a property test with default settings.
// This is the simplest way to use conjecture in tests.
func Quick(t *testing.T, prop Property) {
	t.Helper()
	ct := NewT(t)
	ct.Property(t.Name(), prop)
}

// QuickCheck is a convenience for checking a single-value property.
func QuickCheck[V any](t *testing.T, gen Gen[V], check func(V) error) {
	t.Helper()
	ct := NewT(t)
	Check(ct, t.Name(), gen, check)
}
