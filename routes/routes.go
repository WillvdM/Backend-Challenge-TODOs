package routes

import (
	"github.com/gofiber/fiber/v2"
    "github.com/WillvdM/Backend-Challenge-TODOs/handlers"
)

func SetupRoutes(app *fiber.App) {

	// POST /todos
	// Creates a new todo
	// Call the CreateTodo handler
	app.Post("/todos",handlers.CreateTodo)

	// GET /todos
	// Returns a specific todo by the ID
	// Only echoes ID currently, needs to query db later
	app.Get("/todos",handlers.GetTodoByID)

	// DELETE /todos/:id
	// Deletes a specific todo by the ID
	// Returns no content currently, as no db deletion is inplemented yet
	app.Delete("/todos",handlers.DeleteTodo)
}