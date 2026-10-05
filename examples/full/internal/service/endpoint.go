package service

import (
	"database/sql"

	"github.com/jljl1337/gostarter/pkg/core/queue"
	"github.com/jljl1337/gostarter/pkg/shared/role"
)

type EndpointServiceConfig struct {
	DB           *sql.DB
	IDGenerator  func() string
	QueueManager *queue.QueueManager
	RoleManager  *role.RoleManager
}

type EndpointService struct {
	db           *sql.DB
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
