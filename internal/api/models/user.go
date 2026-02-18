package models

import "time"

// UserInput represents the payload for creating a new user.
type UserInput struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Username string `json:"username"`
}

// User represents the payload that is returned, which lists the user.
type User struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Surname   string     `json:"surname"`
	Username  string     `json:"username"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
