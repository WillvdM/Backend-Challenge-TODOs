package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/WillvdM/Backend-Challenge-TODOs/models"
    //  

)

// Create TODO handles POST /todos
func CreateTodo (c *fiber.Ctx) error {
    var input models.TodoInput

    // JSON body is parsed into TodoInput structure
    if err := c.BodyParser(&input); err!=nil {
        return c.Status(400).JSON(fiber.Map{
            "error" : "Invalid request body",
        
        })
    }

    // Input is returned currently as confirmation 
    return c.JSON(fiber.Map {
        "message" : "Todo created successfully",
        "todo":    input,
    })
}

// GetTodoByID handles GET /todos/:id
func GetTodoByID(c *fiber.Ctx) error {

    // Echoes id parameter
    id :=c.Params("id")

    return c.JSON(fiber.Map {
        "id": id,
    })
}

// Delete TODO handles DELETE /todos/:id
func DeleteTodo (c *fiber.Ctx) error {
    return c.SendStatus(204)
}