package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

/*
NewDB creates a new database connection based on the provided parameters.
It returns a pointer to sqlx.DB and an error if any occurs during the connection
process.
*/
func NewDB(databaseDriver, sqliteDbPath, sqliteDbBusyTimeout, postgresURL string) (*sqlx.DB, error) {
	switch databaseDriver {
	case env.DatabaseDriverPostgreSQL:
		return NewPostgreSQLDB(postgresURL)

	case env.DatabaseDriverSQLite:
		return NewSQLiteDB(sqliteDbPath, sqliteDbBusyTimeout)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", databaseDriver)
	}
}
