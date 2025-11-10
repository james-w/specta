package specta

import (
	"fmt"
	"math"
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
	BytesN(n int) []byte
	Time() time.Time
	TimeAtOffset(d time.Duration) time.Time
	Duration() time.Duration
	ID() string
	UUID() uuid.UUID
}

type Gen struct {
	ctr      atomic.Uint64
	baseTime time.Time
	step     time.Duration
	prefix   string
}

type Option func(*Gen)

func WithStart(start uint64) Option          { return func(g *Gen) { g.ctr.Store(start) } }
func WithBaseTime(t time.Time) Option        { return func(g *Gen) { g.baseTime = t } }
func WithTimeStep(step time.Duration) Option { return func(g *Gen) { g.step = step } }
func WithPrefix(p string) Option             { return func(g *Gen) { g.prefix = p } }

func New(opts ...Option) *Gen {
	g := &Gen{baseTime: time.Unix(0, 0).UTC(), step: time.Second}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *Gen) Next() uint64 { return g.ctr.Add(1) }
func (g *Gen) Bool() bool   { return false }
func (g *Gen) Int() int {
	n := g.Next()
	if n > uint64(math.MaxInt) {
		n %= uint64(math.MaxInt)
	}
	return int(n)
}
func (g *Gen) IntN(max int) int {
	if max <= 0 {
		return 0
	}
	return int(g.Next() % uint64(max))
}
func (g *Gen) Int64() int64 {
	n := g.Next()
	if n > uint64(math.MaxInt64) {
		n %= uint64(math.MaxInt64)
	}
	return int64(n)
}
func (g *Gen) Uint64() uint64   { return g.Next() }
func (g *Gen) Float64() float64 { n := g.Next(); return float64(n) + 0.123 }
func (g *Gen) String() string   { return g.StringWith("str_") }
func (g *Gen) StringWith(prefix string) string {
	return fmt.Sprintf("%s%s%d", g.prefix, prefix, g.Next())
}
func (g *Gen) BytesN(n int) []byte {
	if n <= 0 {
		return nil
	}
	seq := g.Next()
	b := make([]byte, n)
	for i := range(n) {
		b[i] = byte(seq + uint64(i))
	}
	return b
}
func (g *Gen) Time() time.Time                        { return g.baseTime.Add(time.Duration(g.Next()) * g.step) }
func (g *Gen) TimeAtOffset(d time.Duration) time.Time { return g.baseTime.Add(d) }
func (g *Gen) Duration() time.Duration                { return time.Duration(g.Next()) * g.step }
func (g *Gen) ID() string                             { return fmt.Sprintf("%sid_%d", g.prefix, g.Next()) }

func (g *Gen) UUID() uuid.UUID { return DeterministicUUIDFromInt(g.Next()) }

var ns = uuid.MustParse("00000000-0000-0000-0000-000000000000")

func DeterministicUUIDFromInt(i uint64) uuid.UUID {
    return uuid.NewSHA1(ns, []byte(fmt.Sprintf("%d", i)))
}
