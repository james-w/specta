package specta

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/james-w/specta/conjecture"
	"github.com/sanity-io/litter"
)

// Gen is an alias for conjecture.Gen for convenience
type Gen[T any] = conjecture.Gen[T]

// Draw draws a labeled value from a generator and automatically tracks it for error reporting.
// Using labels improves error messages by showing variable names in failures.
//
// Example:
//
//	xs := specta.Draw(t, specta.Slice(specta.Int().Range(0, 100)), "xs")
//
// The generated value will automatically appear in error messages when the property fails.
//
// Error handling:
// - Programming errors (invalid constraints, frozen data) panic with clear messages
// - Filtering errors (overrun during replay, assumption failures) trigger skipTest to skip the iteration
func Draw[V any](t *T, gen Gen[V], label string) V {
	// Create a span for labeled draws
	var spanID int
	if label != "" {
		spanID = t.Data.StartSpan(label)
	}

	value, err := gen.Draw(t.Data)

	if label != "" {
		t.Data.EndSpan(spanID, err != nil)
	}

	// Handle errors internally
	if err != nil {
		switch {
		case errors.Is(err, conjecture.ErrInvalidDraw):
			panic(fmt.Sprintf("invalid generator constraints: %v", err))
		case errors.Is(err, conjecture.ErrFrozen):
			panic("attempted to draw from frozen data")
		case errors.Is(err, conjecture.ErrOverrun):
			// During replay/shrinking, ran out of data - skip this iteration
			panic(skipTest{})
		default:
			// Assumption failure or other filtering - skip this iteration
			panic(skipTest{})
		}
	}

	if t.generatedValues != nil && label != "" {
		// Track the value for error reporting
		t.generatedValues[label] = formatGeneratedValue(value)
		t.generatedValuesOrder = append(t.generatedValuesOrder, label)
	}
	return value
}

// prettyPrinter is configured for clean, readable output in property test error messages
var prettyPrinter = litter.Options{
	StripPackageNames: true, // Remove package names for cleaner output
	HidePrivateFields: false,
	HideZeroValues:    false,
	Compact:           false, // Multiline for readability
	Separator:         " ",
}

// formatGeneratedValue formats a value for display in property test error messages.
// Uses litter for clean multiline formatting with proper handling of circular references.
func formatGeneratedValue(value any) string {
	formatted := strings.TrimSpace(prettyPrinter.Sdump(value))

	// Add indentation to continuation lines for alignment
	// When we print "label = value", multiline values need extra indent
	lines := strings.Split(formatted, "\n")
	if len(lines) <= 1 {
		return formatted
	}

	// First line stays as-is, subsequent lines get 2 spaces of indent
	for i := 1; i < len(lines); i++ {
		lines[i] = "  " + lines[i]
	}

	return strings.Join(lines, "\n")
}

// =============================================================================
// Integer Generators
// =============================================================================

// IntGenerator provides a fluent API for building integer generators
type IntGenerator struct {
	min *int64
	max *int64
}

// Int creates an integer generator with optional constraints
func Int() *IntGenerator {
	return &IntGenerator{}
}

// Range constrains the integer to [min, max]
func (g *IntGenerator) Range(min, max int64) *IntGenerator {
	if min > max {
		panic(fmt.Sprintf("IntGenerator: min (%d) > max (%d)", min, max))
	}
	g.min = &min
	g.max = &max
	return g
}

// Min sets the minimum value (inclusive)
func (g *IntGenerator) Min(min int64) *IntGenerator {
	if g.max != nil && min > *g.max {
		panic(fmt.Sprintf("IntGenerator: min (%d) > existing max (%d)", min, *g.max))
	}
	g.min = &min
	return g
}

// Max sets the maximum value (inclusive)
func (g *IntGenerator) Max(max int64) *IntGenerator {
	if g.min != nil && max < *g.min {
		panic(fmt.Sprintf("IntGenerator: max (%d) < existing min (%d)", max, *g.min))
	}
	g.max = &max
	return g
}

// Positive constrains to positive integers (> 0)
func (g *IntGenerator) Positive() *IntGenerator {
	if g.max != nil && *g.max < 1 {
		panic(fmt.Sprintf("IntGenerator: Positive() conflicts with existing max (%d)", *g.max))
	}
	one := int64(1)
	g.min = &one
	return g
}

// NonNegative constrains to non-negative integers (>= 0)
func (g *IntGenerator) NonNegative() *IntGenerator {
	if g.max != nil && *g.max < 0 {
		panic(fmt.Sprintf("IntGenerator: NonNegative() conflicts with existing max (%d)", *g.max))
	}
	zero := int64(0)
	g.min = &zero
	return g
}

// Negative constrains to negative integers (< 0)
func (g *IntGenerator) Negative() *IntGenerator {
	if g.min != nil && *g.min >= 0 {
		panic(fmt.Sprintf("IntGenerator: Negative() conflicts with existing min (%d)", *g.min))
	}
	negOne := int64(-1)
	g.max = &negOne
	return g
}

// Draw implements Gen[int64]
func (g *IntGenerator) Draw(d conjecture.DataSource) (int64, error) {
	min := int64(-9223372036854775808) // math.MinInt64
	max := int64(9223372036854775807)  // math.MaxInt64
	if g.min != nil {
		min = *g.min
	}
	if g.max != nil {
		max = *g.max
	}

	// Choose shrink target: prefer 0, but use min if 0 is not in range
	shrinkTarget := int64(0)
	if min > 0 {
		shrinkTarget = min
	} else if max < 0 {
		shrinkTarget = max
	}

	return d.DrawInteger(conjecture.IntegerParams{
		Min:          min,
		Max:          max,
		ShrinkToward: shrinkTarget,
	})
}

// String implements Gen[int64]
func (g *IntGenerator) String() string {
	return "IntGenerator"
}

// Filter returns a generator that only produces integers satisfying the predicate
func (g *IntGenerator) Filter(pred func(int64) bool) Gen[int64] {
	return conjecture.Filter(g, pred)
}

// =============================================================================
// Boolean Generators
// =============================================================================

// Bool creates a boolean generator
func Bool() Gen[bool] {
	return conjecture.Boolean()
}

// =============================================================================
// String Generators
// =============================================================================

// StringGenerator provides a fluent API for building string generators
type StringGenerator struct {
	minLen      *int
	maxLen      *int
	charset     stringCharset
	prefix      string
	suffix      string
	exampleHint string // Used for deterministic generation (e.g., "id_" → "id_1", "id_2")
}

type stringCharset int

const (
	charsetAny stringCharset = iota
	charsetASCII
	charsetPrintable
	charsetAlphaNum
	charsetAlpha
)

// String creates a string generator
func String() *StringGenerator {
	return &StringGenerator{
		charset: charsetAny,
	}
}

// MinLen sets the minimum length
func (g *StringGenerator) MinLen(n int) *StringGenerator {
	if g.maxLen != nil && n > *g.maxLen {
		panic(fmt.Sprintf("StringGenerator: MinLen (%d) > existing MaxLen (%d)", n, *g.maxLen))
	}
	g.minLen = &n
	return g
}

// MaxLen sets the maximum length
func (g *StringGenerator) MaxLen(n int) *StringGenerator {
	if g.minLen != nil && n < *g.minLen {
		panic(fmt.Sprintf("StringGenerator: MaxLen (%d) < existing MinLen (%d)", n, *g.minLen))
	}
	affixLen := len(g.prefix) + len(g.suffix)
	if affixLen > n {
		panic(fmt.Sprintf("StringGenerator: MaxLen (%d) < existing prefix (%d) + suffix (%d) = %d",
			n, len(g.prefix), len(g.suffix), affixLen))
	}
	g.maxLen = &n
	return g
}

// Len sets exact length
func (g *StringGenerator) Len(n int) *StringGenerator {
	affixLen := len(g.prefix) + len(g.suffix)
	if affixLen > n {
		panic(fmt.Sprintf("StringGenerator: Len (%d) < existing prefix (%d) + suffix (%d) = %d",
			n, len(g.prefix), len(g.suffix), affixLen))
	}
	g.minLen = &n
	g.maxLen = &n
	return g
}

// NonEmpty ensures string is not empty
func (g *StringGenerator) NonEmpty() *StringGenerator {
	one := 1
	g.minLen = &one
	return g
}

// Prefix adds a prefix to generated strings
func (g *StringGenerator) Prefix(p string) *StringGenerator {
	g.prefix = p
	affixLen := len(g.prefix) + len(g.suffix)
	if g.maxLen != nil && affixLen > *g.maxLen {
		panic(fmt.Sprintf("StringGenerator: prefix (%d) + suffix (%d) = %d > existing MaxLen (%d)",
			len(g.prefix), len(g.suffix), affixLen, *g.maxLen))
	}
	return g
}

// Suffix adds a suffix to generated strings
func (g *StringGenerator) Suffix(s string) *StringGenerator {
	g.suffix = s
	affixLen := len(g.prefix) + len(g.suffix)
	if g.maxLen != nil && affixLen > *g.maxLen {
		panic(fmt.Sprintf("StringGenerator: prefix (%d) + suffix (%d) = %d > existing MaxLen (%d)",
			len(g.prefix), len(g.suffix), affixLen, *g.maxLen))
	}
	return g
}

// ASCII generates only ASCII characters
func (g *StringGenerator) ASCII() *StringGenerator {
	g.charset = charsetASCII
	return g
}

// Printable generates only printable ASCII
func (g *StringGenerator) Printable() *StringGenerator {
	g.charset = charsetPrintable
	return g
}

// AlphaNum generates only alphanumeric characters
func (g *StringGenerator) AlphaNum() *StringGenerator {
	g.charset = charsetAlphaNum
	return g
}

// Alpha generates only alphabetic characters
func (g *StringGenerator) Alpha() *StringGenerator {
	g.charset = charsetAlpha
	return g
}

// ExampleHint sets a hint for deterministic generation.
// When using PrimitivesGen (factory mode), generates strings like "id_1", "id_2" instead of "str_1", "str_2".
// Has no effect in property testing mode (ConjectureData).
func (g *StringGenerator) ExampleHint(hint string) *StringGenerator {
	g.exampleHint = hint
	return g
}

// Draw implements Gen[string]
func (g *StringGenerator) Draw(d conjecture.DataSource) (string, error) {
	// In deterministic mode, merge ExampleHint into prefix
	prefix := g.prefix
	if d.IsDeterministic() && g.exampleHint != "" {
		prefix = g.exampleHint
	}

	// Determine min/max lengths
	minLen := 0
	maxLen := 100
	if g.minLen != nil {
		minLen = *g.minLen
	}
	if g.maxLen != nil {
		maxLen = *g.maxLen
	}

	// If prefix/suffix are present, adjust minLen/maxLen for the base string
	// so the total length (prefix + base + suffix) matches the constraints
	prefixLen := len(prefix)
	suffixLen := len(g.suffix)
	affixLen := prefixLen + suffixLen

	if affixLen > 0 {
		// Adjust lengths to account for prefix/suffix
		minLen -= affixLen
		if minLen < 0 {
			minLen = 0
		}
		maxLen -= affixLen
		if maxLen < 0 {
			maxLen = 0
		}
	}

	// Build codepoint intervals based on charset
	var intervals []conjecture.CodepointInterval
	switch g.charset {
	case charsetASCII:
		intervals = []conjecture.CodepointInterval{{Low: 0, High: 127}}
	case charsetPrintable:
		intervals = []conjecture.CodepointInterval{{Low: 32, High: 126}}
	case charsetAlphaNum:
		intervals = []conjecture.CodepointInterval{
			{Low: '0', High: '9'},
			{Low: 'A', High: 'Z'},
			{Low: 'a', High: 'z'},
		}
	case charsetAlpha:
		intervals = []conjecture.CodepointInterval{
			{Low: 'A', High: 'Z'},
			{Low: 'a', High: 'z'},
		}
	default: // charsetAny
		intervals = []conjecture.CodepointInterval{{Low: 0, High: 127}} // Start with ASCII for now
	}

	baseStr, err := d.DrawString(conjecture.StringParams{
		MinSize:   minLen,
		MaxSize:   maxLen,
		Intervals: intervals,
	})
	if err != nil {
		return "", err
	}

	// Add prefix/suffix if needed
	if prefix != "" || g.suffix != "" {
		return prefix + baseStr + g.suffix, nil
	}

	return baseStr, nil
}

// String implements Gen[string]
func (g *StringGenerator) String() string {
	return "StringGenerator"
}

// Filter returns a generator that only produces strings satisfying the predicate
func (g *StringGenerator) Filter(pred func(string) bool) Gen[string] {
	return conjecture.Filter(g, pred)
}

// =============================================================================
// Float Generators
// =============================================================================

// Float64 creates a float64 generator for [min, max]
func Float64(min, max float64) Gen[float64] {
	return conjecture.Float(min, max)
}

// =============================================================================
// Time Generators
// =============================================================================

// Time creates a time.Time generator
func Time() Gen[time.Time] {
	// Generate as Unix timestamp and convert
	minTime := int64(0)            // 1970-01-01
	maxTime := int64(253402300799) // 9999-12-31

	return conjecture.Map(
		conjecture.Integer(minTime, maxTime),
		func(unix int64) time.Time {
			return time.Unix(unix, 0).UTC()
		},
	)
}

// =============================================================================
// Duration Generators
// =============================================================================

// Duration creates a time.Duration generator
func Duration() Gen[time.Duration] {
	// Generate durations as nanoseconds
	return conjecture.Map(
		conjecture.Integer(0, int64(24*time.Hour)), // 0 to 24 hours
		func(nanos int64) time.Duration {
			return time.Duration(nanos)
		},
	)
}

// =============================================================================
// Bytes Generators
// =============================================================================

// Bytes creates a []byte generator
func Bytes() Gen[[]byte] {
	return conjecture.Bytes(0, 100)
}

// BytesLen creates a []byte generator with specific length
func BytesLen(minLen, maxLen int) Gen[[]byte] {
	return conjecture.Bytes(minLen, maxLen)
}

// =============================================================================
// UUID Generators
// =============================================================================

// UUID creates a UUID string generator
func UUID() Gen[string] {
	// Generate 16 random bytes and format as UUID
	return conjecture.Map(
		conjecture.Bytes(16, 16),
		func(b []byte) string {
			return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
				b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
		},
	)
}

// =============================================================================
// Slice Generators
// =============================================================================

// Slice creates a slice generator
func Slice[T any](elem Gen[T]) *SliceGenerator[T] {
	return &SliceGenerator[T]{
		elem: elem,
	}
}

// SliceGenerator provides fluent API for slice generation
type SliceGenerator[T any] struct {
	elem   Gen[T]
	minLen *int
	maxLen *int
}

// MinLen sets minimum slice length
func (g *SliceGenerator[T]) MinLen(n int) *SliceGenerator[T] {
	g.minLen = &n
	return g
}

// MaxLen sets maximum slice length
func (g *SliceGenerator[T]) MaxLen(n int) *SliceGenerator[T] {
	g.maxLen = &n
	return g
}

// Len sets exact slice length
func (g *SliceGenerator[T]) Len(n int) *SliceGenerator[T] {
	g.minLen = &n
	g.maxLen = &n
	return g
}

// NonEmpty ensures slice has at least one element
func (g *SliceGenerator[T]) NonEmpty() *SliceGenerator[T] {
	one := 1
	g.minLen = &one
	return g
}

// Draw implements Gen[[]T]
func (g *SliceGenerator[T]) Draw(d conjecture.DataSource) ([]T, error) {
	minLen := 0
	maxLen := 100
	if g.minLen != nil {
		minLen = *g.minLen
	}
	if g.maxLen != nil {
		maxLen = *g.maxLen
	}

	return conjecture.Slice(g.elem, minLen, maxLen).Draw(d)
}

// String implements Gen[[]T]
func (g *SliceGenerator[T]) String() string {
	return "SliceGenerator"
}

// Filter returns a generator that only produces slices satisfying the predicate
func (g *SliceGenerator[T]) Filter(pred func([]T) bool) Gen[[]T] {
	return conjecture.Filter(g, pred)
}

// =============================================================================
// Map Generators
// =============================================================================

// MapOf creates a map generator from key and value generators.
func MapOf[K comparable, V any](keyGen Gen[K], valueGen Gen[V]) *MapGenerator[K, V] {
	return &MapGenerator[K, V]{
		keyGen:   keyGen,
		valueGen: valueGen,
	}
}

// MapGenerator provides fluent API for map generation
type MapGenerator[K comparable, V any] struct {
	keyGen   Gen[K]
	valueGen Gen[V]
	minLen   *int
	maxLen   *int
}

// MinLen sets minimum map size
func (g *MapGenerator[K, V]) MinLen(n int) *MapGenerator[K, V] {
	g.minLen = &n
	return g
}

// MaxLen sets maximum map size
func (g *MapGenerator[K, V]) MaxLen(n int) *MapGenerator[K, V] {
	g.maxLen = &n
	return g
}

// NonEmpty ensures map has at least one entry
func (g *MapGenerator[K, V]) NonEmpty() *MapGenerator[K, V] {
	one := 1
	g.minLen = &one
	return g
}

// Draw implements Gen[map[K]V]
func (g *MapGenerator[K, V]) Draw(d conjecture.DataSource) (map[K]V, error) {
	minLen := 0
	maxLen := 100
	if g.minLen != nil {
		minLen = *g.minLen
	}
	if g.maxLen != nil {
		maxLen = *g.maxLen
	}

	return conjecture.MapGen(g.keyGen, g.valueGen, minLen, maxLen).Draw(d)
}

// String implements Gen[map[K]V]
func (g *MapGenerator[K, V]) String() string {
	return "MapGenerator"
}

// Filter returns a generator that only produces maps satisfying the predicate
func (g *MapGenerator[K, V]) Filter(pred func(map[K]V) bool) Gen[map[K]V] {
	return conjecture.Filter(g, pred)
}

// Email returns a generator for email addresses
func Email() Gen[string] {
	return conjecture.Build("Email", func(d conjecture.DataSource) (string, error) {
		user, err := String().ExampleHint("user_").Draw(d)
		if err != nil {
			return "", err
		}
		return user + "@example.com", nil
	})
}

// URL returns a generator for URLs
func URL() Gen[string] {
	return conjecture.Build("URL", func(d conjecture.DataSource) (string, error) {
		path, err := String().ExampleHint("path_").Draw(d)
		if err != nil {
			return "", err
		}
		return "https://example.com/" + path, nil
	})
}

// UUIDString returns a generator for UUID strings (alias for UUID)
func UUIDString() Gen[string] {
	return UUID()
}

// =============================================================================
// Choice Generator (OneOf combinator)
// =============================================================================

// ChoiceGenerator provides fluent API for choosing between alternative generators
type ChoiceGenerator[T any] struct {
	alternatives []Gen[T]
	weights      []int // nil = uniform, non-nil = weighted
}

// Choice creates a generator that chooses uniformly from the given alternatives.
// Earlier alternatives are preferred during shrinking.
//
// Example:
//
//	gen := specta.Choice(
//	    specta.Int().Range(1, 10),
//	    specta.Int().Range(100, 200),
//	)
func Choice[T any](gens ...Gen[T]) *ChoiceGenerator[T] {
	if len(gens) == 0 {
		panic("Choice requires at least one generator")
	}
	// Validate no nil generators
	for i, gen := range gens {
		if gen == nil {
			panic(fmt.Sprintf("Choice: generator at index %d is nil", i))
		}
	}
	return &ChoiceGenerator[T]{alternatives: gens}
}

// Or adds another alternative with equal weight
func (g *ChoiceGenerator[T]) Or(gen Gen[T]) *ChoiceGenerator[T] {
	if gen == nil {
		panic("Choice.Or: generator cannot be nil")
	}
	g.alternatives = append(g.alternatives, gen)
	if g.weights != nil {
		g.weights = append(g.weights, 1) // default weight
	}
	return g
}

// OrWeighted adds an alternative with a specific weight.
// The first call to OrWeighted converts to weighted mode (existing alternatives get weight 1).
//
// Example:
//
//	// 90% small numbers, 10% large numbers
//	gen := specta.Choice(specta.Int().Range(1, 10)).
//	    OrWeighted(specta.Int().Range(100, 200), 1)
func (g *ChoiceGenerator[T]) OrWeighted(gen Gen[T], weight int) *ChoiceGenerator[T] {
	if gen == nil {
		panic("Choice.OrWeighted: generator cannot be nil")
	}
	if weight <= 0 {
		panic(fmt.Sprintf("weight must be positive, got %d", weight))
	}

	// Lazy initialization: first weighted call converts to weighted mode
	if g.weights == nil {
		g.weights = make([]int, len(g.alternatives))
		for i := range g.weights {
			g.weights[i] = 1 // existing alternatives get weight 1
		}
	}

	g.alternatives = append(g.alternatives, gen)
	g.weights = append(g.weights, weight)
	return g
}

// Draw implements Gen[T]
func (g *ChoiceGenerator[T]) Draw(d conjecture.DataSource) (T, error) {
	if g.weights == nil {
		// Uniform: delegate to conjecture.OneOf
		return conjecture.OneOf(g.alternatives...).Draw(d)
	}

	// Weighted: delegate to conjecture.Frequency
	weighted := make([]conjecture.WeightedGen[T], len(g.alternatives))
	for i, gen := range g.alternatives {
		weighted[i] = conjecture.WeightedGen[T]{
			Weight: g.weights[i],
			Gen:    gen,
		}
	}
	return conjecture.Frequency(weighted...).Draw(d)
}

// String implements Gen[T]
func (g *ChoiceGenerator[T]) String() string {
	if g.weights == nil {
		return fmt.Sprintf("Choice(%d alternatives)", len(g.alternatives))
	}
	return fmt.Sprintf("Choice(%d weighted alternatives)", len(g.alternatives))
}

// Filter returns a generator that only produces values satisfying the predicate
func (g *ChoiceGenerator[T]) Filter(pred func(T) bool) Gen[T] {
	return conjecture.Filter(g, pred)
}

// =============================================================================
// Optional Generator (Optional combinator)
// =============================================================================

// Optional returns a generator that produces either a value or nil (pointer with optional probability).
// Returns a pointer: nil for no value, *T for a value.
//
// By default, generates present values 80% of the time (matching conjecture.Optional).
// Pass a custom probability to override:
//
//	specta.Optional(gen)       // 80% present
//	specta.Optional(gen, 0.5)  // 50% present
//	specta.Optional(gen, 0.2)  // 20% present (rare)
//	specta.Optional(gen, 0.95) // 95% present (common)
func Optional[T any](gen Gen[T], presentProbability ...float64) Gen[*T] {
	if gen == nil {
		panic("Optional: generator cannot be nil")
	}
	if len(presentProbability) > 1 {
		panic(fmt.Sprintf("Optional: expected at most 1 probability argument, got %d", len(presentProbability)))
	}
	p := 0.8 // default
	if len(presentProbability) > 0 {
		p = presentProbability[0]
		if p < 0 || p > 1 {
			panic(fmt.Sprintf("presentProbability must be in [0, 1], got %f", p))
		}
	}

	return conjecture.Build(fmt.Sprintf("Optional(p=%.2f)", p), func(d conjecture.DataSource) (*T, error) {
		present, err := d.DrawBoolean(p)
		if err != nil {
			return nil, err
		}
		if !present {
			return nil, nil
		}
		v, err := gen.Draw(d)
		if err != nil {
			return nil, err
		}
		return &v, nil
	})
}

// =============================================================================
// Tuple Generators
// =============================================================================

// Pair generates a tuple of two values.
// Returns a struct with fields A and B.
//
// Example:
//
//	coords := specta.Pair(
//	    specta.Float64(-180, 180),  // longitude
//	    specta.Float64(-90, 90),    // latitude
//	)
//	pair := specta.Draw(pt, coords, "coords")
//	lon := pair.A  // float64
//	lat := pair.B  // float64
func Pair[A, B any](ga Gen[A], gb Gen[B]) Gen[struct {
	A A
	B B
}] {
	if ga == nil || gb == nil {
		panic("Pair: generators cannot be nil")
	}
	return conjecture.Tuple2(ga, gb)
}

// Triple generates a tuple of three values.
// Returns a struct with fields A, B, and C.
//
// Example:
//
//	coords3d := specta.Triple(
//	    specta.Float64(-180, 180),  // longitude
//	    specta.Float64(-90, 90),    // latitude
//	    specta.Float64(0, 10000),   // altitude
//	)
func Triple[A, B, C any](ga Gen[A], gb Gen[B], gc Gen[C]) Gen[struct {
	A A
	B B
	C C
}] {
	if ga == nil || gb == nil || gc == nil {
		panic("Triple: generators cannot be nil")
	}
	return conjecture.Tuple3(ga, gb, gc)
}
