package env

import (
	"os"
	"strconv"
)

/*
MustGetBool retrieves the value of the environment variable named by the
key. If the `key_FILE` environment variable is set, it reads the value from
the specified file. If neither the `key` nor `key_FILE` environment
variables are set, it returns the provided default value.

If there is an error reading from the file specified by `key_FILE` or if
the value cannot be converted to a boolean, it panics.
*/
func MustGetBool(key string, defaultValue bool) bool {
	value, err := GetBool(key, defaultValue)
	if err != nil {
		panic(err)
	}
	return value
}

/*
GetBool retrieves the value of the environment variable named by the key.
If the `key_FILE` environment variable is set, it reads the value from the
specified file. If neither the `key` nor `key_FILE` environment variables
are set, it returns the provided default value.

It returns an error if there is an issue reading from the file specified by
`key_FILE` or if the value cannot be converted to a boolean.
*/
func GetBool(key string, defaultValue bool) (bool, error) {
	defaultStr := "false"
	if defaultValue {
		defaultStr = "true"
	}

	value, err := GetString(key, defaultStr)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(value)
}

/*
MustGetInt retrieves the value of the environment variable named by the key.
If the `key_FILE` environment variable is set, it reads the value from the
specified file. If neither the `key` nor `key_FILE` environment variables
are set, it returns the provided default value.

If there is an error reading from the file specified by `key_FILE` or if
the value cannot be converted to an int, it panics.
*/
func MustGetInt(key string, defaultValue int) int {
	value, err := GetInt(key, defaultValue)
	if err != nil {
		panic(err)
	}
	return value
}

/*
GetInt retrieves the value of the environment variable named by the key.
If the `key_FILE` environment variable is set, it reads the value from the
specified file. If neither the `key` nor `key_FILE` environment variables
are set, it returns the provided default value.

It returns an error if there is an issue reading from the file specified by
`key_FILE` or if the value cannot be converted to an int.
*/
func GetInt(key string, defaultValue int) (int, error) {
	value, err := GetString(key, strconv.Itoa(defaultValue))
	if err != nil {
		return 0, err
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return intValue, nil
}

/*
MustGetString retrieves the value of the environment variable named by the
key. If the `key_FILE` environment variable is set, it reads the value from
the specified file. If neither the `key` nor `key_FILE` environment
variables are set, it returns the provided default value.

If there is an error reading from the file specified by `key_FILE`, it
panics.
*/
func MustGetString(key string, defaultValue string) string {
	value, err := GetString(key, defaultValue)
	if err != nil {
		panic(err)
	}
	return value
}

/*
GetString retrieves the value of the environment variable named by the key.
If the `key_FILE` environment variable is set, it reads the value from the
specified file. If neither the `key` nor `key_FILE` environment variables
are set, it returns the provided default value.

It returns an error if there is an issue reading from the file specified by
`key_FILE`.
*/
func GetString(key string, defaultValue string) (string, error) {
	fileKey := key + "_FILE"
	fileValue, fileExists := os.LookupEnv(fileKey)
	if fileExists {
		// Read from the file
		data, err := os.ReadFile(fileValue)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	value, exists := os.LookupEnv(key)
	if exists {
		return value, nil
	}
	return defaultValue, nil
}
