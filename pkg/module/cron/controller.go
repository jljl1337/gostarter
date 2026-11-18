package cron

import "context"

type Controller struct {
	service *Service
}

func (c *Controller) Name() string {
	return "cron"
}

func (c *Controller) Start() error {
	c.service.start()
	return nil
}

func (c *Controller) Stop(ctx context.Context) error {
	return c.service.stop(ctx)
}
