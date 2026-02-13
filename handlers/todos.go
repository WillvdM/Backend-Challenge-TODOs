package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WillvdM/Backend-Challenge-TODOs/config"
	"github.com/WillvdM/Backend-Challenge-TODOs/db"
	"github.com/WillvdM/Backend-Challenge-TODOs/models"
	"github.com/gofiber/fiber/v2"
)

// CreateTodos handles CREATE /todos
// Referenced in routes/routes.go
func CreateTodos(c *fiber.Ctx) error {
	// Define the input structure request body
	var inputs []struct {
		Title     string  `json:"title"`
		Completed bool    `json:"completed"`
		Assignee  string  `json:"assignee"`
		DueDate   *string `json:"due_date"`
	}

	// Parse the JSON body into the input struct
	if err := c.BodyParser(&inputs); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Prepare a slice to hold the inserted todos.
	inserted := []map[string]interface{}{}

	// Loop through each input todo.
	for _, input := range inputs {

		// Validate if the required field is present
		if strings.TrimSpace(input.Assignee) == "" {

			// If not present, an error is returned.
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Assignee is required for all todos",
			})
		}

		// Parse due_date if provided into the dd-mm-yyyy format
		var dueDate *time.Time
		if input.DueDate != nil && *input.DueDate != "" {
			parsed, err := time.Parse("02-01-2006", *input.DueDate) // dd-mm-yyyy
			if err != nil {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid due_date format. Use dd-mm-yyyy", // dd-mm-yyyy is used
				})
			}
			dueDate = &parsed
		}

		// Insert each todo individually
		var id int

		// After inserting, JSON returns all the values that was inserted
		err := db.DB.QueryRow(`
            INSERT INTO todos (title, completed, assignee, due_date)
            VALUES ($1, $2, $3, $4)
            RETURNING id
        `, input.Title, input.Completed, input.Assignee, dueDate).Scan(&id)

		// If no rows were changed, the a server error is returned
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// DueDate is formatted as a string for the response (optional)
		dueDateStr := ""
		if dueDate != nil {
			dueDateStr = dueDate.Format("02-01-2006")
		}

		// Add inserted todo to response slice.
		inserted = append(inserted, map[string]interface{}{
			"id":        id,
			"title":     input.Title,
			"completed": input.Completed,
			"assignee":  input.Assignee,
			"due_date":  dueDateStr,
		})
	}

	// Step 4: Return all inserted todos in JSON format.
	return c.Status(http.StatusCreated).JSON(inserted)
}

// GetTodos handles GET /todos.
// Referenced in routes/routes.go.
// Fetches list of todos and returns them as JSON to the client.
// Query the db, process each row in the TodoResponse struct, collect them in a slice, the return results as JSON.
func GetTodos(c *fiber.Ctx) error {

	// Offset pagination is used to only return values from a specified point. Default is 0.
	// Prevents invalid input.
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// The limit is set to only return a specified number of data. Default is 10.
	// The limit is converted from string to int.
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if offset < 0 {
		offset = 0
	}

	// The max limit is 100.
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get requested sort field from query parameters , the default is id.
	sortField := c.Query("sort", "id")

	// Get requested sort order from query parameters, the default is ascending.
	order := strings.ToLower(c.Query("order", "asc"))

	// Sort field is validated against the whitelist that was specified in config.yaml.
	// Prevents SQL injection.
	allowed := false
	for _, f := range config.Config.SortFields {
		if f == sortField {
			allowed = true
			break
		}
	}

	// Ensures safe fallback to the default if a field is not allowed.
	if !allowed {
		sortField = "id"
	}

	// Sort order is validated
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// Queries the database with ORDER BY.
	// 'NULLS LAST' ensures that any null timestamp only appears at the bottom of the list.
	query := fmt.Sprintf(`
		SELECT id, title, completed, due_date, completed_at, created_at, assignee, updated_at
		FROM todos
		ORDER BY %s %s 
		LIMIT $1 OFFSET $2
	`, sortField, order)

	// The query is executed.
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {

		// On query failure, return an error.
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Rows are closed to free up resources.
	defer rows.Close()

	// Rows are scanned
	// A slice is declared to hold the results
	todos := []models.TodoResponse{}

	// A for-loop that iterates over each result set row.
	for rows.Next() {
		var todo models.TodoResponse

		// Columns must be scanned in the exact order that the SELECT statement specifies.
		// Pointer types is used for nullable timestamps, as this prevents runtime errors for columns that are NULL.
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.DueDate, &todo.CompletedAt, &todo.CreatedAt, &todo.Assignee, &todo.UpdatedAt)
		if err != nil {

			// A server error is returned on scan failure.
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Add data to the end of the todo array slice, without modifying the previous data.
		todos = append(todos, todo)
	}

	// Check for iteration errors after looping.
	if err := rows.Err(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Calculate the total number of table rows.
	var totalCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM todos").Scan(&totalCount)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Calculate the current page number.
	currentPage := (offset / limit) + 1

	// Caluculate the total number of pages.
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// The response object is constructed
	response := models.TodosResponse{
		Todos:       todos,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	// JSON response is printed neatly for readability using Marshal response.
	data, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Set the content-type header and send the response.
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Send(data)
}

// GetTodoByID handles GET /todo/:id
// Referenced in routes/routes.go
func GetTodoByID(c *fiber.Ctx) error {

	// Get URL id
	idParam := c.Params("id")

	// Convert from string to integer
	id, err := strconv.Atoi(idParam)
	if err != nil {

		// Return error if the TODO id is not a valid integer.
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	// Struct for the database results.
	var todo struct {
		ID        int
		Title     string
		Completed bool
		UpdatedAt time.Time
	}

	// The database is queried for the todo with the given id.
	err = db.DB.QueryRow(
		"SELECT id, title, completed, updated_at FROM TODOS WHERE id = $1",
		id,
	).Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.UpdatedAt)
	if err == sql.ErrNoRows {
		// Return error if no TODO with the given ID exists
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "TODO not found",
		})
	}

	if err != nil {
		// Return a database error.
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return the todos returned by the query as a JSON
	return c.JSON(fiber.Map{
		"id":        todo.ID,
		"title":     todo.Title,
		"completed": todo.Completed,
	})
}

// Refrenced in routes/routes.go
// GetExpiredTodos returns all todos where due_date has passed.
// and the todo is not completed.
func GetExpiredTodo(c *fiber.Ctx) error {

	// Query is executed to retrieve the rows of expired todos
	rows, err := db.DB.Query(`
		SELECT id, title, completed, assignee, due_date, completed_at, created_at, updated_at
		FROM todos
		WHERE due_date < NOW()
		AND (completed = false OR completed IS NULL)
	`)
	if err != nil {

		// Return a server error if the query fails.
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer rows.Close() // Ensures that rows are closed after the function exits, freeing up resources.

	// reponseTodo wraps the model with an ID field, as ID is not part of models.Todo.
	type responseTodo struct {
		ID int `json:"id"`
		models.Todo
	}

	// A slice to store all the expired todos returned by the query.
	var expiredTodos []responseTodo

	// Each row that the query returned is iterated over.
	for rows.Next() {

		// Temporary values are declared to scan DB values.
		var (
			id        int
			todo      models.Todo
			dbDueDate *time.Time // Converted to string later.
		)

		// Scan DB values
		err := rows.Scan(
			&id,
			&todo.Title,
			&todo.Completed,
			&todo.Assignee,
			&dbDueDate,
			&todo.CompletedAt,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {

			// Return a server error if the scanning fails.
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Convert due_date from *time.Time to *string
		if dbDueDate != nil {
			formatted := dbDueDate.Format("02-01-2006")
			todo.DueDate = &formatted
		}

		// The todo is appended to the expired slice.
		expiredTodos = append(expiredTodos, responseTodo{
			ID:   id,
			Todo: todo,
		})
	}

	// List all values in JSON format.
	return c.JSON(fiber.Map{
		"expired_todos": expiredTodos,
	})
}

// DeleteTodo handles DELETE /todos, deleting a todo by the todo ID.
// Referenced in routes/routes.go.
func DeleteTodo(c *fiber.Ctx) error {

	// Extract and validate TODO ID from the URL
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	// Holds an SQL execution result, can check last affected rows or id that was last inserted
	var result sql.Result

	// The deletion logic is executed, based on the config
	switch config.Config.DeletionMode {

	// Mark row as deleted without removing it
	case config.DeletionSoft:
		result, err = db.DB.Exec(
			"Update todos SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id,
		)

	// Permanently remove the row
	case config.DeletionHard:
		result, err = db.DB.Exec(
			"DELETE FROM todos WHERE id = $1", id,
		)

		// Safety: Refuse to delete anything if no deletion mode is specified
	default:
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Deletion mode is invalid",
		})
	}

	// Handle db execution errors and return the error in a JSON messages
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check affected rows
	rows, err := result.RowsAffected()
	if rows == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "TODO not found",
		})
	}

	// Success response in JSON if row is deleted.
	return c.JSON(fiber.Map{
		"message": "TODO deleted successfully",
	})
}

// Handles updating an existing todo item with PATCH
// Refernced in routes/routes.go
func UpdateTodo(c *fiber.Ctx) error {

	// Extract the todo ID from the URL path parameter, convert it to integer.
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {

		// Return error if ID is not an integer.
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid TODO id",
		})
	}

	// Define a struct to hold input fields for partial updates (only provided fields gets updated)
	var input struct {
		Title     *string `json:"title,omitempty"`
		Completed *bool   `json:"completed,omitempty"`
		DueDate   *string `json:"due_date,omitempty"` // string for parsing, leading to easier user input
	}

	// Allow empty PATCH body, but only parse request body if it is not empty.
	if len(c.Body()) != 0 {
		if err := c.BodyParser(&input); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
	}

	// Parse the due_date string into *time.Time if it is provided.
	var dueDate *time.Time
	if input.DueDate != nil && *input.DueDate != "" {
		parsed, err := time.Parse("02-01-2006", *input.DueDate) // Expects dd-mm-yyyy format.
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid due_date format. Use dd-mm-yyyy",
			})
		}
		dueDate = &parsed
	}

	// Build dynamic SET clause
	setParts := []string{}
	args := []interface{}{}
	argID := 1

	// If a title is provided, it will be included.
	if input.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argID))
		args = append(args, *input.Title)
		argID++
	}

	// If completed flag is provided, it will be included.
	if input.Completed != nil {
		setParts = append(setParts, fmt.Sprintf("completed = $%d", argID))
		args = append(args, *input.Completed)
		argID++
		setParts = append(setParts, fmt.Sprintf("completed_at = CASE WHEN $%d = true THEN NOW() ELSE NULL END", argID-1))
	}

	// If due_date is provided, it will be included.
	if dueDate != nil {
		setParts = append(setParts, fmt.Sprintf("due_date = $%d", argID))
		args = append(args, dueDate)
		argID++
	}

	// If no fields were provided in the body, an error is returned.
	if len(setParts) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields provided for update",
		})
	}

	// The final SQL query is dynamically constructed.
	query := fmt.Sprintf(`
		UPDATE todos
		SET %s
		WHERE id = $%d
		RETURNING title, completed_at, due_date, updated_at
	`, strings.Join(setParts, ", "), argID)

	// Append the TODO ID as the last argument for the WHERE clause.
	args = append(args, id)

	// These variables hold the returned values from the query.
	var title string
	var completedAt *time.Time
	var dbDueDate *time.Time
	var updatedAt *time.Time

	// Query is executed, and results scanned.
	err = db.DB.QueryRow(query, args...).Scan(&title, &completedAt, &dbDueDate, &updatedAt)
	if err == sql.ErrNoRows {

		// Return an error if no TODO was found with the given ID.
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "TODO not found",
		})
	}
	if err != nil {

		// Any other database error.
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Format completed_at timestamp as a string if it is present.
	var completedAtStr string
	if completedAt != nil {
		completedAtStr = completedAt.Format("02-01-2006 15:04:05")
	}

	// Format due_date timestamp as a string if it is present.
	var dueDateStr string
	if dbDueDate != nil {
		dueDateStr = dbDueDate.Format("02-01-2006")
	}

	// Format updated_at timestamp as a string if it is present
	var updatedAtStr string
	if updatedAt != nil {
		updatedAtStr = updatedAt.Format("02-01-2006 15:04:05")
	}

	// Return updated TODO fields in JSON format.
	return c.JSON(fiber.Map{
		"id":          id,
		"title":       title,
		"completedAt": completedAtStr,
		"due_date":    dueDateStr,
		"updated_at":  updatedAtStr,
	})
}

// CreateUsers function adds a user record to the database.
// Referenced in routes/routes.go.
func CreateUser(c *fiber.Ctx) error {

	// The inputs struct declares what fields need to be modified for the record/s to be created (name, surname, username).
	var inputs []struct {
		Name     string `json:"name"`
		Surname  string `json:"surname"`
		Username string `json:"username"`
	}

	if err := c.BodyParser(&inputs); err != nil {

		// Returns an error if the input is not in the correct format.
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Returns error if no name, surname or username values are provided.
	if len(inputs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No user created",
		})
	}

	// Stores succefully created IDs
	createdIDs := []string{}

	// for-loop that iterates over each user that was provided in the request body.
	for _, input := range inputs {

		var id string

		// Executes the query that inserts the values into the user table.
		err := db.DB.QueryRow(`
	INSERT INTO users (name, surname, username)
	VALUES ($1, $2, $3)
	RETURNING id
	`, input.Name, input.Surname, input.Username).Scan(&id)

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// The generated ID is collected for the response.
		createdIDs = append(createdIDs, id)
	}

	// Return all created user IDs.
	return c.Status(201).JSON(fiber.Map{
		"message": "User created",
		"ids":     createdIDs,
	})
}

// GetUsers function fecthes all users and lists them in JSON format.
// Referenced in routes/rotutes.go.
func GetUsers(c *fiber.Ctx) error {

	// Query that is executed to return all user records.
	rows, err := db.DB.Query(`
	SELECT id, name, surname, username, created_at, updated_at, deleted_at
	FROM users
	`)

	// Returns an error if the query could not be exectued.
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer rows.Close() // Closes the rows after the function exits, freeing up resources.

	// A users array is declared to hold the results.
	users := []models.User{}

	// The array is iterated over and the results are scanned to userd array.
	for rows.Next() {
		var u models.User

		// Scans all the values from the DB and place it in holders.
		if err := rows.Scan(&u.ID, &u.Name, &u.Surname, &u.Username, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt); err != nil {

			// Returns an error if the user array cannot be collected.
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Add the records scanned from the database to the end of the users array.
		users = append(users, u)
	}

	// Return a list of users in JSON.
	return c.JSON(users)
}

// GetUserById function returns a user record that was searched by the user's ID.
// Referenced in routes/routes.go.
func GetUserById(c *fiber.Ctx) error {
	id := c.Params("id")

	// A container is created for the User model structure that was declared in models/todo.go.
	var u models.User

	// Query that is executed to return the user record.
	err := db.DB.QueryRow(`
	SELECT id, name, surname, username, created_at, updated_at, deleted_at
	FROM users
	WHERE id=$1 
	`, id).Scan(&u.ID, &u.Name, &u.Surname, &u.Username, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt) // Scans the specified values from the database and places it in the "u" container.

	if err == sql.ErrNoRows {
		// Returns an error if no rows were returned.
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found"})
	} else if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return the user record ("u") in JSON.
	return c.JSON(u)
}

// UpdateUser function updates a user's name, surname and username by using their ID as the parameter.
func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	// Define a structure that indicates what fields need to be updated for the record to be updated.
	var input struct {
		Name     *string `json:"name"`
		Surname  *string `json:"surname"`
		Username *string `json:"username"`
	}
	if err := c.BodyParser(&input); err != nil {

		// Return an error if the updated changes are not sent in the correct JSON format.
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	//Fetch the current user and when the user was updated.
	var currentUsername string
	var updatedAt time.Time

	// Query that is executed to gather the rows that must be updated.
	// The results are scanned into the DB.
	err := db.DB.QueryRow(`
	SELECT username, updated_at
	FROM users 
	WHERE id=$1
	`, id).Scan(&currentUsername, &updatedAt)
	if err == sql.ErrNoRows {

		// Returns an error if ID that was entered is not found in the DB.
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found"})
	} else if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Build dynamic SET clause and parameter list for partial updates.
	setParts := []string{}
	args := []interface{}{}
	argID := 1

	// If the name of the user changes, the name field must be changed to the new value.
	if input.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argID))
		args = append(args, *input.Name)
		argID++ // Each time a field is added, the placholder after that is used. $1 is used here.
	}

	// If the surname of the user changes, the "surname" field must be changed to the new value.
	if input.Surname != nil {
		setParts = append(setParts, fmt.Sprintf("surname = $%d", argID))
		args = append(args, *input.Surname)
		argID++ // $2 used here.
	}

	if input.Username != nil && *input.Username != currentUsername {
		if time.Since(updatedAt) < 24*time.Hour { // Set the time a user must wait before updating their username (24 hours).

			// Returns an error if the user tries to update their username without waiting 24 hours after previous update.
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Username can only be updated once every 24 hours",
			})
		}

		// If the username of the user changes, the username record must be changed to new value.
		setParts = append(setParts, fmt.Sprintf("username = $%d", argID))
		args = append(args, *input.Username)
		argID++ // $3 used here.
	}

	// If the updated fields are the same as the current fields, a message is returned without the records being changed.
	if len(setParts) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields to update",
		})
	}

	//Update the updated_at timestamp
	setParts = append(setParts, ("updated_at = NOW()"))

	// Execute the query to update the user table with the new data.
	query := fmt.Sprintf(`
	UPDATE users
	SET %s WHERE id=$%d
	`, strings.Join(setParts, ", "), argID) // The parts that were set earlier (name, surname, username) are joined.
	args = append(args, id)

	_, err = db.DB.Exec(query, args...) // "args" values and  are passed in.
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return the success message in JSON if the user was updated
	return c.JSON(fiber.Map{"message": "User updated"})
}

// The DeleteUser function allows a user to be soft deleted by their id.
// Referenced in routes/routes.go.
func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	// Query that is executed that deletes a user.
	res, err := db.DB.Exec(`
		UPDATE users
		SET deleted_at = NOW()
		WHERE id=$1
		`, id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {

		// Returns an error if no rows were soft deleted.
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Returns a success message in JSON if the user was soft deleted.
	return c.JSON(fiber.Map{"message": "User deleted"})
}
