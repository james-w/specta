package specta

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PropertyPrimitives implements the Primitives interface using a Source for property testing.
// Unlike the deterministic Gen implementation, this generates pseudo-random values based
// on the Source's random byte stream, enabling property-based testing with shrinking.
//
// Key differences from Gen:
//   - Values are pseudo-random (from Source) rather than strictly sequential
//   - Counter (Next) still increments for uniqueness but starts from random position
//   - String prefixes work the same way
//   - Time generation uses random offsets
//   - UUIDs are random (via Source) rather than deterministic
type PropertyPrimitives struct {
	source   *Source
	counter  uint64
	prefix   string
	baseTime time.Time
	timeStep time.Duration
}

// PrimitivesOption configures PropertyPrimitives behavior.
type PrimitivesOption func(*PropertyPrimitives)

// WithPropertyPrefix sets the string prefix for String() generation.
func WithPropertyPrefix(prefix string) PrimitivesOption {
	return func(p *PropertyPrimitives) {
		p.prefix = prefix
	}
}

// WithPropertyBaseTime sets the base time for Time() generation.
func WithPropertyBaseTime(baseTime time.Time) PrimitivesOption {
	return func(p *PropertyPrimitives) {
		p.baseTime = baseTime
	}
}

// WithPropertyTimeStep sets the time increment step.
func WithPropertyTimeStep(step time.Duration) PrimitivesOption {
	return func(p *PropertyPrimitives) {
		p.timeStep = step
	}
}

// NewPropertyPrimitives creates a Primitives implementation backed by a Source.
// This allows existing factories to work in property tests.
//
// Options:
//   - WithPropertyPrefix: Set string prefix (default: "")
//   - WithPropertyBaseTime: Set base time for Time() generation (default: Unix epoch)
//   - WithPropertyTimeStep: Set time increment step (default: 1 second)
//
// Example:
//
//	Property(t, func(t *T) {
//	    p := NewPropertyPrimitives(t.Source)
//	    user := UserFactory{p}.Build()
//	    // Test properties of user...
//	})
func NewPropertyPrimitives(source *Source, opts ...PrimitivesOption) Primitives {
	p := &PropertyPrimitives{
		source:   source,
		counter:  0,
		prefix:   "",
		baseTime: time.Unix(0, 0).UTC(),
		timeStep: time.Second,
	}

	// Apply options
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Next returns a sequentially incrementing counter.
// The counter provides uniqueness within a test run, useful for IDs.
func (p *PropertyPrimitives) Next() uint64 {
	p.counter++
	return p.counter
}

// Bool generates a random boolean value from the Source.
func (p *PropertyPrimitives) Bool() bool {
	return p.source.DrawBits(1) == 1
}

// Int generates a random int from the Source.
func (p *PropertyPrimitives) Int() int {
	return int(p.source.DrawBits(63)) // Use 63 bits to avoid overflow
}

// IntN generates a random int in range [0, max) from the Source.
func (p *PropertyPrimitives) IntN(max int) int {
	if max <= 0 {
		return 0
	}
	bits := p.source.DrawBits(63)
	return int(bits % uint64(max))
}

// Int64 generates a random int64 from the Source.
func (p *PropertyPrimitives) Int64() int64 {
	return int64(p.source.DrawBits(63)) // Use 63 bits to keep positive
}

// Uint64 generates a random uint64 from the Source.
func (p *PropertyPrimitives) Uint64() uint64 {
	return p.source.DrawBits(64)
}

// Float64 generates a random float64 in range [0.0, 1.0) from the Source.
func (p *PropertyPrimitives) Float64() float64 {
	// Generate a random uint64 and convert to float in [0, 1)
	bits := p.source.DrawBits(53) // Use 53 bits for IEEE 754 double precision mantissa
	return float64(bits) / float64(uint64(1)<<53)
}

// String generates a string with the configured prefix and an incrementing counter.
// Format: "{prefix}_{counter}"
func (p *PropertyPrimitives) String() string {
	return p.StringWith(p.prefix)
}

// StringWith generates a string with the given prefix and an incrementing counter.
// Format: "{prefix}_{counter}"
func (p *PropertyPrimitives) StringWith(prefix string) string {
	n := p.Next()
	if prefix == "" {
		return fmt.Sprintf("str_%d", n)
	}
	return fmt.Sprintf("%s_%d", prefix, n)
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

// Time generates a time by adding counter-based offsets to the base time.
// Each call increments the offset by timeStep.
func (p *PropertyPrimitives) Time() time.Time {
	n := p.Next()
	return p.baseTime.Add(time.Duration(n) * p.timeStep)
}

// TimeAtOffset generates a time at the specified offset from the base time.
func (p *PropertyPrimitives) TimeAtOffset(d time.Duration) time.Time {
	return p.baseTime.Add(d)
}

// Duration generates a duration based on the counter and time step.
func (p *PropertyPrimitives) Duration() time.Duration {
	n := p.Next()
	return time.Duration(n) * p.timeStep
}

// ID generates a string ID with "id_" prefix and counter.
func (p *PropertyPrimitives) ID() string {
	return fmt.Sprintf("id_%d", p.Next())
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
