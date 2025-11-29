package conjecture

import (
	"fmt"
	"reflect"
)

// DataSource is the interface that provides drawing primitives for generators.
// Both ConjectureData (for property testing) and deterministic adapters implement this.
type DataSource interface {
	DrawInteger(params IntegerParams) (int64, error)
	DrawBoolean(p float64) (bool, error)
	DrawString(params StringParams) (string, error)
	DrawFloat(params FloatParams) (float64, error)
	DrawBytes(params BytesParams) ([]byte, error)
	DrawBits(n int) uint64
	StartSpan(label string) int
	EndSpan(id int, failed bool)
	IsDeterministic() bool
	Tags() map[string]any
	Tag(key string, value any)
	MarkOverrun()
	MarkInteresting(reason string)
}

// Gen[T] is a composable generator that produces values of type T.
// Generators are the primary interface for users to describe the shape of test inputs.
//
// Generators work by drawing from a DataSource context. The same generator
// can produce different values depending on the underlying choice sequence,
// enabling both random generation and deterministic replay/shrinking.
type Gen[T any] interface {
	// Draw produces a value from the given data context.
	Draw(d DataSource) (T, error)

	// String returns a description of the generator.
	String() string
}

// genImpl is the internal implementation of Gen[T] for function-based generators.
type genImpl[T any] struct {
	draw func(d DataSource) (T, error)
	name string
}

// Draw implements Gen[T].Draw
func (g *genImpl[T]) Draw(d DataSource) (T, error) {
	return g.draw(d)
}

// String implements Gen[T].String
func (g *genImpl[T]) String() string {
	if g.name != "" {
		return g.name
	}
	return fmt.Sprintf("Gen[%s]", reflect.TypeOf((*T)(nil)).Elem().String())
}

// newGen creates a generator from a draw function.
func newGen[T any](name string, draw func(d DataSource) (T, error)) Gen[T] {
	return &genImpl[T]{draw: draw, name: name}
}

// =============================================================================
// Primitive Generators
// =============================================================================

// Boolean returns a generator for boolean values.
// By default, true and false are equally likely.
func Boolean() Gen[bool] {
	return BooleanWithP(0.5)
}

// BooleanWithP returns a generator for booleans where true has probability p.
func BooleanWithP(p float64) Gen[bool] {
	return newGen(fmt.Sprintf("Boolean(p=%.2f)", p), func(d DataSource) (bool, error) {
		return d.DrawBoolean(p)
	})
}

// Integer returns a generator for integers in [min, max].
func Integer(min, max int64) Gen[int64] {
	return IntegerWith(IntegerParams{Min: min, Max: max})
}

// IntegerWith returns a generator with full control over integer generation.
func IntegerWith(params IntegerParams) Gen[int64] {
	return newGen(fmt.Sprintf("Integer(%d, %d)", params.Min, params.Max), func(d DataSource) (int64, error) {
		return d.DrawInteger(params)
	})
}

// Int returns a generator for int values (platform-dependent size).
func Int(min, max int) Gen[int] {
	return Map(Integer(int64(min), int64(max)), func(n int64) int {
		return int(n)
	})
}

// UInt64 returns a generator for uint64 values.
func UInt64(min, max uint64) Gen[uint64] {
	return newGen(fmt.Sprintf("UInt64(%d, %d)", min, max), func(d DataSource) (uint64, error) {
		// Handle the uint64 range by using int64 internally
		// This is safe because we track constraints properly
		v, err := d.DrawInteger(IntegerParams{
			Min: 0,
			Max: int64(max - min),
		})
		return uint64(v) + min, err
	})
}

// Float returns a generator for float64 values in [min, max].
func Float(min, max float64) Gen[float64] {
	return FloatWith(FloatParams{Min: min, Max: max})
}

// FloatWith returns a generator with full control over float generation.
func FloatWith(params FloatParams) Gen[float64] {
	return newGen(fmt.Sprintf("Float(%v, %v)", params.Min, params.Max), func(d DataSource) (float64, error) {
		return d.DrawFloat(params)
	})
}

// String returns a generator for strings with default settings.
func String() Gen[string] {
	return StringWith(StringParams{MinSize: 0, MaxSize: 100})
}

// StringWith returns a generator with full control over string generation.
func StringWith(params StringParams) Gen[string] {
	return newGen(fmt.Sprintf("String(min=%d, max=%d)", params.MinSize, params.MaxSize), func(d DataSource) (string, error) {
		return d.DrawString(params)
	})
}

// ASCIIString returns a generator for printable ASCII strings.
func ASCIIString(minSize, maxSize int) Gen[string] {
	return StringWith(StringParams{
		MinSize:   minSize,
		MaxSize:   maxSize,
		Intervals: DefaultASCIIPrintable(),
	})
}

// Bytes returns a generator for byte slices.
func Bytes(minSize, maxSize int) Gen[[]byte] {
	return newGen(fmt.Sprintf("Bytes(%d, %d)", minSize, maxSize), func(d DataSource) ([]byte, error) {
		return d.DrawBytes(BytesParams{MinSize: minSize, MaxSize: maxSize})
	})
}

// =============================================================================
// Combinators
// =============================================================================

// Just returns a generator that always produces the given value.
// Useful as a base case or for constant values.
func Just[T any](value T) Gen[T] {
	return newGen(fmt.Sprintf("Just(%v)", value), func(d DataSource) (T, error) {
		return value, nil
	})
}

// Map transforms the output of a generator.
func Map[T, U any](g Gen[T], f func(T) U) Gen[U] {
	return newGen(fmt.Sprintf("Map(%s)", g.String()), func(d DataSource) (U, error) {
		v, err := g.Draw(d)
		if err != nil {
			var zero U
			return zero, err
		}
		return f(v), nil
	})
}

// MapErr transforms with a fallible function.
func MapErr[T, U any](g Gen[T], f func(T) (U, error)) Gen[U] {
	return newGen(fmt.Sprintf("MapErr(%s)", g.String()), func(d DataSource) (U, error) {
		v, err := g.Draw(d)
		if err != nil {
			var zero U
			return zero, err
		}
		return f(v)
	})
}

// FlatMap chains generators where the second depends on the first's output.
func FlatMap[T, U any](g Gen[T], f func(T) Gen[U]) Gen[U] {
	return newGen(fmt.Sprintf("FlatMap(%s)", g.String()), func(d DataSource) (U, error) {
		v, err := g.Draw(d)
		if err != nil {
			var zero U
			return zero, err
		}
		return f(v).Draw(d)
	})
}

// Filter returns a generator that only produces values satisfying the predicate.
// Warning: If the predicate rejects most values, this will be slow.
func Filter[T any](g Gen[T], pred func(T) bool) Gen[T] {
	return newGen(fmt.Sprintf("Filter(%s)", g.String()), func(d DataSource) (T, error) {
		// Try a few times before giving up
		for attempt := 0; attempt < 100; attempt++ {
			spanID := d.StartSpan("filter_attempt")
			v, err := g.Draw(d)
			if err != nil {
				d.EndSpan(spanID, true)
				return v, err
			}
			if pred(v) {
				d.EndSpan(spanID, false)
				return v, nil
			}
			// Mark this attempt as discardable
			d.EndSpan(spanID, true)
		}
		var zero T
		return zero, fmt.Errorf("filter: gave up after 100 attempts")
	})
}

// OneOf returns a generator that chooses from the given generators.
// Earlier generators are preferred during shrinking.
func OneOf[T any](gens ...Gen[T]) Gen[T] {
	if len(gens) == 0 {
		panic("OneOf requires at least one generator")
	}
	if len(gens) == 1 {
		return gens[0]
	}

	return newGen(fmt.Sprintf("OneOf(%d choices)", len(gens)), func(d DataSource) (T, error) {
		// Draw index with bias toward 0 (earlier = simpler)
		idx, err := d.DrawInteger(IntegerParams{
			Min:          0,
			Max:          int64(len(gens) - 1),
			ShrinkToward: 0,
		})
		if err != nil {
			var zero T
			return zero, err
		}
		return gens[idx].Draw(d)
	})
}

// Frequency chooses generators with specified weights.
// Higher weights mean more likely to be chosen during generation.
type WeightedGen[T any] struct {
	Weight int
	Gen    Gen[T]
}

// Frequency returns a generator that chooses from weighted generators.
func Frequency[T any](weighted ...WeightedGen[T]) Gen[T] {
	if len(weighted) == 0 {
		panic("Frequency requires at least one generator")
	}

	totalWeight := 0
	for _, w := range weighted {
		totalWeight += w.Weight
	}

	return newGen("Frequency", func(d DataSource) (T, error) {
		target, err := d.DrawInteger(IntegerParams{
			Min: 0,
			Max: int64(totalWeight - 1),
		})
		if err != nil {
			var zero T
			return zero, err
		}

		cumulative := int64(0)
		for _, w := range weighted {
			cumulative += int64(w.Weight)
			if target < cumulative {
				return w.Gen.Draw(d)
			}
		}
		// Shouldn't reach here, but fallback to first
		return weighted[0].Gen.Draw(d)
	})
}

// SampledFrom returns a generator that picks from the given values.
func SampledFrom[T any](values ...T) Gen[T] {
	if len(values) == 0 {
		panic("SampledFrom requires at least one value")
	}

	return newGen(fmt.Sprintf("SampledFrom(%d values)", len(values)), func(d DataSource) (T, error) {
		idx, err := d.DrawInteger(IntegerParams{
			Min:          0,
			Max:          int64(len(values) - 1),
			ShrinkToward: 0,
		})
		if err != nil {
			var zero T
			return zero, err
		}
		return values[idx], nil
	})
}

// =============================================================================
// Collection Generators
// =============================================================================

// Slice returns a generator for slices of T.
func Slice[T any](elem Gen[T], minLen, maxLen int) Gen[[]T] {
	return newGen(fmt.Sprintf("Slice(%s, %d, %d)", elem.String(), minLen, maxLen), func(d DataSource) ([]T, error) {
		// Use continuation booleans for shrinkable length
		// This allows element deletion during shrinking
		result := make([]T, 0, minLen)

		// Draw required elements
		for i := 0; i < minLen; i++ {
			v, err := elem.Draw(d)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}

		// Draw optional elements with continuation booleans
		for len(result) < maxLen {
			// Probability decreases as we get longer - biases toward shorter
			p := 0.9 - float64(len(result)-minLen)*0.1
			if p < 0.3 {
				p = 0.3
			}
			cont, err := d.DrawBoolean(p)
			if err != nil {
				return nil, err
			}
			if !cont {
				break
			}

			v, err := elem.Draw(d)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}

		return result, nil
	})
}

// SliceN returns a generator for slices of exactly n elements.
func SliceN[T any](elem Gen[T], n int) Gen[[]T] {
	return Slice(elem, n, n)
}

// MapGen returns a generator for maps.
func MapGen[K comparable, V any](key Gen[K], value Gen[V], minLen, maxLen int) Gen[map[K]V] {
	return newGen(fmt.Sprintf("Map(%s->%s)", key.String(), value.String()), func(d DataSource) (map[K]V, error) {
		result := make(map[K]V)

		// Generate minimum entries
		// Track attempts to detect when we're stuck due to collisions
		const maxCollisionRetries = 100
		collisionCount := 0

		for len(result) < minLen {
			k, err := key.Draw(d)
			if err != nil {
				return nil, err
			}
			v, err := value.Draw(d)
			if err != nil {
				return nil, err
			}

			oldSize := len(result)
			result[k] = v

			// Detect collision
			if len(result) == oldSize {
				collisionCount++
				if collisionCount >= maxCollisionRetries {
					// Too many collisions - mark this test as overrun
					// Return what we have so far; Property will see overrun status and skip
					d.MarkOverrun()
					return result, nil
				}
			} else {
				collisionCount = 0 // Reset on successful unique key
			}
		}

		// Optional additional entries
		for len(result) < maxLen {
			cont, err := d.DrawBoolean(0.7)
			if err != nil {
				return nil, err
			}
			if !cont {
				break
			}

			k, err := key.Draw(d)
			if err != nil {
				return nil, err
			}
			v, err := value.Draw(d)
			if err != nil {
				return nil, err
			}
			result[k] = v
		}

		return result, nil
	})
}

// Optional returns a generator that may or may not produce a value.
// Returns a pointer: nil for no value, *T for a value.
func Optional[T any](g Gen[T]) Gen[*T] {
	return newGen(fmt.Sprintf("Optional(%s)", g.String()), func(d DataSource) (*T, error) {
		present, err := d.DrawBoolean(0.8) // Bias toward present
		if err != nil {
			return nil, err
		}
		if !present {
			return nil, nil
		}
		v, err := g.Draw(d)
		if err != nil {
			return nil, err
		}
		return &v, nil
	})
}

// =============================================================================
// Struct Building
// =============================================================================

// Tuple2 generates a pair of values.
func Tuple2[A, B any](ga Gen[A], gb Gen[B]) Gen[struct {
	A A
	B B
}] {
	type T = struct {
		A A
		B B
	}
	return newGen(fmt.Sprintf("Tuple2(%s, %s)", ga.String(), gb.String()), func(d DataSource) (T, error) {
		a, err := ga.Draw(d)
		if err != nil {
			return T{}, err
		}
		b, err := gb.Draw(d)
		if err != nil {
			return T{}, err
		}
		return T{A: a, B: b}, nil
	})
}

// Tuple3 generates a triple of values.
func Tuple3[A, B, C any](ga Gen[A], gb Gen[B], gc Gen[C]) Gen[struct {
	A A
	B B
	C C
}] {
	type T = struct {
		A A
		B B
		C C
	}
	return newGen(fmt.Sprintf("Tuple3(%s, %s, %s)", ga.String(), gb.String(), gc.String()), func(d DataSource) (T, error) {
		a, err := ga.Draw(d)
		if err != nil {
			return T{}, err
		}
		b, err := gb.Draw(d)
		if err != nil {
			return T{}, err
		}
		c, err := gc.Draw(d)
		if err != nil {
			return T{}, err
		}
		return T{A: a, B: b, C: c}, nil
	})
}

// Build creates a generator that calls a constructor with generated arguments.
// This is the most flexible way to generate structs.
func Build[T any](
	label string,
	constructor func(d DataSource) (T, error),
) Gen[T] {
	return newGen(label, constructor)
}

// =============================================================================
// Recursive Generators
// =============================================================================

// Recursive creates a generator for recursive data structures.
// The base generator produces leaves, and extend wraps the recursive case.
func Recursive[T any](base Gen[T], extend func(Gen[T]) Gen[T], maxDepth int) Gen[T] {
	// Use deferred evaluation to handle recursion
	var rec Gen[T]
	rec = newGen(fmt.Sprintf("Recursive(%s, depth=%d)", base.String(), maxDepth), func(d DataSource) (T, error) {
		// Bias toward base case as we recurse (via shrink ordering)
		useBase, err := d.DrawBoolean(0.3)
		if err != nil {
			var zero T
			return zero, err
		}
		if useBase {
			return base.Draw(d)
		}
		return extend(rec).Draw(d)
	})

	// Wrap with depth limiting
	return newGen(rec.String(), func(d DataSource) (T, error) {
		// Track depth in data tags
		depth, _ := d.Tags()["recursive_depth"].(int)
		if depth >= maxDepth {
			return base.Draw(d)
		}
		d.Tag("recursive_depth", depth+1)
		defer d.Tag("recursive_depth", depth)
		return rec.Draw(d)
	})
}

// Lazy creates a generator that is evaluated lazily.
// Useful for forward references in mutually recursive structures.
func Lazy[T any](makeGen func() Gen[T]) Gen[T] {
	var cached Gen[T]
	return newGen("Lazy", func(d DataSource) (T, error) {
		if cached == nil {
			cached = makeGen()
		}
		return cached.Draw(d)
	})
}
