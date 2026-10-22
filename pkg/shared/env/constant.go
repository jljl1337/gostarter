package env

const (
	DatabaseDriverSQLite             = "sqlite"
	DatabaseDriverPostgreSQL         = "postgresql"
	PasswordHashingAlgorithmArgon2id = "argon2id"
	PasswordHashingAlgorithmBcrypt   = "bcrypt"
	AlphaNumericCharset              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	QueueTaskStatusPending   = "pending"
	QueueTaskStatusRunning   = "running"
	QueueTaskStatusSucceeded = "succeeded"
	QueueTaskStatusFailed    = "failed"
)

var (
	// Variables set at build time

	Version   = "dev"
	CommitSHA = "unknown"

	// Internal variables

	ConstantsSet bool = false

	// Plain variables from environment variables

	DatabaseDriver              string
	DataDir                     string
	LiveDataDir                 string
	BackupDataDir               string
	SQLiteDir                   string
	SQLiteLiveFileName          string
	SQLiteBusyTimeout           string
	SQLiteBackupEnabled         bool
	SQLiteBackupCronSchedule    string
	SQLiteBackupFileName        string
	PostgreSQLURL               string
	SessionCleanupEnabled       bool
	SessionCleanupCronSchedule  string
	LogLevel                    int
	LogHealthCheck              bool
	Port                        string
	SocketPath                  string
	SocketPerm                  string
	GracefulShutdownTimeoutSec  int
	CORSOrigins                 string
	PasswordHashingAlgorithm    string
	PasswordArgon2idMemory      int
	PasswordArgon2idIterations  int
	PasswordArgon2idParallelism int
	PasswordArgon2idSaltLength  int
	PasswordArgon2idKeyLength   int
	PasswordBcryptCost          int
	SessionCookieName           string
	SessionCookieHTTPOnly       bool
	SessionCookieSecure         bool
	SessionCookieSameSite       string
	SessionTokenLength          int
	SessionTokenCharset         string
	SessionLifetimeMin          int
	SessionRefreshThresholdMin  int
	PreSessionLifetimeMin       int
	CSRFTokenLength             int
	CSRFTokenCharset            string
)

func MustSetConstants(files ...string) {
	MustSetConstantsWithPrefix("", files...)
}

func MustSetConstantsWithPrefix(prefix string, files ...string) {
	MustLoadOptionalEnvFile(files...)

	DatabaseDriver = MustGetString(addPrefix("DATABASE_DRIVER", prefix), DatabaseDriverSQLite)
	DataDir = MustGetString(addPrefix("DATA_DIR", prefix), "data")
	LiveDataDir = MustGetString(addPrefix("LIVE_DATA_DIR", prefix), "live")
	BackupDataDir = MustGetString(addPrefix("BACKUP_DATA_DIR", prefix), "backup")
	SQLiteDir = MustGetString(addPrefix("SQLITE_DIR", prefix), "db")
	SQLiteLiveFileName = MustGetString(addPrefix("SQLITE_LIVE_FILE_NAME", prefix), "live.db")
	SQLiteBusyTimeout = MustGetString(addPrefix("SQLITE_BUSY_TIMEOUT", prefix), "30000")
	SQLiteBackupEnabled = MustGetBool(addPrefix("SQLITE_BACKUP_ENABLED", prefix), true)
	SQLiteBackupCronSchedule = MustGetString(addPrefix("SQLITE_BACKUP_CRON_SCHEDULE", prefix), "0 0 * * *")
	SQLiteBackupFileName = MustGetString(addPrefix("SQLITE_BACKUP_FILE_NAME", prefix), "backup.db")
	PostgreSQLURL = MustGetString(addPrefix("POSTGRESQL_URL", prefix), "")
	SessionCleanupEnabled = MustGetBool(addPrefix("SESSION_CLEANUP_ENABLED", prefix), true)
	SessionCleanupCronSchedule = MustGetString(addPrefix("SESSION_CLEANUP_CRON_SCHEDULE", prefix), "0 0 * * 0")
	LogLevel = MustGetInt(addPrefix("LOG_LEVEL", prefix), 0)
	LogHealthCheck = MustGetBool(addPrefix("LOG_HEALTH_CHECK", prefix), false)
	Port = MustGetString(addPrefix("PORT", prefix), "3000")
	SocketPath = MustGetString(addPrefix("SOCKET_PATH", prefix), "")
	SocketPerm = MustGetString(addPrefix("SOCKET_PERM", prefix), "0666")
	GracefulShutdownTimeoutSec = MustGetInt(addPrefix("GRACEFUL_SHUTDOWN_TIMEOUT_SEC", prefix), 30)
	CORSOrigins = MustGetString(addPrefix("CORS_ORIGINS", prefix), "*")
	PasswordHashingAlgorithm = MustGetString(addPrefix("PASSWORD_HASHING_ALGORITHM", prefix), PasswordHashingAlgorithmArgon2id)
	PasswordArgon2idMemory = MustGetInt(addPrefix("PASSWORD_ARGON2ID_MEMORY", prefix), 64*1024)
	PasswordArgon2idIterations = MustGetInt(addPrefix("PASSWORD_ARGON2ID_ITERATIONS", prefix), 3)
	PasswordArgon2idParallelism = MustGetInt(addPrefix("PASSWORD_ARGON2ID_PARALLELISM", prefix), 1)
	PasswordArgon2idSaltLength = MustGetInt(addPrefix("PASSWORD_ARGON2ID_SALT_LENGTH", prefix), 16)
	PasswordArgon2idKeyLength = MustGetInt(addPrefix("PASSWORD_ARGON2ID_KEY_LENGTH", prefix), 32)
	PasswordBcryptCost = MustGetInt(addPrefix("PASSWORD_BCRYPT_COST", prefix), 12)
	SessionCookieName = MustGetString(addPrefix("SESSION_COOKIE_NAME", prefix), "session_token")
	SessionCookieHTTPOnly = MustGetBool(addPrefix("SESSION_COOKIE_HTTP_ONLY", prefix), true)
	SessionCookieSecure = MustGetBool(addPrefix("SESSION_COOKIE_SECURE", prefix), false)
	SessionCookieSameSite = MustGetString(addPrefix("SESSION_COOKIE_SAME_SITE", prefix), "lax")
	SessionTokenLength = MustGetInt(addPrefix("SESSION_TOKEN_LENGTH", prefix), 32)
	SessionTokenCharset = MustGetString(addPrefix("SESSION_TOKEN_CHARSET", prefix), AlphaNumericCharset)
	SessionLifetimeMin = MustGetInt(addPrefix("SESSION_LIFETIME_MIN", prefix), 60*24*7)
	SessionRefreshThresholdMin = MustGetInt(addPrefix("SESSION_REFRESH_THRESHOLD_MIN", prefix), 60*24)
	PreSessionLifetimeMin = MustGetInt(addPrefix("PRE_SESSION_LIFETIME_MIN", prefix), 15)
	CSRFTokenLength = MustGetInt(addPrefix("CSRF_TOKEN_LENGTH", prefix), 32)
	CSRFTokenCharset = MustGetString(addPrefix("CSRF_TOKEN_CHARSET", prefix), AlphaNumericCharset)

	ConstantsSet = true
}

func addPrefix(str, prefix string) string {
	return prefix + str
}
