package specta

import "fmt"

// IntGenerator generates int64 values with optional range constraints.
type IntGenerator struct {
	min *int64
	max *int64
}

// Int creates a new integer generator.
// By default it generates values across the full int64 range.
// Use Range() to constrain the values.
func Int() *IntGenerator {
	return &IntGenerator{}
}

// Range constrains the generator to produce values between min and max (inclusive).
func (g *IntGenerator) Range(min, max int64) *IntGenerator {
	if min > max {
		panic(fmt.Sprintf("IntGenerator.Range: min (%d) must be <= max (%d)", min, max))
	}
	g.min = &min
	g.max = &max
	return g
}

// Draw generates an int64 value from the source and records it in the log.
// The label is used to identify this value in error messages.
func (g *IntGenerator) Draw(t *T, label string) int64 {
	// Draw 64 random bits
	bits := t.Source.DrawBits(64)

	// Convert to int64
	value := int64(bits)

	// Apply range constraints if specified
	if g.min != nil && g.max != nil {
		// Map the value to [min, max] range
		rangeSize := uint64(*g.max - *g.min + 1)
		if rangeSize > 0 {
			offset := bits % rangeSize
			value = *g.min + int64(offset)
		} else {
			// Range is just a single value
			value = *g.min
		}
	}

	// Log the generated value
	t.Source.WriteLog(fmt.Sprintf("Int(%s)=%d ", label, value))

	return value
}
