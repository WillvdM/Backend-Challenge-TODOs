//Seperate request data from DB data

package models

import "time"

// Represents data that is expected to be entered when creating or updating a todo
type TodoInput struct {
	Title     string     `json:"title,omitempty"` // Leave field if it is empty
	Completed bool       `json:"completed,omitempty"`
	CreatedAt time.Time  `json:"created_at"` // Not shown as a field
	Assignee  string     `json:"assignee"`
	DueDate   *time.Time `json:"due_date,omitempty"`
}

// This represents a todo item in API responses.
// The struct field are serialized to JSON when it is returned.
// Serialization: process of converting a data structure or object into a format that can be stored or transmitted and later deconstructed.
type TodoResponse struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Completed   bool       `json:"completed"`
	Assignee    string     `json:"assignee"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Represents a paginated list of todos returned by the API
// Wraps todo data together with pagination metadata
type TodosResponse struct {
	Todos       []TodoResponse `json:"todos"`
	CurrentPage int            `json:"current_page"`
	TotalPages  int            `json:"total_pages"`
}
