package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

/*
NewPostgreSQLDBFromEnv creates a new database connection based on the
environment variables defined in the env package. It returns a pointer to
sql.DB and an error if any occurs during the connection process.
*/
func NewPostgreSQLDBFromEnv() (*sql.DB, error) {
	return NewPostgreSQLDB(env.PostgreSQLURL)
}

/*
NewPostgreSQLDB creates a new PostgreSQL database connection using the
provided URL. It returns a pointer to sql.DB and an error if any occurs
during the connection process.
*/
func NewPostgreSQLDB(url string) (*sql.DB, error) {
	if url == "" {
		return nil, fmt.Errorf("PostgreSQL URL is missing")
	}
	return sql.Open("pgx", url)
}
