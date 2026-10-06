package service

import (
	"context"
	"database/sql"

	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/generator"

	"github.com/jljl1337/gostarter/examples/full/internal/repository"
)

type SchedulerService struct {
	db *sql.DB
}

func NewSchedulerService(db *sql.DB) *SchedulerService {
	return &SchedulerService{
		db: db,
	}
}

func (s *SchedulerService) DeleteExpiredNotes(ctx context.Context) (int64, error) {
	queries, err := repository.NewQueriesFromEnv(ctx, s.db)
	if err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to create queries: %v", err)
	}
	defer queries.Close()

	if err = queries.BeginTx(ctx); err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to begin transaction: %v", err)
	}
	defer queries.RollbackTx(ctx)

	deadline := generator.MinutesBeforeNowISO8601(1)
	deleted, err := queries.DeleteNotesByUpdatedAt(ctx, deadline)
	if err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to delete expired notes: %v", err)
	}

	if err := queries.CommitTx(ctx); err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to commit transaction: %v", err)
	}

	return deleted, nil
}
