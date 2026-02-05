package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/WillvdM/Backend-Challenge-TODOs/config"
	"github.com/WillvdM/Backend-Challenge-TODOs/db"
	"github.com/WillvdM/Backend-Challenge-TODOs/models"
	"github.com/gofiber/fiber/v2"
)

// This represents a todo item in API responses.
// The struct field are serialized to JSON when it is returned.
// Serialization: process of converting a data structure or object into a format that can be stored or transmitted and later deconstructed.
type TodoResponse struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Used in routes/routes.go
func CreateTodo(c *fiber.Ctx) error {

	// Inserts a new todo into the database, returns the created todo with ID it generated.
	var input []models.TodoInput

	// JSON body is parsed into TodoInput structure, return an error if invalid.
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// NB!!!! Still need to change this.
	// Will still do research regarding if this code is the best for inserting multiple todos at once, code was copied.
	var examples []string

	// Will do research if interface is the best method.
	var values []interface{}

	//Magic number?
	counter := 1

	// for loop, copied but not researched.
	for _, todo := range input {
		example := fmt.Sprintf("($%d,$%d)", counter, counter+1)
		counter += 2 //Magic number?
		examples = append(examples, example)
		values = append(values, todo.Title, todo.Completed)
	}

	// Allows script to run until actual data is available.
	placeholder := strings.Join(examples, ",")
	query := fmt.Sprintf("INSERT INTO todos (title, completed) VALUES %s RETURNING id", placeholder)

	// Run the DB query and extract the query result rows. If no rows are returned, an error code is returned.
	rows, err := db.DB.Query(query, values...)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Extract the IDs from the query rows.
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Adds elements to the end of a slice, if underlying array capacity is exceeded.
		ids = append(ids, id)
	}

	// Return success message if the TODO was created.
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"ids": ids,
	})
}

// Used in routes/routes.go
// Fetches list of todos and returns them as JSON to the client.

// Query the db, process each row in the TodoResponse struct, collect them in a slice, the return results as JSON
func GetTodos(c *fiber.Ctx) error {

	// Retrieve a paginated list of todos from the database, return them as JSON.
	// - offset (optional, default 0): number of records to skip
	offset, _ := strconv.Atoi(c.Query("offset", "0")) // Coneverts from strings to integers
	// - limit  (optional, default size is 10, max is 100): number of records to return
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Validate the parameters, offset should not be negative
	if offset < 0 {
		offset = 0
	}

	// Limit should be between 1 and 100, avoids excessive queries
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Execute query with paramaterized inputs, prevents SQL injection
	// Retrive all the values and order them by id
	rows, err := db.DB.Query("SELECT id, title, completed FROM todos ORDER BY id LIMIT $1 OFFSET $2", limit, offset) // Paramterized SQL query
	if err != nil {

		// Return error if the query fails, with the error in JSON format
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Ensure the rows are closed when function returns.
	defer rows.Close()

	var totalCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM todos").Scan(&totalCount)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Calculate current page using offset and limit (1-based indexing).
	// Total number of pages calculated by using ceiling division.
	currentPage := (offset / limit) + 1

	// Total number of pages calculated by using ceiling division.
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Initializes and collects the results into an empty slice.
	todos := []models.TodoResponse{}

	// Each row is processed and appended to the response slice.
	for rows.Next() {
		var todo models.TodoResponse

		// Scan the row values into the TodoResponse struct and return an error if the rows are unchganged (scanning fails).
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Adds elements to the end of a slice, if underlying array capacity is exceeded (dynamically resizing collection).
		todos = append(todos, todo)
	}

	// Check errors encountered during the iteration proccess.
	if err := rows.Err(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Todos and metadate are wrapped into a single response object.
	response := models.TodosResponse{
		Todos:       todos,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	// The above response is marshalled into prettyJSON, making it look neater.
	data, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Set the HeaderContentType, JSON data is sent as an HTTP response.
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	// All todos are returned.
	return c.Send(data)
}

// Used in routes/routes.go
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

// Used in routes/routes.go
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

// Used in routes/routes.go
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

	// If an error occurs the server error 500 message is returned, as well as an error description in the JSON.
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
