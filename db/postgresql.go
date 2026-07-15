package db

import (
	"fmt"

	"github.com/jljl1337/gostarter/env"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

/*
NewDBFromEnv creates a new database connection based on the environment
variables defined in the env package. It returns a pointer to sqlx.DB and an
error if any occurs during the connection process.
*/
func NewPostgreSQLDBFromEnv() (*sqlx.DB, error) {
	return NewPostgreSQLDB(env.PostgresURL)
}

/*
NewPostgreSQLDB creates a new PostgreSQL database connection using the
provided URL. It returns a pointer to sqlx.DB and an error if any occurs
during the connection process.
*/
func NewPostgreSQLDB(url string) (*sqlx.DB, error) {
	if url == "" {
		return nil, fmt.Errorf("PostgreSQL URL is missing")
	}
	return sqlx.Open("postgres", url)
}
