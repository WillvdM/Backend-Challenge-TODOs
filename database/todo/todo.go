package todo

import "time"

// Todo represents a task in the system.
// Todo is used internally by the repository layer.
type Todo struct {
	ID          int
	Title       string
	Completed   *bool
	Assignee    string
	DueDate     *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *string
}
