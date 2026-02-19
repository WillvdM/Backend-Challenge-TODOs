package models

// CreateTodoRequest represents the payload for creating a new TODO item.
type CreateTodoRequest struct {
	Title    string  `json:"title"`
	Assginee string  `json:"assignee"`
	DueDate  *string `json:"due_date,omitempty"`
}

// UpdateTodoRequest represents the payload for updating an existing TODO item.
// All fields are optional, only provided fields will be updated.
type UpdateTodoRequest struct {
	Title     *string `json:"title,omitempty"`
	Completed *bool   `json:"completed,omitempty"`
	DueDate   *string `json:"due_date,omitempty"`
}

// TodoResponse represents a todo returned by the API.
// TodoResponse includes additional fields that are only relevant in responses.
type TodoResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Completed   bool    `json:"completed"`
	Assignee    string  `json:"assignee"`
	DueDate     *string `json:"due_date,omitempty"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

// TodosResponse represents a paginated list of todos.
// TodosRespone is used when returning mulitple todos from the API.
type TodosResponse struct {
	Todos       []TodoResponse `json:"todos"`
	CurrentPage int            `json:"current_page"`
	TotalPages  int            `json:"total_pages"`
}
