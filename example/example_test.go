package example_test

import (
    "fmt"
    "testing"
    "github.com/james-w/gomatchers"
    "github.com/james-w/gomatchers/example/factory"
)

func TestFoo(t *testing.T) {
    p := gomatchers.New()

	child := factory.User().Name("foo").Build(p)

    fmt.Printf("%+v\n", child)

	fmt.Printf("%+v\n", factory.Parent().ChildFromRecipe(
		factory.ActiveUser().Name("bar"),
	).Many(2, p))

	fmt.Printf("%+v\n", factory.Parent().Child(child).Many(2, p))
}

