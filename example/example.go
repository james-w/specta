package example

//go:generate go run ../cmd/main.go -config testgen.yaml

type UserView struct {
	ID     string
	Name   string
	Active bool
	Score  int
}

type Parent struct {
	Child UserView
}
