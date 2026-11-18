package cron

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron/v2"
)

type Service struct {
	scheduler gocron.Scheduler
}

func NewService() (*Service, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &Service{
		scheduler: scheduler,
	}, nil
}

func (s *Service) AddJob(task func(context.Context), optionList ...JobOption) error {
	job := &Job{
		task: task,
	}

	for _, option := range optionList {
		if err := option(job); err != nil {
			return fmt.Errorf("failed to apply job option: %w", err)
		}
	}

	_, err := s.scheduler.NewJob(
		job.definition,
		gocron.NewTask(job.task),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	return nil
}

func (s *Service) start() {
	s.scheduler.Start()
}

func (s *Service) stop(ctx context.Context) error {
	return s.scheduler.ShutdownWithContext(ctx)
}
