package service

import (
	"database/sql"
	"regexp"

	"github.com/jljl1337/gostarter/pkg/shared/crypto"
	"github.com/jljl1337/gostarter/pkg/shared/role"
)

type EndpointServiceConfig struct {
	DB             *sql.DB
	IDGenerator    func() string
	UsernameRegex  *regexp.Regexp
	PasswordRegex  *regexp.Regexp
	HashingManager *crypto.HashingManager
	RoleManager    *role.RoleManager
}

type EndpointService struct {
	db             *sql.DB
	idGenerator    func() string
	usernameRegex  *regexp.Regexp
	passwordRegex  *regexp.Regexp
	hashingManager *crypto.HashingManager
	roleManager    *role.RoleManager
}

func NewEndpointService(config EndpointServiceConfig) *EndpointService {
	return &EndpointService{
		db:             config.DB,
		idGenerator:    config.IDGenerator,
		usernameRegex:  config.UsernameRegex,
		passwordRegex:  config.PasswordRegex,
		hashingManager: config.HashingManager,
		roleManager:    config.RoleManager,
	}
}

func (s *EndpointService) NewID() string {
	return s.idGenerator()
}
