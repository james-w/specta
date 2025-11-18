package specta

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Generator produces values of type Value from a Source.
// Generators can be composed and constrained to control the distribution of values.
// They integrate with shrinking - the Source tracks choices for later simplification.
type Generator[Value any] interface {
	// Draw generates a value from the source and logs it with the given label.
	// The label is used to identify this value in error messages and logs.
	Draw(s Source, label string) Value
}

// providerGenerator adapts a Provider[T] to a Generator[T].
type providerGenerator[T any] struct {
	provider Provider[T]
}

// GeneratorFromProvider converts a Provider[T] into a Generator[T].
// This is useful for using existing providers with generators that expect Generator[T].
func GeneratorFromProvider[T any](p Provider[T]) Generator[T] {
	return &providerGenerator[T]{provider: p}
}

// Draw implements Generator[T] by calling the Provider.
func (g *providerGenerator[T]) Draw(s Source, label string) T {
	return g.provider(s)
}

// IntGenerator generates int64 values with optional range constraints.
type IntGenerator struct {
	min               *int64
	max               *int64
	filterFn          func(int64) bool
	maxFilterAttempts int
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

// Filter constrains the generator to only produce values that satisfy the predicate.
// The generator will retry up to 100 times (by default) to find a matching value.
// If no matching value is found after max attempts, the test iteration is skipped (via t.Assume).
//
// Use Filter for predicates that pass frequently (>50% of values).
// For rare conditions, use t.Assume() instead to skip test iterations directly.
//
// Example:
//
//	primes := Int().Range(1, 100).Filter(isPrime)
func (g *IntGenerator) Filter(fn func(int64) bool) *IntGenerator {
	g.filterFn = fn
	if g.maxFilterAttempts == 0 {
		g.maxFilterAttempts = 100
	}
	return g
}

// Draw generates an int64 value from the source and records it in the log.
// The label is used to identify this value in error messages.
func (g *IntGenerator) Draw(s Source, label string) int64 {
	// Draw bits from source
	bits := s.DrawBits(64)

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

	if s.IsDeterministic() {
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

		// Log and return the value
		s.WriteLog(fmt.Sprintf("Int(%s)=%d ", label, value))
		return value
	} else {
		maxAttempts := 1
		if g.filterFn != nil {
			maxAttempts = g.maxFilterAttempts
		}

		for attempt := 0; attempt < maxAttempts; attempt++ {
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

			// Check filter predicate if present
			if g.filterFn != nil && !g.filterFn(value) {
				continue // Try again
			}

			// Log the generated value
			s.WriteLog(fmt.Sprintf("Int(%s)=%d ", label, value))

			return value
		}
	}

	// Filter exhausted max attempts - skip this test iteration
	panic(skipTest{})
}

// StringGenerator generates string values with optional constraints.
type StringGenerator struct {
	minLen            *int
	maxLen            *int
	prefix            string
	suffix            string
	charset           stringCharset
	exampleHint       string
	filterFn          func(string) bool
	maxFilterAttempts int
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

// ExampleHint provides a soft prefix hint for deterministic generation.
// This is used when generating friendly example values with Gen.
// If a hard Prefix() constraint is set, it takes precedence over the hint.
func (g *StringGenerator) ExampleHint(hint string) *StringGenerator {
	g.exampleHint = hint
	return g
}

// Filter constrains the generator to only produce strings that satisfy the predicate.
// The generator will retry up to 100 times (by default) to find a matching value.
// If no matching value is found after max attempts, the test iteration is skipped (via t.Assume).
//
// Use Filter for predicates that pass frequently (>50% of values).
// For rare conditions, use t.Assume() instead to skip test iterations directly.
//
// Example:
//
//	validEmails := String().AlphaNum().Filter(func(s string) bool {
//	    return strings.Contains(s, "@") && len(s) > 5
//	})
func (g *StringGenerator) Filter(fn func(string) bool) *StringGenerator {
	g.filterFn = fn
	if g.maxFilterAttempts == 0 {
		g.maxFilterAttempts = 100
	}
	return g
}

// Draw generates a string value from the source and records it in the log.
func (g *StringGenerator) Draw(s Source, label string) string {
	if s.IsDeterministic() {
		return g.drawDeterministic(s, label)
	}
	return g.drawRandom(s, label)
}

// drawDeterministic generates friendly, predictable strings like "user_1", "user_2"
func (g *StringGenerator) drawDeterministic(s Source, label string) string {
	// Get counter value
	counter := s.DrawBits(64)

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

	s.WriteLog(fmt.Sprintf("String(%s)=%q ", label, result))
	return result
}

// drawRandom generates adversarial strings exploring full type space
func (g *StringGenerator) drawRandom(s Source, label string) string {
	// Determine length
	minLen := 0
	if g.minLen != nil {
		minLen = *g.minLen
	}
	maxLen := 100 // Default max length
	if g.maxLen != nil {
		maxLen = *g.maxLen
	}

	maxAttempts := 1
	if g.filterFn != nil {
		maxAttempts = g.maxFilterAttempts
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
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
			if g.filterFn == nil || g.filterFn(result) {
				s.WriteLog(fmt.Sprintf("String(%s)=%q ", label, result))
				return result
			}
			continue // Try again with filter
		}

		// Generate random length for the middle part
		var randomLen int
		if randomMaxLen == randomMinLen {
			randomLen = randomMinLen
		} else {
			lengthRange := randomMaxLen - randomMinLen + 1
			randomLen = randomMinLen + int(s.DrawBits(32)%uint64(lengthRange))
		}

		// Generate random bytes based on charset
		var randomBytes []byte
		if randomLen > 0 {
			randomBytes = make([]byte, randomLen)
			for i := 0; i < randomLen; i++ {
				randomBytes[i] = g.generateByte(s)
			}
		}

		result := g.prefix + string(randomBytes) + g.suffix

		// Check filter predicate if present
		if g.filterFn != nil && !g.filterFn(result) {
			continue // Try again
		}

		s.WriteLog(fmt.Sprintf("String(%s)=%q ", label, result))
		return result
	}

	// Filter exhausted max attempts - skip this test iteration
	panic(skipTest{})
}

func (g *StringGenerator) generateByte(s Source) byte {
	switch g.charset {
	case charsetAny:
		// Any byte value
		return byte(s.DrawBits(8))

	case charsetASCII:
		// ASCII: 0x00-0x7F
		return byte(s.DrawBits(7))

	case charsetPrintable:
		// Printable ASCII: 0x20 (' ') to 0x7E ('~')
		return byte(0x20 + s.DrawBits(7)%(0x7F-0x20))

	case charsetAlphaNum:
		// Alphanumeric: a-z, A-Z, 0-9 (62 characters)
		const alphaNum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		idx := s.DrawBits(6) % 62
		return alphaNum[idx]

	case charsetAlpha:
		// Alphabetic: a-z, A-Z (52 characters)
		const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		idx := s.DrawBits(6) % 52
		return alpha[idx]

	default:
		return byte(s.DrawBits(8))
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

// BoolGenerator generates boolean values.
type BoolGenerator struct{}

// Bool creates a new boolean generator.
func Bool() *BoolGenerator {
	return &BoolGenerator{}
}

// Draw generates a boolean value from the source.
func (g *BoolGenerator) Draw(s Source, label string) bool {
	var value bool
	if s.IsDeterministic() {
		// Deterministic mode: always return false (simple, predictable)
		value = false
	} else {
		// Random mode: draw a random bit
		value = s.DrawBits(1) == 1
	}
	s.WriteLog(fmt.Sprintf("Bool(%s)=%v ", label, value))
	return value
}

// Float64Generator generates float64 values.
type Float64Generator struct{}

// Float64 creates a new float64 generator.
func Float64() *Float64Generator {
	return &Float64Generator{}
}

// Draw generates a float64 value from the source.
func (g *Float64Generator) Draw(s Source, label string) float64 {
	if s.IsDeterministic() {
		// Deterministic: small sequential values
		counter := s.DrawBits(32)
		value := float64(counter) + 0.5
		s.WriteLog(fmt.Sprintf("Float64(%s)=%v ", label, value))
		return value
	}
	// Random: full float64 space including special values
	bits := s.DrawBits(64)
	value := math.Float64frombits(bits)
	s.WriteLog(fmt.Sprintf("Float64(%s)=%v ", label, value))
	return value
}

// TimeGenerator generates time.Time values.
type TimeGenerator struct {
	baseTime time.Time
}

// Time creates a new time generator.
// Default base time is Unix epoch.
func Time() *TimeGenerator {
	return &TimeGenerator{
		baseTime: time.Unix(0, 0).UTC(),
	}
}

// BaseTime sets the base time for generation.
func (g *TimeGenerator) BaseTime(t time.Time) *TimeGenerator {
	g.baseTime = t
	return g
}

// Draw generates a time.Time value from the source.
func (g *TimeGenerator) Draw(s Source, label string) time.Time {
	if s.IsDeterministic() {
		// Deterministic: sequential times from base
		offset := s.DrawBits(32)
		value := g.baseTime.Add(time.Duration(offset) * time.Second)
		s.WriteLog(fmt.Sprintf("Time(%s)=%v ", label, value.Format(time.RFC3339)))
		return value
	}
	// Random: full time range
	bits := s.DrawBits(64)
	value := time.Unix(int64(bits), 0).UTC()
	s.WriteLog(fmt.Sprintf("Time(%s)=%v ", label, value.Format(time.RFC3339)))
	return value
}

// DurationGenerator generates time.Duration values.
type DurationGenerator struct{}

// Duration creates a new duration generator.
func Duration() *DurationGenerator {
	return &DurationGenerator{}
}

// Draw generates a time.Duration value from the source.
func (g *DurationGenerator) Draw(s Source, label string) time.Duration {
	if s.IsDeterministic() {
		// Deterministic: sequential durations
		counter := s.DrawBits(32)
		value := time.Duration(counter) * time.Second
		s.WriteLog(fmt.Sprintf("Duration(%s)=%v ", label, value))
		return value
	}
	// Random: full duration range
	bits := s.DrawBits(64)
	value := time.Duration(int64(bits))
	s.WriteLog(fmt.Sprintf("Duration(%s)=%v ", label, value))
	return value
}

// BytesGenerator generates byte slices.
type BytesGenerator struct {
	length *int
}

// Bytes creates a new bytes generator.
// Default generates 16 bytes for deterministic mode, 0-100 for random.
func Bytes() *BytesGenerator {
	return &BytesGenerator{}
}

// Len sets the exact length of the byte slice.
func (g *BytesGenerator) Len(n int) *BytesGenerator {
	g.length = &n
	return g
}

// Draw generates a byte slice from the source.
func (g *BytesGenerator) Draw(s Source, label string) []byte {
	var length int
	if g.length != nil {
		length = *g.length
	} else if s.IsDeterministic() {
		length = 16 // Default deterministic length
	} else {
		length = int(s.DrawBits(7)) // 0-127 bytes
	}

	if length == 0 {
		s.WriteLog(fmt.Sprintf("Bytes(%s)=[] ", label))
		return []byte{}
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = byte(s.DrawBits(8))
	}
	s.WriteLog(fmt.Sprintf("Bytes(%s)=[%d bytes] ", label, length))
	return result
}

// UUIDGenerator generates UUID values.
type UUIDGenerator struct{}

// UUID creates a new UUID generator.
func UUID() *UUIDGenerator {
	return &UUIDGenerator{}
}

// Draw generates a UUID from the source.
func (g *UUIDGenerator) Draw(s Source, label string) uuid.UUID {
	if s.IsDeterministic() {
		// Deterministic: SHA1-based UUID from counter
		counter := s.DrawBits(64)
		value := DeterministicUUIDFromInt(counter)
		s.WriteLog(fmt.Sprintf("UUID(%s)=%v ", label, value))
		return value
	}
	// Random: random bytes as UUID
	var data [16]byte
	for i := 0; i < 16; i++ {
		data[i] = byte(s.DrawBits(8))
	}
	// Set version 4 and variant bits
	data[6] = (data[6] & 0x0f) | 0x40 // Version 4
	data[8] = (data[8] & 0x3f) | 0x80 // Variant is 10
	value := uuid.Must(uuid.FromBytes(data[:]))
	s.WriteLog(fmt.Sprintf("UUID(%s)=%v ", label, value))
	return value
}

// SliceGenerator generates slices of values using an element generator.
type SliceGenerator[T any] struct {
	elementGen        Generator[T]
	minLen            *int
	maxLen            *int
	filterFn          func([]T) bool
	maxFilterAttempts int
}

// Slice creates a new slice generator using the given element generator.
// By default it generates slices with 0-100 elements.
// Use MinLen/MaxLen/Len to constrain the size.
//
// Example:
//
//	// Generate slices of integers between 1-10
//	ints := Slice(Int().Range(1, 10)).MinLen(5).MaxLen(20)
func Slice[T any](elementGen Generator[T]) *SliceGenerator[T] {
	return &SliceGenerator[T]{
		elementGen: elementGen,
	}
}

// validate checks that the generator's constraints are consistent.
func (g *SliceGenerator[T]) validate() {
	if g.minLen != nil && g.maxLen != nil && *g.minLen > *g.maxLen {
		panic(fmt.Sprintf("SliceGenerator: MinLen(%d) > MaxLen(%d)", *g.minLen, *g.maxLen))
	}
}

// MinLen constrains the generator to produce slices with at least minLen elements.
func (g *SliceGenerator[T]) MinLen(minLen int) *SliceGenerator[T] {
	if minLen < 0 {
		panic(fmt.Sprintf("SliceGenerator.MinLen: minLen (%d) must be >= 0", minLen))
	}
	g.minLen = &minLen
	g.validate()
	return g
}

// MaxLen constrains the generator to produce slices with at most maxLen elements.
func (g *SliceGenerator[T]) MaxLen(maxLen int) *SliceGenerator[T] {
	if maxLen < 0 {
		panic(fmt.Sprintf("SliceGenerator.MaxLen: maxLen (%d) must be >= 0", maxLen))
	}
	g.maxLen = &maxLen
	g.validate()
	return g
}

// Len constrains the generator to produce slices with exactly len elements.
func (g *SliceGenerator[T]) Len(len int) *SliceGenerator[T] {
	if len < 0 {
		panic(fmt.Sprintf("SliceGenerator.Len: len (%d) must be >= 0", len))
	}
	g.minLen = &len
	g.maxLen = &len
	g.validate()
	return g
}

// NonEmpty constrains the generator to produce non-empty slices.
func (g *SliceGenerator[T]) NonEmpty() *SliceGenerator[T] {
	one := 1
	g.minLen = &one
	g.validate()
	return g
}

// Filter constrains the generator to only produce slices that satisfy the predicate.
// The generator will retry up to 100 times (by default) to find a matching value.
// If no matching value is found after max attempts, the test iteration is skipped.
//
// Use Filter for predicates that pass frequently (>50% of values).
// For rare conditions, use t.Assume() instead to skip test iterations directly.
//
// Example:
//
//	uniqueInts := Slice(Int()).Filter(func(s []int) bool {
//	    seen := make(map[int]bool)
//	    for _, v := range s {
//	        if seen[v] { return false }
//	        seen[v] = true
//	    }
//	    return true
//	})
func (g *SliceGenerator[T]) Filter(fn func([]T) bool) *SliceGenerator[T] {
	g.filterFn = fn
	if g.maxFilterAttempts == 0 {
		g.maxFilterAttempts = 100
	}
	return g
}

// Draw generates a slice from the source.
func (g *SliceGenerator[T]) Draw(s Source, label string) []T {
	// Determine length bounds
	minLen := 0
	maxLen := 100 // Default max length
	if g.minLen != nil {
		minLen = *g.minLen
	}
	if g.maxLen != nil {
		maxLen = *g.maxLen
	}

	maxAttempts := 1
	if g.filterFn != nil {
		maxAttempts = g.maxFilterAttempts
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Generate a length
		var length int
		if s.IsDeterministic() {
			// Deterministic mode: use counter for predictable lengths
			// Start with minLen and increment
			counter := s.DrawBits(64)
			if minLen == maxLen {
				length = minLen
			} else {
				rangeSize := maxLen - minLen + 1
				length = minLen + int(counter%uint64(rangeSize))
			}
		} else {
			// Random mode: random length in range
			if minLen == maxLen {
				length = minLen
			} else {
				lengthRange := maxLen - minLen + 1
				length = minLen + int(s.DrawBits(32)%uint64(lengthRange))
			}
		}

		// Generate elements
		result := make([]T, length)
		for i := 0; i < length; i++ {
			result[i] = g.elementGen.Draw(s, fmt.Sprintf("%s[%d]", label, i))
		}

		// Check filter predicate if present
		if g.filterFn != nil && !g.filterFn(result) {
			continue // Try again
		}

		s.WriteLog(fmt.Sprintf("Slice(%s)=[%d elements] ", label, length))
		return result
	}

	// Filter exhausted max attempts - skip this test iteration
	panic(skipTest{})
}

// MapGenerator generates maps with key-value pairs using generators for keys and values.
type MapGenerator[K comparable, V any] struct {
	keyGen            Generator[K]
	valueGen          Generator[V]
	minLen            *int
	maxLen            *int
	filterFn          func(map[K]V) bool
	maxFilterAttempts int
}

// Map creates a new map generator using the given key and value generators.
// By default it generates maps with 0-100 entries.
// Use MinLen/MaxLen/Len to constrain the size.
//
// Example:
//
//	// Generate maps from string keys to int values
//	m := Map(String().AlphaNum(), Int().Range(0, 100)).MinLen(1).MaxLen(10)
func Map[K comparable, V any](keyGen Generator[K], valueGen Generator[V]) *MapGenerator[K, V] {
	return &MapGenerator[K, V]{
		keyGen:   keyGen,
		valueGen: valueGen,
	}
}

// validate checks that the generator's constraints are consistent.
func (g *MapGenerator[K, V]) validate() {
	if g.minLen != nil && g.maxLen != nil && *g.minLen > *g.maxLen {
		panic(fmt.Sprintf("MapGenerator: MinLen(%d) > MaxLen(%d)", *g.minLen, *g.maxLen))
	}
}

// MinLen constrains the generator to produce maps with at least minLen entries.
func (g *MapGenerator[K, V]) MinLen(minLen int) *MapGenerator[K, V] {
	if minLen < 0 {
		panic(fmt.Sprintf("MapGenerator.MinLen: minLen (%d) must be >= 0", minLen))
	}
	g.minLen = &minLen
	g.validate()
	return g
}

// MaxLen constrains the generator to produce maps with at most maxLen entries.
func (g *MapGenerator[K, V]) MaxLen(maxLen int) *MapGenerator[K, V] {
	if maxLen < 0 {
		panic(fmt.Sprintf("MapGenerator.MaxLen: maxLen (%d) must be >= 0", maxLen))
	}
	g.maxLen = &maxLen
	g.validate()
	return g
}

// Len constrains the generator to produce maps with exactly len entries.
func (g *MapGenerator[K, V]) Len(len int) *MapGenerator[K, V] {
	if len < 0 {
		panic(fmt.Sprintf("MapGenerator.Len: len (%d) must be >= 0", len))
	}
	g.minLen = &len
	g.maxLen = &len
	g.validate()
	return g
}

// NonEmpty constrains the generator to produce non-empty maps.
func (g *MapGenerator[K, V]) NonEmpty() *MapGenerator[K, V] {
	one := 1
	g.minLen = &one
	g.validate()
	return g
}

// Filter constrains the generator to only produce maps that satisfy the predicate.
// The generator will retry up to 100 times (by default) to find a matching value.
// If no matching value is found after max attempts, the test iteration is skipped.
//
// Use Filter for predicates that pass frequently (>50% of values).
// For rare conditions, use t.Assume() instead to skip test iterations directly.
//
// Example:
//
//	nonEmptyValues := Map(String(), Int()).Filter(func(m map[string]int) bool {
//	    for _, v := range m {
//	        if v == 0 { return false }
//	    }
//	    return true
//	})
func (g *MapGenerator[K, V]) Filter(fn func(map[K]V) bool) *MapGenerator[K, V] {
	g.filterFn = fn
	if g.maxFilterAttempts == 0 {
		g.maxFilterAttempts = 100
	}
	return g
}

// Draw generates a map from the source.
func (g *MapGenerator[K, V]) Draw(s Source, label string) map[K]V {
	// Determine length bounds
	minLen := 0
	maxLen := 100 // Default max length
	if g.minLen != nil {
		minLen = *g.minLen
	}
	if g.maxLen != nil {
		maxLen = *g.maxLen
	}

	maxAttempts := 1
	if g.filterFn != nil {
		maxAttempts = g.maxFilterAttempts
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Generate a target size
		var targetSize int
		if s.IsDeterministic() {
			// Deterministic mode: use counter for predictable sizes
			counter := s.DrawBits(64)
			if minLen == maxLen {
				targetSize = minLen
			} else {
				rangeSize := maxLen - minLen + 1
				targetSize = minLen + int(counter%uint64(rangeSize))
			}
		} else {
			// Random mode: random size in range
			if minLen == maxLen {
				targetSize = minLen
			} else {
				sizeRange := maxLen - minLen + 1
				targetSize = minLen + int(s.DrawBits(32)%uint64(sizeRange))
			}
		}

		// Generate entries until we have targetSize unique keys
		// Note: We may need to generate more than targetSize entries if keys collide
		result := make(map[K]V, targetSize)
		attempts := 0
		maxKeyAttempts := targetSize * 10 // Allow some retries for key collisions

		for len(result) < targetSize && attempts < maxKeyAttempts {
			key := g.keyGen.Draw(s, fmt.Sprintf("%s[key-%d]", label, len(result)))
			value := g.valueGen.Draw(s, fmt.Sprintf("%s[val-%d]", label, len(result)))
			result[key] = value // Overwrites if key already exists
			attempts++
		}

		// If we couldn't generate enough unique keys, we have what we have
		// This is acceptable behavior - maps naturally deduplicate keys

		// Check filter predicate if present
		if g.filterFn != nil && !g.filterFn(result) {
			continue // Try again
		}

		s.WriteLog(fmt.Sprintf("Map(%s)=[%d entries] ", label, len(result)))
		return result
	}

	// Filter exhausted max attempts - skip this test iteration
	panic(skipTest{})
}
