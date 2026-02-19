package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/WillvdM/Backend-Challenge-TODOs/internal/api/models"
	repository "github.com/WillvdM/Backend-Challenge-TODOs/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// UserHandler wraps a repository.
// Separates data access logic from business logic.
type UserHandler struct {
	Repo *repository.UserRepository
}

// NewUserHandler constructs a new UserHandler.
func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{Repo: repo}
}

// The CreateUser function adds a single user or multiple user records to the database.
// Provides a useful response message if the user was created.
func (handler *UserHandler) CreateUser(c *fiber.Ctx) error {

	var input []models.UserInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(input) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No user created",
		})
	}

	var createdIDs []string

	for _, user := range input {
		id, err := handler.Repo.Create(user)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		createdIDs = append(createdIDs, id)
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "User created",
		"ids":     createdIDs,
	})
}

// GetUsers function fecthes all users and lists them in JSON format.
func (handler *UserHandler) GetUsers(c *fiber.Ctx) error {

	users, err := handler.Repo.List()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(users)
}

// GetUserById function returns a user record that was searched by the user's ID.
func (handler *UserHandler) GetUserById(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := handler.Repo.GetById(id)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found"})
	} else if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(user)
}

// UpdateUser function updates a user's name, surname and username by using their ID as the parameter.
// Provides a useful message if a user has been updated.
func (handler *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var input models.UserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := handler.Repo.Update(id, input)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found"})
	} else if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "User updated"})
}

// The DeleteUser function allows a user to be deleted by their id.
// If deletion is hard, the user is permanently deleted. If deletion is soft, they are soft deleted (set deleted_at).
// Provides a useful response if a user has been soft deleted.
func (handler *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	err := handler.Repo.Delete(id)

	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "User deleted"})
}
