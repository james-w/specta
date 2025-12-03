package conjecture_test

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/james-w/specta/conjecture"
)

// =============================================================================
// Basic Examples
// =============================================================================

func TestIntegerProperty(t *testing.T) {
	// Simple property: x + 0 == x
	conjecture.QuickCheck(t, conjecture.Integer(0, 1000), func(x int64) error {
		if x+0 != x {
			return errors.New("additive identity violated")
		}
		return nil
	})
}

func TestSliceReverse(t *testing.T) {
	ct := conjecture.NewT(t)

	// Property: reversing a slice twice gives the original
	conjecture.Check(ct, "reverse_reverse_identity",
		conjecture.Slice(conjecture.Integer(-100, 100), 0, 50),
		func(xs []int64) error {
			original := make([]int64, len(xs))
			copy(original, xs)

			// Reverse once
			reverse(xs)
			// Reverse again
			reverse(xs)

			for i := range xs {
				if xs[i] != original[i] {
					return errors.New("reverse(reverse(xs)) != xs")
				}
			}
			return nil
		})
}

func reverse(xs []int64) {
	for i, j := 0, len(xs)-1; i < j; i, j = i+1, j-1 {
		xs[i], xs[j] = xs[j], xs[i]
	}
}

func TestSortIdempotent(t *testing.T) {
	ct := conjecture.NewT(t)

	// Property: sorting a sorted slice gives the same slice
	conjecture.Check(ct, "sort_idempotent",
		conjecture.Slice(conjecture.Integer(0, 1000), 0, 100),
		func(xs []int64) error {
			sorted1 := make([]int64, len(xs))
			copy(sorted1, xs)
			sort.Slice(sorted1, func(i, j int) bool { return sorted1[i] < sorted1[j] })

			sorted2 := make([]int64, len(sorted1))
			copy(sorted2, sorted1)
			sort.Slice(sorted2, func(i, j int) bool { return sorted2[i] < sorted2[j] })

			for i := range sorted1 {
				if sorted1[i] != sorted2[i] {
					return errors.New("sort(sort(xs)) != sort(xs)")
				}
			}
			return nil
		})
}

// =============================================================================
// Composite Generator Examples
// =============================================================================

// User represents a domain object
type User struct {
	ID    int64
	Name  string
	Email string
	Age   int
}

// UserGen creates a generator for User objects
func UserGen() conjecture.Gen[User] {
	return conjecture.Build("User", func(d conjecture.DataSource) (User, error) {
		id, err := d.DrawInteger(conjecture.IntegerParams{Min: 1, Max: 1000000})
		if err != nil {
			return User{}, err
		}

		name, err := d.DrawString(conjecture.StringParams{
			MinSize:   1,
			MaxSize:   50,
			Intervals: conjecture.DefaultASCIIPrintable(),
		})
		if err != nil {
			return User{}, err
		}

		age, err := d.DrawInteger(conjecture.IntegerParams{Min: 0, Max: 150})
		if err != nil {
			return User{}, err
		}

		return User{
			ID:    id,
			Name:  name,
			Email: name + "@example.com",
			Age:   int(age),
		}, nil
	})
}

func TestUserValidation(t *testing.T) {
	ct := conjecture.NewT(t)

	conjecture.Check(ct, "user_has_valid_email",
		UserGen(),
		func(u User) error {
			if len(u.Email) == 0 {
				return errors.New("email cannot be empty")
			}
			// Email should contain @
			for _, c := range u.Email {
				if c == '@' {
					return nil
				}
			}
			return errors.New("email must contain @")
		})
}

// =============================================================================
// OneOf and Recursive Examples
// =============================================================================

// Expr represents a simple arithmetic expression
type Expr interface {
	Eval() int64
}

type Lit struct{ Value int64 }
type Add struct{ Left, Right Expr }
type Mul struct{ Left, Right Expr }

func (l Lit) Eval() int64 { return l.Value }
func (a Add) Eval() int64 { return a.Left.Eval() + a.Right.Eval() }
func (m Mul) Eval() int64 { return m.Left.Eval() * m.Right.Eval() }

// ExprGen creates a generator for expression trees
func ExprGen(maxDepth int) conjecture.Gen[Expr] {
	litGen := conjecture.Map(conjecture.Integer(-10, 10), func(n int64) Expr {
		return Lit{Value: n}
	})

	if maxDepth <= 0 {
		return litGen
	}

	subGen := ExprGen(maxDepth - 1)

	addGen := conjecture.Build("Add", func(d conjecture.DataSource) (Expr, error) {
		left, err := subGen.Draw(d)
		if err != nil {
			return nil, err
		}
		right, err := subGen.Draw(d)
		if err != nil {
			return nil, err
		}
		return Add{Left: left, Right: right}, nil
	})

	mulGen := conjecture.Build("Mul", func(d conjecture.DataSource) (Expr, error) {
		left, err := subGen.Draw(d)
		if err != nil {
			return nil, err
		}
		right, err := subGen.Draw(d)
		if err != nil {
			return nil, err
		}
		return Mul{Left: left, Right: right}, nil
	})

	// Prefer simpler expressions during shrinking
	return conjecture.OneOf(litGen, addGen, mulGen)
}

func TestExpressionEval(t *testing.T) {
	ct := conjecture.NewT(t)

	// Property: evaluation is deterministic
	conjecture.Check(ct, "eval_deterministic",
		ExprGen(3),
		func(e Expr) error {
			v1 := e.Eval()
			v2 := e.Eval()
			if v1 != v2 {
				return errors.New("expression evaluation is non-deterministic")
			}
			return nil
		})
}

// =============================================================================
// Stateful Testing Example
// =============================================================================

// Counter is a simple stateful object
type Counter struct {
	value int
}

func (c *Counter) Inc()     { c.value++ }
func (c *Counter) Dec()     { c.value-- }
func (c *Counter) Get() int { return c.value }
func (c *Counter) Reset()   { c.value = 0 }

// Operation represents an operation on the counter
type Operation int

const (
	OpInc Operation = iota
	OpDec
	OpReset
)

func OperationGen() conjecture.Gen[Operation] {
	return conjecture.Map(conjecture.Integer(0, 2), func(n int64) Operation {
		return Operation(n)
	})
}

func TestCounterStateMachine(t *testing.T) {
	ct := conjecture.NewT(t)

	// Property: counter value matches expected after sequence of operations
	conjecture.Check(ct, "counter_state_machine",
		conjecture.Slice(OperationGen(), 0, 100),
		func(ops []Operation) error {
			c := &Counter{}
			expected := 0

			for _, op := range ops {
				switch op {
				case OpInc:
					c.Inc()
					expected++
				case OpDec:
					c.Dec()
					expected--
				case OpReset:
					c.Reset()
					expected = 0
				}
			}

			if c.Get() != expected {
				return errors.New("counter value mismatch")
			}
			return nil
		})
}

// =============================================================================
// Filter and Precondition Examples
// =============================================================================

func TestFilteredProperty(t *testing.T) {
	ct := conjecture.NewT(t)

	// Generate only even numbers
	evenGen := conjecture.Filter(
		conjecture.Integer(0, 100),
		func(n int64) bool { return n%2 == 0 },
	)

	conjecture.Check(ct, "even_numbers_are_even",
		evenGen,
		func(n int64) error {
			if n%2 != 0 {
				return errors.New("generated number is not even")
			}
			return nil
		})
}

// =============================================================================
// Map Generator Examples
// =============================================================================

func TestMapProperties(t *testing.T) {
	ct := conjecture.NewT(t)

	mapGen := conjecture.MapGen(
		conjecture.ASCIIString(1, 10),
		conjecture.Integer(0, 1000),
		0, 20,
	)

	conjecture.Check(ct, "map_has_expected_keys",
		mapGen,
		func(m map[string]int64) error {
			// All keys should be non-empty ASCII strings
			for k := range m {
				if len(k) == 0 {
					return errors.New("empty key found")
				}
			}
			return nil
		})
}

// =============================================================================
// Demonstrating Shrinking
// =============================================================================

// Example_shrinking demonstrates how shrinking finds minimal counterexamples.
// This property intentionally fails to show the shrinking output.
func Example_shrinking() {
	// Use a fixed seed so output is deterministic
	data := conjecture.NewConjectureData(conjecture.WithSeed(1))

	// Draw a slice with values 0-99 (uses uniform distribution for ranges ≤100)
	xs, _ := conjecture.Slice(conjecture.Integer(0, 99), 1, 10).Draw(data)

	// Check if any value >= 50
	for _, x := range xs {
		if x >= 50 {
			data.MarkInteresting("found value >= 50")
			break
		}
	}

	// If interesting, shrink it
	if data.Status() == conjecture.StatusInteresting {
		shrinkResult := conjecture.Shrink(
			context.Background(),
			data.Sequence(),
			func(d *conjecture.ConjectureData) bool {
				xs, err := conjecture.Slice(conjecture.Integer(0, 99), 1, 10).Draw(d)
				if err != nil {
					return false
				}
				for _, x := range xs {
					if x >= 50 {
						d.MarkInteresting("found value >= 50")
						return true
					}
				}
				return false
			},
		)

		fmt.Printf("Property failed:\nfound value >= 50\n\nFailing example:\n")
		seq := shrinkResult.Sequence
		for i := 0; i < seq.Len() && i < 10; i++ {
			c := seq.Get(i)
			fmt.Printf("[%d] %s: %v\n", i, c.Type, c.Value)
		}
	}

	// Output:
	// Property failed:
	// found value >= 50
	//
	// Failing example:
	// [0] integer: 50
	// [1] boolean: false
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkGeneration(b *testing.B) {
	gen := conjecture.Slice(conjecture.Integer(0, 1000), 0, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
		_, _ = gen.Draw(d)
	}
}

func BenchmarkStringGeneration(b *testing.B) {
	gen := conjecture.StringWith(conjecture.StringParams{
		MinSize:   0,
		MaxSize:   100,
		Intervals: conjecture.DefaultUnicode(),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
		_, _ = gen.Draw(d)
	}
}
