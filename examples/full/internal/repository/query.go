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

func NewQueriesTx(ctx context.Context, db *repository.Database, opts *sql.TxOptions) (*Queries, func(), func() error, error) {
	gsQueries, rollback, commit, err := repository.NewQueriesTx(ctx, db, opts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create queries with transaction: %w", err)
	}

	queries := &Queries{
		Queries: *gsQueries,
	}

	return queries, rollback, commit, nil
}

func NewQueries(db *repository.Database) (*Queries, error) {
	gsQueries, err := repository.NewQueries(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create queries: %w", err)
	}

	queries := &Queries{
		Queries: *gsQueries,
	}

	return queries, nil
}
