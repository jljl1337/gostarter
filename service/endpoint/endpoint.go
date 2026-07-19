package endpoint

import (
	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/crypto"
	"github.com/jljl1337/gostarter/validation"
)

type EndpointService struct {
	db                *sqlx.DB
	hashingManager    *crypto.HashingManager
	validationManager *validation.ValidationManager
}

func NewEndpointService(db *sqlx.DB, hashingManager *crypto.HashingManager, validationManager *validation.ValidationManager) *EndpointService {
	return &EndpointService{
		db:                db,
		hashingManager:    hashingManager,
		validationManager: validationManager,
	}
}
