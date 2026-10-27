package service

import (
	"context"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
)

type SchedulerService struct {
	db *repository.Database
}

func NewSchedulerService(db *repository.Database) *SchedulerService {
	return &SchedulerService{
		db: db,
	}
}

func (s *SchedulerService) BackupSQLiteDBFromEnv(ctx context.Context) error {
	return repository.BackupSQLiteDBFromEnv(ctx, s.db)
}

func (s *SchedulerService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	queries, rollback, commit, err := repository.NewQueriesTx(ctx, s.db, nil)
	if err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to create queries with transaction: %w", err)
	}
	defer rollback()

	now := generator.NowISO8601()
	deleted, err := queries.DeleteSessionByExpiresAt(ctx, now)
	if err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to delete expired sessions: %w", err)
	}

	if err := commit(); err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to commit transaction: %w", err)
	}

	return deleted, nil
}
