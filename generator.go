package specta

import (
	"fmt"
	"math"
	"strings"
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

// Range constrains the generator to produce values between min and max (inclusive).
func (g *IntGenerator) Range(min, max int64) *IntGenerator {
	if min > max {
		panic(fmt.Sprintf("IntGenerator.Range: min (%d) must be <= max (%d)", min, max))
	}
	g.min = &min
	g.max = &max
	return g
}

// Min constrains the generator to produce values >= min.
func (g *IntGenerator) Min(min int64) *IntGenerator {
	g.min = &min
	return g
}

// Max constrains the generator to produce values <= max.
func (g *IntGenerator) Max(max int64) *IntGenerator {
	g.max = &max
	return g
}

// Positive constrains the generator to produce values > 0.
func (g *IntGenerator) Positive() *IntGenerator {
	one := int64(1)
	g.min = &one
	return g
}

// NonNegative constrains the generator to produce values >= 0.
func (g *IntGenerator) NonNegative() *IntGenerator {
	zero := int64(0)
	g.min = &zero
	return g
}

// Negative constrains the generator to produce values < 0.
func (g *IntGenerator) Negative() *IntGenerator {
	negOne := int64(-1)
	g.max = &negOne
	return g
}

// Draw generates an int64 value from the source and records it in the log.
// The label is used to identify this value in error messages.
func (g *IntGenerator) Draw(t *T, label string) int64 {
	// Draw bits from source
	bits := t.Source.DrawBits(64)

	// Determine effective min and max
	min := int64(math.MinInt64)
	max := int64(math.MaxInt64)
	if g.min != nil {
		min = *g.min
	}
	if g.max != nil {
		max = *g.max
	}

	var value int64

	if t.Source.IsDeterministic() {
		// Deterministic mode: use counter directly for small, predictable values
		// This gives us 0, 1, 2, 3... which is friendly for debugging
		value = int64(bits)

		// Apply constraints
		if g.min != nil || g.max != nil {
			// Wrap value into [min, max] range
			rangeSize := uint64(max - min + 1)
			if rangeSize == 0 {
				// Full range, value is already good
				value = int64(bits)
			} else {
				// Map counter into constrained range
				offset := uint64(value) % rangeSize
				value = min + int64(offset)
			}
		}
	} else {
		// Random mode: full exploration of int64 space
		value = int64(bits)

		// Apply constraints
		if g.min != nil || g.max != nil {
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
	}

	// Log the generated value
	t.Source.WriteLog(fmt.Sprintf("Int(%s)=%d ", label, value))

	return value
}

// StringGenerator generates string values with optional constraints.
type StringGenerator struct {
	minLen      *int
	maxLen      *int
	prefix      string
	suffix      string
	charset     stringCharset
	exampleHint string
}

type stringCharset int

const (
	charsetAny stringCharset = iota // Any bytes (including invalid UTF-8)
	charsetASCII                     // ASCII characters (0x00-0x7F)
	charsetPrintable                 // Printable ASCII (0x20-0x7E)
	charsetAlphaNum                  // Alphanumeric (a-z, A-Z, 0-9)
	charsetAlpha                     // Alphabetic (a-z, A-Z)
)

// String creates a new string generator.
// By default it generates any byte sequence (including invalid UTF-8).
func String() *StringGenerator {
	return &StringGenerator{
		charset: charsetAny,
	}
}

// MinLen constrains the generator to produce strings with at least minLen bytes.
func (g *StringGenerator) MinLen(minLen int) *StringGenerator {
	if minLen < 0 {
		panic(fmt.Sprintf("StringGenerator.MinLen: minLen (%d) must be >= 0", minLen))
	}
	g.minLen = &minLen
	return g
}

// MaxLen constrains the generator to produce strings with at most maxLen bytes.
func (g *StringGenerator) MaxLen(maxLen int) *StringGenerator {
	if maxLen < 0 {
		panic(fmt.Sprintf("StringGenerator.MaxLen: maxLen (%d) must be >= 0", maxLen))
	}
	g.maxLen = &maxLen
	return g
}

// Len constrains the generator to produce strings with exactly len bytes.
func (g *StringGenerator) Len(len int) *StringGenerator {
	if len < 0 {
		panic(fmt.Sprintf("StringGenerator.Len: len (%d) must be >= 0", len))
	}
	g.minLen = &len
	g.maxLen = &len
	return g
}

// NonEmpty constrains the generator to produce non-empty strings.
func (g *StringGenerator) NonEmpty() *StringGenerator {
	one := 1
	g.minLen = &one
	return g
}

// Prefix constrains the generator to produce strings starting with the given prefix.
func (g *StringGenerator) Prefix(prefix string) *StringGenerator {
	g.prefix = prefix
	return g
}

// Suffix constrains the generator to produce strings ending with the given suffix.
func (g *StringGenerator) Suffix(suffix string) *StringGenerator {
	g.suffix = suffix
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

// ExampleHint provides a soft prefix hint for deterministic generation.
// This is used when generating friendly example values with Gen.
// If a hard Prefix() constraint is set, it takes precedence over the hint.
func (g *StringGenerator) ExampleHint(hint string) *StringGenerator {
	g.exampleHint = hint
	return g
}

// Draw generates a string value from the source and records it in the log.
func (g *StringGenerator) Draw(t *T, label string) string {
	if t.Source.IsDeterministic() {
		return g.drawDeterministic(t, label)
	}
	return g.drawRandom(t, label)
}

// drawDeterministic generates friendly, predictable strings like "user_1", "user_2"
func (g *StringGenerator) drawDeterministic(t *T, label string) string {
	// Get counter value
	counter := t.Source.DrawBits(64)

	// Determine effective prefix (hard constraint or soft hint)
	effectivePrefix := g.prefix
	if effectivePrefix == "" && g.exampleHint != "" {
		effectivePrefix = g.exampleHint
	}

	// Generate content from counter respecting charset
	content := g.formatCounter(counter, g.charset)

	// Build with prefix + content + suffix
	result := effectivePrefix + content
	if g.suffix != "" {
		result += g.suffix
	}

	// Enforce length constraints
	result = g.enforceLength(result, g.charset)

	t.Source.WriteLog(fmt.Sprintf("String(%s)=%q ", label, result))
	return result
}

// drawRandom generates adversarial strings exploring full type space
func (g *StringGenerator) drawRandom(t *T, label string) string {
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

// formatCounter formats a counter value as a string respecting charset constraints
func (g *StringGenerator) formatCounter(counter uint64, charset stringCharset) string {
	switch charset {
	case charsetAlpha:
		// For alpha, convert to base-26 using letters
		if counter == 0 {
			return "a"
		}
		result := ""
		for counter > 0 {
			result = string(rune('a'+(counter%26))) + result
			counter /= 26
		}
		return result
	case charsetAlphaNum, charsetPrintable, charsetASCII, charsetAny:
		// For others, use decimal representation (digits are valid in all these)
		return fmt.Sprintf("%d", counter)
	default:
		return fmt.Sprintf("%d", counter)
	}
}

// enforceLength pads or truncates the string to meet length constraints
func (g *StringGenerator) enforceLength(s string, charset stringCharset) string {
	minLen := 0
	if g.minLen != nil {
		minLen = *g.minLen
	}
	maxLen := 100
	if g.maxLen != nil {
		maxLen = *g.maxLen
	}

	// Truncate if too long
	if len(s) > maxLen {
		// Try to preserve suffix if present
		if g.suffix != "" && len(g.suffix) < maxLen {
			// Truncate from the middle, keep suffix
			prefixPart := s[:maxLen-len(g.suffix)]
			return prefixPart + g.suffix
		}
		return s[:maxLen]
	}

	// Pad if too short
	if len(s) < minLen {
		padding := minLen - len(s)
		padChar := g.getPadChar(charset)
		return s + strings.Repeat(string(padChar), padding)
	}

	return s
}

// getPadChar returns an appropriate padding character for the charset
func (g *StringGenerator) getPadChar(charset stringCharset) rune {
	switch charset {
	case charsetAlpha:
		return 'a'
	case charsetAlphaNum, charsetPrintable, charsetASCII:
		return '0'
	default:
		return '0'
	}
}
