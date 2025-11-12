package example

//go:generate specta -config testgen.yaml

type UserView struct {
	ID     string
	Name   string
	Active bool
	Score  int
}

type Parent struct {
	Child UserView
}
