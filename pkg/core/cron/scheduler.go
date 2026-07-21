package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/jljl1337/gostarter/pkg/core/service/cron"
	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/log"
)

type Scheduler struct {
	scheduler        gocron.Scheduler
	schedulerService *cron.SchedulerService
}

func NewSchedulerFromEnv(schedulerService *cron.SchedulerService) (*Scheduler, error) {
	scheduler, err := NewScheduler(schedulerService)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Database backup job (only for SQLite)
	if env.DatabaseDriver == "sqlite" {
		if env.EnableSQLiteBackup {
			err = AddSQLiteBackupJob(scheduler)
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
		err = AddSessionCleanupJob(scheduler)
		if err != nil {
			return nil, fmt.Errorf("failed to create session cleanup cron job: %w", err)
		}
	} else {
		log.Warn("Session cleanup cron job not scheduled")
	}
	return scheduler, nil
}

func NewScheduler(schedulerService *cron.SchedulerService) (*Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &Scheduler{
		scheduler:        scheduler,
		schedulerService: schedulerService,
	}, nil
}

func (s *Scheduler) AddJobWithSeconds(cronSchedule string, task func(context.Context)) error {
	return s.AddJob(cronSchedule, true, task)
}

func (s *Scheduler) AddJobWithoutSeconds(cronSchedule string, task func(context.Context)) error {
	return s.AddJob(cronSchedule, false, task)
}

func (s *Scheduler) AddJob(cronSchedule string, withSeconds bool, task func(context.Context)) error {
	_, err := s.scheduler.NewJob(
		gocron.CronJob(
			cronSchedule,
			withSeconds,
		),
		gocron.NewTask(task),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	return nil
}

func (s *Scheduler) Start() {
	s.scheduler.Start()
}

func (s *Scheduler) Shutdown() error {
	return s.scheduler.Shutdown()
}

func AddSQLiteBackupJob(scheduler *Scheduler) error {
	return scheduler.AddJobWithoutSeconds(env.SQLiteBackupCronSchedule, func(ctx context.Context) {
		log.Info("Starting database backup")

		start := time.Now()

		if err := scheduler.schedulerService.BackupSQLiteDBFromEnv(ctx); err != nil {
			log.Errorf("Failed to backup database: %v", err)
			return
		}

		log.Infof("Database backup completed in %s", time.Since(start).String())
	})
}

func AddSessionCleanupJob(scheduler *Scheduler) error {
	return scheduler.AddJobWithoutSeconds(env.SessionCleanupCronSchedule, func(ctx context.Context) {
		log.Info("Starting session cleanup")

		start := time.Now()

		rows, err := scheduler.schedulerService.CleanupExpiredSessions(ctx)
		if err != nil {
			log.Errorf("Failed to cleanup expired sessions: %v", err)
			return
		}

		log.Infof("Session cleanup completed in %s, %d sessions deleted", time.Since(start).String(), rows)
	})
}
