package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

	//Global database connection pool
var DB *sql.DB

	// Connect initializes and verifies the PostgreSQL connection
func Connect () {
	connStr := "host = localhost port = 5432 user = postgres password = postgres dbname = TODOs sslmode = disable"
	var err error	

	// sql.Open prepares the database connection
	DB, err = sql.Open ("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Tests the connection to the database
	if err = DB.Ping(); err!= nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to PostgresSQL")
}
