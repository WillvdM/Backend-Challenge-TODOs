package routes

import (
	"github.com/gofiber/fiber/v2"
    "github.com/WillvdM/Backend-Challenge-TODOs/handlers"
)

func SetupRoutes(app *fiber.App) {
	app.Post("/todos",handlers.CreateTodo)
	app.Get("/todos",handlers.GetTodoByID)
	app.Delete("/todos",handlers.DeleteTodo)
}