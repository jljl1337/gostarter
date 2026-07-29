package gostarter

import "github.com/jljl1337/gostarter/pkg/shared/env"

func MustSetConstantsWithPrefix() {
	env.MustSetConstants(true)
}

func MustSetConstantsWithoutPrefix() {
	env.MustSetConstants(false)
}

func MustGetBool(key string, defaultValue bool) bool {
	return env.MustGetBool(key, defaultValue)
}

func GetBool(key string, defaultValue bool) (bool, error) {
	return env.GetBool(key, defaultValue)
}

func MustGetInt(key string, defaultValue int) int {
	return env.MustGetInt(key, defaultValue)
}

func GetInt(key string, defaultValue int) (int, error) {
	return env.GetInt(key, defaultValue)
}

func MustGetString(key string, defaultValue string) string {
	return env.MustGetString(key, defaultValue)
}

func GetString(key string, defaultValue string) (string, error) {
	return env.GetString(key, defaultValue)
}
