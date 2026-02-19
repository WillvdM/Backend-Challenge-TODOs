package user

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// UserRepository handles all database operations related to user.
type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository creates and returns a new UserRepository instance.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// Create inserts a new user in the database.
func (repo *UserRepository) Create(user UserInput) (string, error) {
	var id string
	err := repo.DB.QueryRow(`
	INSERT INTO users (name, surname, username)
	VALUES ($1, $2, $3)
	RETURNING id
	`, user.Name, user.Surname, user.Username).Scan(&id)
	return id, err
}

// List gets all users and lists them in JSON format.
func (repo *UserRepository) List() ([]User, error) {
	rows, err := repo.DB.Query(`
	SELECT id, name, surname, username, created_at, updated_at, deleted_at
	FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Surname, &user.Username, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// GetById retrieves a single user by their ID and returns them in JSON format.
func (repo *UserRepository) GetById(id string) (User, error) {
	var user User
	err := repo.DB.QueryRow(`
	SELECT id, name, surname, username, created_at, updated_at, deleted_at
	FROM users
	WHERE id = $1
	`, id).Scan(&user.ID, &user.Name, &user.Surname, &user.Username, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	return user, err
}

// Update allows the provided fields to be udpated by specifying the TODO id.
// A time is set that a user must wait before updating their username (24 hours).
// Returns an error if the user tries to update their username without waiting 24 hours after previous update.
func (repo *UserRepository) Update(id string, input UserInput) error {
	var setParts []string
	var args []interface{}
	argID := 1

	var lastUsernameUpdate sql.NullTime
	err := repo.DB.QueryRow("SELECT updated_at FROM users WHERE id=$1", id).Scan(&lastUsernameUpdate)
	if err != nil {
		return err
	}

	if input.Username != "" && lastUsernameUpdate.Valid {
		if time.Since(lastUsernameUpdate.Time) < 24*time.Hour {
			return fmt.Errorf("username can only be updated once every 24 hours")
		}
	}

	if input.Name != "" {
		setParts = append(setParts, `name = $`+strconv.Itoa(argID))
		args = append(args, input.Name)
		argID++
	}

	if input.Surname != "" {
		setParts = append(setParts, `surname = $`+strconv.Itoa(argID))
		args = append(args, input.Surname)
		argID++
	}

	if input.Username != "" {
		setParts = append(setParts, `username = $`+strconv.Itoa(argID))
		args = append(args, input.Username)
		argID++
	}

	if len(setParts) == 0 {
		return nil
	}

	query := "UPDATE users SET " + strings.Join(setParts, ", ") + ", updated_at = NOW() WHERE id=$" + strconv.Itoa(argID)
	args = append(args, id)

	res, err := repo.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete removes a user by their id.
func (repo *UserRepository) Delete(id string) error {
	res, err := repo.DB.Exec(`
	UPDATE users
	set deleted_at = NOW()
	WHERE id=$1	
	`, id)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
