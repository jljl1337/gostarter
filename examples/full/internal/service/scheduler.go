package service

import (
	"context"

	gsRepository "github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/generator"

	"github.com/jljl1337/gostarter/examples/full/internal/repository"
)

type SchedulerService struct {
	db *gsRepository.Database
}

func NewSchedulerService(db *gsRepository.Database) *SchedulerService {
	return &SchedulerService{
		db: db,
	}
}

func (s *SchedulerService) DeleteExpiredNotes(ctx context.Context) (int64, error) {
	queries, rollback, commit, err := repository.NewQueriesTx(ctx, s.db, nil)
	if err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to create queries with transaction: %w", err)
	}
	defer rollback()

	deadline := generator.MinutesBeforeNowISO8601(1)
	deleted, err := queries.DeleteNotesByUpdatedAt(ctx, deadline)
	if err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to delete expired notes: %w", err)
	}

	if err := commit(); err != nil {
		return 0, service.NewServiceErrorf(service.ErrCodeInternal, "failed to commit transaction: %w", err)
	}

	return deleted, nil
}
