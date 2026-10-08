package db

import (
	"database/sql"
	"fmt"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

/*
NewDBFromEnv creates a new database connection based on the environment
variables defined in the env package. It returns a pointer to sql.DB and an
error if any occurs during the connection process.
*/
func NewDBFromEnv() (*sql.DB, error) {
	switch env.DatabaseDriver {
	case env.DatabaseDriverPostgreSQL:
		return NewPostgreSQLDBFromEnv()

	case env.DatabaseDriverTurso:
		return NewTursoDBFromEnv()

	default:
		return nil, fmt.Errorf("unsupported database type: %s", env.DatabaseDriver)
	}
}
