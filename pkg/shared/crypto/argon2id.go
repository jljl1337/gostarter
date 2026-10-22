package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

type Argon2idHasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func NewArgon2idHasherFromEnv() *Argon2idHasher {
	return &Argon2idHasher{
		memory:      env.Variables.PasswordArgon2idMemory,
		iterations:  env.Variables.PasswordArgon2idIterations,
		parallelism: env.Variables.PasswordArgon2idParallelism,
		saltLength:  env.Variables.PasswordArgon2idSaltLength,
		keyLength:   env.Variables.PasswordArgon2idKeyLength,
	}
}

func (h *Argon2idHasher) Name() string {
	return env.PasswordHashingAlgorithmArgon2id
}

func (h *Argon2idHasher) Hash(password string) (string, error) {
	// Generate a cryptographically secure random salt
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Passwords should be passed to argon2.IDKey as raw bytes
	hash := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, h.parallelism, h.keyLength)

	// Base64 encode the salt and hash without padding for standard PHC format
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return full formatted string including parameters
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.memory, h.iterations, h.parallelism, b64Salt, b64Hash)

	return encoded, nil
}

func (h *Argon2idHasher) IsValidHash(hash string) bool {
	// Check if the hash starts with the expected prefix for Argon2id
	if len(hash) < 9 || hash[:9] != "$argon2id" {
		return false
	}

	return true
}

func (h *Argon2idHasher) Compare(hash, password string) (bool, error) {
	version, memory, iterations, parallelism, salt, hashBytes, err := h.SplitHash(hash)
	if err != nil {
		return false, fmt.Errorf("failed to split hash: %w", err)
	}

	if version != argon2.Version {
		return false, fmt.Errorf("argon2id version mismatch, expected %d, got %d", argon2.Version, version)
	}

	// Decode the base64 encoded salt and hash
	saltDecoded, err := base64.RawStdEncoding.DecodeString(string(salt))
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}
	hashDecoded, err := base64.RawStdEncoding.DecodeString(string(hashBytes))
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Generate a new hash with the same parameters and salt
	newHash := argon2.IDKey([]byte(password), saltDecoded, iterations, memory, parallelism, uint32(len(hashDecoded)))

	// Compare the newly generated hash with the original hash
	return subtle.ConstantTimeCompare(newHash, hashDecoded) == 1, nil
}

func (h *Argon2idHasher) CompareParameters(hash string) (bool, error) {
	// TODO: check salt (and hash length?) as well
	version, memory, iterations, parallelism, _, _, err := h.SplitHash(hash)
	if err != nil {
		return false, fmt.Errorf("failed to split hash: %w", err)
	}

	// Compare the parameters with the current hasher's parameters
	if version != argon2.Version ||
		memory != h.memory ||
		iterations != h.iterations ||
		parallelism != h.parallelism {
		return false, nil
	}

	return true, nil
}

/*
SplitHash splits the given Argon2id hash string into its components:
version, memory, iterations, parallelism, salt, and hash bytes. It returns
an error if the hash format is invalid.
*/
func (h *Argon2idHasher) SplitHash(hash string) (version int, memory, iterations uint32, parallelism uint8, salt, hashBytes []byte, err error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return 0, 0, 0, 0, nil, nil, fmt.Errorf("failed to parse hash: invalid argon2id format")
	}

	if _, err = fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return 0, 0, 0, 0, nil, nil, fmt.Errorf("failed to parse hash version: %w", err)
	}
	if _, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return 0, 0, 0, 0, nil, nil, fmt.Errorf("failed to parse hash parameters: %w", err)
	}

	return version, memory, iterations, parallelism, []byte(parts[4]), []byte(parts[5]), nil
}
