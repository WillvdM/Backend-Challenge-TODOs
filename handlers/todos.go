package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/WillvdM/Backend-Challenge-TODOs/models"
     "github.com/WillvdM/Backend-Challenge-TODOs/db"


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

        var id int
    err := db.DB.QueryRow(
        "INSERT INTO todos (title, completed) VALUES ($1, $2) RETURNING id",
        input.Title,
        input.Completed,
    ).Scan (&id)

        if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": err.Error(),
    })
}
    
    // Return created todo
    todo := map[string]interface{} {
        "id": id,
        "title": input.Title,
        "completed": input.Completed,
    }
    return c.Status(201).JSON(todo)
}



// GetTodoByID handles GET /todos/:id
func GetTodos(c *fiber.Ctx) error {
    rows, err :=db.DB.Query("SELECT id , title, completed FROM todos ")
   if err != nil {
    return c.Status(500).JSON(fiber.Map{"error": err.Error()})
   }
   defer rows.Close()

   // Collect the results
   todos := []map[string]interface{}{}
   for rows.Next() {
    var id int
    var title string
    var completed bool

    err := rows.Scan(&id, &title, &completed)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    todo := map[string]interface{} {
        "id": id,
        "title": title,
        "completed": completed,
    }
    todos = append (todos, todo)
   }
   return c.JSON(todos)
}

