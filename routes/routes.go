package routes

import (
	"github.com/WillvdM/Backend-Challenge-TODOs/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {

	// POST /todos
	app.Post("/todos", handlers.CreateTodo)

	// GET /todos
	app.Get("/todos", handlers.GetTodos)

	// GET /todos by ID
	app.Get("/todos/:id", handlers.GetTodoByID)

	// DELETE /todos
	app.Delete("/todos/:id", handlers.DeleteTodo)

	// UPDATE /todos
	app.Put("todos/:id", handlers.UpdateTodo)
}
