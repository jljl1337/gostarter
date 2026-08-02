package env

import "net/http"

const (
	DatabaseDriverSQLite             = "sqlite"
	DatabaseDriverPostgreSQL         = "postgresql"
	DatabaseDriverMySQL              = "mysql"
	PasswordHashingAlgorithmArgon2id = "argon2id"
	PasswordHashingAlgorithmBcrypt   = "bcrypt"
	AlphaNumericCharset              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	// Variables set at build time

	Version   = "dev"
	CommitSHA = "unknown"

	// Internal variables

	ConstantsSet bool = false

	// Plain variables from environment variables

	DataDir                     string
	LiveDataDir                 string
	BackupDataDir               string
	SQLiteDir                   string
	SQLiteLiveFileName          string
	SQLiteDbBusyTimeout         string
	SQLiteBackupEnabled         bool
	SQLiteBackupCronSchedule    string
	SQLiteBackupFileName        string
	PostgreSQLURL               string
	MySQLURL                    string
	MySQLMaxLifetimeMin         int
	MySQLMaxOpenConns           int
	MySQLMaxIdleConns           int
	SessionCleanupEnabled       bool
	SessionCleanupCronSchedule  string
	LogLevel                    int
	LogHealthCheck              bool
	Port                        string
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
	SessionCookieHttpOnly       bool
	SessionCookieSecure         bool
	SessionTokenLength          int
	SessionTokenCharset         string
	SessionLifetimeMin          int
	SessionRefreshThresholdMin  int
	PreSessionLifetimeMin       int
	CSRFTokenLength             int
	CSRFTokenCharset            string

	// Derived variables

	DatabaseDriver        string
	SessionCookieSameSite http.SameSite
)

func MustSetConstants(addPrefix bool) { // TODO: list of warning
	MustLoadOptionalEnvFile()

	databaseDriver := MustGetString(prefix("DATABASE_DRIVER", addPrefix), DatabaseDriverSQLite)
	DataDir = MustGetString(prefix("DATA_DIR", addPrefix), "data")
	LiveDataDir = MustGetString(prefix("LIVE_DATA_DIR", addPrefix), "live")
	BackupDataDir = MustGetString(prefix("BACKUP_DATA_DIR", addPrefix), "backup")
	SQLiteDir = MustGetString(prefix("SQLITE_DIR", addPrefix), "db")
	SQLiteLiveFileName = MustGetString(prefix("SQLITE_LIVE_FILE_NAME", addPrefix), "live.db")
	SQLiteDbBusyTimeout = MustGetString(prefix("SQLITE_BUSY_TIMEOUT", addPrefix), "30000")
	SQLiteBackupEnabled = MustGetBool(prefix("SQLITE_BACKUP_ENABLED", addPrefix), true)
	SQLiteBackupCronSchedule = MustGetString(prefix("SQLITE_BACKUP_CRON_SCHEDULE", addPrefix), "0 0 * * *")
	SQLiteBackupFileName = MustGetString(prefix("SQLITE_BACKUP_FILE_NAME", addPrefix), "backup.db")
	PostgreSQLURL = MustGetString(prefix("POSTGRESQL_URL", addPrefix), "")
	MySQLURL = MustGetString(prefix("MYSQL_URL", addPrefix), "")
	MySQLMaxLifetimeMin = MustGetInt(prefix("MYSQL_MAX_LIFETIME_MIN", addPrefix), 3)
	MySQLMaxOpenConns = MustGetInt(prefix("MYSQL_MAX_OPEN_CONNS", addPrefix), 10)
	MySQLMaxIdleConns = MustGetInt(prefix("MYSQL_MAX_IDLE_CONNS", addPrefix), 10)
	SessionCleanupEnabled = MustGetBool(prefix("SESSION_CLEANUP_ENABLED", addPrefix), true)
	SessionCleanupCronSchedule = MustGetString(prefix("SESSION_CLEANUP_CRON_SCHEDULE", addPrefix), "0 0 * * 0")
	LogLevel = MustGetInt(prefix("LOG_LEVEL", addPrefix), 0)
	LogHealthCheck = MustGetBool(prefix("LOG_HEALTH_CHECK", addPrefix), false)
	Port = MustGetString(prefix("PORT", addPrefix), "3000")
	GracefulShutdownTimeoutSec = MustGetInt(prefix("GRACEFUL_SHUTDOWN_TIMEOUT_SEC", addPrefix), 30)
	CORSOrigins = MustGetString(prefix("CORS_ORIGINS", addPrefix), "*")
	PasswordHashingAlgorithm = MustGetString(prefix("PASSWORD_HASHING_ALGORITHM", addPrefix), PasswordHashingAlgorithmArgon2id)
	PasswordArgon2idMemory = MustGetInt(prefix("PASSWORD_ARGON2ID_MEMORY", addPrefix), 64*1024)
	PasswordArgon2idIterations = MustGetInt(prefix("PASSWORD_ARGON2ID_ITERATIONS", addPrefix), 3)
	PasswordArgon2idParallelism = MustGetInt(prefix("PASSWORD_ARGON2ID_PARALLELISM", addPrefix), 1)
	PasswordArgon2idSaltLength = MustGetInt(prefix("PASSWORD_ARGON2ID_SALT_LENGTH", addPrefix), 16)
	PasswordArgon2idKeyLength = MustGetInt(prefix("PASSWORD_ARGON2ID_KEY_LENGTH", addPrefix), 32)
	PasswordBcryptCost = MustGetInt(prefix("PASSWORD_BCRYPT_COST", addPrefix), 12)
	SessionCookieName = MustGetString(prefix("SESSION_COOKIE_NAME", addPrefix), "session_token")
	SessionCookieHttpOnly = MustGetBool(prefix("SESSION_COOKIE_HTTP_ONLY", addPrefix), true)
	SessionCookieSecure = MustGetBool(prefix("SESSION_COOKIE_SECURE", addPrefix), false)
	sessionCookieSameSite := MustGetString(prefix("SESSION_COOKIE_SAME_SITE", addPrefix), "lax")
	SessionTokenLength = MustGetInt(prefix("SESSION_TOKEN_LENGTH", addPrefix), 32)
	SessionTokenCharset = MustGetString(prefix("SESSION_TOKEN_CHARSET", addPrefix), AlphaNumericCharset)
	SessionLifetimeMin = MustGetInt(prefix("SESSION_LIFETIME_MIN", addPrefix), 60*24*7)
	SessionRefreshThresholdMin = MustGetInt(prefix("SESSION_REFRESH_THRESHOLD_MIN", addPrefix), 60*24)
	PreSessionLifetimeMin = MustGetInt(prefix("PRE_SESSION_LIFETIME_MIN", addPrefix), 15)
	CSRFTokenLength = MustGetInt(prefix("CSRF_TOKEN_LENGTH", addPrefix), 32)
	CSRFTokenCharset = MustGetString(prefix("CSRF_TOKEN_CHARSET", addPrefix), AlphaNumericCharset)

	switch databaseDriver {
	case DatabaseDriverPostgreSQL:
		DatabaseDriver = DatabaseDriverPostgreSQL
	case DatabaseDriverSQLite:
		DatabaseDriver = DatabaseDriverSQLite
	case DatabaseDriverMySQL:
		DatabaseDriver = DatabaseDriverMySQL
	default:
		DatabaseDriver = DatabaseDriverSQLite
	}

	switch sessionCookieSameSite {
	case "lax":
		SessionCookieSameSite = http.SameSiteLaxMode
	case "strict":
		SessionCookieSameSite = http.SameSiteStrictMode
	case "none":
		SessionCookieSameSite = http.SameSiteNoneMode
	default:
		SessionCookieSameSite = http.SameSiteNoneMode
	}

	ConstantsSet = true
}

func prefix(str string, add bool) string {
	if add {
		return "GOSTARTER_" + str
	}
	return str
}
