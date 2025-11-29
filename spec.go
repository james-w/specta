package specta

import (
	"errors"
	"fmt"
	"math"

	"github.com/james-w/specta/conjecture"
)

// Generator is an alias for Gen[T] - used in generated code for convenience
type Generator[T any] = Gen[T]

// DataSource is re-exported from conjecture for user convenience.
// Users can write custom generators without importing conjecture package.
type DataSource = conjecture.DataSource

// Build creates a new generator from a constructor function.
// Re-exported from conjecture for user convenience.
func Build[T any](label string, constructor func(d DataSource) (T, error)) Gen[T] {
	return conjecture.Build(label, constructor)
}

// Map transforms a generator's output using a mapping function.
// Re-exported from conjecture for user convenience.
func Map[T, U any](gen Gen[T], fn func(T) U) Gen[U] {
	return conjecture.Map(gen, fn)
}

// Just creates a generator that always returns the same literal value.
// Re-exported from conjecture for user convenience.
func Just[T any](value T) Gen[T] {
	return conjecture.Just(value)
}

// Maybe holds an optional Gen.
type Maybe[V any] struct {
	gen Gen[V]
	set bool
}

// Some creates a Maybe with a generator.
func Some[V any](gen Gen[V]) Maybe[V] { return Maybe[V]{gen: gen, set: true} }

// IsSet returns true if this Maybe has a value.
func (m Maybe[V]) IsSet() bool {
	return m.set
}

// GetValue draws the value from this Maybe's generator.
// Panics if the Maybe is not set. Use IsSet() to check first.
func (m Maybe[V]) GetValue(s Source, label string) V {
	if !m.set {
		panic("GetValue called on unset Maybe")
	}
	return executeGenerator(s, label, m.gen)
}

// GetWithGenerator returns the set value if present, otherwise draws from the default generator.
// This is used by generated factory code to support composable generator defaults.
// The Source parameter can be either *PrimitivesGen (deterministic) or *conjecture.ConjectureData (property testing).
func (m Maybe[V]) GetWithGenerator(s Source, label string, defGen Gen[V]) V {
	// Determine which generator to use
	gen := defGen
	if m.set {
		gen = m.gen // User override
	}

	// Execute based on Source runtime type
	return executeGenerator(s, label, gen)
}

// executeGenerator runs a Gen[V] with the given Source, handling both PrimitivesGen and ConjectureData.
func executeGenerator[V any](s Source, label string, gen Gen[V]) (result V) {
	// Catch panics from generators (e.g., from DrawBits on type mismatch)
	defer func() {
		if r := recover(); r != nil {
			// Re-panic with skipTest for known error types
			if err, ok := r.(error); ok {
				switch {
				case errors.Is(err, conjecture.ErrOverrun):
					panic(skipTest{})
				case errors.Is(err, conjecture.ErrFrozen):
					panic("attempted to draw from frozen data")
				default:
					// Type mismatch or other replay errors - skip this iteration
					panic(skipTest{})
				}
			}
			// Unknown panic type - re-panic as-is
			panic(r)
		}
	}()

	// Path 1: ConjectureData (property testing)
	if cd, ok := s.(*conjecture.ConjectureData); ok {
		value, err := gen.Draw(cd)
		if err != nil {
			// Handle errors the same as Draw() function in generator.go
			switch {
			case errors.Is(err, conjecture.ErrOverrun):
				// During replay/shrinking, ran out of data - skip this iteration
				panic(skipTest{})
			case errors.Is(err, conjecture.ErrInvalidDraw):
				panic(fmt.Sprintf("invalid generator constraints for %s: %v", label, err))
			case errors.Is(err, conjecture.ErrFrozen):
				panic("attempted to draw from frozen data")
			default:
				// Assumption failure or other filtering - skip this iteration
				panic(skipTest{})
			}
		}
		return value
	}

	// Path 2: dataSourceWrapper (wrapped DataSource from Gen() methods)
	if wrapper, ok := s.(*dataSourceWrapper); ok {
		value, err := gen.Draw(wrapper.DataSource)
		if err != nil {
			// Same error handling as Path 1
			switch {
			case errors.Is(err, conjecture.ErrOverrun):
				// During replay/shrinking, ran out of data - skip this iteration
				panic(skipTest{})
			case errors.Is(err, conjecture.ErrInvalidDraw):
				panic(fmt.Sprintf("invalid generator constraints for %s: %v", label, err))
			case errors.Is(err, conjecture.ErrFrozen):
				panic("attempted to draw from frozen data")
			default:
				// Assumption failure or other filtering - skip this iteration
				panic(skipTest{})
			}
		}
		return value
	}

	// Path 3: PrimitivesGen (factory testing) - use adapter
	if pg, ok := s.(*PrimitivesGen); ok {
		adapter := wrapPrimitivesForGen(pg)
		value, err := gen.Draw(adapter)
		if err != nil {
			panic(fmt.Sprintf("Generator %s failed in factory mode: %v", label, err))
		}
		return value
	}

	panic(fmt.Sprintf("Unknown Source type: %T", s))
}

// primitivesGenAdapter wraps PrimitivesGen to provide ConjectureData-like interface.
// This enables deterministic factory generation using the Gen[T] API.
type primitivesGenAdapter struct {
	gen  *PrimitivesGen
	tags map[string]any
}

func wrapPrimitivesForGen(pg *PrimitivesGen) *primitivesGenAdapter {
	return &primitivesGenAdapter{gen: pg, tags: make(map[string]any)}
}

// DrawInteger implements conjecture.ConjectureData interface for deterministic generation.
func (a *primitivesGenAdapter) DrawInteger(params conjecture.IntegerParams) (int64, error) {
	counter := int64(a.gen.Next())
	if params.Min == params.Max {
		return params.Min, nil
	}

	// Handle full int64 range overflow (when Max - Min + 1 == 0 due to overflow)
	if params.Min == math.MinInt64 && params.Max == math.MaxInt64 {
		// Use counter as-is for full range
		return counter, nil
	}

	rangeSize := params.Max - params.Min + 1
	return params.Min + (counter % rangeSize), nil
}

// DrawBoolean implements conjecture.ConjectureData interface for deterministic generation.
func (a *primitivesGenAdapter) DrawBoolean(p float64) (bool, error) {
	return a.gen.Bool(), nil // Deterministic: always false
}

// DrawString implements conjecture.ConjectureData interface for deterministic generation.
func (a *primitivesGenAdapter) DrawString(params conjecture.StringParams) (string, error) {
	counter := a.gen.Next()
	return fmt.Sprintf("str_%d", counter), nil
}

// DrawFloat implements conjecture.ConjectureData interface for deterministic generation.
func (a *primitivesGenAdapter) DrawFloat(params conjecture.FloatParams) (float64, error) {
	return a.gen.Float64(), nil
}

// DrawBytes implements conjecture.ConjectureData interface for deterministic generation.
func (a *primitivesGenAdapter) DrawBytes(params conjecture.BytesParams) ([]byte, error) {
	size := (params.MinSize + params.MaxSize) / 2
	if size == 0 {
		size = params.MinSize
	}
	return a.gen.BytesN(size), nil
}

// StartSpan implements conjecture.ConjectureData interface.
func (a *primitivesGenAdapter) StartSpan(label string) int {
	a.gen.StartInterval(label)
	return 0
}

// EndSpan implements conjecture.ConjectureData interface.
func (a *primitivesGenAdapter) EndSpan(id int, failed bool) {
	a.gen.EndInterval()
}

// IsDeterministic returns true since this adapter provides deterministic generation.
func (a *primitivesGenAdapter) IsDeterministic() bool {
	return true
}

// DrawBits implements Source interface for generators that need raw bits.
func (a *primitivesGenAdapter) DrawBits(n int) uint64 {
	return a.gen.DrawBits(n)
}

// Tags returns the tag map for storing generator state.
func (a *primitivesGenAdapter) Tags() map[string]any {
	return a.tags
}

// Tag sets a tag value.
func (a *primitivesGenAdapter) Tag(key string, value any) {
	a.tags[key] = value
}

// MarkOverrun marks that we've run out of choices (no-op for deterministic generation).
func (a *primitivesGenAdapter) MarkOverrun() {
	// No-op for deterministic generation
}

// MarkInteresting marks an interesting test case (no-op for deterministic generation).
func (a *primitivesGenAdapter) MarkInteresting(reason string) {
	// No-op for deterministic generation
}

// Opt applies to a spec S (not the final instance).
type Opt[S any] func(*S)

// BuildWithSpec constructs T from Source plus the collected spec S.
type BuildWithSpec[T any, S any] func(s Source, spec S) T

// SpecFactory builds values from specs.
type SpecFactory[T any, S any] struct {
	S       Source
	NewSpec func() S
	Build   BuildWithSpec[T, S]
}

func NewSpecFactory[T any, S any](s Source, newSpec func() S, build BuildWithSpec[T, S]) *SpecFactory[T, S] {
	return &SpecFactory[T, S]{S: s, NewSpec: newSpec, Build: build}
}

func (f *SpecFactory[T, S]) Make(opts ...Opt[S]) T {
	s := f.NewSpec()
	for _, o := range opts {
		o(&s)
	}
	return f.Build(f.S, s)
}

func (f *SpecFactory[T, S]) Many(n int, opts ...Opt[S]) []T {
	if n <= 0 {
		return nil
	}
	out := make([]T, n)
	for i := range n {
		s := f.NewSpec()
		for _, o := range opts {
			o(&s)
		}
		out[i] = f.Build(f.S, s)
	}
	return out
}

// Spec option composition helpers.
func SetLit[S any, V any](assign func(*S, Maybe[V]), v V) Opt[S] {
	return func(s *S) { assign(s, Some(conjecture.Just(v))) }
}
func SetWith[S any, V any](assign func(*S, Maybe[V]), gen Gen[V]) Opt[S] {
	return func(s *S) { assign(s, Some(gen)) }
}
func Compose[S any](opts ...Opt[S]) Opt[S] {
	return func(s *S) {
		for _, o := range opts {
			o(s)
		}
	}
}

// FromSpecGen creates a Gen[T] from a BuildWithSpec function and spec options.
// This is used for generating nested custom types.
func FromSpecGen[T any, S any](build BuildWithSpec[T, S], newSpec func() S, opts ...Opt[S]) Gen[T] {
	return conjecture.Build("FromSpec", func(d conjecture.DataSource) (T, error) {
		spec := newSpec()
		for _, opt := range opts {
			opt(&spec)
		}
		// DataSource can be converted to Source for build
		return build(asSource(d), spec), nil
	})
}

// PtrOfGen creates a Gen[*T] from a Gen[T].
// This is used for generating pointers to custom types.
func PtrOfGen[T any](gen Gen[T]) Gen[*T] {
	return conjecture.Build("PtrOf", func(d conjecture.DataSource) (*T, error) {
		val, err := gen.Draw(d)
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
}

// asSource converts a DataSource to a Source for use with Build functions.
func asSource(d conjecture.DataSource) Source {
	// If it's already a Source, return it directly
	if s, ok := d.(Source); ok {
		return s
	}
	// Otherwise, wrap it
	return &dataSourceWrapper{d}
}

// AsSource converts a DataSource to a Source (exported version for generated code).
func AsSource(d conjecture.DataSource) Source {
	return asSource(d)
}

type dataSourceWrapper struct {
	conjecture.DataSource
}

func (w *dataSourceWrapper) StartInterval(label string) {
	w.StartSpan(label)
}

func (w *dataSourceWrapper) EndInterval() {
	w.EndSpan(0, false)
}

func (w *dataSourceWrapper) WriteLog(msg string) {
	// No-op for wrapped sources
}
