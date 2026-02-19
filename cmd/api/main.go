package main

import (
	"log"

	"github.com/WillvdM/Backend-Challenge-TODOs/api"
	"github.com/WillvdM/Backend-Challenge-TODOs/config"
	"github.com/WillvdM/Backend-Challenge-TODOs/database"
	"github.com/WillvdM/Backend-Challenge-TODOs/database/todo"
	"github.com/WillvdM/Backend-Challenge-TODOs/database/user"

	"github.com/gofiber/fiber/v2"
)

func main() {

	// Load application config from config.yaml.
	config.LoadConfig()

	// Connect to the database using string config.
	database, err := database.New(config.Config.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	todoRepo := todo.NewTodoRepository(database)
	userRepo := user.NewUserRepository(database)

	// Initialize fiber HTTP server
	app := fiber.New()
	api.SetupRoutes(app, todoRepo, userRepo)
	log.Fatal(app.Listen("localhost:3000"))
}
