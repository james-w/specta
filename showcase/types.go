package showcase

//go:generate specta -config testgen.yaml

import (
	"fmt"
	"strings"
	"time"
)

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

// BankAccount demonstrates constructor-based factory generation
// with unexported fields that should not be directly settable
type BankAccount struct {
	id      string
	name    string
	balance int
}

// NewBankAccount creates a new bank account with the given name and initial balance
func NewBankAccount(name string, balance int) BankAccount {
	return BankAccount{
		id:      "account_" + name, // Generated ID
		name:    name,
		balance: balance,
	}
}

// GetName returns the account name (needed for testing)
func (b BankAccount) GetName() string {
	return b.name
}

// GetBalance returns the account balance (needed for testing)
func (b BankAccount) GetBalance() int {
	return b.balance
}

// Email represents a validated email address
// This demonstrates constructor-based factories with error returns
type Email struct {
	address string
}

// NewEmail creates a new Email with validation
// Returns an error if the email format is invalid
func NewEmail(address string) (Email, error) {
	if address == "" {
		return Email{}, fmt.Errorf("email address cannot be empty")
	}
	if !strings.Contains(address, "@") {
		return Email{}, fmt.Errorf("invalid email format: %q", address)
	}
	return Email{address: address}, nil
}

// GetAddress returns the email address (needed for testing)
func (e Email) GetAddress() string {
	return e.address
}

// Account represents a user account with an associated user
// This demonstrates constructor-based factories with custom type parameters
type Account struct {
	id     string
	user   User
	status string
}

// NewAccount creates a new Account for the given user
func NewAccount(user User, status string) Account {
	return Account{
		id:     "account_" + user.ID,
		user:   user,
		status: status,
	}
}

// GetUser returns the associated user
func (a Account) GetUser() User {
	return a.user
}

// GetStatus returns the account status
func (a Account) GetStatus() string {
	return a.status
}
