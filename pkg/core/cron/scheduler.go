package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/log"
)

type Job struct {
	CronSchedule string
	WithSeconds  bool
	Task         func(context.Context)
}

type Scheduler struct {
	scheduler gocron.Scheduler
}

func NewScheduler() (*Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &Scheduler{
		scheduler: scheduler,
	}, nil
}

func (s *Scheduler) AddJob(job Job) error {
	_, err := s.scheduler.NewJob(
		gocron.CronJob(
			job.CronSchedule,
			job.WithSeconds,
		),
		gocron.NewTask(job.Task),
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

func (s *Scheduler) Shutdown(ctx context.Context) error {
	return s.scheduler.ShutdownWithContext(ctx)
}

func DefaultSchedulerJobFromEnv(schedulerService *service.SchedulerService) []Job {
	var jobList []Job

	// Database backup job (only for Turso)
	if env.DatabaseDriver == env.DatabaseDriverTurso {
		if env.TursoBackupEnabled {
			job := TursoBackupJob(schedulerService)
			jobList = append(jobList, job)
		} else {
			log.Warn("Database backup cron job is available but not scheduled")
		}
	} else {
		log.Info("Database backup is only available for Turso, skip adding cron job")
	}

	// Session cleanup job
	if env.SessionCleanupEnabled {
		job := SessionCleanupJob(schedulerService)
		jobList = append(jobList, job)
	} else {
		log.Warn("Session cleanup cron job not scheduled")
	}
	return jobList
}

func TursoBackupJob(schedulerService *service.SchedulerService) Job {
	return Job{
		CronSchedule: env.TursoBackupCronSchedule,
		WithSeconds:  false,
		Task: func(ctx context.Context) {
			log.Info("Starting database backup")

			start := time.Now()

			if err := schedulerService.BackupTursoDBFromEnv(ctx); err != nil {
				log.Errorf("Failed to backup database: %v", err)
				return
			}

			log.Infof("Database backup completed in %s", time.Since(start).String())
		},
	}
}

func SessionCleanupJob(schedulerService *service.SchedulerService) Job {
	return Job{
		CronSchedule: env.SessionCleanupCronSchedule,
		WithSeconds:  false,
		Task: func(ctx context.Context) {
			log.Info("Starting session cleanup")

			start := time.Now()

			rows, err := schedulerService.CleanupExpiredSessions(ctx)
			if err != nil {
				log.Errorf("Failed to cleanup expired sessions: %v", err)
				return
			}

			log.Infof("Session cleanup completed in %s, %d sessions deleted", time.Since(start).String(), rows)
		},
	}
}
