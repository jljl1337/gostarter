package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgreSQLDB creates a new PostgreSQL database connection using the
// provided URL. It returns a pointer to sql.DB and an error if any occurs
// during the connection process.
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
