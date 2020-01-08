package repo

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const Statement = "SELECT balance_update($1, $2)"

func OpenDB(dbUser, dbName, dbPass string) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}
