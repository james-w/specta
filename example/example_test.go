package example_test

import (
    "testing"
    "github.com/james-w/specta"
    "github.com/james-w/specta/example/factory"
)

func TestFoo(t *testing.T) {
    p := specta.New()

	child := factory.UserView().Name("foo").Build(p)

	specta.AssertThat(t, child, factory.UserViewMatches().Name(specta.Equal("bar")).Matcher())
}

