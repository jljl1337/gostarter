package service

import (
	"context"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/db"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
	"github.com/jmoiron/sqlx"
)

type SchedulerService struct {
	db *sqlx.DB
}

func NewSchedulerService(db *sqlx.DB) *SchedulerService {
	return &SchedulerService{
		db: db,
	}
}

func (s *SchedulerService) BackupSQLiteDBFromEnv(ctx context.Context) error {
	return db.BackupSQLiteDBFromEnv(s.db)
}

func (s *SchedulerService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	queries := repository.NewQueries(tx)

	now := generator.NowISO8601()
	deleted, err := queries.DeleteSessionByExpiresAt(ctx, now)
	if err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to delete expired sessions: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to commit transaction: %v", err)
	}

	return deleted, nil
}
