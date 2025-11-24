package specta

// Provider returns a value (literal or computed from Source).
type Provider[V any] func(s Source) V

// Lit turns a literal into a Provider.
func Lit[V any](v V) Provider[V] { return func(Source) V { return v } }

// Maybe holds an optional Provider.
type Maybe[V any] struct {
	fn  Provider[V]
	set bool
}

func Some[V any](fn Provider[V]) Maybe[V] { return Maybe[V]{fn: fn, set: true} }

func (m Maybe[V]) Get(s Source, def Provider[V]) V {
	if m.set {
		return m.fn(s)
	}
	return def(s)
}

// IsSet returns true if this Maybe has a value.
func (m Maybe[V]) IsSet() bool {
	return m.set
}

// Value returns the value from this Maybe, evaluating it with the given Source.
// Panics if the Maybe is not set.
func (m Maybe[V]) Value(s Source) V {
	if !m.set {
		panic("Maybe.Value called on unset Maybe")
	}
	return m.fn(s)
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
	return func(s *S) { assign(s, Some(Lit(v))) }
}
func SetWith[S any, V any](assign func(*S, Maybe[V]), pv Provider[V]) Opt[S] {
	return func(s *S) { assign(s, Some(pv)) }
}
func Compose[S any](opts ...Opt[S]) Opt[S] {
	return func(s *S) {
		for _, o := range opts {
			o(s)
		}
	}
}

// Nesting helpers
func FromSpec[T any, S any](build BuildWithSpec[T, S], newSpec func() S, opts ...Opt[S]) Provider[T] {
	return func(s Source) T {
		spec := newSpec()
		for _, o := range opts {
			o(&spec)
		}
		return build(s, spec)
	}
}
func FromSpecN[T any, S any](n int, build BuildWithSpec[T, S], newSpec func() S, opts ...Opt[S]) Provider[[]T] {
	return func(s Source) []T {
		if n <= 0 {
			return nil
		}
		out := make([]T, n)
		for i := range n {
			spec := newSpec()
			for _, o := range opts {
				o(&spec)
			}
			out[i] = build(s, spec)
		}
		return out
	}
}
func FromSpecEach[T any, S any](n int, build BuildWithSpec[T, S], newSpec func() S, edit func(i int, s *S)) Provider[[]T] {
	return func(s Source) []T {
		if n <= 0 {
			return nil
		}
		out := make([]T, n)
		for i := range n {
			spec := newSpec()
			if edit != nil {
				edit(i, &spec)
			}
			out[i] = build(s, spec)
		}
		return out
	}
}
func PtrOf[T any](prov Provider[T]) Provider[*T] {
	return func(s Source) *T { v := prov(s); return &v }
}
func SliceOf[T any](provs ...Provider[T]) Provider[[]T] {
	return func(s Source) []T {
		out := make([]T, len(provs))
		for i, pv := range provs {
			out[i] = pv(s)
		}
		return out
	}
}
