package service

import (
	"github.com/jljl1337/gostarter/pkg/core/queue"
	"github.com/jljl1337/gostarter/pkg/shared/role"
	"github.com/jmoiron/sqlx"
)

type EndpointService struct {
	db           *sqlx.DB
	idGenerator  func() string
	queueManager *queue.QueueManager
	roleManager  *role.RoleManager
}

func NewEndpointService(db *sqlx.DB, idGenerator func() string, queueManager *queue.QueueManager, roleManager *role.RoleManager) *EndpointService {
	return &EndpointService{
		db:           db,
		idGenerator:  idGenerator,
		queueManager: queueManager,
		roleManager:  roleManager,
	}
}
