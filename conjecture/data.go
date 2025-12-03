package conjecture

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
)

// Status represents the outcome of a test execution.
type Status uint8

const (
	StatusValid       Status = iota // Test ran and returned a result
	StatusInvalid                   // Test was filtered out (assume failed)
	StatusOverrun                   // Ran out of data during replay
	StatusInteresting               // Test found a failure
)

func (s Status) String() string {
	switch s {
	case StatusValid:
		return "valid"
	case StatusInvalid:
		return "invalid"
	case StatusOverrun:
		return "overrun"
	case StatusInteresting:
		return "interesting"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

var (
	// ErrOverrun is returned when replaying exhausts the choice sequence.
	ErrOverrun = errors.New("conjecture: choice sequence overrun")

	// ErrInvalidDraw is returned when draw constraints are violated.
	ErrInvalidDraw = errors.New("conjecture: invalid draw constraints")

	// ErrFrozen is returned when attempting to draw on frozen data.
	ErrFrozen = errors.New("conjecture: data is frozen")
)

// DataMode indicates whether we're generating new data or replaying existing data.
type DataMode uint8

const (
	ModeGenerate DataMode = iota // Generate new random choices
	ModeReplay                   // Replay from existing choice sequence
)

// ConjectureData is the core context for test case generation.
// It manages the choice sequence, provides primitive drawing operations,
// and tracks the hierarchical structure of generated data.
type ConjectureData struct {
	// Mode determines whether we're generating or replaying
	mode DataMode

	// The choice sequence being built (generate) or consumed (replay)
	sequence *ChoiceSequence

	// Current position in sequence during replay
	index int

	// Random source for generation
	random *PCGRandom

	// Status after test execution
	status Status

	// Stack of open spans for hierarchical structure
	spanStack []int

	// Whether the data is frozen (no more draws allowed)
	frozen bool

	// Maximum allowed choices (prevents infinite loops)
	maxChoices int

	// Tags for debugging and observability
	tags map[string]any
}

// DataOption configures a ConjectureData instance.
type DataOption func(*ConjectureData)

// WithMaxChoices sets the maximum number of choices allowed.
func WithMaxChoices(n int) DataOption {
	return func(d *ConjectureData) {
		d.maxChoices = n
	}
}

// WithSeed sets the random seed for generation.
func WithSeed(seed uint64) DataOption {
	return func(d *ConjectureData) {
		d.random = NewPCGRandom(seed)
	}
}

// NewConjectureData creates a new data context for generating test cases.
func NewConjectureData(opts ...DataOption) *ConjectureData {
	d := &ConjectureData{
		mode:       ModeGenerate,
		sequence:   NewChoiceSequence(),
		random:     NewPCGRandom(0), // Will be seeded by engine
		status:     StatusValid,
		spanStack:  make([]int, 0, 8),
		maxChoices: 10000, // Reasonable default
		tags:       make(map[string]any),
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// ForReplay creates a data context that replays an existing choice sequence.
func ForReplay(seq *ChoiceSequence, opts ...DataOption) *ConjectureData {
	d := &ConjectureData{
		mode:       ModeReplay,
		sequence:   seq.Clone(),
		index:      0,
		status:     StatusValid,
		spanStack:  make([]int, 0, 8),
		maxChoices: seq.Len() + 1000, // Allow some growth
		tags:       make(map[string]any),
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Sequence returns the current choice sequence.
func (d *ConjectureData) Sequence() *ChoiceSequence {
	return d.sequence
}

// Status returns the current status.
func (d *ConjectureData) Status() Status {
	return d.status
}

// SetStatus sets the status. Used by test execution.
func (d *ConjectureData) SetStatus(s Status) {
	d.status = s
}

// Freeze prevents further draws and finalizes the data.
func (d *ConjectureData) Freeze() {
	d.frozen = true
}

// IsFrozen returns whether the data is frozen.
func (d *ConjectureData) IsFrozen() bool {
	return d.frozen
}

// Tag sets a tag for observability.
func (d *ConjectureData) Tag(key string, value any) {
	d.tags[key] = value
}

// Tags returns all tags.
func (d *ConjectureData) Tags() map[string]any {
	return d.tags
}

// StartSpan begins a labeled span in the choice sequence.
// Returns a handle that must be passed to EndSpan.
func (d *ConjectureData) StartSpan(label string) int {
	spanID := d.sequence.StartSpan(label)
	d.spanStack = append(d.spanStack, spanID)
	return spanID
}

// EndSpan closes the span with the given ID.
// If discard is true, marks the span as discardable during shrinking.
func (d *ConjectureData) EndSpan(spanID int, discard bool) {
	d.sequence.EndSpan(spanID, discard)
	// Pop from stack
	for i := len(d.spanStack) - 1; i >= 0; i-- {
		if d.spanStack[i] == spanID {
			d.spanStack = append(d.spanStack[:i], d.spanStack[i+1:]...)
			break
		}
	}
}

// recordChoice adds a choice to the sequence during generation.
func (d *ConjectureData) recordChoice(c Choice) error {
	if d.frozen {
		return ErrFrozen
	}
	if d.sequence.Len() >= d.maxChoices {
		d.status = StatusOverrun
		return ErrOverrun
	}
	d.sequence.Append(c)
	return nil
}

// replayChoice retrieves the next choice during replay.
func (d *ConjectureData) replayChoice(expectedType ChoiceType) (Choice, error) {
	if d.frozen {
		return Choice{}, ErrFrozen
	}
	if d.index >= d.sequence.Len() {
		d.status = StatusOverrun
		return Choice{}, ErrOverrun
	}

	c := d.sequence.Get(d.index)
	if c.Type != expectedType {
		// Type mismatch during replay - sequence is invalid
		d.status = StatusOverrun
		return Choice{}, fmt.Errorf("type mismatch at index %d: expected %s, got %s",
			d.index, expectedType, c.Type)
	}

	d.index++
	return c, nil
}

// DrawBoolean draws a boolean value.
// p is the probability of true (default 0.5).
func (d *ConjectureData) DrawBoolean(p float64) (bool, error) {
	if p <= 0 {
		return false, nil
	}
	if p >= 1 {
		return true, nil
	}

	constraints := Constraints{Probability: p}

	if d.mode == ModeReplay {
		c, err := d.replayChoice(ChoiceBoolean)
		if err != nil {
			return false, err
		}
		return c.Value.(bool), nil
	}

	// Generate: sample with probability p
	value := d.random.Float64() < p

	err := d.recordChoice(Choice{
		Type:        ChoiceBoolean,
		Value:       value,
		Constraints: constraints,
	})
	return value, err
}

// DrawBooleanForced draws a boolean with a forced value.
// Used for shrinking exploration.
func (d *ConjectureData) DrawBooleanForced(value bool) error {
	return d.recordChoice(Choice{
		Type:        ChoiceBoolean,
		Value:       value,
		Forced:      true,
		Constraints: Constraints{Probability: 1},
	})
}

// IntegerParams configures integer drawing.
type IntegerParams struct {
	Min          int64
	Max          int64
	ShrinkToward int64 // Value to shrink toward (default: Min)
	Weights      []int // Optional weights for biased sampling
}

// DrawInteger draws an integer in [min, max].
// The value shrinks toward shrinkToward (default: min).
func (d *ConjectureData) DrawInteger(params IntegerParams) (int64, error) {
	if params.Max < params.Min {
		return 0, ErrInvalidDraw
	}

	constraints := Constraints{
		MinInt:       params.Min,
		MaxInt:       params.Max,
		ShrinkToward: params.ShrinkToward,
	}

	if d.mode == ModeReplay {
		c, err := d.replayChoice(ChoiceInteger)
		if err != nil {
			return 0, err
		}
		value := c.Value.(int64)
		// Clamp to current constraints (may differ from original)
		if value < params.Min {
			value = params.Min
		}
		if value > params.Max {
			value = params.Max
		}
		return value, nil
	}

	// Generate with bias toward interesting values
	value := d.drawIntegerBiased(params)

	err := d.recordChoice(Choice{
		Type:        ChoiceInteger,
		Value:       value,
		Constraints: constraints,
	})
	return value, err
}

// drawIntegerBiased generates an integer with bias toward interesting values.
func (d *ConjectureData) drawIntegerBiased(params IntegerParams) int64 {
	min, max := params.Min, params.Max

	if min == max {
		return min
	}

	// Calculate range size
	rangeSize := uint64(max - min + 1)

	// Small ranges: uniform is fine
	if rangeSize <= 100 {
		return min + int64(d.random.Uint64n(rangeSize))
	}

	// Larger ranges: bias toward interesting values
	// Use bits to decide sampling strategy
	probe := d.random.Uint64()

	switch {
	case probe < 0x2000000000000000: // ~12.5% - return shrinkToward if in range
		target := params.ShrinkToward
		if target >= min && target <= max {
			return target
		}
		fallthrough

	case probe < 0x4000000000000000: // ~12.5% - small values near min
		offset := d.geometricInt(8) // Small offset
		value := min + offset
		if value <= max {
			return value
		}
		return max

	case probe < 0x6000000000000000: // ~12.5% - boundary values
		boundaries := []int64{min, max, 0, -1, 1}
		for _, b := range boundaries {
			if b >= min && b <= max && d.random.Float64() < 0.3 {
				return b
			}
		}
		fallthrough

	case probe < 0x8000000000000000: // ~12.5% - powers of 2
		// Find a power of 2 in range
		for exp := 0; exp < 64; exp++ {
			pow := int64(1) << exp
			if pow >= min && pow <= max && d.random.Float64() < 0.2 {
				return pow
			}
			if -pow >= min && -pow <= max && d.random.Float64() < 0.2 {
				return -pow
			}
		}
		fallthrough

	default: // ~50% - uniform
		return min + int64(d.random.Uint64n(rangeSize))
	}
}

// geometricInt returns a geometrically distributed integer.
// Used for biasing toward small values.
func (d *ConjectureData) geometricInt(mean float64) int64 {
	// Geometric distribution via inverse CDF
	u := d.random.Float64()
	if u == 0 {
		return 0
	}
	p := 1.0 / (mean + 1)
	return int64(math.Log(u) / math.Log(1-p))
}

// FloatParams configures float drawing.
type FloatParams struct {
	Min           float64
	Max           float64
	AllowNaN      bool
	AllowInfinity bool
}

// DrawFloat draws a float64 value.
func (d *ConjectureData) DrawFloat(params FloatParams) (float64, error) {
	constraints := Constraints{
		MinFloat:      params.Min,
		MaxFloat:      params.Max,
		AllowNaN:      params.AllowNaN,
		AllowInfinity: params.AllowInfinity,
	}

	if d.mode == ModeReplay {
		c, err := d.replayChoice(ChoiceFloat)
		if err != nil {
			return 0, err
		}
		return c.Value.(float64), nil
	}

	value := d.drawFloatBiased(params)

	err := d.recordChoice(Choice{
		Type:        ChoiceFloat,
		Value:       value,
		Constraints: constraints,
	})
	return value, err
}

// drawFloatBiased generates a float with bias toward interesting values.
func (d *ConjectureData) drawFloatBiased(params FloatParams) float64 {
	min, max := params.Min, params.Max

	probe := d.random.Uint64()

	// Occasionally generate special values
	if params.AllowNaN && probe < 0x0800000000000000 { // ~3%
		return math.NaN()
	}
	if params.AllowInfinity && probe < 0x1000000000000000 { // ~3%
		if d.random.Float64() < 0.5 {
			return math.Inf(1)
		}
		return math.Inf(-1)
	}

	// Bias toward zero
	if min <= 0 && max >= 0 && probe < 0x2000000000000000 { // ~12.5%
		return 0.0
	}

	// Bias toward integral values
	if probe < 0x4000000000000000 { // ~12.5%
		intVal := float64(int64(d.random.Float64()*(max-min) + min))
		if intVal >= min && intVal <= max {
			return intVal
		}
	}

	// Bias toward simple fractions
	if probe < 0x5000000000000000 { // ~6%
		denominators := []float64{2, 4, 5, 8, 10}
		denom := denominators[d.random.Uint64n(uint64(len(denominators)))]
		numer := math.Floor(d.random.Float64()*(max-min)*denom+min*denom) / denom
		if numer >= min && numer <= max {
			return numer
		}
	}

	// Uniform in range
	return d.random.Float64()*(max-min) + min
}

// StringParams configures string drawing.
type StringParams struct {
	MinSize   int
	MaxSize   int                 // -1 for unbounded (uses internal limit)
	Intervals []CodepointInterval // Allowed codepoint ranges
}

// DefaultASCIIPrintable returns intervals for printable ASCII.
func DefaultASCIIPrintable() []CodepointInterval {
	return []CodepointInterval{{0x20, 0x7E}}
}

// DefaultUnicode returns intervals for common Unicode (excluding surrogates).
func DefaultUnicode() []CodepointInterval {
	return []CodepointInterval{
		{0x0000, 0xD7FF},
		{0xE000, 0xFFFF},
		{0x10000, 0x10FFFF},
	}
}

// DrawString draws a string value.
func (d *ConjectureData) DrawString(params StringParams) (string, error) {
	maxSize := params.MaxSize
	if maxSize < 0 {
		maxSize = 1000 // Default limit
	}

	intervals := params.Intervals
	if len(intervals) == 0 {
		intervals = DefaultUnicode()
	}

	constraints := Constraints{
		MinSize:   params.MinSize,
		MaxSize:   maxSize,
		Intervals: intervals,
	}

	if d.mode == ModeReplay {
		c, err := d.replayChoice(ChoiceString)
		if err != nil {
			return "", err
		}
		return c.Value.(string), nil
	}

	value := d.drawStringBiased(params.MinSize, maxSize, intervals)

	err := d.recordChoice(Choice{
		Type:        ChoiceString,
		Value:       value,
		Constraints: constraints,
	})
	return value, err
}

// drawStringBiased generates a string with bias toward interesting values.
func (d *ConjectureData) drawStringBiased(minSize, maxSize int, intervals []CodepointInterval) string {
	// Bias toward smaller strings
	var length int
	probe := d.random.Uint64()

	switch {
	case probe < 0x2000000000000000 && minSize == 0: // ~12.5% empty
		length = 0
	case probe < 0x4000000000000000: // ~12.5% minimum
		length = minSize
	case probe < 0x6000000000000000: // ~12.5% small
		length = minSize + int(d.geometricInt(3))
		if length > maxSize {
			length = maxSize
		}
	default: // ~62.5% biased toward smaller
		avgLen := float64(maxSize-minSize) / 4 // Bias toward smaller
		length = minSize + int(d.geometricInt(avgLen))
		if length > maxSize {
			length = maxSize
		}
	}

	// Build the string
	runes := make([]rune, length)
	for i := 0; i < length; i++ {
		runes[i] = d.drawCodepoint(intervals)
	}

	return string(runes)
}

// drawCodepoint draws a single codepoint from the intervals.
func (d *ConjectureData) drawCodepoint(intervals []CodepointInterval) rune {
	// Calculate total codepoints available
	var total int64
	for _, iv := range intervals {
		total += int64(iv.High - iv.Low + 1)
	}

	if total == 0 {
		return '?'
	}

	// Pick a random codepoint
	idx := d.random.Int64n(total)

	// Find which interval it falls in
	for _, iv := range intervals {
		size := int64(iv.High - iv.Low + 1)
		if idx < size {
			return iv.Low + rune(idx)
		}
		idx -= size
	}

	return intervals[0].Low
}

// BytesParams configures bytes drawing.
type BytesParams struct {
	MinSize int
	MaxSize int // -1 for unbounded
}

// DrawBytes draws a byte slice.
func (d *ConjectureData) DrawBytes(params BytesParams) ([]byte, error) {
	maxSize := params.MaxSize
	if maxSize < 0 {
		maxSize = 1000
	}

	constraints := Constraints{
		MinSize: params.MinSize,
		MaxSize: maxSize,
	}

	if d.mode == ModeReplay {
		c, err := d.replayChoice(ChoiceBytes)
		if err != nil {
			return nil, err
		}
		return c.Value.([]byte), nil
	}

	// Bias toward smaller byte slices
	var length int
	probe := d.random.Uint64()

	switch {
	case probe < 0x2000000000000000 && params.MinSize == 0:
		length = 0
	case probe < 0x4000000000000000:
		length = params.MinSize
	default:
		length = params.MinSize + int(d.geometricInt(float64(maxSize-params.MinSize)/4))
		if length > maxSize {
			length = maxSize
		}
	}

	value := make([]byte, length)
	for i := range value {
		// Bias toward smaller byte values (helps shrinking)
		if d.random.Float64() < 0.3 {
			value[i] = byte(d.random.Uint64n(16)) // Small values
		} else {
			value[i] = byte(d.random.Uint64n(256))
		}
	}

	err := d.recordChoice(Choice{
		Type:        ChoiceBytes,
		Value:       value,
		Constraints: constraints,
	})
	return value, err
}

// Assume marks the current test case as invalid if cond is false.
// Returns an error that should cause the test to stop.
func (d *ConjectureData) Assume(cond bool) error {
	if !cond {
		d.status = StatusInvalid
		return errors.New("assumption failed")
	}
	return nil
}

// MarkInteresting marks the current test as having found a failure.
func (d *ConjectureData) MarkInteresting(reason string) {
	d.status = StatusInteresting
	d.Tag("failure_reason", reason)
}

// MarkOverrun marks the current test as having overrun its choice sequence.
// This happens when constraints can't be satisfied (e.g., can't generate enough unique map keys).
func (d *ConjectureData) MarkOverrun() {
	d.status = StatusOverrun
}

// PCGRandom is a simple PCG random number generator.
// Used instead of math/rand for reproducibility across Go versions.
type PCGRandom struct {
	state uint64
	inc   uint64
}

// NewPCGRandom creates a new PCG random generator with the given seed.
func NewPCGRandom(seed uint64) *PCGRandom {
	r := &PCGRandom{}
	r.Seed(seed)
	return r
}

// Seed initializes the generator.
func (r *PCGRandom) Seed(seed uint64) {
	r.state = 0
	r.inc = (seed << 1) | 1
	r.Uint64()
	r.state += seed
	r.Uint64()
}

// Uint64 returns a random uint64.
func (r *PCGRandom) Uint64() uint64 {
	oldstate := r.state
	r.state = oldstate*6364136223846793005 + r.inc
	xorshifted := uint32(((oldstate >> 18) ^ oldstate) >> 27)
	rot := uint32(oldstate >> 59)
	return uint64(bits.RotateLeft32(xorshifted, -int(rot)))<<32 |
		uint64(bits.RotateLeft32(xorshifted, -int(rot+16)))
}

// Uint64n returns a random uint64 in [0, n).
func (r *PCGRandom) Uint64n(n uint64) uint64 {
	if n == 0 {
		return 0
	}
	return r.Uint64() % n
}

// Int64n returns a random int64 in [0, n).
func (r *PCGRandom) Int64n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	return int64(r.Uint64n(uint64(n)))
}

// Float64 returns a random float64 in [0, 1).
func (r *PCGRandom) Float64() float64 {
	return float64(r.Uint64()>>11) / (1 << 53)
}

// Source interface implementation for backwards compatibility with existing Build() signatures.
// This allows ConjectureData to be passed where Source is expected.

// DrawBits draws n random bits as a uint64.
// Implements the Source interface.
func (d *ConjectureData) DrawBits(n int) uint64 {
	if n <= 0 || n > 64 {
		panic("DrawBits: n must be between 1 and 64")
	}
	// Use DrawInteger to maintain consistency with choice tracking
	maxVal := (uint64(1) << n) - 1
	val, err := d.DrawInteger(IntegerParams{Min: 0, Max: int64(maxVal), ShrinkToward: 0})
	if err != nil {
		// Panic with the error - executeGenerator's defer/recover will catch it
		panic(err)
	}
	return uint64(val)
}

// IsDeterministic returns false for ConjectureData (it's for property testing, not deterministic factories).
// Implements the Source interface.
func (d *ConjectureData) IsDeterministic() bool {
	return false
}

// WriteLog stores a log message in the tags for debugging.
// Implements the Source interface.
func (d *ConjectureData) WriteLog(msg string) {
	if d.tags == nil {
		d.tags = make(map[string]any)
	}
	if existing, ok := d.tags["log"]; ok {
		d.tags["log"] = existing.(string) + msg
	} else {
		d.tags["log"] = msg
	}
}

// StartInterval marks the beginning of a labeled interval.
// Implements the Source interface by delegating to StartSpan.
func (d *ConjectureData) StartInterval(label string) {
	d.StartSpan(label)
}

// EndInterval marks the end of the current interval.
// Implements the Source interface by delegating to EndSpan.
func (d *ConjectureData) EndInterval() {
	if len(d.spanStack) > 0 {
		topID := d.spanStack[len(d.spanStack)-1]
		d.EndSpan(topID, false)
	}
}
