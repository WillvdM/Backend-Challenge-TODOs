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

// GetTodos handles GET /todos.
// Referenced in routes/routes.go.
// Fetches list of todos and returns them as JSON to the client.
// Query the db, process each row in the TodoResponse struct, collect them in a slice, the return results as JSON.
func GetTodos(c *fiber.Ctx) error {

	// Offset pagination is used to only return values from a specified point. Default is 0.
	// Prevents invalid input.
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// The limit is set to only return a specified number of data. Default is 10.
	// The limit is converted from string to int.
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if offset < 0 {
		offset = 0
	}

	// The max limit is 100.
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get requested sort field from query parameters , the default is id.
	sortField := c.Query("sort", "id")

	// Get requested sort order from query parameters, the default is ascending.
	order := strings.ToLower(c.Query("order", "asc"))

	// Sort field is validated against the whitelist that was specified in config.yaml.
	// Prevents SQL injection.
	allowed := false
	for _, f := range config.Config.SortFields {
		if f == sortField {
			allowed = true
			break
		}
	}

	// Ensures safe fallback to the default if a field is not allowed.
	if !allowed {
		sortField = "id"
	}

	// Sort order is validated
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// Queries the database with ORDER BY.
	// 'NULLS LAST' ensures that any null timestamp only appears at the bottom of the list.
	query := fmt.Sprintf(`
		SELECT id, title, completed, due_date, completed_at, created_at, assignee
		FROM todos
		ORDER BY %s %s 
		LIMIT $1 OFFSET $2
	`, sortField, order)

	// The query is executed.
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {

		// On query failure, return an error.
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Rows are closed to free up resources.
	defer rows.Close()

	// Rows are scanned
	// A slice is declared to hold the results
	todos := []models.TodoResponse{}

	// A for-loop that iterates over each result set row.
	for rows.Next() {
		var todo models.TodoResponse

		// Columns must be scanned in the exact order that the SELECT statement specifies.
		// Pointer types is used for nullable timestamps, as this prevents runtime errors for columns that are NULL.
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.DueDate, &todo.CompletedAt, &todo.CreatedAt, &todo.Assignee)
		if err != nil {

			// A server error is returned on scan failure.
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Add data to the end of the todo array slice, without modifying the previous data.
		todos = append(todos, todo)
	}

	// Check for iteration errors after looping.
	if err := rows.Err(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Calculate the total number of table rows.
	var totalCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM todos").Scan(&totalCount)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Calculate the current page number.
	currentPage := (offset / limit) + 1

	// Caluculate the total number of pages.
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// The response object is constructed
	response := models.TodosResponse{
		Todos:       todos,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	// JSON response is printed neatly for readability using Marshal response.
	data, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Set the content-type header and send the response.
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Send(data)
}

// GetTodoByID handles GET /todo/:id
// Referenced in routes/routes.go
func GetTodoByID(c *fiber.Ctx) error {

	// Get URL id
	idParam := c.Params("id")

	// Convert from string to integer
	id, err := strconv.Atoi(idParam)
	if err != nil {

		// Return error if the TODO id is not a valid integer.
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	// Struct for the database results.
	var todo struct {
		ID        int
		Title     string
		Completed bool
	}

	// The database is queried for the todo with the given id.
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

// GetExpiredTodo handles GET /todos/expired
// Refrenced in routes/routes.go
func GetExpiredTodo(c *fiber.Ctx) error {

	// This SQL query fetches expired todos.
	// The most overdue todo appears first.
	query := `
	SELECT id, title, completed, assignee, due_date, completed_at, created_at
	FROM todos
	WHERE due_date IS NOT NULL
	AND due_date < NOW()
	AND (completed IS NULL OR completed = false)
	ORDER BY due_date ASC
	`

	// The query is executed.
	rows, err := db.DB.Query(query)
	if err != nil {

		// Return a server error if the query fails.
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Ensure that the rows are closed after processing, increasing performance.
	defer rows.Close()

	// The expired slice stores all expired toods.
	expired := []models.TodoResponse{}

	// Each row that the query returned is iterated over.
	for rows.Next() {
		var todo models.TodoResponse

		// Temporary varaibles for DB column scanning.
		var title string
		var assignee string
		var completed bool
		var dueDate *string // Changed from *time.Time to string for easier user input.
		var completedAt *time.Time
		var createdAt *time.Time

		// The database row is scanned into temporary variables.
		if err := rows.Scan(&todo.ID, &title, &completed, &assignee, &dueDate, &completedAt, &createdAt); err != nil {

			// Return a server error if the scanning fails.
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// The scanned values are mapped into the TodoResponse struct.
		todo.Title = &title
		todo.Completed = &completed
		todo.Assignee = &assignee
		todo.DueDate = dueDate
		todo.CompletedAt = completedAt
		todo.CreatedAt = createdAt

		// The todo is appended to the expired slice.
		expired = append(expired, todo)
	}

	// List all values in JSON format.
	return c.JSON(fiber.Map{"expired_todos": expired})

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

// Handles updating an existing todo item with PATCH
// Refernced in routes/routes.go
func UpdateTodo(c *fiber.Ctx) error {

	// Extract the todo ID from the URL path parameter, convert it to integer.
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {

		// Return error if ID is not an integer.
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	// Define a struct to hold input fields for partial updates (only provided fields gets updated)
	var input struct {
		Title     *string `json:"title,omitempty"`
		Completed *bool   `json:"completed,omitempty"`
		DueDate   *string `json:"due_date,omitempty"` // string for parsing, leading to easier user input
	}

	// Allow empty PATCH body, but only parse request body if it is not empty.
	if len(c.Body()) != 0 {
		if err := c.BodyParser(&input); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
	}

	// Parse the due_date string into *time.Time if it is provided.
	var dueDate *time.Time
	if input.DueDate != nil && *input.DueDate != "" {
		parsed, err := time.Parse("02-01-2006", *input.DueDate) // Expects dd-mm-yyyy format.
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid due_date format. Use dd-mm-yyyy",
			})
		}
		dueDate = &parsed
	}

	// Build dynamic SET clause
	setParts := []string{}
	args := []interface{}{}
	argID := 1

	// If a title is provided, it will be included.
	if input.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argID))
		args = append(args, *input.Title)
		argID++
	}

	// If completed flag is provided, it will be included.
	if input.Completed != nil {
		setParts = append(setParts, fmt.Sprintf("completed = $%d", argID))
		args = append(args, *input.Completed)
		argID++
		setParts = append(setParts, fmt.Sprintf("completed_at = CASE WHEN $%d = true THEN NOW() ELSE NULL END", argID-1))
	}

	// If due_date is provided, it will be included.
	if dueDate != nil {
		setParts = append(setParts, fmt.Sprintf("due_date = $%d", argID))
		args = append(args, dueDate)
		argID++
	}

	// If no fields were provided in the body, an error is returned.
	if len(setParts) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields provided for update",
		})
	}

	// The final SQL query is dynamically constructed.
	query := fmt.Sprintf(`
		UPDATE todos
		SET %s
		WHERE id = $%d
		RETURNING title, completed_at, due_date
	`, strings.Join(setParts, ", "), argID)

	// Append the TODO ID as the last argument for the WHERE clause.
	args = append(args, id)

	// These variables hold the returned values from the query.
	var title string
	var completedAt *time.Time
	var dbDueDate *time.Time

	// Query is executed, and results scanned.
	err = db.DB.QueryRow(query, args...).Scan(&title, &completedAt, &dbDueDate)
	if err == sql.ErrNoRows {

		// Return an error if no TODO was found with the given ID.
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "TODO not found",
		})
	}
	if err != nil {

		// ANy other database error.
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Format completed_at timestamp as a string if it is present.
	var completedAtStr string
	if completedAt != nil {
		completedAtStr = completedAt.Format("02-01-2006 15:04:05")
	}

	// Format due_date timestamp as a string if it is present.
	var dueDateStr string
	if dbDueDate != nil {
		dueDateStr = dbDueDate.Format("02-01-2006")
	}

	// Return updated TODO fields in JSON format.
	return c.JSON(fiber.Map{
		"id":          id,
		"title":       title,
		"completedAt": completedAtStr,
		"due_date":    dueDateStr,
	})
}
