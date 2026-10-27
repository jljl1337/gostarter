package service

import (
	"github.com/jljl1337/gostarter/pkg/core/queue"
	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/role"
)

type EndpointServiceConfig struct {
	DB           *repository.Database
	IDGenerator  func() string
	QueueManager *queue.QueueManager
	RoleManager  *role.RoleManager
}

type EndpointService struct {
	db           *repository.Database
	idGenerator  func() string
	queueManager *queue.QueueManager
	roleManager  *role.RoleManager
}

func NewEndpointService(config EndpointServiceConfig) *EndpointService {
	return &EndpointService{
		db:           config.DB,
		idGenerator:  config.IDGenerator,
		queueManager: config.QueueManager,
		roleManager:  config.RoleManager,
	}
}
