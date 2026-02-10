package models

import "time"

// Todo represents the core TODO entity.
// This struct is embedded in other models.
// Todo contains fields shared across input, database, and output.
type Todo struct {
	Title       string     `json:"title,omitempty"`
	Completed   bool       `json:"completed,omitempty"`
	Assignee    string     `json:"assignee"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at"`
}

// TodoInput represents the payload used when creating or updating a todo.
// It embeds Todo to reuse common fields.
type TodoInput struct {
	Todo
}

// TodoResponse represents a todo returned by the API.
// TodoResponse includes additional fields that are only relevant in responses.
type TodoResponse struct {
	ID int `json:"id"`
	Todo
}

// TodosResponses represent a paginated list of todos.
// TodosRespones is used when returning mulitple todos from the API.
type TodosResponse struct {
	Todos       []TodoResponse `json:"todos"`
	CurrentPage int            `json:"current_page"`
	TotalPages  int            `json:"total_pages"`
}
