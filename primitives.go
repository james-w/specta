package specta

import (
	"fmt"
	"math"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Primitives interface {
	Next() uint64
	Bool() bool
	Int() int
	IntN(max int) int
	Int64() int64
	Uint64() uint64
	Float64() float64
	String() string
	StringWith(prefix string) string
	Bytes() []byte
	BytesN(n int) []byte
	Time() time.Time
	TimeAtOffset(d time.Duration) time.Time
	Duration() time.Duration
	ID() string
	UUID() uuid.UUID
}

// PrimitivesGen is a deterministic generator for factory test data
type PrimitivesGen struct {
	ctr      atomic.Uint64
	baseTime time.Time
	step     time.Duration
	prefix   string
	logging  bool
	log      strings.Builder
}

type PrimitivesOption func(*PrimitivesGen)

func WithStart(start uint64) PrimitivesOption   { return func(g *PrimitivesGen) { g.ctr.Store(start) } }
func WithBaseTime(t time.Time) PrimitivesOption { return func(g *PrimitivesGen) { g.baseTime = t } }
func WithTimeStep(step time.Duration) PrimitivesOption {
	return func(g *PrimitivesGen) { g.step = step }
}
func WithPrefix(p string) PrimitivesOption { return func(g *PrimitivesGen) { g.prefix = p } }

func New(opts ...PrimitivesOption) *PrimitivesGen {
	g := &PrimitivesGen{baseTime: time.Unix(0, 0).UTC(), step: time.Second}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *PrimitivesGen) Next() uint64 { return g.ctr.Add(1) }

// Bool always returns false for deterministic, minimal test data generation.
// This is intentional - Gen is for factories, not property testing.
// For random booleans, use PropertyPrimitives.
func (g *PrimitivesGen) Bool() bool { return false }
func (g *PrimitivesGen) Int() int {
	n := g.Next()
	if n > uint64(math.MaxInt) {
		n %= uint64(math.MaxInt)
	}
	return int(n)
}
func (g *PrimitivesGen) IntN(max int) int {
	if max <= 0 {
		return 0
	}
	return int(g.Next() % uint64(max))
}
func (g *PrimitivesGen) Int64() int64 {
	n := g.Next()
	if n > uint64(math.MaxInt64) {
		n %= uint64(math.MaxInt64)
	}
	return int64(n)
}
func (g *PrimitivesGen) Uint64() uint64   { return g.Next() }
func (g *PrimitivesGen) Float64() float64 { n := g.Next(); return float64(n) + 0.123 }
func (g *PrimitivesGen) String() string   { return g.StringWith("str_") }
func (g *PrimitivesGen) StringWith(prefix string) string {
	return fmt.Sprintf("%s%s%d", g.prefix, prefix, g.Next())
}
func (g *PrimitivesGen) Bytes() []byte {
	// Default to 16 bytes for deterministic generation
	return g.BytesN(16)
}
func (g *PrimitivesGen) BytesN(n int) []byte {
	if n <= 0 {
		return nil
	}
	seq := g.Next()
	b := make([]byte, n)
	for i := range n {
		b[i] = byte(seq + uint64(i))
	}
	return b
}
func (g *PrimitivesGen) Time() time.Time                        { return g.baseTime.Add(time.Duration(g.Next()) * g.step) }
func (g *PrimitivesGen) TimeAtOffset(d time.Duration) time.Time { return g.baseTime.Add(d) }
func (g *PrimitivesGen) Duration() time.Duration                { return time.Duration(g.Next()) * g.step }
func (g *PrimitivesGen) ID() string                             { return fmt.Sprintf("%sid_%d", g.prefix, g.Next()) }

func (g *PrimitivesGen) UUID() uuid.UUID { return DeterministicUUIDFromInt(g.Next()) }

// DrawBits implements Source interface.
// Returns bits from the counter for deterministic, predictable generation.
func (g *PrimitivesGen) DrawBits(n int) uint64 {
	if n <= 0 || n > 64 {
		panic("DrawBits: n must be between 1 and 64")
	}
	value := g.Next()
	mask := uint64((1 << n) - 1)
	return value & mask
}

// WriteLog implements Source interface.
// Appends to the log if logging is enabled.
func (g *PrimitivesGen) WriteLog(msg string) {
	if g.logging {
		g.log.WriteString(msg)
	}
}

// IsDeterministic implements Source interface.
// Returns true because Gen produces predictable, friendly values.
func (g *PrimitivesGen) IsDeterministic() bool {
	return true
}

// StartInterval implements Source interface.
// No-op for Gen since it doesn't track intervals (only used for property testing).
func (g *PrimitivesGen) StartInterval(label string) {}

// EndInterval implements Source interface.
// No-op for Gen since it doesn't track intervals (only used for property testing).
func (g *PrimitivesGen) EndInterval() {}

var ns = uuid.MustParse("00000000-0000-0000-0000-000000000000")

func DeterministicUUIDFromInt(i uint64) uuid.UUID {
	return uuid.NewSHA1(ns, []byte(fmt.Sprintf("%d", i)))
}
