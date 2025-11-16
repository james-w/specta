package specta

import (
	"fmt"
	"math"
)

// Generator produces values of type Value from a Source.
// Generators can be composed and constrained to control the distribution of values.
// They integrate with shrinking - the Source tracks choices for later simplification.
type Generator[Value any] interface {
	// Draw generates a value from the source and logs it with the given label.
	// The label is used to identify this value in error messages and logs.
	Draw(t *T, label string) Value
}

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

// validate checks that the generator's constraints are consistent.
// Panics if constraints are impossible to satisfy.
func (g *IntGenerator) validate() {
	if g.min != nil && g.max != nil && *g.min > *g.max {
		panic(fmt.Sprintf("IntGenerator: Min(%d) > Max(%d)", *g.min, *g.max))
	}
}

// Range constrains the generator to produce values between min and max (inclusive).
func (g *IntGenerator) Range(min, max int64) *IntGenerator {
	g.min = &min
	g.max = &max
	g.validate()
	return g
}

// Min constrains the generator to produce values >= min.
func (g *IntGenerator) Min(min int64) *IntGenerator {
	g.min = &min
	g.validate()
	return g
}

// Max constrains the generator to produce values <= max.
func (g *IntGenerator) Max(max int64) *IntGenerator {
	g.max = &max
	g.validate()
	return g
}

// Positive constrains the generator to produce values > 0.
func (g *IntGenerator) Positive() *IntGenerator {
	one := int64(1)
	g.min = &one
	g.validate()
	return g
}

// NonNegative constrains the generator to produce values >= 0.
func (g *IntGenerator) NonNegative() *IntGenerator {
	zero := int64(0)
	g.min = &zero
	g.validate()
	return g
}

// Negative constrains the generator to produce values < 0.
func (g *IntGenerator) Negative() *IntGenerator {
	negOne := int64(-1)
	g.max = &negOne
	g.validate()
	return g
}

// Draw generates an int64 value from the source and records it in the log.
// The label is used to identify this value in error messages.
func (g *IntGenerator) Draw(t *T, label string) int64 {
	// Draw 64 random bits
	bits := t.Source.DrawBits(64)

	// Convert to int64
	value := int64(bits)

	// Determine effective min and max
	min := int64(math.MinInt64)
	max := int64(math.MaxInt64)
	if g.min != nil {
		min = *g.min
	}
	if g.max != nil {
		max = *g.max
	}

	// If we have any constraints, map the value into [min, max]
	if g.min != nil || g.max != nil {
		// Calculate range size
		// Use big.Int for calculations to avoid overflow
		rangeSize := uint64(max - min + 1)
		if rangeSize == 0 {
			// Special case: range wraps around (e.g., MinInt64 to MaxInt64)
			// In this case, any value is valid
			value = int64(bits)
		} else {
			// Map bits to [0, rangeSize) then add min
			offset := bits % rangeSize
			value = min + int64(offset)
		}
	}

	// Log the generated value
	t.Source.WriteLog(fmt.Sprintf("Int(%s)=%d ", label, value))

	return value
}

// StringGenerator generates string values with optional constraints.
type StringGenerator struct {
	minLen  *int
	maxLen  *int
	prefix  string
	suffix  string
	charset stringCharset
}

type stringCharset int

const (
	charsetAny       stringCharset = iota // Any bytes (including invalid UTF-8)
	charsetASCII                          // ASCII characters (0x00-0x7F)
	charsetPrintable                      // Printable ASCII (0x20-0x7E)
	charsetAlphaNum                       // Alphanumeric (a-z, A-Z, 0-9)
	charsetAlpha                          // Alphabetic (a-z, A-Z)
)

// String creates a new string generator.
// By default it generates any byte sequence (including invalid UTF-8).
func String() *StringGenerator {
	return &StringGenerator{
		charset: charsetAny,
	}
}

// validate checks that the generator's constraints are consistent.
// Panics if constraints are impossible to satisfy.
func (g *StringGenerator) validate() {
	// Check minLen <= maxLen
	if g.minLen != nil && g.maxLen != nil && *g.minLen > *g.maxLen {
		panic(fmt.Sprintf("StringGenerator: MinLen(%d) > MaxLen(%d)", *g.minLen, *g.maxLen))
	}

	// Check that prefix+suffix doesn't exceed maxLen
	fixedLen := len(g.prefix) + len(g.suffix)
	if g.maxLen != nil && fixedLen > *g.maxLen {
		panic(fmt.Sprintf("StringGenerator: Prefix(%q) + Suffix(%q) = %d bytes > MaxLen(%d)",
			g.prefix, g.suffix, fixedLen, *g.maxLen))
	}
}

// MinLen constrains the generator to produce strings with at least minLen bytes.
func (g *StringGenerator) MinLen(minLen int) *StringGenerator {
	if minLen < 0 {
		panic(fmt.Sprintf("StringGenerator.MinLen: minLen (%d) must be >= 0", minLen))
	}
	g.minLen = &minLen
	g.validate()
	return g
}

// MaxLen constrains the generator to produce strings with at most maxLen bytes.
func (g *StringGenerator) MaxLen(maxLen int) *StringGenerator {
	if maxLen < 0 {
		panic(fmt.Sprintf("StringGenerator.MaxLen: maxLen (%d) must be >= 0", maxLen))
	}
	g.maxLen = &maxLen
	g.validate()
	return g
}

// Len constrains the generator to produce strings with exactly len bytes.
func (g *StringGenerator) Len(len int) *StringGenerator {
	if len < 0 {
		panic(fmt.Sprintf("StringGenerator.Len: len (%d) must be >= 0", len))
	}
	g.minLen = &len
	g.maxLen = &len
	g.validate()
	return g
}

// NonEmpty constrains the generator to produce non-empty strings.
func (g *StringGenerator) NonEmpty() *StringGenerator {
	one := 1
	g.minLen = &one
	g.validate()
	return g
}

// Prefix constrains the generator to produce strings starting with the given prefix.
func (g *StringGenerator) Prefix(prefix string) *StringGenerator {
	g.prefix = prefix
	g.validate()
	return g
}

// Suffix constrains the generator to produce strings ending with the given suffix.
func (g *StringGenerator) Suffix(suffix string) *StringGenerator {
	g.suffix = suffix
	g.validate()
	return g
}

// ASCII constrains the generator to produce ASCII strings (bytes 0x00-0x7F).
func (g *StringGenerator) ASCII() *StringGenerator {
	g.charset = charsetASCII
	return g
}

// Printable constrains the generator to produce printable ASCII strings (bytes 0x20-0x7E).
func (g *StringGenerator) Printable() *StringGenerator {
	g.charset = charsetPrintable
	return g
}

// AlphaNum constrains the generator to produce alphanumeric strings (a-z, A-Z, 0-9).
func (g *StringGenerator) AlphaNum() *StringGenerator {
	g.charset = charsetAlphaNum
	return g
}

// Alpha constrains the generator to produce alphabetic strings (a-z, A-Z).
func (g *StringGenerator) Alpha() *StringGenerator {
	g.charset = charsetAlpha
	return g
}

// Draw generates a string value from the source and records it in the log.
func (g *StringGenerator) Draw(t *T, label string) string {
	// Determine length
	minLen := 0
	if g.minLen != nil {
		minLen = *g.minLen
	}
	maxLen := 100 // Default max length
	if g.maxLen != nil {
		maxLen = *g.maxLen
	}

	// Account for prefix/suffix in length calculation
	prefixLen := len(g.prefix)
	suffixLen := len(g.suffix)
	fixedLen := prefixLen + suffixLen

	// Adjust min/max for the random part
	randomMinLen := minLen - fixedLen
	if randomMinLen < 0 {
		randomMinLen = 0
	}
	randomMaxLen := maxLen - fixedLen
	if randomMaxLen < 0 {
		// Prefix+suffix already exceeds maxLen, just use prefix+suffix
		result := g.prefix + g.suffix
		t.Source.WriteLog(fmt.Sprintf("String(%s)=%q ", label, result))
		return result
	}

	// Generate random length for the middle part
	var randomLen int
	if randomMaxLen == randomMinLen {
		randomLen = randomMinLen
	} else {
		lengthRange := randomMaxLen - randomMinLen + 1
		randomLen = randomMinLen + int(t.Source.DrawBits(32)%uint64(lengthRange))
	}

	// Generate random bytes based on charset
	var randomBytes []byte
	if randomLen > 0 {
		randomBytes = make([]byte, randomLen)
		for i := 0; i < randomLen; i++ {
			randomBytes[i] = g.generateByte(t)
		}
	}

	result := g.prefix + string(randomBytes) + g.suffix
	t.Source.WriteLog(fmt.Sprintf("String(%s)=%q ", label, result))
	return result
}

func (g *StringGenerator) generateByte(t *T) byte {
	switch g.charset {
	case charsetAny:
		// Any byte value
		return byte(t.Source.DrawBits(8))

	case charsetASCII:
		// ASCII: 0x00-0x7F
		return byte(t.Source.DrawBits(7))

	case charsetPrintable:
		// Printable ASCII: 0x20 (' ') to 0x7E ('~')
		return byte(0x20 + t.Source.DrawBits(7)%(0x7F-0x20))

	case charsetAlphaNum:
		// Alphanumeric: a-z, A-Z, 0-9 (62 characters)
		const alphaNum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		idx := t.Source.DrawBits(6) % 62
		return alphaNum[idx]

	case charsetAlpha:
		// Alphabetic: a-z, A-Z (52 characters)
		const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		idx := t.Source.DrawBits(6) % 52
		return alpha[idx]

	default:
		return byte(t.Source.DrawBits(8))
	}
}
