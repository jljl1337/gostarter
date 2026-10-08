package service

import (
	"context"
	"database/sql"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/db"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
)

type SchedulerService struct {
	db *sql.DB
}

func NewSchedulerService(db *sql.DB) *SchedulerService {
	return &SchedulerService{
		db: db,
	}
}

func (s *SchedulerService) BackupTursoDBFromEnv(ctx context.Context) error {
	return db.BackupTursoDBFromEnv(ctx, s.db)
}

func (s *SchedulerService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	queries, err := repository.NewQueriesFromEnv(ctx, s.db)
	if err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to create queries: %v", err)
	}
	defer queries.Close()

	if err = queries.BeginTx(ctx); err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to begin transaction: %v", err)
	}
	defer queries.RollbackTx(ctx)

	now := generator.NowISO8601()
	deleted, err := queries.DeleteSessionByExpiresAt(ctx, now)
	if err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to delete expired sessions: %v", err)
	}

	if err := queries.CommitTx(ctx); err != nil {
		return 0, NewServiceErrorf(ErrCodeInternal, "failed to commit transaction: %v", err)
	}

	return deleted, nil
}
