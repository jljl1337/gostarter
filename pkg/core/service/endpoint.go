package service

import (
	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/shared/crypto"
	"github.com/jljl1337/gostarter/pkg/shared/role"
	"github.com/jljl1337/gostarter/pkg/shared/validation"
)

type EndpointService struct {
	db                *sqlx.DB
	idGenerator       func() string
	hashingManager    *crypto.HashingManager
	validationManager *validation.ValidationManager
	roleManager       *role.RoleManager
}

func NewEndpointService(db *sqlx.DB, idGenerator func() string, hashingManager *crypto.HashingManager, validationManager *validation.ValidationManager, roleManager *role.RoleManager) *EndpointService {
	return &EndpointService{
		db:                db,
		idGenerator:       idGenerator,
		hashingManager:    hashingManager,
		validationManager: validationManager,
		roleManager:       roleManager,
	}
}

func (s *EndpointService) NewID() string {
	return s.idGenerator()
}
