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

func TestLoadIntoStruct(t *testing.T) {
	type config struct {
		Name    string `env:"GOSTARTER_TEST_LOAD_NAME"`
		Port    int    `env:"GOSTARTER_TEST_LOAD_PORT" default:"8080"`
		Debug   bool   `env:"GOSTARTER_TEST_LOAD_DEBUG" default:"true"`
		Ignored string
		hidden  string `env:"GOSTARTER_TEST_LOAD_HIDDEN"`
	}

	t.Run("populates fields from environment and defaults", func(t *testing.T) {
		t.Setenv("GOSTARTER_TEST_LOAD_NAME", "example")
		t.Setenv("GOSTARTER_TEST_LOAD_HIDDEN", "ignored")

		cfg := config{Ignored: "untouched"}

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
		if cfg.Ignored != "untouched" {
			t.Errorf("expected Ignored to be %q, got %q", "untouched", cfg.Ignored)
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
