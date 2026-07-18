package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/db"
	"github.com/jljl1337/gostarter/env"
	"github.com/jljl1337/gostarter/generator"
	"github.com/jljl1337/gostarter/repository"
)

func NewSchedulerFromEnv(dbInstance *sqlx.DB) (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Database backup job (only for SQLite)
	if env.DatabaseDriver == "sqlite" {
		if env.EnableSQLiteBackup {
			_, err = AddSQLiteBackupJob(scheduler, dbInstance)
			if err != nil {
				return nil, fmt.Errorf("failed to create cron job: %w", err)
			}
		} else {
			slog.Warn("SQLite Database backup cron job not scheduled")
		}
	} else {
		slog.Info("Database backup not available for PostgreSQL, skip adding cron job")
	}

	// Session cleanup job
	if env.EnableSessionCleanup {
		_, err = AddSessionCleanupJob(scheduler, dbInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to create session cleanup cron job: %w", err)
		}
	} else {
		slog.Warn("Session cleanup cron job not scheduled")
	}

	return scheduler, nil
}

func AddSQLiteBackupJob(scheduler gocron.Scheduler, dbInstance *sqlx.DB) (gocron.Job, error) {
	return scheduler.NewJob(
		gocron.CronJob(
			env.SQLiteBackupCronSchedule,
			false,
		),
		gocron.NewTask(
			func() {
				slog.Info("Starting database backup")

				start := time.Now()

				if err := db.BackupSQLiteDBFromEnv(dbInstance); err != nil {
					slog.Error("Failed to backup database: " + err.Error())
					return
				}

				slog.Info("Database backup completed in " + time.Since(start).String())
			},
		),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
}

func AddSessionCleanupJob(scheduler gocron.Scheduler, dbInstance *sqlx.DB) (gocron.Job, error) {
	return scheduler.NewJob(
		gocron.CronJob(
			env.SessionCleanupCronSchedule,
			false,
		),
		gocron.NewTask(
			func() {
				slog.Info("Starting session cleanup")

				start := time.Now()

				now := generator.NowISO8601()
				queries := repository.NewQueries(dbInstance)
				rows, err := queries.DeleteSessionByExpiresAt(context.Background(), now)
				if err != nil {
					slog.Error("Failed to cleanup expired sessions: " + err.Error())
					return
				}

				slog.Info(fmt.Sprintf("Session cleanup completed in %s, %d sessions deleted", time.Since(start).String(), rows))
			},
		),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
}
