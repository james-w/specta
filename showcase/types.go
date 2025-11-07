package showcase

//go:generate go run ../cmd/main.go -config testgen.yaml

import "time"

// Address represents a physical address
type Address struct {
	Street  string
	City    string
	State   string
	ZipCode string
	Country string
}

// User represents a user account
type User struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Active    bool
	Address   Address
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Product represents an item for sale
type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	InStock     bool
	CreatedAt   time.Time
}

// OrderItem represents a single item in an order
type OrderItem struct {
	Product  Product
	Quantity int
	Price    float64
}

// Order represents a customer order
type Order struct {
	ID        string
	User      User
	Items     []OrderItem
	Total     float64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BlogPost demonstrates nested user and timestamps
type BlogPost struct {
	ID          string
	Title       string
	Content     string
	Author      User
	Published   bool
	PublishedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Comment demonstrates multiple levels of nesting
type Comment struct {
	ID        string
	Post      BlogPost
	Author    User
	Content   string
	CreatedAt time.Time
}
