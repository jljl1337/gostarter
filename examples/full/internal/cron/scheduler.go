package cron

import (
	"context"

	"github.com/jljl1337/gostarter/pkg/core/cron"
	"github.com/jljl1337/gostarter/pkg/shared/log"

	"github.com/jljl1337/gostarter/examples/full/internal/env"
	"github.com/jljl1337/gostarter/examples/full/internal/service"
)

func DeleteNotesJob(schedulerService *service.SchedulerService) cron.Job {
	return cron.Job{
		CronSchedule: env.DeleteNotesCronSchedule,
		WithSeconds:  false,
		Task: func(ctx context.Context) {
			deleted, err := schedulerService.DeleteExpiredNotes(ctx)
			if err != nil {
				log.Errorf("Failed to delete expired notes: %v", err)
				return
			}
			log.Infof("Deleted %d expired notes", deleted)
		},
	}
}
