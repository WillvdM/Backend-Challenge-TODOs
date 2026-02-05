//Seperate request data from DB data

package models

import "time"

// Represents data that is expected to be entered when creating or updating a todo
type TodoInput struct {
	Title     string    `json:"title,omitempty"` // Leave field if it is empty
	Completed bool      `json:"completed,omitempty"`
	CreatedAt time.Time `json:"created_at"` // Not shown as a field
}

// Represents a SINGLE todo item returned in API responses.
type TodoResponse struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Represents a paginated list of todos returned by the API
// Wraps todo data together with pagination metadata
type TodosResponse struct {
	Todos       []TodoResponse `json:"todos"`
	CurrentPage int            `json:"current_page"`
	TotalPages  int            `json:"total_pages"`
}
