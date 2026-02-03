package main

import (
	"log"

	"github.com/WillvdM/Backend-Challenge-TODOs/db"
	"github.com/WillvdM/Backend-Challenge-TODOs/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {

	//Create a new Fiber instance
	app := fiber.New()

	// Establish PostgreSQL connection
	db.Connect()

	// Registers all API routes with Fiber app
	routes.SetupRoutes(app)

	//Starts the HTTP server on the port, exits if it fails
	log.Fatal(app.Listen("localhost:3000"))
}
