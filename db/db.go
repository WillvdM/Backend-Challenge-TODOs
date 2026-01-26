package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect () {
	connStr := "host = localhost port = 5432 user = postrgres password = postgres dbname = TODOs sslmode = disable"
	var err error	
	DB, err = sql.Open ("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to PostgresSQL")
}
