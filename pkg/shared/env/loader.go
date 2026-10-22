package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/joho/godotenv"
)

// EnvLoader is a utility for loading environment variables into a struct. It
// can also read environment variables from .env files and supports default
// values for struct fields.
//
// The Prefix field allows you to specify a prefix for environment variable
// names. If set, all environment variable lookups will be prefixed with this
// value followed by an underscore. For example, if Prefix is set to "MYAPP",
// then the environment variable "DATABASE_URL" will be looked up as
// "MYAPP_DATABASE_URL".
type EnvLoader struct {
	Prefix string
}

// ReadOptionalEnvFiles calls [ReadEnvFiles] but ignores the error if the file
// doesn't exist.
func (e *EnvLoader) ReadOptionalEnvFiles(files ...string) error {
	// It's okay if the .env file doesn't exist, we can still proceed
	if err := e.ReadEnvFiles(files...); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error loading .env files: %w", err)
	}

	return nil
}

// ReadEnvFiles reads environment variables from the specified .env files.
func (e *EnvLoader) ReadEnvFiles(files ...string) error {
	return godotenv.Load(files...)
}

// LoadIntoStruct loads environment variables into the provided struct. It uses
// the "env" struct tag to determine the environment variable name. If the tag
// is not present, it converts the struct field name to screaming snake case.
// It also supports default values specified in the "default" struct tag.
//
// [GetString] is used to retrieve the environment variable values.
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

		if !valueField.CanSet() {
			continue
		}

		envKey, ok := typeField.Tag.Lookup("env")
		if !ok {
			envKey = screamingSnakeCase(typeField.Name)
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

func screamingSnakeCase(name string) string {
	runes := []rune(name)

	var builder strings.Builder
	builder.Grow(len(name) + 4)

	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextIsLower) {
				builder.WriteByte('_')
			}
		}

		builder.WriteRune(unicode.ToUpper(r))
	}

	return builder.String()
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

// GetString retrieves the value of the environment variable specified by key.
// If a corresponding *_FILE environment variable is set, it reads the value
// from the specified file. Otherwise, if the environment variable is not set,
// it returns the provided defaultValue. The `key` parameter is
// case-insensitive and will be converted to uppercase.
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
