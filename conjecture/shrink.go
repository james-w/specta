package conjecture

import (
	"context"
	"sort"
)

// Shrinker attempts to find a simpler choice sequence that still triggers
// the same failure. It operates by applying a series of transformation passes
// to the choice sequence and checking if the test still fails.
type Shrinker struct {
	// The test function to check
	test func(*ConjectureData) bool // Returns true if still "interesting" (failing)

	// Current best (smallest) failing sequence
	current *ChoiceSequence

	// Statistics
	calls     int
	shrinks   int
	unchanged int // Passes with no improvement

	// Configuration
	maxCalls      int
	maxUnchanged  int
	stopOnPlateau bool
}

// ShrinkResult contains the outcome of shrinking.
type ShrinkResult struct {
	Sequence  *ChoiceSequence
	Calls     int  // Total test calls during shrinking
	Shrinks   int  // Successful shrink steps
	Exhausted bool // True if all passes ran without improvement
}

// ShrinkOption configures the shrinker.
type ShrinkOption func(*Shrinker)

// WithMaxShrinkCalls sets the maximum number of test calls during shrinking.
func WithMaxShrinkCalls(n int) ShrinkOption {
	return func(s *Shrinker) {
		s.maxCalls = n
	}
}

// Shrink attempts to find a simpler failing example.
// The test function should return true if the input is still "interesting"
// (i.e., still triggers the failure we're trying to minimize).
func Shrink(
	ctx context.Context,
	initial *ChoiceSequence,
	test func(*ConjectureData) bool,
	opts ...ShrinkOption,
) ShrinkResult {
	s := &Shrinker{
		test:          test,
		current:       initial.Clone(),
		maxCalls:      10000,
		maxUnchanged:  5,
		stopOnPlateau: true,
	}

	for _, opt := range opts {
		opt(s)
	}

	// Run shrinking passes until no more progress
	for s.unchanged < s.maxUnchanged && s.calls < s.maxCalls {
		select {
		case <-ctx.Done():
			return ShrinkResult{
				Sequence:  s.current,
				Calls:     s.calls,
				Shrinks:   s.shrinks,
				Exhausted: false,
			}
		default:
		}

		improved := s.runAllPasses(ctx)
		if !improved {
			s.unchanged++
		} else {
			s.unchanged = 0
		}
	}

	return ShrinkResult{
		Sequence:  s.current,
		Calls:     s.calls,
		Shrinks:   s.shrinks,
		Exhausted: s.unchanged >= s.maxUnchanged,
	}
}

// trySequence tests a candidate sequence and updates current if it's better.
func (s *Shrinker) trySequence(candidate *ChoiceSequence) bool {
	s.calls++

	// Must be simpler than current
	if candidate.Compare(s.current) >= 0 {
		return false
	}

	// Check if still interesting
	data := ForReplay(candidate)
	if s.test(data) && data.Status() == StatusInteresting {
		s.current = candidate.Clone()
		s.shrinks++
		return true
	}

	return false
}

// runAllPasses runs all shrinking passes once.
// Returns true if any progress was made.
func (s *Shrinker) runAllPasses(ctx context.Context) bool {
	improved := false

	// Order matters: more aggressive passes first, then fine-grained
	passes := []func(context.Context) bool{
		s.passDeleteSpans,
		s.passDeleteChunks,
		s.passZeroChunks,
		s.passMinimizeGroups, // Try minimizing groups of equal values together
		s.passMinimizeIndividual,
		s.passSwapAdjacent,
		s.passRedistribute,
	}

	for _, pass := range passes {
		select {
		case <-ctx.Done():
			return improved
		default:
		}

		if pass(ctx) {
			improved = true
		}
	}

	return improved
}

// passDeleteSpans attempts to delete marked spans.
func (s *Shrinker) passDeleteSpans(ctx context.Context) bool {
	improved := false
	spans := s.current.Spans()

	// Sort by start position descending (delete from end first to preserve indices)
	sort.Slice(spans, func(i, j int) bool {
		return spans[i].Start > spans[j].Start
	})

	for _, span := range spans {
		select {
		case <-ctx.Done():
			return improved
		default:
		}

		if span.End <= span.Start {
			continue
		}

		candidate := s.current.DeleteRange(span.Start, span.End)
		if s.trySequence(candidate) {
			improved = true
			// Restart since indices changed
			return improved
		}
	}

	return improved
}

// passDeleteChunks tries deleting chunks of various sizes.
func (s *Shrinker) passDeleteChunks(ctx context.Context) bool {
	improved := false
	n := s.current.Len()

	// Try different chunk sizes, largest first
	for chunkSize := n / 2; chunkSize >= 1; chunkSize /= 2 {
		// Try deleting from end first (more likely to work)
		for start := n - chunkSize; start >= 0; start -= chunkSize {
			select {
			case <-ctx.Done():
				return improved
			default:
			}

			end := start + chunkSize
			if end > n {
				end = n
			}

			candidate := s.current.DeleteRange(start, end)
			if s.trySequence(candidate) {
				improved = true
				n = s.current.Len()
				// Restart with new length
				break
			}
		}
	}

	// Also try deleting individual elements
	for i := n - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			return improved
		default:
		}

		candidate := s.current.DeleteRange(i, i+1)
		if s.trySequence(candidate) {
			improved = true
		}
	}

	return improved
}

// passZeroChunks replaces chunks with zero values.
func (s *Shrinker) passZeroChunks(ctx context.Context) bool {
	improved := false
	n := s.current.Len()

	// Try zeroing individual choices
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			return improved
		default:
		}

		c := s.current.Get(i)
		zeroChoice := makeZeroChoice(c)
		if zeroChoice.Value == c.Value {
			continue // Already minimal
		}

		candidate := s.current.ReplaceChoice(i, zeroChoice)
		if s.trySequence(candidate) {
			improved = true
		}
	}

	// Try zeroing chunks
	for chunkSize := 2; chunkSize <= n/2; chunkSize *= 2 {
		for start := 0; start+chunkSize <= n; start++ {
			select {
			case <-ctx.Done():
				return improved
			default:
			}

			candidate := s.current.Clone()
			allZero := true
			for i := start; i < start+chunkSize; i++ {
				c := candidate.Get(i)
				zeroChoice := makeZeroChoice(c)
				if zeroChoice.Value != c.Value {
					allZero = false
				}
				candidate.choices[i] = zeroChoice
			}

			if !allZero && s.trySequence(candidate) {
				improved = true
			}
		}
	}

	return improved
}

// makeZeroChoice returns the zero value for a choice's type.
func makeZeroChoice(c Choice) Choice {
	zero := c // Copy constraints
	switch c.Type {
	case ChoiceBoolean:
		zero.Value = false
	case ChoiceInteger:
		// Zero is the shrink target
		target := c.Constraints.ShrinkToward
		if target < c.Constraints.MinInt {
			target = c.Constraints.MinInt
		}
		if target > c.Constraints.MaxInt {
			target = c.Constraints.MaxInt
		}
		zero.Value = target
	case ChoiceFloat:
		if c.Constraints.MinFloat <= 0 && c.Constraints.MaxFloat >= 0 {
			zero.Value = 0.0
		} else {
			zero.Value = c.Constraints.MinFloat
		}
	case ChoiceString:
		if c.Constraints.MinSize == 0 {
			zero.Value = ""
		} else {
			// Minimal string with preferred shrink target (preferably 'a' for readability)
			r := findShrinkTarget(c.Constraints.Intervals)
			zero.Value = string(repeatRune(r, c.Constraints.MinSize))
		}
	case ChoiceBytes:
		zero.Value = make([]byte, c.Constraints.MinSize)
	}
	return zero
}

func repeatRune(r rune, n int) []rune {
	result := make([]rune, n)
	for i := range result {
		result[i] = r
	}
	return result
}

// passMinimizeGroups tries to minimize groups of choices with equal values together.
// This implements Hypothesis's minimize_duplicated_choices pass, which is crucial
// for tests that depend on relationships between values (e.g., tests that only
// fail when x == y).
//
// Example: If a test draws y=3 and ls=[3], and fails when "y in ls", then
// shrinking them independently gets stuck (y can't shrink because 3 must be in ls,
// and ls can't shrink because removing 3 would make y no longer in ls). But
// shrinking both together to y=0, ls=[0] works.
func (s *Shrinker) passMinimizeGroups(ctx context.Context) bool {
	improved := false

	// Group choices by (type, value) - identical values of the same type
	type groupKey struct {
		typ   ChoiceType
		value int64 // For integers; could extend to other types
	}
	groups := make(map[groupKey][]int) // key -> indices

	for i := 0; i < s.current.Len(); i++ {
		c := s.current.Get(i)
		if c.Type != ChoiceInteger {
			continue // Only handle integers for now
		}

		key := groupKey{
			typ:   c.Type,
			value: c.Value.(int64),
		}
		groups[key] = append(groups[key], i)
	}

	// For each group with 2+ identical values, try shrinking together
	for _, indices := range groups {
		if len(indices) < 2 {
			continue // Single value, will be handled by passMinimizeIndividual
		}

		select {
		case <-ctx.Done():
			return improved
		default:
		}

		// Get current value and constraints from first occurrence
		c := s.current.Get(indices[0])
		current := c.Value.(int64)
		target := c.Constraints.ShrinkToward

		// Clamp target to valid range
		if target < c.Constraints.MinInt {
			target = c.Constraints.MinInt
		}
		if target > c.Constraints.MaxInt {
			target = c.Constraints.MaxInt
		}

		if current == target {
			continue // Already at target
		}

		// Binary search toward target, updating ALL occurrences together
		lo, hi := target, current
		if lo > hi {
			lo, hi = hi, lo
		}

		for hi-lo > 1 {
			mid := lo + (hi-lo)/2

			candidate := s.current.Clone()
			for _, idx := range indices {
				candidate.choices[idx].Value = mid
			}

			if s.trySequence(candidate) {
				improved = true
				hi = mid
			} else {
				lo = mid
			}
		}

		// Try target directly as final attempt
		candidate := s.current.Clone()
		for _, idx := range indices {
			candidate.choices[idx].Value = target
		}
		if s.trySequence(candidate) {
			improved = true
		}
	}

	return improved
}

// passMinimizeIndividual uses binary search to minimize individual values.
func (s *Shrinker) passMinimizeIndividual(ctx context.Context) bool {
	improved := false

	for i := 0; i < s.current.Len(); i++ {
		select {
		case <-ctx.Done():
			return improved
		default:
		}

		c := s.current.Get(i)

		switch c.Type {
		case ChoiceInteger:
			if s.minimizeInteger(i, c) {
				improved = true
			}
		case ChoiceString:
			if s.minimizeString(i, c) {
				improved = true
			}
		case ChoiceBytes:
			if s.minimizeBytes(i, c) {
				improved = true
			}
		}
	}

	return improved
}

// minimizeInteger uses binary search to find the smallest working integer.
func (s *Shrinker) minimizeInteger(idx int, c Choice) bool {
	current := c.Value.(int64)
	target := c.Constraints.ShrinkToward

	// Clamp target to valid range
	if target < c.Constraints.MinInt {
		target = c.Constraints.MinInt
	}
	if target > c.Constraints.MaxInt {
		target = c.Constraints.MaxInt
	}

	if current == target {
		return false
	}

	improved := false

	// Binary search toward target
	lo := target
	hi := current

	if lo > hi {
		lo, hi = hi, lo
	}

	for hi-lo > 1 {
		mid := lo + (hi-lo)/2

		candidate := s.current.Clone()
		candidate.choices[idx].Value = mid

		if s.trySequence(candidate) {
			improved = true
			hi = mid
			current = mid
		} else {
			lo = mid
		}
	}

	// Try the target directly
	if current != target {
		candidate := s.current.Clone()
		candidate.choices[idx].Value = target
		if s.trySequence(candidate) {
			improved = true
		}
	}

	return improved
}

// minimizeString tries to make strings shorter and simpler.
func (s *Shrinker) minimizeString(idx int, c Choice) bool {
	current := c.Value.(string)
	improved := false

	// Try shorter lengths
	minLen := c.Constraints.MinSize
	for length := len(current) - 1; length >= minLen; length-- {
		candidate := s.current.Clone()
		candidate.choices[idx].Value = current[:length]
		if s.trySequence(candidate) {
			improved = true
			current = current[:length]
		}
	}

	// Try replacing characters with simpler ones
	runes := []rune(current)
	for i := 0; i < len(runes); i++ {
		// Try the preferred shrink target (preferably 'a' for readability)
		if len(c.Constraints.Intervals) > 0 {
			simpler := findShrinkTarget(c.Constraints.Intervals)
			if runes[i] != simpler {
				newRunes := make([]rune, len(runes))
				copy(newRunes, runes)
				newRunes[i] = simpler
				candidate := s.current.Clone()
				candidate.choices[idx].Value = string(newRunes)
				if s.trySequence(candidate) {
					improved = true
					runes[i] = simpler
				}
			}
		}
	}

	return improved
}

// findShrinkTarget returns the preferred codepoint to shrink towards.
// Prefers 'a' (lowercase letter) for readability if it's in the allowed intervals,
// otherwise returns the minimum codepoint.
func findShrinkTarget(intervals []CodepointInterval) rune {
	targetChar := 'a' // Prefer lowercase 'a' for readability

	// Check if 'a' is in any interval
	for _, interval := range intervals {
		if targetChar >= interval.Low && targetChar <= interval.High {
			return targetChar
		}
	}

	// 'a' not available, use minimum codepoint
	if len(intervals) > 0 {
		return intervals[0].Low
	}

	return 0
}

// minimizeBytes tries to make byte slices shorter and simpler.
func (s *Shrinker) minimizeBytes(idx int, c Choice) bool {
	current := c.Value.([]byte)
	improved := false

	// Try shorter lengths
	minLen := c.Constraints.MinSize
	for length := len(current) - 1; length >= minLen; length-- {
		candidate := s.current.Clone()
		candidate.choices[idx].Value = current[:length]
		if s.trySequence(candidate) {
			improved = true
			current = current[:length]
		}
	}

	// Try zeroing individual bytes
	for i := 0; i < len(current); i++ {
		if current[i] != 0 {
			newBytes := make([]byte, len(current))
			copy(newBytes, current)
			newBytes[i] = 0
			candidate := s.current.Clone()
			candidate.choices[idx].Value = newBytes
			if s.trySequence(candidate) {
				improved = true
				current = newBytes
			}
		}
	}

	return improved
}

// passSwapAdjacent tries swapping adjacent choices if it produces a simpler sequence.
func (s *Shrinker) passSwapAdjacent(ctx context.Context) bool {
	improved := false
	n := s.current.Len()

	for i := 0; i < n-1; i++ {
		select {
		case <-ctx.Done():
			return improved
		default:
		}

		// Only swap if same type
		c1 := s.current.Get(i)
		c2 := s.current.Get(i + 1)
		if c1.Type != c2.Type {
			continue
		}

		// Only swap if it produces a simpler sequence
		if compareChoices(c2, c1) >= 0 {
			continue
		}

		candidate := s.current.Clone()
		candidate.choices[i], candidate.choices[i+1] = candidate.choices[i+1], candidate.choices[i]

		if s.trySequence(candidate) {
			improved = true
		}
	}

	return improved
}

// passRedistribute tries to redistribute values between adjacent choices.
// For example, if we have [10, 5] and shrink toward 0, we might try [5, 10]
// or [7, 8] to see if the total can be preserved while making parts simpler.
func (s *Shrinker) passRedistribute(ctx context.Context) bool {
	improved := false
	n := s.current.Len()

	for i := 0; i < n-1; i++ {
		select {
		case <-ctx.Done():
			return improved
		default:
		}

		c1 := s.current.Get(i)
		c2 := s.current.Get(i + 1)

		// Only redistribute integers with same constraints
		if c1.Type != ChoiceInteger || c2.Type != ChoiceInteger {
			continue
		}
		if c1.Constraints.MinInt != c2.Constraints.MinInt ||
			c1.Constraints.MaxInt != c2.Constraints.MaxInt {
			continue
		}

		v1 := c1.Value.(int64)
		v2 := c2.Value.(int64)
		total := v1 + v2
		target := c1.Constraints.ShrinkToward

		// Try moving value from c1 to c2 to make c1 closer to target
		newV1 := target
		newV2 := total - newV1

		if newV2 >= c2.Constraints.MinInt && newV2 <= c2.Constraints.MaxInt {
			candidate := s.current.Clone()
			candidate.choices[i].Value = newV1
			candidate.choices[i+1].Value = newV2

			if s.trySequence(candidate) {
				improved = true
			}
		}
	}

	return improved
}

// AdaptiveShrink uses adaptive pass selection based on what's working.
func AdaptiveShrink(
	ctx context.Context,
	initial *ChoiceSequence,
	test func(*ConjectureData) bool,
) ShrinkResult {
	// Start with basic shrinking
	result := Shrink(ctx, initial, test)

	// If we have budget remaining, try more aggressive techniques
	if result.Exhausted && result.Calls < 5000 {
		// Try again with higher limits
		result = Shrink(ctx, result.Sequence, test,
			WithMaxShrinkCalls(10000-result.Calls))
	}

	return result
}
