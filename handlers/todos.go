package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WillvdM/Backend-Challenge-TODOs/config"
	"github.com/WillvdM/Backend-Challenge-TODOs/db"
	"github.com/WillvdM/Backend-Challenge-TODOs/models"
	"github.com/gofiber/fiber/v2"
)

// CreateTodos handles CREATE /todos
// Referenced in routes/routes.go
func CreateTodos(c *fiber.Ctx) error {
	// Define the input structure request body
	var inputs []struct {
		Title     string  `json:"title"`
		Completed bool    `json:"completed"`
		Assignee  string  `json:"assignee"`
		DueDate   *string `json:"due_date"`
	}

	// Parse the JSON body into the input struct
	if err := c.BodyParser(&inputs); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Prepare a slice to hold the inserted todos.
	inserted := []map[string]interface{}{}

	// Loop through each input todo.
	for _, input := range inputs {

		// Validate if the required field is present
		if strings.TrimSpace(input.Assignee) == "" {

			// If not present, an error is returned.
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Assignee is required for all todos",
			})
		}

		// Parse due_date if provided into the dd-mm-yyyy format
		var dueDate *time.Time
		if input.DueDate != nil && *input.DueDate != "" {
			parsed, err := time.Parse("02-01-2006", *input.DueDate) // dd-mm-yyyy
			if err != nil {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid due_date format. Use dd-mm-yyyy", // dd-mm-yyyy is used
				})
			}
			dueDate = &parsed
		}

		// Insert each todo individually
		var id int

		// After inserting, JSON returns all the values that was inserted
		err := db.DB.QueryRow(`
            INSERT INTO todos (title, completed, assignee, due_date)
            VALUES ($1, $2, $3, $4)
            RETURNING id
        `, input.Title, input.Completed, input.Assignee, dueDate).Scan(&id)

		// If no rows were changed, the a server error is returned
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// DueDate is formatted as a string for the response (optional)
		dueDateStr := ""
		if dueDate != nil {
			dueDateStr = dueDate.Format("02-01-2006")
		}

		// Add inserted todo to response slice.
		inserted = append(inserted, map[string]interface{}{
			"id":        id,
			"title":     input.Title,
			"completed": input.Completed,
			"assignee":  input.Assignee,
			"due_date":  dueDateStr,
		})
	}

	// Step 4: Return all inserted todos in JSON format.
	return c.Status(http.StatusCreated).JSON(inserted)
}

// GetTodos handles GET /todos
// Referenced in routes/routes.go
// Fetches list of todos and returns them as JSON to the client.
// Query the db, process each row in the TodoResponse struct, collect them in a slice, the return results as JSON
func GetTodos(c *fiber.Ctx) error {
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if offset < 0 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// --- Sorting ---
	sortField := c.Query("sort", "id")
	order := strings.ToLower(c.Query("order", "asc"))

	allowed := false
	for _, f := range config.Config.SortFields {
		if f == sortField {
			allowed = true
			break
		}
	}
	if !allowed {
		sortField = "id"
	}

	if order != "asc" && order != "desc" {
		order = "asc"
	}

	query := fmt.Sprintf(`
		SELECT id, title, completed, assignee, due_date, completed_at
		FROM todos
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, sortField, order)

	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer rows.Close() // <- now reachable

	todos := []models.TodoResponse{}

	for rows.Next() {
		var todo models.TodoResponse
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.Assignee, &todo.DueDate, &todo.CompletedAt)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var totalCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM todos").Scan(&totalCount)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	currentPage := (offset / limit) + 1
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	response := models.TodosResponse{
		Todos:       todos,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	data, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Send(data)
}

func GetTodoByID(c *fiber.Ctx) error {

	// Get URL id
	idParam := c.Params("id")

	// Convert from string to integer
	id, err := strconv.Atoi(idParam)
	if err != nil {

		// Return error if the TODO id is not a valid integer
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	// Struct for the database results before it is converted to TodoResponse
	var todo struct {
		ID        int
		Title     string
		Completed bool
	}

	// The database is queried for the todo with the given id
	err = db.DB.QueryRow(
		"SELECT id, title, completed FROM TODOS WHERE id = $1",
		id,
	).Scan(&todo.ID, &todo.Title, &todo.Completed)
	if err != nil {

		// Return error if no TODO with the given ID exists
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "TDOO not found",
		})
	}

	// Return the found todos as a JSON
	return c.JSON(fiber.Map{
		"id":        todo.ID,
		"title":     todo.Title,
		"completed": todo.Completed,
	})
}

// DeleteTodo handles DELETE /todos, deleting a todo by the todo ID.
// Referenced in routes/routes.go.
func DeleteTodo(c *fiber.Ctx) error {

	// Extract and validate TODO ID from the URL
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	// Holds an SQL execution result, can check last affected rows or id that was last inserted
	var result sql.Result

	// The deletion logic is executed, based on the config
	switch config.Config.DeletionMode {

	// Mark row as deleted without removing it
	case config.DeletionSoft:
		result, err = db.DB.Exec(
			"Update todos SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id,
		)

	// Permanently remove the row
	case config.DeletionHard:
		result, err = db.DB.Exec(
			"DELETE FROM todos WHERE id = $1", id,
		)

		// Safety: Refuse to delete anything if no deletion mode is specified
	default:
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Deletion mode is invalid",
		})
	}

	// Handle db execution errors and return the error in a JSON messages
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check affected rows
	rows, err := result.RowsAffected()
	if rows == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "TODO not found",
		})
	}

	// Success response if row is deleted
	return c.JSON(fiber.Map{
		"message": "TODO deleted successfully",
	})
}

// Handles updating an existing todo item
// Refernced in routes/routes.go
func UpdateTodo(c *fiber.Ctx) error {

	// Read 'id' path parameter from the URL
	idParam := c.Params("id")

	// Convert the ID from string to integer
	// If conversion fails, ID is invalid
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	// Define a struct for parsing the request body JSON.
	var input struct {
		models.Todo
		DueDate *string `json:"due_date"`
	}

	// Parse the request body JSON into the input struct.
	if err := c.BodyParser(&input); err != nil {

		// Return this error if the structure of request is incorrect.
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Parse due_date string into a time.Time pointer.
	// Only parse if the field is not nil and not empty.
	var dueDate *time.Time
	if input.DueDate != nil && *input.DueDate != "" {
		parsed, err := time.Parse("02-01-2006", *input.DueDate) // Expects dd-mm-yyyy format
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid due_date format. Use dd-mm-yyyy",
			})
		}

		// dueDate pointer is set for database insertion
		dueDate = &parsed
	}

	// Query is executed, parameters is passed in
	query := `
		UPDATE todos
		SET
			title = $1,
			completed = $2,
			completed_at = CASE WHEN $2 = true THEN NOW() ELSE NULL END,
			due_date = $3
		WHERE id = $4
	`

	result, err := db.DB.Exec(
		query,
		input.Title,
		input.Completed,
		dueDate,
		id,
	)

	// Check affected rows. If not affected rows, it means the TODO does not exist and an error is returned.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// If the ID is not found, return not found error.
	if rowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "TODO not found",
		})
	}

	// due_Date is prepared for a JSON response.
	// If it is not nil, it is formatted as a dd-mm-yyyy string.
	var dueDateStr string
	if dueDate != nil {
		dueDateStr = dueDate.Format("02-01-2006")
	}

	// Updated TODO is returned for confirmation, with these fields included
	return c.JSON(fiber.Map{
		"id":          id,
		"title":       input.Title,
		"completedAt": input.Completed,
		"due_date":    dueDateStr,
	})
}
