package queue

import "context"

type Controller struct {
	service *Service
}

func (c *Controller) Name() string {
	return "queue"
}

func (c *Controller) Start() error {
	return c.service.start()
}

func (c *Controller) Stop(ctx context.Context) error {
	return c.service.stop(ctx)
}
