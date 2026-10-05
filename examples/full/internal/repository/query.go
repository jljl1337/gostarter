package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jljl1337/gostarter/pkg/core/repository"
)

type Queries struct {
	repository.Queryer
}

func NewQueries(ctx context.Context, db *sql.DB) (*Queries, error) {
	queryer, err := repository.NewQueryerFromEnv(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create queryer: %w", err)
	}

	return &Queries{
		Queryer: *queryer,
	}, nil
}
