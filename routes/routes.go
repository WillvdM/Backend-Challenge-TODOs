package routes

import (
	"github.com/WillvdM/Backend-Challenge-TODOs/handlers"
	"github.com/gofiber/fiber/v2"
)

// Confgures HTTP routes for the application.
// Each of the routes has a handler funcrion associated that processes requests.
func SetupRoutes(app *fiber.App) {

	// POST /todos (Create todos).
	app.Post("/todos", handlers.CreateTodos)

	// GET /todos (Retrieve paginated list of todos).
	app.Get("/todos", handlers.GetTodos)

	// GET /todos by ID (Retrieve todos by ID).
	app.Get("/todos/:id", handlers.GetTodoByID)

	// DELETE /todos (Remove todos by ID).
	app.Delete("/todos/:id", handlers.DeleteTodo)

	// UPDATE /todos (Update todos by ID).
	app.Put("/todos/:id", handlers.UpdateTodo)
}
