package service

import (
	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/shared/crypto"
	"github.com/jljl1337/gostarter/pkg/shared/role"
	"github.com/jljl1337/gostarter/pkg/shared/validation"
)

type EndpointServiceConfig struct {
	DB                *sqlx.DB
	IDGenerator       func() string
	HashingManager    *crypto.HashingManager
	ValidationManager *validation.ValidationManager
	RoleManager       *role.RoleManager
}

type EndpointService struct {
	db                *sqlx.DB
	idGenerator       func() string
	hashingManager    *crypto.HashingManager
	validationManager *validation.ValidationManager
	roleManager       *role.RoleManager
}

func NewEndpointService(config EndpointServiceConfig) *EndpointService {
	return &EndpointService{
		db:                config.DB,
		idGenerator:       config.IDGenerator,
		hashingManager:    config.HashingManager,
		validationManager: config.ValidationManager,
		roleManager:       config.RoleManager,
	}
}

func (s *EndpointService) NewID() string {
	return s.idGenerator()
}
