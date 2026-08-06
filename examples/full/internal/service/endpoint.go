package service

import (
	"github.com/jljl1337/gostarter/pkg/core/queue"
	"github.com/jmoiron/sqlx"
)

type EndpointService struct {
	db           *sqlx.DB
	idGenerator  func() string
	queueManager *queue.QueueManager
}

func NewEndpointService(db *sqlx.DB, idGenerator func() string, queueManager *queue.QueueManager) *EndpointService {
	return &EndpointService{
		db:           db,
		idGenerator:  idGenerator,
		queueManager: queueManager,
	}
}
