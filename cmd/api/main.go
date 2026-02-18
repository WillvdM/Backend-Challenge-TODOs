package main

import (
	"log"

	"github.com/WillvdM/Backend-Challenge-TODOs/config"
	"github.com/WillvdM/Backend-Challenge-TODOs/internal/api/routes"
	"github.com/WillvdM/Backend-Challenge-TODOs/internal/repository"
	"github.com/gofiber/fiber/v2"
)

func main() {

	// Load application config from config.yaml.
	config.LoadConfig()

	// Connect to the database using string config.
	database, err := repository.New(config.Config.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	todoRepo := repository.NewTodoRepository(database)
	userRepo := repository.NewUserRepository(database)

	// Initialize fiber HTTp server
	app := fiber.New()
	routes.SetupRoutes(app, todoRepo, userRepo)
	log.Fatal(app.Listen("localhost:3000"))
}
