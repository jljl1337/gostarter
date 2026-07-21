package cron

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
	queries := repository.NewQueries(s.db)
	now := generator.NowISO8601()
	return queries.DeleteSessionByExpiresAt(ctx, now)
}
