package crypto

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/jljl1337/gostarter/env"
)

type BcryptHasher struct {
	cost int
}

func NewBcryptHasherFromEnv() *BcryptHasher {
	return NewBcryptHasher(env.PasswordBcryptCost)
}

func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Name() string {
	return env.PasswordHashingAlgorithmBcrypt
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(bytes), err
}

func (h *BcryptHasher) IsValidHash(hash string) bool {
	// Check if the hash starts with the expected prefix for bcrypt
	if len(hash) < 4 || (hash[:4] != "$2a$" && hash[:4] != "$2b$" && hash[:4] != "$2y$") {
		return false
	}

	return true
}

func (h *BcryptHasher) Compare(hash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil, err
}

func (h *BcryptHasher) CompareParameters(hash string) (bool, error) {
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return false, err
	}
	return cost == h.cost, nil
}
