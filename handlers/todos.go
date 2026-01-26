package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/WillvdM/Backend-Challenge-TODOs/models"
    //  

)
func CreateTodo (c *fiber.Ctx) error {
    var input models.TodoInput

    if err := c.BodyParser(&input); err!=nil {
        return c.Status(400).JSON(fiber.Map{
            "error" : "Invalid request body",
        
        })
    }

    return c.JSON(fiber.Map {
        "message" : "Todo created successfully",
        "todo":    input,
    })
}

func GetTodoByID(c *fiber.Ctx) error {
    id :=c.Params("id")

    return c.JSON(fiber.Map {
        "id": id,
    })
}

func DeleteTodo (c *fiber.Ctx) error {
    return c.SendStatus(204)
}