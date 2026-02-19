package api

import (
	todoAPI "github.com/WillvdM/Backend-Challenge-TODOs/api/todo"
	userAPI "github.com/WillvdM/Backend-Challenge-TODOs/api/user"
	"github.com/WillvdM/Backend-Challenge-TODOs/database/todo"
	"github.com/WillvdM/Backend-Challenge-TODOs/database/user"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes confgures HTTP routes for the application.
// Each of the routes has a handler function associated that processes requests.
func SetupRoutes(app *fiber.App, todoRepo *todo.TodoRepository, userRepo *user.UserRepository) {

	todoHandler := todoAPI.NewTodoHandler(todoRepo)
	userHandler := userAPI.NewUserHandler(userRepo)

	// POST /todos (Create todos).
	app.Post("/todos", todoHandler.CreateTodos)

	// GET /todos (Retrieve paginated list of todos).
	app.Get("/todos", todoHandler.GetTodos)

	// GET /todos/expired (Retrieve todos that are expired)
	app.Get("todos/expired", todoHandler.GetExpiredTodo)

	// GET /todos by ID:/id (Retrieve todos by ID).
	app.Get("/todos/:id", todoHandler.GetTodoByID)

	// DELETE /todos/:id (Remove todos by ID).
	app.Delete("/todos/:id", todoHandler.DeleteTodo)

	// UPDATE /todos/:id (Update todos by ID).
	app.Patch("/todos/:id", todoHandler.UpdateTodo)

	// POST /users (Create users)
	app.Post("/users", userHandler.CreateUser)

	// GET /users (Retrieve all users)
	app.Get("/users", userHandler.GetUsers)

	// GET /users/:id (Retrieve users by ID)
	app.Get("/users/:id", userHandler.GetUserById)

	// DELETE /users/:id (Delete users by ID)
	app.Delete("/users/:id", userHandler.DeleteUser)

	// PATCH /users/:id (Update users by ID)
	app.Patch("/users/:id", userHandler.UpdateUser)

}
