package migration

import "context"

type Controller struct {
	service *Service
}

func (c *Controller) Name() string {
	return "migration"
}

func (c *Controller) Start() error {
	return c.service.migrateAll(context.Background())
}

func (c *Controller) Stop(ctx context.Context) error {
	return nil
}
