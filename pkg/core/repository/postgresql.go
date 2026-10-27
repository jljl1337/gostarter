package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

/*
NewDBFromEnv creates a new database connection based on the environment
variables defined in the env package. It returns a pointer to sql.DB and an
error if any occurs during the connection process.
*/
func NewPostgreSQLDBFromEnv() (*Database, error) {
	return NewPostgreSQLDB(env.Variables.PostgreSQLURL)
}

/*
NewPostgreSQLDB creates a new PostgreSQL database connection using the
provided URL. It returns a pointer to sql.DB and an error if any occurs
during the connection process.
*/
func NewPostgreSQLDB(url string) (*Database, error) {
	if url == "" {
		return nil, fmt.Errorf("PostgreSQL URL is missing")
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	return &Database{
		DB:        db,
		VarSyntax: DollarSign,
	}, nil
}
