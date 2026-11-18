package database

import "context"

type Controller struct {
	service *Service
}

func (c *Controller) Name() string {
	return "database"
}

func (c *Controller) Start() error {
	return c.service.pingAll()
}

func (c *Controller) Stop(ctx context.Context) error {
	return c.service.closeAll(ctx)
}
