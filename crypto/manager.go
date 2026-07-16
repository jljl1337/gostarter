package crypto

import (
	"fmt"

	"github.com/jljl1337/gostarter/env"
)

type HashingManager struct {
	hasherMap     map[string]Hasher
	defaultHasher Hasher
}

func NewHashingManager(hasherList []Hasher) (*HashingManager, error) {
	if len(hasherList) == 0 {
		return nil, fmt.Errorf("hasher list cannot be empty")
	}

	hasherMap := make(map[string]Hasher)
	for _, hasher := range hasherList {
		hasherMap[hasher.Name()] = hasher
	}

	defaultHasherName := env.PasswordHashingAlgorithm
	_, exists := hasherMap[defaultHasherName]
	if !exists {
		defaultHasherName = hasherList[0].Name()
	}

	return &HashingManager{
		hasherMap:     hasherMap,
		defaultHasher: hasherMap[defaultHasherName],
	}, nil
}

func (m *HashingManager) HashPassword(password string) (string, error) {
	return m.defaultHasher.Hash(password)
}

func (m *HashingManager) ComparePassword(hash, password string) (bool, error) {
	hasher, err := m.getHasherForHash(hash)
	if err != nil {
		return false, err
	}

	return hasher.Compare(hash, password)
}

func (m *HashingManager) CheckIfNeedsRehash(hash string) (bool, error) {
	hasher, err := m.getHasherForHash(hash)
	if err != nil {
		return false, err
	}

	if hasher.Name() != m.defaultHasher.Name() {
		return true, nil
	}

	return hasher.CompareParameters(hash)
}

func (m *HashingManager) getHasherForHash(hash string) (Hasher, error) {
	for _, hasher := range m.hasherMap {
		if hasher.IsValidHash(hash) {
			return hasher, nil
		}
	}
	return nil, fmt.Errorf("no suitable hasher found for the provided hash")
}
