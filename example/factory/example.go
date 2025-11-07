package factory

import (
	"github.com/james-w/gomatchers"
	"github.com/james-w/gomatchers/example"
)

type UserSpec struct {
	ID     gomatchers.Maybe[string]
	Name   gomatchers.Maybe[string]
	Active gomatchers.Maybe[bool]
	Score  gomatchers.Maybe[int]
}

func NewUserSpec() UserSpec { return UserSpec{} }


func NewUserFactory(p gomatchers.Primitives) *gomatchers.SpecFactory[example.UserView, UserSpec] {
	return gomatchers.NewSpecFactory(p, NewUserSpec, BuildUserView)
}

var (
	defUserID   = func(p gomatchers.Primitives) string { return p.ID() }
	defUserName = func(p gomatchers.Primitives) string { return p.StringWith("name_") }
	defActive   = func(p gomatchers.Primitives) bool   { return p.Bool() }
	defScore    = func(p gomatchers.Primitives) int    { return p.IntN(1000) }
)

func BuildUserView(p gomatchers.Primitives, s UserSpec) example.UserView {
	id := s.ID.Get(p, defUserID)
	name := s.Name.Get(p, defUserName)
	active := s.Active.Get(p, defActive)
	score := s.Score.Get(p, defScore)

	return example.UserView{
		ID:     id,
		Name:   name,
		Active: active,
		Score:  score,
	}
}

func WithName(v string) gomatchers.Opt[UserSpec] {
	return gomatchers.SetLit(func(s *UserSpec, m gomatchers.Maybe[string]) { s.Name = m }, v)
}

func WithActive(v bool) gomatchers.Opt[UserSpec] {
	return gomatchers.SetLit(func(s *UserSpec, m gomatchers.Maybe[bool]) { s.Active = m }, v)
}

type UserRecipe struct{ opts []gomatchers.Opt[UserSpec] }

func User() UserRecipe { return UserRecipe{} }

func (r UserRecipe) Name(v string) UserRecipe   { r.opts = append(r.opts, WithName(v)); return r }
func (r UserRecipe) Active(v bool) UserRecipe   { r.opts = append(r.opts, WithActive(v)); return r }

// Provider for nesting into parents (defers evaluation; consumes Primitives later)
func (r UserRecipe) Provider() gomatchers.Provider[example.UserView] {
	return gomatchers.FromSpec(BuildUserView, NewUserSpec, r.opts...)
}

// Build now if you need a concrete value (rare in composing tests)
func (r UserRecipe) Build(p gomatchers.Primitives) example.UserView {
	return NewUserFactory(p).Make(r.opts...)
}

func (r UserRecipe) Many(n int, p gomatchers.Primitives) []example.UserView {
	return NewUserFactory(p).Many(n, r.opts...)
}

func ActiveUser() UserRecipe { return User().Active(true) }



type ParentSpec struct {
	Child gomatchers.Maybe[example.UserView]
}

func NewParentSpec() ParentSpec { return ParentSpec{} }

var (
	defChild = gomatchers.FromSpec(BuildUserView, NewUserSpec)
)

func BuildParent(p gomatchers.Primitives, s ParentSpec) example.Parent {
	child := s.Child.Get(p, defChild)

	return example.Parent{
		Child:    child,
	}
}

func NewParentFactory(p gomatchers.Primitives) *gomatchers.SpecFactory[example.Parent, ParentSpec] {
	return gomatchers.NewSpecFactory(p, NewParentSpec, BuildParent)
}

func WithChild(v example.UserView) gomatchers.Opt[ParentSpec] {
	return gomatchers.SetLit(func(s *ParentSpec, m gomatchers.Maybe[example.UserView]) { s.Child = m }, v)
}

func WithChildFromRecipe(rec UserRecipe) gomatchers.Opt[ParentSpec] {
	return gomatchers.SetWith(
		func(s *ParentSpec, m gomatchers.Maybe[example.UserView]) { s.Child = m },
		rec.Provider(),
	)
}

type ParentRecipe struct{ opts []gomatchers.Opt[ParentSpec] }

func Parent() ParentRecipe { return ParentRecipe{} }

// Child sets a literal child to use as the child value in the created instance
func (r ParentRecipe) Child(child example.UserView) ParentRecipe   { r.opts = append(r.opts, WithChild(child)); return r }
// ChildFromRecipe sets a recipe that will be constructed and the result used as the child value when an instance is created
func (r ParentRecipe) ChildFromRecipe(v UserRecipe) ParentRecipe   { r.opts = append(r.opts, WithChildFromRecipe(v)); return r }

// Provider for nesting into parents (defers evaluation; consumes Primitives later)
func (r ParentRecipe) Provider() gomatchers.Provider[example.Parent] {
	return gomatchers.FromSpec(BuildParent, NewParentSpec, r.opts...)
}

// Build now if you need a concrete value (rare in composing tests)
func (r ParentRecipe) Build(p gomatchers.Primitives) example.Parent {
	return NewParentFactory(p).Make(r.opts...)
}

func (r ParentRecipe) Many(n int, p gomatchers.Primitives) []example.Parent {
	return NewParentFactory(p).Many(n, r.opts...)
}
