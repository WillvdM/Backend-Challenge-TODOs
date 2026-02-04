package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/WillvdM/Backend-Challenge-TODOs/config"
	"github.com/WillvdM/Backend-Challenge-TODOs/db"
	"github.com/WillvdM/Backend-Challenge-TODOs/models"
	"github.com/gofiber/fiber/v2"
)

// CreateTodo handles POST /todos. Receives JSON input, inserts a new todo into the database, returns the created todo with ID it generated.
func CreateTodo(c *fiber.Ctx) error {
	var input []models.TodoInput

	// JSON body is parsed into TodoInput structure
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// NB!!!! Still need to change this
	// Will still do research regarding if this code is the best for inserting multiple todos at once, code was copied

	var examples []string
	var values []interface{}
	/* Will do research if interface is the best method
	//Magic number? */
	counter := 1

	for _, todo := range input {
		example := fmt.Sprintf("($%d,$%d)", counter, counter+1)
		counter += 2 //Magic number?
		examples = append(examples, example)
		values = append(values, todo.Title, todo.Completed)
	}

	placeholder := strings.Join(examples, ",") // Unclear about what placeholder does?
	query := fmt.Sprintf("INSERT INTO todos (title, completed) VALUES %s RETURNING id", placeholder)

	// Run the DB query and extract the query result rows
	rows, err := db.DB.Query(query, values...)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Extract the IDs from the query rows
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		ids = append(ids, id)
	}

	// Return success message if the TODO was created
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"ids": ids,
	})
}

// GetTodoByID handles GET /todos/:id
func GetTodos(c *fiber.Ctx) error {
	rows, err := db.DB.Query("SELECT id , title, completed FROM todos ")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	// Collect the results
	todos := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var title string
		var completed bool

		// Scan the row values and return an error if the rows are unchganged
		err := rows.Scan(&id, &title, &completed)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// TODO object is built from row data
		todo := map[string]interface{}{
			"id":        id,
			"title":     title,
			"completed": completed,
		}

		// Append the data
		todos = append(todos, todo)
	}

	// All todos are returned
	return c.JSON(todos)
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

	// Temporary struct for the database results
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

func DeleteTodo(c *fiber.Ctx) error {

	// Extract and validate TODO ID from the URL
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	// TODO cannot be deleted if the conifguration was not loaded properly
	if !config.Loaded {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Deletion config not loaded",
		})
	}

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

	// Handle db execution errors
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

	// Success response
	return c.JSON(fiber.Map{
		"message": "TODO deleted successfully",
		"mode":    config.Config.DeletionMode,
	})
}

// Update TODO by ID
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

	// Return this error if the structure of request is incorrect
	var input models.TodoInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// SQL UPDATE statement is executed, updates given TODOs
	result, err := db.DB.Exec(
		"UPDATE TODOs SET title = $1, completed = $2 WHERE id = $3",
		input.Title,
		input.Completed,
		id,
	)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check affected rows. If not affected rows, it means the TODO does not exist
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// If the ID is not found, return not found error
	if rowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "TODO not found",
		})
	}

	// Updated TODO is returned for confirmation
	return c.JSON(fiber.Map{
		"id":        "id",
		"title":     input.Title,
		"completed": input.Completed,
	})
}
