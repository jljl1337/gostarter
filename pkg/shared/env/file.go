package env

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

/*
MustLoadOptionalEnvFile attempts to load environment variables from a .env file
if it exists. If the file does not exist, it will not return an error,
allowing the application to proceed with existing environment variables.
If any other error occurs while loading the .env file, it will panic.
*/
func MustLoadOptionalEnvFile() {
	if err := LoadOptionalEnvFile(); err != nil {
		panic(err)
	}
}

/*
LoadOptionalEnvFile attempts to load environment variables from a .env file
if it exists. If the file does not exist, it will not return an error,
allowing the application to proceed with existing environment variables.
*/
func LoadOptionalEnvFile() error {
	// It's okay if the .env file doesn't exist, we can proceed with existing env vars
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
