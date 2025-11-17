package specta

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// PropertyPrimitives implements the Primitives interface using a Source for property testing.
// Unlike the deterministic Gen implementation, this generates fully random values from the
// complete type space, enabling property-based testing that finds bugs through exploration.
//
// Key differences from Gen:
//   - All values explore the FULL type space (empty strings, negatives, NaN, etc.)
//   - String() and ID() return any possible string (not just readable "str_1" format)
//   - Int() includes negatives, not just positives
//   - Float64() includes ±Inf, NaN, subnormals, not just [0, 1)
//   - Time() includes zero time, distant past/future, not just near epoch
//   - Next() returns random uint64 values (not sequential)
//
// This adversarial generation finds bugs from unexpected values (unicode in strings,
// division by zero from zero/negative ints, NaN propagation, etc.).
//
// Use generators to constrain values: String().Prefix("user"), Int().Range(0, 100)
type PropertyPrimitives struct {
	source   Source
	baseTime time.Time
}

// PrimitivesOption configures PropertyPrimitives behavior.
type PrimitivesOption func(*PropertyPrimitives)

// WithPropertyBaseTime sets the base time for TimeAtOffset() generation.
// This does NOT affect Time() which explores the full time space.
func WithPropertyBaseTime(baseTime time.Time) PrimitivesOption {
	return func(p *PropertyPrimitives) {
		p.baseTime = baseTime
	}
}

// NewPropertyPrimitives creates a Primitives implementation backed by a Source.
// This generates values from the FULL type space for adversarial testing.
//
// Options:
//   - WithPropertyBaseTime: Set base time for TimeAtOffset() (default: Unix epoch)
//
// Example:
//
//	Property(t, func(t *T) {
//	    p := t.Primitives()  // Convenience method
//	    user := UserRecipe{}.Build(p)
//	    // Test properties of user...
//	    // Will find bugs from empty strings, unicode, negatives, etc.
//	})
func NewPropertyPrimitives(source Source, opts ...PrimitivesOption) Primitives {
	p := &PropertyPrimitives{
		source:   source,
		baseTime: time.Unix(0, 0).UTC(),
	}

	// Apply options
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Next returns a random uint64 value from the Source.
// Unlike Gen which returns sequential counters, this returns random values.
func (p *PropertyPrimitives) Next() uint64 {
	return p.source.DrawBits(64)
}

// Bool generates a random boolean value from the Source.
func (p *PropertyPrimitives) Bool() bool {
	return p.source.DrawBits(1) == 1
}

// Int generates a random int from the full range of int values.
// This includes negative values, zero, and positive values.
func (p *PropertyPrimitives) Int() int {
	// Generate full 64-bit value then cast to int
	// This will be 32-bit or 64-bit depending on platform
	bits := p.source.DrawBits(64)
	return int(int64(bits)) // Cast through int64 to get sign extension
}

// IntN generates a random int in range [0, max) from the Source.
func (p *PropertyPrimitives) IntN(max int) int {
	if max <= 0 {
		return 0
	}
	bits := p.source.DrawBits(63)
	return int(bits % uint64(max))
}

// Int64 generates a random int64 from the full range including negative values.
// This includes math.MinInt64, negative values, zero, positive values, and math.MaxInt64.
func (p *PropertyPrimitives) Int64() int64 {
	return int64(p.source.DrawBits(64))
}

// Uint64 generates a random uint64 from the Source.
func (p *PropertyPrimitives) Uint64() uint64 {
	return p.source.DrawBits(64)
}

// Float64 generates a random float64 from the full range including special values.
// This includes: negative values, -0.0, 0.0, positive values, ±Inf, NaN, subnormals.
func (p *PropertyPrimitives) Float64() float64 {
	// Use different strategies to get variety including edge cases
	strategy := p.IntN(20)

	switch {
	case strategy == 0: // NaN
		return math.NaN()
	case strategy == 1: // +Inf
		return math.Inf(1)
	case strategy == 2: // -Inf
		return math.Inf(-1)
	case strategy == 3: // 0.0
		return 0.0
	case strategy == 4: // -0.0
		return math.Copysign(0.0, -1.0)
	case strategy == 5: // MaxFloat64
		return math.MaxFloat64
	case strategy == 6: // -MaxFloat64
		return -math.MaxFloat64
	case strategy == 7: // SmallestNonzeroFloat64
		return math.SmallestNonzeroFloat64
	case strategy == 8: // -SmallestNonzeroFloat64
		return -math.SmallestNonzeroFloat64
	default:
		// Random float from bit pattern (includes subnormals, all ranges)
		bits := p.source.DrawBits(64)
		return math.Float64frombits(bits)
	}
}

// String generates a random string from the Source.
// This explores the FULL space of possible Go strings, including invalid UTF-8.
// Go strings are just []byte and don't require valid UTF-8 encoding.
// To constrain strings, use generators like String().UTF8() or String().ASCII().
func (p *PropertyPrimitives) String() string {
	return string(p.Bytes())
}

// StringWith generates a random string with the given prefix.
// Format: "{prefix}_{random_bytes}" where random_bytes is from full space.
// This is a constraint - use it when you explicitly need a prefix.
func (p *PropertyPrimitives) StringWith(prefix string) string {
	suffix := p.String()
	if prefix == "" {
		return suffix
	}
	return fmt.Sprintf("%s_%s", prefix, suffix)
}

// Bytes generates a random byte slice from the Source.
// Length is random from 0 to 100 bytes, exploring the full range.
func (p *PropertyPrimitives) Bytes() []byte {
	// Random length from 0 to 100 bytes
	length := p.IntN(101)
	return p.BytesN(length)
}

// BytesN generates n random bytes from the Source.
func (p *PropertyPrimitives) BytesN(n int) []byte {
	if n <= 0 {
		return []byte{}
	}

	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = byte(p.source.DrawBits(8))
	}
	return result
}

// Time generates a random time from the full range of time.Time values.
// This includes: zero time, distant past, distant future, and everything in between.
func (p *PropertyPrimitives) Time() time.Time {
	// Use different strategies to get variety including edge cases
	strategy := p.IntN(10)

	switch {
	case strategy == 0: // Zero time
		return time.Time{}
	case strategy == 1: // Unix epoch
		return time.Unix(0, 0).UTC()
	case strategy == 2: // Distant past (year 1)
		return time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	case strategy == 3: // Distant future (year 9999)
		return time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	default:
		// Random time by generating random Unix timestamp
		// Use full int64 range for seconds (-292 billion years to +292 billion years)
		sec := int64(p.source.DrawBits(64))
		nsec := int64(p.IntN(1000000000)) // 0 to 999,999,999 nanoseconds
		return time.Unix(sec, nsec).UTC()
	}
}

// TimeAtOffset generates a time at the specified offset from the base time.
func (p *PropertyPrimitives) TimeAtOffset(d time.Duration) time.Time {
	return p.baseTime.Add(d)
}

// Duration generates a random duration from the full range of time.Duration values.
// This includes: negative durations, zero, positive durations, math.MinInt64, math.MaxInt64.
func (p *PropertyPrimitives) Duration() time.Duration {
	// time.Duration is an int64 of nanoseconds
	// Generate full int64 range
	return time.Duration(int64(p.source.DrawBits(64)))
}

// ID generates a random string from the full string space.
// This is identical to String() - property testing explores all possible strings.
// To constrain IDs, use generators like String().Prefix("id_") or String().AlphaNum().
func (p *PropertyPrimitives) ID() string {
	return p.String()
}

// UUID generates a random UUID using bytes from the Source.
func (p *PropertyPrimitives) UUID() uuid.UUID {
	bytes := p.BytesN(16)

	// Set version 4 (random) and variant bits per RFC 4122
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant 10

	var id uuid.UUID
	copy(id[:], bytes)
	return id
}

// Ensure PropertyPrimitives implements Primitives interface
var _ Primitives = (*PropertyPrimitives)(nil)
