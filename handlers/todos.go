package handlers

import (
	"strconv"

	"github.com/WillvdM/Backend-Challenge-TODOs/db"
	"github.com/WillvdM/Backend-Challenge-TODOs/models"
	"github.com/gofiber/fiber/v2"
)

// Create TODO handles POST /todos. Receives JSON input, inserts a new todo into the database, returns the created todo with ID it generated.
func CreateTodo(c *fiber.Ctx) error {
	var input models.TodoInput

	// JSON body is parsed into TodoInput structure
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Save all at once
	var id int
	err := db.DB.QueryRow(
		"INSERT INTO todos (title, completed) VALUES ($1, $2) RETURNING id",
		input.Title,
		input.Completed,
	).Scan(&id)

	// Error handling for database insertion
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return created todo
	todo := map[string]interface{}{
		"id":        id,
		"title":     input.Title,
		"completed": input.Completed,
	}
	return c.Status(201).JSON(todo)
}

// GetTodoByID handles GET /todos/:id
func GetTodos(c *fiber.Ctx) error {
	rows, err := db.DB.Query("SELECT id , title, completed FROM todos ")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	// Collect the results
	todos := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var title string
		var completed bool

		// Scan the row values
		err := rows.Scan(&id, &title, &completed)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
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

func DeleteTodo(c *fiber.Ctx) error {
	// Get URL ID
	idParam := c.Params("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	result, err := db.DB.Exec("DELETE FROM todos WHERE id=$1", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "TODO not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "TODO deleted successfully",
	})
}
