package env

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeTempFile(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return path
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()

	previous := make(map[string]string, len(keys))
	existed := make(map[string]bool, len(keys))
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		previous[key] = value
		existed[key] = ok

		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("failed to unset %s: %v", key, err)
		}
	}

	t.Cleanup(func() {
		for _, key := range keys {
			if existed[key] {
				_ = os.Setenv(key, previous[key])
				continue
			}
			_ = os.Unsetenv(key)
		}
	})
}

func TestReadOptionalEnvFiles(t *testing.T) {
	t.Run("loads existing files", func(t *testing.T) {
		const keyA = "GOSTARTER_TEST_OPTIONAL_A"
		const keyB = "GOSTARTER_TEST_OPTIONAL_B"
		unsetEnv(t, keyA, keyB)

		fileA := writeTempFile(t, "a.env", keyA+"=alpha")
		fileB := writeTempFile(t, "b.env", keyB+"=beta")

		loader := &EnvLoader{}
		if err := loader.ReadOptionalEnvFiles(fileA, fileB); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := os.Getenv(keyA); got != "alpha" {
			t.Errorf("expected %s to be %q, got %q", keyA, "alpha", got)
		}
		if got := os.Getenv(keyB); got != "beta" {
			t.Errorf("expected %s to be %q, got %q", keyB, "beta", got)
		}
	})

	t.Run("ignores missing files", func(t *testing.T) {
		const key = "GOSTARTER_TEST_OPTIONAL_EXISTING"
		unsetEnv(t, key)

		existing := writeTempFile(t, "existing.env", key+"=present")
		missing := filepath.Join(t.TempDir(), "missing.env")

		loader := &EnvLoader{}
		if err := loader.ReadOptionalEnvFiles(existing, missing); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := os.Getenv(key); got != "present" {
			t.Errorf("expected %s to be %q, got %q", key, "present", got)
		}
	})

	t.Run("returns error for malformed file", func(t *testing.T) {
		path := writeTempFile(t, "malformed.env", `KEY="unterminated`)

		loader := &EnvLoader{}
		err := loader.ReadOptionalEnvFiles(path)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if !strings.Contains(err.Error(), "error loading .env files") {
			t.Errorf("expected wrapped error, got %q", err)
		}
	})

	t.Run("loads default .env file", func(t *testing.T) {
		const key = "GOSTARTER_TEST_OPTIONAL_DEFAULT"
		unsetEnv(t, key)

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(key+"=default"), 0o600); err != nil {
			t.Fatalf("failed to write .env file: %v", err)
		}
		t.Chdir(dir)

		loader := &EnvLoader{}
		if err := loader.ReadOptionalEnvFiles(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := os.Getenv(key); got != "default" {
			t.Errorf("expected %s to be %q, got %q", key, "default", got)
		}
	})
}

func TestReadEnvFiles(t *testing.T) {
	t.Run("loads existing file", func(t *testing.T) {
		const key = "GOSTARTER_TEST_REQUIRED"
		unsetEnv(t, key)

		path := writeTempFile(t, "required.env", key+"=value")

		loader := &EnvLoader{}
		if err := loader.ReadEnvFiles(path); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := os.Getenv(key); got != "value" {
			t.Errorf("expected %s to be %q, got %q", key, "value", got)
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing.env")

		loader := &EnvLoader{}
		if err := loader.ReadEnvFiles(path); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

func TestGetString(t *testing.T) {
	t.Run("returns default when unset", func(t *testing.T) {
		loader := &EnvLoader{}

		got, err := loader.GetString("GOSTARTER_TEST_UNSET", "fallback")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "fallback" {
			t.Errorf("expected %q, got %q", "fallback", got)
		}
	})

	t.Run("returns environment value when set", func(t *testing.T) {
		const key = "GOSTARTER_TEST_VALUE"
		t.Setenv(key, "from-env")

		loader := &EnvLoader{}
		got, err := loader.GetString(key, "fallback")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "from-env" {
			t.Errorf("expected %q, got %q", "from-env", got)
		}
	})

	t.Run("applies prefix and uppercases key", func(t *testing.T) {
		const key = "GOSTARTER_TEST_PREFIX_KEY"
		t.Setenv(key, "prefixed")

		loader := &EnvLoader{Prefix: "gostarter_test_prefix"}
		got, err := loader.GetString("key", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "prefixed" {
			t.Errorf("expected %q, got %q", "prefixed", got)
		}
	})

	t.Run("file key takes precedence over direct value", func(t *testing.T) {
		const key = "GOSTARTER_TEST_PRECEDENCE"
		path := writeTempFile(t, "secret.txt", "from-file")
		t.Setenv(key, "from-env")
		t.Setenv(key+"_FILE", path)

		loader := &EnvLoader{}
		got, err := loader.GetString(key, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "from-file" {
			t.Errorf("expected %q, got %q", "from-file", got)
		}
	})

	t.Run("reads file content verbatim", func(t *testing.T) {
		const key = "GOSTARTER_TEST_FILE_CONTENT"
		content := "line one\nline two\n"
		path := writeTempFile(t, "content.txt", content)
		t.Setenv(key+"_FILE", path)

		loader := &EnvLoader{}
		got, err := loader.GetString(key, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != content {
			t.Errorf("expected %q, got %q", content, got)
		}
	})

	t.Run("applies prefix to file key", func(t *testing.T) {
		const key = "GOSTARTER_TEST_PREFIX_FILE_TOKEN"
		path := writeTempFile(t, "token.txt", "file-token")
		t.Setenv(key+"_FILE", path)

		loader := &EnvLoader{Prefix: "gostarter_test_prefix_file"}
		got, err := loader.GetString("token", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "file-token" {
			t.Errorf("expected %q, got %q", "file-token", got)
		}
	})

	t.Run("returns error when file cannot be read", func(t *testing.T) {
		const key = "GOSTARTER_TEST_UNREADABLE"
		t.Setenv(key+"_FILE", filepath.Join(t.TempDir(), "missing.txt"))

		loader := &EnvLoader{}
		if _, err := loader.GetString(key, ""); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

func TestScreamingSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{name: "single word", field: "Port", want: "PORT"},
		{name: "camel case", field: "LiveDataDir", want: "LIVE_DATA_DIR"},
		{name: "acronym before word", field: "HTTPServerURL", want: "HTTP_SERVER_URL"},
		{name: "trailing acronym", field: "CORSOrigins", want: "CORS_ORIGINS"},
		{name: "embedded acronym", field: "CSRFTokenLength", want: "CSRF_TOKEN_LENGTH"},
		{name: "acronym suffix", field: "APIKey", want: "API_KEY"},
		{name: "consecutive capitals", field: "UserID", want: "USER_ID"},
		{name: "digit", field: "Int8", want: "INT8"},
		{name: "digit boundary", field: "S3Bucket", want: "S3_BUCKET"},
		{name: "already upper", field: "PORT", want: "PORT"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := screamingSnakeCase(test.field); got != test.want {
				t.Errorf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestLoadIntoStruct(t *testing.T) {
	type config struct {
		Name          string `env:"GOSTARTER_TEST_LOAD_NAME"`
		Port          int    `env:"GOSTARTER_TEST_LOAD_PORT" default:"8080"`
		Debug         bool   `env:"GOSTARTER_TEST_LOAD_DEBUG" default:"true"`
		FieldNameOnly string
		Timeout       int    `default:"5"`
		hidden        string `env:"GOSTARTER_TEST_LOAD_HIDDEN"`
	}

	t.Run("populates fields from environment and defaults", func(t *testing.T) {
		t.Setenv("GOSTARTER_TEST_LOAD_NAME", "example")
		t.Setenv("GOSTARTER_TEST_LOAD_HIDDEN", "ignored")
		t.Setenv("FIELD_NAME_ONLY", "from-field-name")

		var cfg config

		loader := &EnvLoader{}
		if err := loader.LoadIntoStruct(&cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Name != "example" {
			t.Errorf("expected Name to be %q, got %q", "example", cfg.Name)
		}
		if cfg.Port != 8080 {
			t.Errorf("expected Port to be %d, got %d", 8080, cfg.Port)
		}
		if !cfg.Debug {
			t.Error("expected Debug to be true")
		}
		if cfg.FieldNameOnly != "from-field-name" {
			t.Errorf("expected FieldNameOnly to be %q, got %q", "from-field-name", cfg.FieldNameOnly)
		}
		if cfg.Timeout != 5 {
			t.Errorf("expected Timeout to be %d, got %d", 5, cfg.Timeout)
		}
		if cfg.hidden != "" {
			t.Errorf("expected unexported field to be empty, got %q", cfg.hidden)
		}
	})

	t.Run("overrides defaults with environment values", func(t *testing.T) {
		t.Setenv("GOSTARTER_TEST_LOAD_NAME", "example")
		t.Setenv("GOSTARTER_TEST_LOAD_PORT", "9090")
		t.Setenv("GOSTARTER_TEST_LOAD_DEBUG", "false")

		var cfg config
		loader := &EnvLoader{}
		if err := loader.LoadIntoStruct(&cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Port != 9090 {
			t.Errorf("expected Port to be %d, got %d", 9090, cfg.Port)
		}
		if cfg.Debug {
			t.Error("expected Debug to be false")
		}
	})

	t.Run("populates numeric types from defaults", func(t *testing.T) {
		type numericConfig struct {
			Int8       int8       `env:"GOSTARTER_TEST_LOAD_INT8" default:"-8"`
			Int16      int16      `env:"GOSTARTER_TEST_LOAD_INT16" default:"-16"`
			Int32      int32      `env:"GOSTARTER_TEST_LOAD_INT32" default:"-32"`
			Int64      int64      `env:"GOSTARTER_TEST_LOAD_INT64" default:"-64"`
			Uint       uint       `env:"GOSTARTER_TEST_LOAD_UINT" default:"1"`
			Uint8      uint8      `env:"GOSTARTER_TEST_LOAD_UINT8" default:"8"`
			Uint16     uint16     `env:"GOSTARTER_TEST_LOAD_UINT16" default:"16"`
			Uint32     uint32     `env:"GOSTARTER_TEST_LOAD_UINT32" default:"32"`
			Uint64     uint64     `env:"GOSTARTER_TEST_LOAD_UINT64" default:"64"`
			Uintptr    uintptr    `env:"GOSTARTER_TEST_LOAD_UINTPTR" default:"10"`
			Float32    float32    `env:"GOSTARTER_TEST_LOAD_FLOAT32" default:"1.5"`
			Float64    float64    `env:"GOSTARTER_TEST_LOAD_FLOAT64" default:"-2.5"`
			Complex64  complex64  `env:"GOSTARTER_TEST_LOAD_COMPLEX64" default:"1+2i"`
			Complex128 complex128 `env:"GOSTARTER_TEST_LOAD_COMPLEX128" default:"3+4i"`
		}

		var cfg numericConfig
		loader := &EnvLoader{}
		if err := loader.LoadIntoStruct(&cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Int8 != -8 || cfg.Int16 != -16 || cfg.Int32 != -32 || cfg.Int64 != -64 {
			t.Errorf("unexpected signed ints: %+v", cfg)
		}
		if cfg.Uint != 1 || cfg.Uint8 != 8 || cfg.Uint16 != 16 || cfg.Uint32 != 32 || cfg.Uint64 != 64 || cfg.Uintptr != 10 {
			t.Errorf("unexpected unsigned ints: %+v", cfg)
		}
		if cfg.Float32 != 1.5 || cfg.Float64 != -2.5 {
			t.Errorf("unexpected floats: %+v", cfg)
		}
		if cfg.Complex64 != 1+2i || cfg.Complex128 != 3+4i {
			t.Errorf("unexpected complex numbers: %+v", cfg)
		}
	})

	t.Run("returns error for unsupported field type", func(t *testing.T) {
		type unsupported struct {
			Values []string `env:"GOSTARTER_TEST_LOAD_UNSUPPORTED"`
		}
		t.Setenv("GOSTARTER_TEST_LOAD_UNSUPPORTED", "value")

		var cfg unsupported
		loader := &EnvLoader{}
		if err := loader.LoadIntoStruct(&cfg); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("returns error for invalid int", func(t *testing.T) {
		type invalidInt struct {
			Port int `env:"GOSTARTER_TEST_LOAD_INVALID_INT"`
		}
		t.Setenv("GOSTARTER_TEST_LOAD_INVALID_INT", "not-a-number")

		var cfg invalidInt
		loader := &EnvLoader{}
		if err := loader.LoadIntoStruct(&cfg); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("returns error for invalid bool", func(t *testing.T) {
		type invalidBool struct {
			Debug bool `env:"GOSTARTER_TEST_LOAD_INVALID_BOOL"`
		}
		t.Setenv("GOSTARTER_TEST_LOAD_INVALID_BOOL", "not-a-bool")

		var cfg invalidBool
		loader := &EnvLoader{}
		if err := loader.LoadIntoStruct(&cfg); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("returns error for invalid targets", func(t *testing.T) {
		var nilConfig *config

		tests := []struct {
			name   string
			target any
		}{
			{name: "nil", target: nil},
			{name: "non-pointer", target: config{}},
			{name: "nil pointer", target: nilConfig},
			{name: "pointer to non-struct", target: new(int)},
		}

		loader := &EnvLoader{}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if err := loader.LoadIntoStruct(test.target); err == nil {
					t.Fatal("expected an error, got nil")
				}
			})
		}
	})
}

func TestSetField(t *testing.T) {
	loader := &EnvLoader{}

	tests := []struct {
		name  string
		typ   reflect.Type
		value string
		check func(t *testing.T, field reflect.Value)
	}{
		{
			name:  "string",
			typ:   reflect.TypeOf(""),
			value: "hello",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.String(); got != "hello" {
					t.Errorf("expected %q, got %q", "hello", got)
				}
			},
		},
		{
			name:  "int",
			typ:   reflect.TypeOf(int(0)),
			value: "42",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Int(); got != 42 {
					t.Errorf("expected %d, got %d", 42, got)
				}
			},
		},
		{
			name:  "int8",
			typ:   reflect.TypeOf(int8(0)),
			value: "-12",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Int(); got != -12 {
					t.Errorf("expected %d, got %d", -12, got)
				}
			},
		},
		{
			name:  "int16",
			typ:   reflect.TypeOf(int16(0)),
			value: "-1234",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Int(); got != -1234 {
					t.Errorf("expected %d, got %d", -1234, got)
				}
			},
		},
		{
			name:  "int32",
			typ:   reflect.TypeOf(int32(0)),
			value: "123456",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Int(); got != 123456 {
					t.Errorf("expected %d, got %d", 123456, got)
				}
			},
		},
		{
			name:  "int64",
			typ:   reflect.TypeOf(int64(0)),
			value: "-1234567890",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Int(); got != -1234567890 {
					t.Errorf("expected %d, got %d", -1234567890, got)
				}
			},
		},
		{
			name:  "uint",
			typ:   reflect.TypeOf(uint(0)),
			value: "42",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Uint(); got != 42 {
					t.Errorf("expected %d, got %d", 42, got)
				}
			},
		},
		{
			name:  "uint8",
			typ:   reflect.TypeOf(uint8(0)),
			value: "255",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Uint(); got != 255 {
					t.Errorf("expected %d, got %d", 255, got)
				}
			},
		},
		{
			name:  "uint16",
			typ:   reflect.TypeOf(uint16(0)),
			value: "65535",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Uint(); got != 65535 {
					t.Errorf("expected %d, got %d", 65535, got)
				}
			},
		},
		{
			name:  "uint32",
			typ:   reflect.TypeOf(uint32(0)),
			value: "4294967295",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Uint(); got != 4294967295 {
					t.Errorf("expected %d, got %d", 4294967295, got)
				}
			},
		},
		{
			name:  "uint64",
			typ:   reflect.TypeOf(uint64(0)),
			value: "123456789",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Uint(); got != 123456789 {
					t.Errorf("expected %d, got %d", 123456789, got)
				}
			},
		},
		{
			name:  "uintptr",
			typ:   reflect.TypeOf(uintptr(0)),
			value: "42",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Uint(); got != 42 {
					t.Errorf("expected %d, got %d", 42, got)
				}
			},
		},
		{
			name:  "float32",
			typ:   reflect.TypeOf(float32(0)),
			value: "3.5",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Float(); got != 3.5 {
					t.Errorf("expected %v, got %v", 3.5, got)
				}
			},
		},
		{
			name:  "float64",
			typ:   reflect.TypeOf(float64(0)),
			value: "-1.25",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Float(); got != -1.25 {
					t.Errorf("expected %v, got %v", -1.25, got)
				}
			},
		},
		{
			name:  "complex64",
			typ:   reflect.TypeOf(complex64(0)),
			value: "1+2i",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Complex(); got != 1+2i {
					t.Errorf("expected %v, got %v", 1+2i, got)
				}
			},
		},
		{
			name:  "complex128",
			typ:   reflect.TypeOf(complex128(0)),
			value: "-3.5+4.25i",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Complex(); got != -3.5+4.25i {
					t.Errorf("expected %v, got %v", -3.5+4.25i, got)
				}
			},
		},
		{
			name:  "bool",
			typ:   reflect.TypeOf(false),
			value: "true",
			check: func(t *testing.T, field reflect.Value) {
				if got := field.Bool(); !got {
					t.Error("expected true, got false")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := reflect.New(test.typ).Elem()
			if err := loader.setField(field, test.value); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			test.check(t, field)
		})
	}

	invalid := []struct {
		name  string
		typ   reflect.Type
		value string
	}{
		{name: "invalid int", typ: reflect.TypeOf(int(0)), value: "abc"},
		{name: "int8 overflow", typ: reflect.TypeOf(int8(0)), value: "128"},
		{name: "int16 overflow", typ: reflect.TypeOf(int16(0)), value: "32768"},
		{name: "int32 overflow", typ: reflect.TypeOf(int32(0)), value: "2147483648"},
		{name: "negative uint", typ: reflect.TypeOf(uint(0)), value: "-1"},
		{name: "uint8 overflow", typ: reflect.TypeOf(uint8(0)), value: "256"},
		{name: "uint16 overflow", typ: reflect.TypeOf(uint16(0)), value: "65536"},
		{name: "uint32 overflow", typ: reflect.TypeOf(uint32(0)), value: "4294967296"},
		{name: "invalid uint", typ: reflect.TypeOf(uint64(0)), value: "abc"},
		{name: "invalid uintptr", typ: reflect.TypeOf(uintptr(0)), value: "abc"},
		{name: "invalid float32", typ: reflect.TypeOf(float32(0)), value: "abc"},
		{name: "invalid float64", typ: reflect.TypeOf(float64(0)), value: "abc"},
		{name: "invalid complex64", typ: reflect.TypeOf(complex64(0)), value: "1+"},
		{name: "invalid complex128", typ: reflect.TypeOf(complex128(0)), value: "1+"},
		{name: "invalid bool", typ: reflect.TypeOf(false), value: "maybe"},
		{name: "unsupported type", typ: reflect.TypeOf([]string{}), value: "value"},
	}

	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			field := reflect.New(test.typ).Elem()
			if err := loader.setField(field, test.value); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
