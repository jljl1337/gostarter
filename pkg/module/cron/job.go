package cron

import (
	"context"

	"github.com/go-co-op/gocron/v2"
)

type Job struct {
	definition gocron.JobDefinition
	task       func(context.Context)
}

type CronSchedule struct {
	CronTab     string
	WithSeconds bool
}

type JobOption func(*Job) error

func WithCronSchedule(cronSchedule CronSchedule) JobOption {
	return func(job *Job) error {
		job.definition = gocron.CronJob(cronSchedule.CronTab, cronSchedule.WithSeconds)
		return nil
	}
}
