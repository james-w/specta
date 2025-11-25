package specta

// shrinker.go implements Hypothesis-style multi-pass shrinking for property-based testing.
//
// The shrinking algorithm operates on the byte stream that was used to generate a failing test case.
// By trying simpler byte streams (shorter, with smaller values), we can find minimal counterexamples.
//
// Key concepts:
// - ShrinkBuffer: Mutable buffer for trying shrink modifications
// - Shrinker: Orchestrates multiple shrinking passes
// - Passes: Different strategies for simplifying the byte stream
//   * Adaptive block deletion: Remove chunks of bytes
//   * Minimize bytes: Reduce individual byte values
//   * Minimize duplicates: Reduce repeated byte sequences
//   * Sort/reorder: Try lexicographically smaller orderings
//
// This is based on the Hypothesis shrinking algorithm described in:
// https://hypothesis.readthedocs.io/en/latest/details.html#shrinking

// ShrinkBuffer is a mutable buffer for trying shrinking modifications.
// It wraps a byte slice and provides operations for deletion, minimization, and reordering.
type ShrinkBuffer struct {
	data []byte
}

// NewShrinkBuffer creates a new shrink buffer from the given data.
func NewShrinkBuffer(data []byte) *ShrinkBuffer {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return &ShrinkBuffer{data: dataCopy}
}

// Data returns a copy of the current buffer contents.
func (sb *ShrinkBuffer) Data() []byte {
	result := make([]byte, len(sb.data))
	copy(result, sb.data)
	return result
}

// Len returns the length of the buffer.
func (sb *ShrinkBuffer) Len() int {
	return len(sb.data)
}

// DeleteRange deletes bytes from [start, end).
// Returns true if the deletion was successful, false if indices are invalid.
func (sb *ShrinkBuffer) DeleteRange(start, end int) bool {
	if start < 0 || end > len(sb.data) || start >= end {
		return false
	}

	sb.data = append(sb.data[:start], sb.data[end:]...)
	return true
}

// MinimizeByte attempts to reduce the value at index by delta.
// Returns true if successful, false if the result would be negative or index is invalid.
func (sb *ShrinkBuffer) MinimizeByte(index int, delta byte) bool {
	if index < 0 || index >= len(sb.data) {
		return false
	}

	if sb.data[index] < delta {
		return false
	}

	sb.data[index] -= delta
	return true
}

// SetByte sets the byte at index to value.
// Returns true if successful, false if index is invalid.
func (sb *ShrinkBuffer) SetByte(index int, value byte) bool {
	if index < 0 || index >= len(sb.data) {
		return false
	}

	sb.data[index] = value
	return true
}

// Swap swaps the bytes at indices i and j.
// Returns true if successful, false if indices are invalid.
func (sb *ShrinkBuffer) Swap(i, j int) bool {
	if i < 0 || i >= len(sb.data) || j < 0 || j >= len(sb.data) {
		return false
	}

	sb.data[i], sb.data[j] = sb.data[j], sb.data[i]
	return true
}

// Shrinker orchestrates the shrinking process using multiple passes.
// It maintains the best (shortest/simplest) failing example found so far.
type Shrinker struct {
	// current is the best failing example found so far
	current []byte

	// intervals tracks the structure of the original failing case
	// (used for structure-aware passes like deleteValue, sort, etc.)
	intervals []Interval

	// testFunc runs the property test on a byte stream and returns true if it fails
	testFunc func([]byte) bool

	// calls tracks the number of test executions (for reporting and limits)
	calls int

	// maxCalls is the maximum number of shrink attempts before giving up
	maxCalls int
}

// NewShrinker creates a new shrinker for the given failing test case.
// testFunc should return true if the test still fails for the given byte stream.
func NewShrinker(data []byte, intervals []Interval, testFunc func([]byte) bool) *Shrinker {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	intervalsCopy := make([]Interval, len(intervals))
	copy(intervalsCopy, intervals)

	return &Shrinker{
		current:   dataCopy,
		intervals: intervalsCopy,
		testFunc:  testFunc,
		calls:     0,
		maxCalls:  10000, // Default limit, can be made configurable
	}
}

// Current returns the best failing example found so far.
func (s *Shrinker) Current() []byte {
	result := make([]byte, len(s.current))
	copy(result, s.current)
	return result
}

// Calls returns the number of test executions performed during shrinking.
func (s *Shrinker) Calls() int {
	return s.calls
}

// test attempts a candidate shrink.
// If the test still fails, it becomes the new current best.
// Returns true if the candidate failed (and was accepted as better).
func (s *Shrinker) test(candidate []byte) bool {
	if s.calls >= s.maxCalls {
		return false
	}

	s.calls++

	// Test must still fail
	if !s.testFunc(candidate) {
		return false
	}

	// Shorter is always better
	// For same length, lexicographically smaller is better
	if len(candidate) < len(s.current) {
		s.current = candidate
		return true
	}

	if len(candidate) == len(s.current) {
		// Check if lexicographically smaller
		for i := 0; i < len(candidate); i++ {
			if candidate[i] < s.current[i] {
				s.current = candidate
				return true
			}
			if candidate[i] > s.current[i] {
				return false
			}
		}
	}

	return false
}

// Shrink runs all shrinking passes and returns the minimized failing example.
// This is the main entry point for shrinking.
//
// It orchestrates multiple passes, running them in rounds until a fixed point is reached
// (no more shrinking possible) or the call budget is exhausted.
func (s *Shrinker) Shrink() []byte {
	// Phase 5: Multi-round orchestration
	// Run passes in rounds until we reach a fixed point
	maxRounds := 10 // Limit rounds to prevent infinite loops
	round := 0

	for round < maxRounds {
		previousSize := len(s.current)
		previousLex := make([]byte, len(s.current))
		copy(previousLex, s.current)

		// Round 1: Aggressive deletion
		// Try to remove as much as possible first
		s.passAdaptiveDeleteBlocks()

		// Round 2: Minimize values
		// Make remaining bytes as small as possible
		s.passMinimizeIndividualBytes()
		s.passMinimizeDuplicates()

		// Round 3: Structural optimization
		// Reorder to be lexicographically smaller
		s.passSortPairs()

		// Check if we made progress
		madeProgress := len(s.current) < previousSize ||
			isLexicographicallySmaller(s.current, previousLex)

		if !madeProgress {
			// Reached fixed point, no more shrinking possible
			break
		}

		round++

		// Check call budget
		if s.calls >= s.maxCalls {
			break
		}
	}

	return s.Current()
}

// passAdaptiveDeleteBlocks attempts to delete chunks of bytes from the buffer.
// It uses adaptive delta debugging: tries to delete increasingly large chunks,
// and when a deletion works, does a binary search in that region to minimize.
//
// This is Hypothesis's most effective pass - it can quickly reduce large failing
// examples by removing irrelevant data.
func (s *Shrinker) passAdaptiveDeleteBlocks() {
	// Try increasingly large block sizes: 8, 7, 6, ..., 2, 1
	// Larger blocks first because they make more progress
	for blockSize := 8; blockSize >= 1; blockSize-- {
		s.deleteBlocksOfSize(blockSize)
	}
}

// deleteBlocksOfSize tries to delete all non-overlapping blocks of the given size.
// When a deletion succeeds, it does a binary search to minimize the deletion.
func (s *Shrinker) deleteBlocksOfSize(blockSize int) {
	i := 0
	for i+blockSize <= len(s.current) {
		// Try deleting bytes [i, i+blockSize)
		buf := NewShrinkBuffer(s.current)
		buf.DeleteRange(i, i+blockSize)

		if s.test(buf.Data()) {
			// Deletion worked! The block was unnecessary.
			// Now binary search within [i, i+blockSize) to find the minimal deletion.
			s.minimizeDeletion(i, blockSize)

			// Don't advance i - there might be more deletions at this position
			// (since current has changed)
		} else {
			// Deletion didn't work, move to next block
			i += blockSize
		}

		// Respect the call limit
		if s.calls >= s.maxCalls {
			return
		}
	}
}

// minimizeDeletion does a binary search to find the smallest contiguous deletion
// starting at position start that still causes the test to fail.
//
// This is the "adaptive" part of adaptive delta debugging - when we successfully
// delete a block, we try to delete smaller sub-blocks to be more precise.
func (s *Shrinker) minimizeDeletion(start, initialSize int) {
	// We know deleting [start, start+initialSize) works
	// Try to find a smaller deletion within that range

	// Binary search on the size
	left, right := 1, initialSize

	for left < right {
		mid := (left + right) / 2

		buf := NewShrinkBuffer(s.current)
		buf.DeleteRange(start, start+mid)

		if s.test(buf.Data()) {
			// Smaller deletion works, try even smaller
			right = mid
		} else {
			// Need a larger deletion
			left = mid + 1
		}

		if s.calls >= s.maxCalls {
			return
		}
	}

	// Apply the minimal deletion
	buf := NewShrinkBuffer(s.current)
	buf.DeleteRange(start, start+left)
	s.test(buf.Data()) // This should succeed and update s.current
}

// passMinimizeIndividualBytes tries to reduce each byte value toward 0.
// This helps find simpler failing examples (e.g., reducing [100] to [1] if both fail).
//
// Uses binary search for efficiency: if we can reduce a byte by 128, try reducing by 64, etc.
func (s *Shrinker) passMinimizeIndividualBytes() {
	// Try to minimize each byte position
	for i := 0; i < len(s.current); i++ {
		s.minimizeByteAt(i)

		if s.calls >= s.maxCalls {
			return
		}
	}
}

// minimizeByteAt tries to reduce the byte at position i using a sparse sampling strategy.
// It tries a small set of candidate values and returns as soon as any improvement is found.
// Multiple passes allow the algorithm to converge to minimal values.
//
// Strategy:
// - For small values (≤16): try all values 0 to n-1 (fast anyway)
// - For larger values: try small values (0-4) and values near current (n-1, n-2, n-4, n-8, n-16, n-32)
// - Exit immediately on first improvement (other passes will continue minimizing)
func (s *Shrinker) minimizeByteAt(i int) {
	current := s.current[i]
	if current == 0 {
		return
	}

	// For small values, just try all of them (fast)
	if current <= 16 {
		for v := uint8(0); v < current; v++ {
			buf := NewShrinkBuffer(s.current)
			buf.SetByte(i, v)
			if s.test(buf.Data()) {
				// Found improvement, let next pass continue
				return
			}
			if s.calls >= s.maxCalls {
				return
			}
		}
		return
	}

	// For larger values, use sparse sampling
	// Try small absolute values first (most likely to be minimal)
	for v := uint8(0); v <= 4 && v < current; v++ {
		buf := NewShrinkBuffer(s.current)
		buf.SetByte(i, v)
		if s.test(buf.Data()) {
			return
		}
		if s.calls >= s.maxCalls {
			return
		}
	}

	// Try values distributed through the range, smallest first
	// This provides good coverage while keeping tests bounded
	// We test in order from smallest to largest to find minimal values first
	candidates := []uint8{
		current / 4,     // Quarter point
		current / 2,     // Midpoint
		current * 3 / 4, // Three-quarter point
		current - 64,
		current - 32,
		current - 16,
		current - 8,
		current - 4,
		current - 2,
		current - 1,
	}

	for _, v := range candidates {
		if v >= current || v <= 4 {
			// Skip if out of range or already tried in small values
			continue
		}
		buf := NewShrinkBuffer(s.current)
		buf.SetByte(i, v)
		if s.test(buf.Data()) {
			return
		}
		if s.calls >= s.maxCalls {
			return
		}
	}
}

// passMinimizeDuplicates looks for runs of identical bytes and tries to reduce their values.
// For example, [10, 10, 10] might shrink to [5, 5, 5] or [0, 0, 0].
//
// This is particularly effective for generated collections where the same value appears multiple times.
func (s *Shrinker) passMinimizeDuplicates() {
	// Find runs of duplicate bytes
	i := 0
	for i < len(s.current) {
		// Find the length of the run starting at i
		runStart := i
		runValue := s.current[i]
		runLength := 1

		for i+runLength < len(s.current) && s.current[i+runLength] == runValue {
			runLength++
		}

		// If we have a run of length >= 2, try to minimize the value
		if runLength >= 2 {
			s.minimizeDuplicateRun(runStart, runLength, runValue)
		}

		i += runLength

		if s.calls >= s.maxCalls {
			return
		}
	}
}

// minimizeDuplicateRun tries to reduce all bytes in a run of duplicates.
// This is more efficient than minimizing them individually because we can
// reduce all instances at once.
func (s *Shrinker) minimizeDuplicateRun(start, length int, currentValue byte) {
	if currentValue == 0 {
		return
	}

	// Binary search for the smallest value
	buf := NewShrinkBuffer(s.current)
	for i := start; i < start+length; i++ {
		buf.SetByte(i, 0)
	}

	if s.test(buf.Data()) {
		// Can set all to 0
		return
	}

	// Binary search between 1 and currentValue
	left, right := uint8(1), currentValue

	for left < right {
		mid := left + (right-left)/2

		buf := NewShrinkBuffer(s.current)
		for i := start; i < start+length; i++ {
			buf.SetByte(i, mid)
		}

		if s.test(buf.Data()) {
			// mid works, try smaller
			right = mid
		} else {
			// mid doesn't work, need larger
			left = mid + 1
		}

		if s.calls >= s.maxCalls {
			return
		}
	}

	// Apply the minimal value to all bytes in the run
	buf = NewShrinkBuffer(s.current)
	for i := start; i < start+length; i++ {
		buf.SetByte(i, left)
	}
	s.test(buf.Data())
}

// passSortPairs tries to sort adjacent intervals to be lexicographically smaller.
// This is particularly effective for collections (lists, sets) where order doesn't matter
// but we want to find the simplest counterexample.
//
// For example, if a test fails with [5, 3, 7], sorting to [3, 5, 7] produces a simpler example.
func (s *Shrinker) passSortPairs() {
	// If we have no interval information, skip this pass
	if len(s.intervals) == 0 {
		return
	}

	// Try to sort each adjacent pair of intervals
	for i := 0; i < len(s.intervals)-1; i++ {
		s.sortIntervalPair(i, i+1)

		if s.calls >= s.maxCalls {
			return
		}
	}
}

// sortIntervalPair tries to swap two adjacent intervals if it makes the result
// lexicographically smaller.
func (s *Shrinker) sortIntervalPair(i, j int) {
	// Get the two intervals (they should be adjacent in the original data)
	interval1 := s.intervals[i]
	interval2 := s.intervals[j]

	// Extract the byte ranges for each interval
	// Note: After previous shrinking passes, the intervals may no longer match
	// the current data length. We need to be careful here.

	// For now, we'll skip structural passes if intervals don't match current data
	// This is a limitation we can improve in Phase 5
	if interval1.End > len(s.current) || interval2.End > len(s.current) {
		return
	}

	// Get the byte sequences for each interval
	bytes1 := s.current[interval1.Start:interval1.End]
	bytes2 := s.current[interval2.Start:interval2.End]

	// Check if swapping would make it lexicographically smaller
	// Only swap if bytes2 < bytes1 (this would make the overall sequence smaller)
	if !isLexicographicallySmaller(bytes2, bytes1) {
		return
	}

	// Try swapping the intervals
	buf := NewShrinkBuffer(s.current)

	// Copy bytes2 to position of bytes1
	for k := 0; k < len(bytes2); k++ {
		buf.SetByte(interval1.Start+k, bytes2[k])
	}

	// Copy bytes1 to position of bytes2
	for k := 0; k < len(bytes1); k++ {
		buf.SetByte(interval2.Start+k, bytes1[k])
	}

	s.test(buf.Data())
}

// isLexicographicallySmaller returns true if a is lexicographically smaller than b.
func isLexicographicallySmaller(a, b []byte) bool {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}

	// If all compared bytes are equal, shorter is smaller
	return len(a) < len(b)
}
