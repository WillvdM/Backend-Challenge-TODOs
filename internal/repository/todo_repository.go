package repository

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/WillvdM/Backend-Challenge-TODOs/internal/domain"
)

// TodoRepository wraps a sql.DB connection.
// Handles all TODO-related database operations.
type TodoRepository struct {
	DB *sql.DB
}

// NewTodoRepository creates a new TodoRepository with a database connection.
func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{DB: db}
}

// Create inserts a new TODO into the database.
// Returns the generated row ID.
func (repo *TodoRepository) Create(todo domain.Todo) (int, error) {
	var id int
	err := repo.DB.QueryRow(`
	INSERT INTO todos (title, completed, assignee, due_date)
	VALUES ($1, $2, $3, $4)
	RETURNING id
 `, todo.Title, todo.Completed, todo.Assignee, todo.DueDate).Scan(&id)
	return id, err
}

// GetById retrieves a single TODO by its ID.
func (repo *TodoRepository) GetById(id int) (domain.Todo, error) {
	var todo domain.Todo
	err := repo.DB.QueryRow(`
	SELECT id, title, completed, due_date, completed_at, created_at, assignee, updated_at, deleted_at
	FROM todos
	WHERE id = $1
	`, id).Scan(
		&todo.ID, &todo.Title, &todo.Completed, &todo.DueDate, &todo.CompletedAt, &todo.CreatedAt, &todo.Assignee, &todo.UpdatedAt, &todo.DeletedAt)
	return todo, err
}

// Update dynamically updates the fields of a TODO.
// Only provided fields are updated.
// Automatically sets the update_at time to the created_at time upon creation.
func (repo *TodoRepository) Update(id int, todo domain.Todo) error {
	var setParts []string
	var args []interface{}
	argID := 1

	if todo.Title != "" {
		setParts = append(setParts, "title=$"+strconv.Itoa(argID))
		args = append(args, todo.Title)
		argID++
	}

	if todo.Completed != nil {
		setParts = append(setParts, "completed=$"+strconv.Itoa(argID))
		args = append(args, todo.Completed)
		argID++
		setParts = append(setParts, "completed_at=CASE WHEN $"+strconv.Itoa(argID-1)+"=true THEN NOW() ELSE NULL END")
	}

	if todo.DueDate != nil {
		setParts = append(setParts, "due_date=$"+strconv.Itoa(argID))
		args = append(args, todo.DueDate)
		argID++
	}

	if len(setParts) == 0 {
		return nil
	}

	query := "UPDATE todos SET " + strings.Join(setParts, ", ") + " , updated_at=NOW() WHERE id=$" + strconv.Itoa(argID)
	args = append(args, id)

	_, err := repo.DB.Exec(query, args...)
	return err
}

// Delete deletes a TODO based on the ID provided.
// If hard is true, perform a permanent delete. Otherwise, perform a soft delete (set deleted_at).

func (repo *TodoRepository) Delete(id int, hard bool) error {
	if hard {
		_, err := repo.DB.Exec(`
		DELETE FROM todos 
		WHERE id = $1
		`, id)
		return err
	} else {
		_, err := repo.DB.Exec(`
		UPDATE todos SET deleted_at=NOW()
		WHERE id = $1 AND deleted_at IS NULL
		`, id)
		return err
	}
}

// List returns a paginated list of TODOs.
// Support dynamic sorting the TODOs.
func (repo *TodoRepository) List(offset, limit int, sortField, order string) ([]domain.Todo, error) {
	query := `
	SELECT id, title, completed, due_date, completed_at, created_at, assignee, updated_at, deleted_at
	FROM todos
	ORDER BY ` + sortField + " " + order + ` 
	LIMIT $1 OFFSET $2`

	rows, err := repo.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []domain.Todo
	for rows.Next() {
		var todo domain.Todo
		if err := rows.Scan(
			&todo.ID, &todo.Title, &todo.Completed, &todo.DueDate, &todo.CompletedAt, &todo.CreatedAt, &todo.Assignee, &todo.UpdatedAt, &todo.DeletedAt); err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}
	return todos, nil
}

// GetExpired returns TODOs whose due_date has passed, which are not marked as completed.
func (repo *TodoRepository) GetExpired() ([]domain.Todo, error) {
	rows, err := repo.DB.Query(`
	SELECT id, title, completed, due_date, completed_at, created_at, assignee, updated_at
	FROM todos
	WHERE due_date < NOW() AND completed = false
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var expired []domain.Todo
	for rows.Next() {
		var todo domain.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.DueDate, &todo.CompletedAt, &todo.CreatedAt, &todo.Assignee, &todo.UpdatedAt); err != nil {
			return nil, err
		}

		expired = append(expired, todo)
	}
	return expired, nil
}
