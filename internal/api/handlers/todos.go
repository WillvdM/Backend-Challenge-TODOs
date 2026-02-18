package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WillvdM/Backend-Challenge-TODOs/config"
	"github.com/WillvdM/Backend-Challenge-TODOs/internal/api/models"
	"github.com/WillvdM/Backend-Challenge-TODOs/internal/domain"
	"github.com/WillvdM/Backend-Challenge-TODOs/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// TodoHandler wraps TodoRepository.
type TodoHandler struct {
	Repo *repository.TodoRepository
}

// NewTodoHandler constructs a new TodoHandler instance with a repository reference.
func NewTodoHandler(repo *repository.TodoRepository) *TodoHandler {
	return &TodoHandler{Repo: repo}
}

// CreateTodos handles POST /todos
// Accepts an array of TODOs, validates inputs, parses the due_date and inserts each into the database.
func (todoHandler *TodoHandler) CreateTodos(c *fiber.Ctx) error {
	var inputs []struct {
		Title     string  `json:"title"`
		Completed bool    `json:"completed"`
		Assignee  string  `json:"assignee"`
		DueDate   *string `json:"due_date"`
	}

	if err := c.BodyParser(&inputs); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	inserted := []map[string]interface{}{}

	for _, input := range inputs {
		if strings.TrimSpace(input.Assignee) == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Assignee is required"})
		}

		var dueDate *time.Time
		if input.DueDate != nil && *input.DueDate != "" {
			parsed, err := time.Parse("02-01-2006", *input.DueDate)
			if err != nil {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid due_date format. Use dd-mm-yyyy"})
			}
			dueDate = &parsed
		}
		id, err := todoHandler.Repo.Create(domain.Todo{
			Title:     input.Title,
			Completed: &input.Completed,
			Assignee:  input.Assignee,
			DueDate:   dueDate,
		})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		inserted = append(inserted, fiber.Map{
			"id":        id,
			"title":     input.Title,
			"completed": input.Completed,
			"assignee":  input.Assignee,
		})
	}
	return c.Status(http.StatusCreated).JSON(inserted)
}

// GetTodos handles GET /todos
// Pagination, sorting is supported.
// Formats all timestamps before returning them as JSON.
func (todoHandler *TodoHandler) GetTodos(c *fiber.Ctx) error {
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if offset < 0 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	sortField := c.Query("sort", "id")
	order := strings.ToLower(c.Query("order", "asc"))

	allowed := false
	for _, f := range config.Config.SortFields {
		if f == sortField {
			allowed = true
			break
		}
	}
	if !allowed {
		sortField = "id"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	todos, err := todoHandler.Repo.List(offset, limit, sortField, order)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var totalCount int
	err = todoHandler.Repo.DB.QueryRow("SELECT COUNT(*) FROM todos").Scan(&totalCount)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	currentPage := (offset / limit) + 1
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	var todoResponses []models.TodoResponse

	for _, todo := range todos {
		var dueDateStr *string

		if todo.DueDate != nil {
			formatted := todo.DueDate.Format("02-01-2006")
			dueDateStr = &formatted
		}

		createdStr := todo.CreatedAt.Format("02-01-2006 15:04:05")
		updatedStr := todo.UpdatedAt.Format("02-01-2006 15:04:05")
		if todo.UpdatedAt.IsZero() {
			updatedStr = createdStr
		}

		var deletedStr *string
		if todo.DeletedAt != nil && *todo.DeletedAt != "" {
			t, err := time.Parse(time.RFC3339, *todo.DeletedAt)
			if err != nil {
				formatted := t.Format("02-01-2006 15:04:05")
				deletedStr = &formatted
			} else {
				deletedStr = todo.DeletedAt
			}
		}

		var completedAtStr *string
		if todo.Completed != nil && *todo.Completed && todo.CompletedAt != nil {
			formatted := todo.CompletedAt.Format("02-01-2006 15:04:05")
			completedAtStr = &formatted
		}
		todoResponses = append(todoResponses, models.TodoResponse{
			ID:          todo.ID,
			Title:       todo.Title,
			Completed:   todo.Completed != nil && *todo.Completed,
			Assignee:    todo.Assignee,
			DueDate:     dueDateStr,
			CreatedAt:   &createdStr,
			UpdatedAt:   &updatedStr,
			DeletedAt:   deletedStr,
			CompletedAt: completedAtStr,
		})
	}

	response := models.TodosResponse{
		Todos:       todoResponses,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	data, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Send(data)
}

// GetTodoByID handles GET /todos/:id
// Returns a single TODO by its ID.
func (todoHandler *TodoHandler) GetTodoByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid TODO id"})
	}

	todo, err := todoHandler.Repo.GetById(id)
	if err == sql.ErrNoRows {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "TODO not found",
		})
	} else if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(todo)
}

// GetExpiredTodo handles GET /todos/expired.
// Returns TODOs whose due_date has passed, and the TODO is not yet completed.
func (todoHandler *TodoHandler) GetExpiredTodo(c *fiber.Ctx) error {
	expired, err := todoHandler.Repo.GetExpired()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error()})
	}

	return c.JSON(expired)
}

// DeleteTodo handles DELETE /todos/:id
// Supports soft delete (flags as deleted_at) or hard delete (pemranently removes row)
func (todoHandler *TodoHandler) DeleteTodo(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid TODO id"})
	}

	err = todoHandler.Repo.Delete(id, config.Config.DeletionMode == config.DeletionHard)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "TODO deleted successfully",
	})
}

// UpdateTodo handles PATCH /todos/:id.
// Updates title, completed and due date, while setting timestamps as needed.
func (todoHandler *TodoHandler) UpdateTodo(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid TODO id"})
	}

	var input struct {
		Title     *string `json:"title,omitempty"`
		Completed *bool   `json:"completed,omitempty"`
		DueDate   *string `json:"due_date,omitempty"`
	}

	if len(c.Body()) != 0 {
		if err := c.BodyParser(&input); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
	}

	var dueDate *time.Time
	if input.DueDate != nil && *input.DueDate != "" {
		parsed, err := time.Parse("02-01-2006", *input.DueDate)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid due_date format. Use dd-mm-yyyy",
			})
		}
		dueDate = &parsed
	}

	todo := domain.Todo{
		DueDate: dueDate,
	}

	if input.Title != nil {
		todo.Title = *input.Title
	}
	if input.Completed != nil {
		todo.Completed = input.Completed
	}
	if err := todoHandler.Repo.Update(id, todo); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "TODO updated"})
}
