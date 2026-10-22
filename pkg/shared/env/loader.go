package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type EnvLoader struct {
	Prefix string
}

func (e *EnvLoader) ReadOptionalEnvFiles(files ...string) error {
	// It's okay if the .env file doesn't exist, we can still proceed
	if err := godotenv.Load(files...); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error loading .env files: %w", err)
	}

	return nil
}

func (e *EnvLoader) ReadEnvFiles(files ...string) error {
	return godotenv.Load(files...)
}

func (e *EnvLoader) LoadIntoStruct(target any) error {
	pointer := reflect.ValueOf(target)
	if pointer.Kind() != reflect.Pointer || pointer.IsNil() {
		return fmt.Errorf("expected non-nil pointer to struct, got %T", target)
	}

	value := pointer.Elem()
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, got pointer to %s", value.Kind())
	}

	structType := value.Type()
	for i := 0; i < structType.NumField(); i++ {
		typeField := structType.Field(i)
		valueField := value.Field(i)

		envKey, ok := typeField.Tag.Lookup("env")
		if !ok {
			// TODO: assume env key to be screaming snake case of field name
			continue
		}
		if !valueField.CanSet() {
			continue
		}

		defaultValue, _ := typeField.Tag.Lookup("default")
		envValue, err := e.GetString(envKey, defaultValue)
		if err != nil {
			return fmt.Errorf("error getting string value for key %s: %w", envKey, err)
		}

		if err := e.setField(valueField, envValue); err != nil {
			return fmt.Errorf("field %s (env %s): %w", typeField.Name, envKey, err)
		}
	}

	return nil
}

func (e *EnvLoader) setField(field reflect.Value, value string) error {
	switch field.Kind() {

	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid %s %q: %w", field.Kind(), value, err)
		}
		field.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid %s %q: %w", field.Kind(), value, err)
		}
		field.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid %s %q: %w", field.Kind(), value, err)
		}
		field.SetFloat(n)

	case reflect.Complex64, reflect.Complex128:
		n, err := strconv.ParseComplex(value, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid %s %q: %w", field.Kind(), value, err)
		}
		field.SetComplex(n)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool %q: %w", value, err)
		}
		field.SetBool(b)

	default:
		return fmt.Errorf("unsupported type %s", field.Kind())
	}

	return nil
}

func (e *EnvLoader) GetString(key string, defaultValue string) (string, error) {
	if e.Prefix != "" {
		key = e.Prefix + "_" + key
		key = strings.ToUpper(key)
	}

	fileKey := key + "_FILE"
	fileKey = strings.ToUpper(fileKey)
	fileValue, fileEnvExists := os.LookupEnv(fileKey)

	if fileEnvExists {
		data, err := os.ReadFile(fileValue)
		if err != nil {
			return "", fmt.Errorf("error reading from file %s: %w", fileValue, err)
		}

		return string(data), nil
	}

	value, exists := os.LookupEnv(key)
	if exists {
		return value, nil
	}

	return defaultValue, nil
}
