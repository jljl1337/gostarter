package service

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/generator"

	"github.com/jljl1337/gostarter/examples/full/internal/repository"
)

type SchedulerService struct {
	db *sqlx.DB
}

func NewSchedulerService(db *sqlx.DB) *SchedulerService {
	return &SchedulerService{
		db: db,
	}
}

func (s *SchedulerService) DeleteExpiredNotes(ctx context.Context) (int64, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	queries := repository.NewQueries(tx)

	deadline := generator.MinutesBeforeNowISO8601(1)
	deleted, err := queries.DeleteNotesByUpdatedAt(ctx, deadline)
	if err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to delete expired notes: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to commit transaction: %v", err)
	}

	return deleted, nil
}
