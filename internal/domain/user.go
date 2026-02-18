package domain

import "time"

// User represents a system user in the domain layer.
// User is used internally by repostitories and business logic.
type User struct {
	ID        string
	Name      string
	Surname   string
	Username  string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}
