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
 	app.Get("/todos",handlers.GetTodos)
}