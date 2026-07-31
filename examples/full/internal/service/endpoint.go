package service

import "github.com/jmoiron/sqlx"

type EndpointService struct {
	db          *sqlx.DB
	idGenerator func() string
}

func NewEndpointService(db *sqlx.DB, idGenerator func() string) *EndpointService {
	return &EndpointService{
		db:          db,
		idGenerator: idGenerator,
	}
}
