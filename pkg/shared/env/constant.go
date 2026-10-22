package env

import "fmt"

const (
	DatabaseDriverSQLite             = "sqlite"
	DatabaseDriverPostgreSQL         = "postgresql"
	PasswordHashingAlgorithmArgon2id = "argon2id"
	PasswordHashingAlgorithmBcrypt   = "bcrypt"
	CharsetAlphaNumeric              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CharsetAlphabetic                = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetAlphabeticLowercase       = "abcdefghijklmnopqrstuvwxyz"
	CharsetAlphabeticUppercase       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetNumeric                   = "0123456789"

	QueueTaskStatusPending   = "pending"
	QueueTaskStatusRunning   = "running"
	QueueTaskStatusSucceeded = "succeeded"
	QueueTaskStatusFailed    = "failed"
)

// Variables set at build time
var (
	Version   = "dev"
	CommitSHA = "unknown"
)

// Variables loaded from environment variables
var (
	Variables Config
)

type Config struct {
	DatabaseDriver              string `env:"DATABASE_DRIVER" default:"sqlite"`
	DataDir                     string `env:"DATA_DIR" default:"data"`
	LiveDataDir                 string `env:"LIVE_DATA_DIR" default:"live"`
	BackupDataDir               string `env:"BACKUP_DATA_DIR" default:"backup"`
	SQLiteDir                   string `env:"SQLITE_DIR" default:"db"`
	SQLiteLiveFileName          string `env:"SQLITE_LIVE_FILE_NAME" default:"live.db"`
	SQLiteBusyTimeout           string `env:"SQLITE_BUSY_TIMEOUT" default:"30000"`
	SQLiteBackupEnabled         bool   `env:"SQLITE_BACKUP_ENABLED" default:"true"`
	SQLiteBackupCronSchedule    string `env:"SQLITE_BACKUP_CRON_SCHEDULE" default:"0 0 * * *"`
	SQLiteBackupFileName        string `env:"SQLITE_BACKUP_FILE_NAME" default:"backup.db"`
	PostgreSQLURL               string `env:"POSTGRESQL_URL" default:""`
	SessionCleanupEnabled       bool   `env:"SESSION_CLEANUP_ENABLED" default:"true"`
	SessionCleanupCronSchedule  string `env:"SESSION_CLEANUP_CRON_SCHEDULE" default:"0 0 * * 0"`
	LogLevel                    int    `env:"LOG_LEVEL" default:"0"`
	LogHealthCheck              bool   `env:"LOG_HEALTH_CHECK" default:"false"`
	Port                        string `env:"PORT" default:"3000"`
	SocketPath                  string `env:"SOCKET_PATH" default:""`
	SocketPerm                  string `env:"SOCKET_PERM" default:"0666"`
	GracefulShutdownTimeoutSec  int    `env:"GRACEFUL_SHUTDOWN_TIMEOUT_SEC" default:"30"`
	CORSOrigins                 string `env:"CORS_ORIGINS" default:"*"`
	PasswordHashingAlgorithm    string `env:"PASSWORD_HASHING_ALGORITHM" default:"argon2id"`
	PasswordArgon2idMemory      uint32 `env:"PASSWORD_ARGON2ID_MEMORY" default:"65536"`
	PasswordArgon2idIterations  uint32 `env:"PASSWORD_ARGON2ID_ITERATIONS" default:"3"`
	PasswordArgon2idParallelism uint8  `env:"PASSWORD_ARGON2ID_PARALLELISM" default:"1"`
	PasswordArgon2idSaltLength  uint32 `env:"PASSWORD_ARGON2ID_SALT_LENGTH" default:"16"`
	PasswordArgon2idKeyLength   uint32 `env:"PASSWORD_ARGON2ID_KEY_LENGTH" default:"32"`
	PasswordBcryptCost          int    `env:"PASSWORD_BCRYPT_COST" default:"12"`
	SessionCookieName           string `env:"SESSION_COOKIE_NAME" default:"session_token"`
	SessionCookieHTTPOnly       bool   `env:"SESSION_COOKIE_HTTP_ONLY" default:"true"`
	SessionCookieSecure         bool   `env:"SESSION_COOKIE_SECURE" default:"false"`
	SessionCookieSameSite       string `env:"SESSION_COOKIE_SAME_SITE" default:"lax"`
	SessionTokenLength          int    `env:"SESSION_TOKEN_LENGTH" default:"32"`
	SessionTokenCharset         string `env:"SESSION_TOKEN_CHARSET" default:"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"`
	SessionLifetimeMin          int    `env:"SESSION_LIFETIME_MIN" default:"10080"`
	SessionRefreshThresholdMin  int    `env:"SESSION_REFRESH_THRESHOLD_MIN" default:"1440"`
	PreSessionLifetimeMin       int    `env:"PRE_SESSION_LIFETIME_MIN" default:"15"`
	CSRFTokenLength             int    `env:"CSRF_TOKEN_LENGTH" default:"32"`
	CSRFTokenCharset            string `env:"CSRF_TOKEN_CHARSET" default:"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"`
}

func SetConstants(files ...string) error {
	return SetConstantsWithPrefix("", files...)
}

func SetConstantsWithPrefix(prefix string, files ...string) error {
	err := LoadOptionalEnvFile(files...)
	if err != nil {
		return fmt.Errorf("failed to load env files: %w", err)
	}

	loader := &EnvLoader{Prefix: prefix}
	if err := loader.LoadIntoStruct(&Variables); err != nil {
		return fmt.Errorf("failed to load env variables into struct: %w", err)
	}

	return nil
}
