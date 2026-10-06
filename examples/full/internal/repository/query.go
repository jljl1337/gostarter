package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jljl1337/gostarter/pkg/core/repository"
)

type Queries struct {
	repository.Queries
}

func NewQueriesFromEnv(ctx context.Context, db *sql.DB) (*Queries, error) {
	queries, err := repository.NewQueriesFromEnv(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create queryer: %w", err)
	}

	return &Queries{
		Queries: *queries,
	}, nil
}
