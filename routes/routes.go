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
}
