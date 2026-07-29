package gostarter

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/shared/db"
	"github.com/jljl1337/gostarter/pkg/shared/env"
)

/*
NewDBFromEnv creates a new database connection based on the environment
variables defined in the env package. It returns a pointer to sqlx.DB and an
error if any occurs during the connection process.
*/
func NewDBFromEnv() (*sqlx.DB, error) {
	switch env.DatabaseDriver {
	case env.DatabaseDriverPostgreSQL:
		return db.NewPostgreSQLDBFromEnv()

	case env.DatabaseDriverSQLite:
		return db.NewSQLiteDBFromEnv()

	default:
		return nil, fmt.Errorf("unsupported database type: %s", env.DatabaseDriver)
	}
}
