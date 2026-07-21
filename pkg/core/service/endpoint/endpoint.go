package endpoint

import (
	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/shared/crypto"
	"github.com/jljl1337/gostarter/pkg/shared/validation"
)

type EndpointService struct {
	db                *sqlx.DB
	idGenerator       func() string
	hashingManager    *crypto.HashingManager
	validationManager *validation.ValidationManager
}

func NewEndpointService(db *sqlx.DB, idGenerator func() string, hashingManager *crypto.HashingManager, validationManager *validation.ValidationManager) *EndpointService {
	return &EndpointService{
		db:                db,
		idGenerator:       idGenerator,
		hashingManager:    hashingManager,
		validationManager: validationManager,
	}
}

func (s *EndpointService) NewID() string {
	return s.idGenerator()
}
