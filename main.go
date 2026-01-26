package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
    "github.com/WillvdM/Backend-Challenge-TODOs/routes"
	"github.com/WillvdM/Backend-Challenge-TODOs/db"
)
func main() {
	app := fiber.New()

	db.Connect ()

	routes.SetupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}