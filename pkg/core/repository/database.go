package repository

import (
	"database/sql"
	"fmt"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

// VarSyntax defines the syntax used for variable placeholders in SQL queries.
type VarSyntax string

const (
	DollarSign   VarSyntax = "$"
	QuestionMark VarSyntax = "?"
	ColonNamed   VarSyntax = ":"
	AtSign       VarSyntax = "@"
)

type Database struct {
	DB        *sql.DB
	VarSyntax VarSyntax
}

func (d *Database) Ping() error {
	if d.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	return d.DB.Ping()
}

func (d *Database) Close() error {
	if d.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	return d.DB.Close()
}

func NewDatabaseFromEnv() (*Database, error) {
	switch env.Variables.DatabaseEngine {
	case env.DatabaseDriverPostgreSQL:
		return NewPostgreSQLDBFromEnv()
	case env.DatabaseDriverSQLite:
		return NewSQLiteDBFromEnv()
	}

	return nil, fmt.Errorf("unsupported database type: %s", env.Variables.DatabaseEngine)
}
