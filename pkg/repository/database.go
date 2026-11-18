package repository

import (
	"context"
	"database/sql"
	"fmt"
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

func (d *Database) Close(ctx context.Context) error {
	if d.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	return d.DB.Close()
}
