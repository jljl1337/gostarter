package env

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// MustLoadOptionalEnvFile calls [LoadOptionalEnvFile] and panics if an
// error occurs.
func MustLoadOptionalEnvFile(files ...string) {
	if err := LoadOptionalEnvFile(files...); err != nil {
		panic(err)
	}
}

// LoadOptionalEnvFile attempts to load environment variables from a list of
// .env files. If a file does not exist, it will be ignored, allowing the
// application to proceed with existing environment variables. If any other
// error occurs while loading the .env files, it will return that error.
//
// If no files are provided, it will attempt to load a default .env file in the
// current working directory.
func LoadOptionalEnvFile(files ...string) error {
	// It's okay if the .env file doesn't exist, we can proceed with existing env vars
	if err := godotenv.Load(files...); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
