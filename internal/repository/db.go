package repository

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// New creates a new database connection using the provided connection string.
// Returns a clear error if the connection fails.
func New(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
