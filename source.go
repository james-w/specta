package specta

import (
	"math/rand/v2"
	"strings"
)

// Source provides data for generators.
// It has two modes:
//   - Deterministic (Gen): counter-based, produces friendly values like "user_1"
//   - Random (randomSource): PRNG-based, explores full type space for property testing
type Source interface {
	// DrawBits generates n random bits as a uint64.
	DrawBits(n int) uint64

	// WriteLog appends a message to the log if logging is enabled.
	WriteLog(msg string)

	// IsDeterministic returns true if this source generates predictable,
	// friendly values (like Gen). Returns false for random property testing.
	IsDeterministic() bool
}

// randomSource provides random data for property testing with integrated shrinking support.
// It tracks the byte stream used for generation, enabling automatic shrinking by
// trying simpler byte streams.
type randomSource struct {
	rng     *rand.Rand
	data    []byte
	index   int
	logging bool
	log     strings.Builder
}

// NewSource creates a new randomSource with the given seed for deterministic generation.
func NewSource(seed int64) *randomSource {
	return &randomSource{
		rng:  rand.New(rand.NewPCG(uint64(seed), uint64(seed^0x123456789abcdef))),
		data: make([]byte, 0, 1024),
	}
}

// NewSourceFromData creates a randomSource that replays from the given byte stream.
// This is used during shrinking and reproduction.
func NewSourceFromData(data []byte) *randomSource {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return &randomSource{
		data:  dataCopy,
		index: 0,
	}
}

// DrawBits generates n random bits as a uint64.
// The bits are drawn from the random number generator and appended to the data stream.
// During replay (when created with NewSourceFromData), bits are read from the existing data.
func (s *randomSource) DrawBits(n int) uint64 {
	if n <= 0 || n > 64 {
		panic("DrawBits: n must be between 1 and 64")
	}

	// If we're replaying from existing data, read from it
	if s.rng == nil {
		// Read bytes from existing data stream
		bytesNeeded := (n + 7) / 8
		if s.index+bytesNeeded > len(s.data) {
			// If we run out of data during replay, return 0
			// This can happen when shrinking reduces the byte stream
			return 0
		}

		var result uint64
		for i := 0; i < bytesNeeded && s.index < len(s.data); i++ {
			result = (result << 8) | uint64(s.data[s.index])
			s.index++
		}

		// Mask to requested number of bits
		mask := uint64((1 << n) - 1)
		return result & mask
	}

	// Generate new random bits
	var result uint64
	bytesNeeded := (n + 7) / 8

	for i := 0; i < bytesNeeded; i++ {
		b := byte(s.rng.Uint32() & 0xFF)
		s.data = append(s.data, b)
		result = (result << 8) | uint64(b)
	}

	// Mask to requested number of bits
	mask := uint64((1 << n) - 1)
	return result & mask
}

// IsDeterministic returns false for randomSource (it's for property testing).
func (s *randomSource) IsDeterministic() bool {
	return false
}

// Data returns the complete byte stream generated so far.
// This is used for shrinking.
func (s *randomSource) Data() []byte {
	result := make([]byte, len(s.data))
	copy(result, s.data)
	return result
}

// EnableLogging turns on logging of generated values.
// This is used during the final replay to show what values were generated.
func (s *randomSource) EnableLogging() {
	s.logging = true
	s.log.Reset()
}

// Log returns the logged output from generators.
func (s *randomSource) Log() string {
	return s.log.String()
}

// WriteLog appends to the log if logging is enabled.
// This is called by generators to record what they produced.
func (s *randomSource) WriteLog(msg string) {
	if s.logging {
		s.log.WriteString(msg)
	}
}
