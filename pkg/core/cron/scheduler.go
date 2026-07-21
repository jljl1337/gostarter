package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/db"
	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
	"github.com/jljl1337/gostarter/pkg/shared/log"
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
			log.Warn("SQLite Database backup cron job not scheduled")
		}
	} else {
		log.Info("Database backup not available for PostgreSQL, skip adding cron job")
	}

	// Session cleanup job
	if env.EnableSessionCleanup {
		_, err = AddSessionCleanupJob(scheduler, dbInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to create session cleanup cron job: %w", err)
		}
	} else {
		log.Warn("Session cleanup cron job not scheduled")
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
				log.Info("Starting database backup")

				start := time.Now()

				if err := db.BackupSQLiteDBFromEnv(dbInstance); err != nil {
					log.Errorf("Failed to backup database: %v", err)
					return
				}

				log.Infof("Database backup completed in %s", time.Since(start).String())
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
				log.Info("Starting session cleanup")

				start := time.Now()

				now := generator.NowISO8601()
				queries := repository.NewQueries(dbInstance)
				rows, err := queries.DeleteSessionByExpiresAt(context.Background(), now)
				if err != nil {
					log.Errorf("Failed to cleanup expired sessions: %v", err)
					return
				}

				log.Infof("Session cleanup completed in %s, %d sessions deleted", time.Since(start).String(), rows)
			},
		),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
}
