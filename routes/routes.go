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

	// GET /todos that are expired
	app.Get("todos/expired", handlers.GetExpiredTodo)

	// GET /todos by ID (Retrieve todos by ID).
	app.Get("/todos/:id", handlers.GetTodoByID)

	// DELETE /todos (Remove todos by ID).
	app.Delete("/todos/:id", handlers.DeleteTodo)

	// UPDATE /todos (Update todos by ID).
	app.Patch("/todos/:id", handlers.UpdateTodo)

	// POST /users (Create users)
	app.Post("/users", handlers.CreateUser)

	// GET /users (Retrieve users)
	app.Get("/users", handlers.GetUsers)

	// GET /users (Retrieve users by ID)
	app.Get("/users/:id", handlers.GetUserById)

	// DELETE /users (Delete users by ID)
	app.Delete("/users/:id", handlers.DeleteUser)

	// PATCH /users (Update users by ID)
	app.Patch("/users/:id", handlers.UpdateUser)

}
