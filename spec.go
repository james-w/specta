package gomatchers

// Provider returns a value (literal or computed from Primitives).
type Provider[V any] func(p Primitives) V

// Lit turns a literal into a Provider.
func Lit[V any](v V) Provider[V] { return func(Primitives) V { return v } }

// Maybe holds an optional Provider.
type Maybe[V any] struct {
	fn  Provider[V]
	set bool
}

func Some[V any](fn Provider[V]) Maybe[V] { return Maybe[V]{fn: fn, set: true} }

func (m Maybe[V]) Get(p Primitives, def Provider[V]) V {
	if m.set {
		return m.fn(p)
	}
	return def(p)
}

// Opt applies to a spec S (not the final instance).
type Opt[S any] func(*S)

// BuildWithSpec constructs T from Primitives plus the collected spec S.
type BuildWithSpec[T any, S any] func(p Primitives, s S) T

// SpecFactory builds values from specs.
type SpecFactory[T any, S any] struct {
	P       Primitives
	NewSpec func() S
	Build   BuildWithSpec[T, S]
}

func NewSpecFactory[T any, S any](p Primitives, newSpec func() S, build BuildWithSpec[T, S]) *SpecFactory[T, S] {
	return &SpecFactory[T, S]{P: p, NewSpec: newSpec, Build: build}
}

func (f *SpecFactory[T, S]) Make(opts ...Opt[S]) T {
	s := f.NewSpec()
	for _, o := range opts {
		o(&s)
	}
	return f.Build(f.P, s)
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
		out[i] = f.Build(f.P, s)
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
	return func(p Primitives) T {
		s := newSpec()
		for _, o := range opts {
			o(&s)
		}
		return build(p, s)
	}
}
func FromSpecN[T any, S any](n int, build BuildWithSpec[T, S], newSpec func() S, opts ...Opt[S]) Provider[[]T] {
	return func(p Primitives) []T {
		if n <= 0 {
			return nil
		}
		out := make([]T, n)
		for i := range n {
			s := newSpec()
			for _, o := range opts {
				o(&s)
			}
			out[i] = build(p, s)
		}
		return out
	}
}
func FromSpecEach[T any, S any](n int, build BuildWithSpec[T, S], newSpec func() S, edit func(i int, s *S)) Provider[[]T] {
	return func(p Primitives) []T {
		if n <= 0 {
			return nil
		}
		out := make([]T, n)
		for i := range n {
			s := newSpec()
			if edit != nil {
				edit(i, &s)
			}
			out[i] = build(p, s)
		}
		return out
	}
}
func PtrOf[T any](prov Provider[T]) Provider[*T] {
	return func(p Primitives) *T { v := prov(p); return &v }
}
func SliceOf[T any](provs ...Provider[T]) Provider[[]T] {
	return func(p Primitives) []T {
		out := make([]T, len(provs))
		for i, pv := range provs {
			out[i] = pv(p)
		}
		return out
	}
}

