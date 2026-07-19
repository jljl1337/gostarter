package env

import "net/http"

const (
	DatabaseDriverSQLite             = "sqlite"
	DatabaseDriverPostgreSQL         = "postgresql"
	PasswordHashingAlgorithmArgon2id = "argon2id"
	PasswordHashingAlgorithmBcrypt   = "bcrypt"
)

var (
	// Variables set at build time

	Version = "dev"

	// Internal variables

	ConstantsSet bool = false

	// Plain variables from environment variables

	DataDir                     string
	LiveDataDir                 string
	BackupDataDir               string
	SQLiteDir                   string
	LiveSQLiteFileName          string
	EnableSQLiteBackup          bool
	BackupSQLiteFileName        string
	PostgresURL                 string
	SQLiteDbBusyTimeout         string
	SQLiteBackupDbPath          string
	SQLiteBackupCronSchedule    string
	EnableSessionCleanup        bool
	SessionCleanupCronSchedule  string
	LogLevel                    int
	LogHealthCheck              bool
	Port                        string
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

func MustSetConstants() { // TODO: list of warning
	MustLoadOptionalEnvFile()

	DataDir = MustGetString("GOSTARTER_DATA_DIR", "data")
	LiveDataDir = MustGetString("GOSTARTER_LIVE_DATA_DIR", "live")
	BackupDataDir = MustGetString("GOSTARTER_BACKUP_DATA_DIR", "backup")
	SQLiteDir = MustGetString("GOSTARTER_SQLITE_DIR", "db")
	LiveSQLiteFileName = MustGetString("GOSTARTER_LIVE_SQLITE_FILE_NAME", "live.db")
	EnableSQLiteBackup = MustGetBool("GOSTARTER_ENABLE_SQLITE_BACKUP", true)
	BackupSQLiteFileName = MustGetString("GOSTARTER_BACKUP_SQLITE_FILE_NAME", "backup.db")
	databaseDriver := MustGetString("GOSTARTER_DATABASE_DRIVER", "sqlite")
	PostgresURL = MustGetString("GOSTARTER_POSTGRES_URL", "")
	SQLiteDbBusyTimeout = MustGetString("GOSTARTER_SQLITE_BUSY_TIMEOUT", "30000")
	SQLiteBackupCronSchedule = MustGetString("GOSTARTER_SQLITE_BACKUP_CRON_SCHEDULE", "0 0 * * *")
	EnableSessionCleanup = MustGetBool("GOSTARTER_ENABLE_SESSION_CLEANUP", true)
	SessionCleanupCronSchedule = MustGetString("GOSTARTER_SESSION_CLEANUP_CRON_SCHEDULE", "0 0 * * 0")
	LogLevel = MustGetInt("GOSTARTER_LOG_LEVEL", 0)
	LogHealthCheck = MustGetBool("GOSTARTER_LOG_HEALTH_CHECK", false)
	Port = MustGetString("GOSTARTER_PORT", "3000")
	CORSOrigins = MustGetString("GOSTARTER_CORS_ORIGINS", "*")
	PasswordHashingAlgorithm = MustGetString("GOSTARTER_PASSWORD_HASHING_ALGORITHM", PasswordHashingAlgorithmArgon2id)
	PasswordArgon2idMemory = MustGetInt("GOSTARTER_PASSWORD_ARGON2ID_MEMORY", 64*1024)
	PasswordArgon2idIterations = MustGetInt("GOSTARTER_PASSWORD_ARGON2ID_ITERATIONS", 3)
	PasswordArgon2idParallelism = MustGetInt("GOSTARTER_PASSWORD_ARGON2ID_PARALLELISM", 1)
	PasswordArgon2idSaltLength = MustGetInt("GOSTARTER_PASSWORD_ARGON2ID_SALT_LENGTH", 16)
	PasswordArgon2idKeyLength = MustGetInt("GOSTARTER_PASSWORD_ARGON2ID_KEY_LENGTH", 32)
	PasswordBcryptCost = MustGetInt("GOSTARTER_PASSWORD_BCRYPT_COST", 12)
	SessionCookieName = MustGetString("GOSTARTER_SESSION_COOKIE_NAME", "issho_session_token")
	SessionCookieHttpOnly = MustGetBool("GOSTARTER_SESSION_COOKIE_HTTP_ONLY", true)
	SessionCookieSecure = MustGetBool("GOSTARTER_SESSION_COOKIE_SECURE", false)
	sessionCookieSameSite := MustGetString("GOSTARTER_SESSION_COOKIE_SAME_SITE", "lax")
	SessionTokenLength = MustGetInt("GOSTARTER_SESSION_TOKEN_LENGTH", 32)
	SessionTokenCharset = MustGetString("GOSTARTER_SESSION_TOKEN_CHARSET", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	SessionLifetimeMin = MustGetInt("GOSTARTER_SESSION_LIFETIME_MIN", 60*24*7)
	SessionRefreshThresholdMin = MustGetInt("GOSTARTER_SESSION_REFRESH_THRESHOLD_MIN", 60*24)
	PreSessionLifetimeMin = MustGetInt("GOSTARTER_PRE_SESSION_LIFETIME_MIN", 15)
	CSRFTokenLength = MustGetInt("GOSTARTER_CSRF_TOKEN_LENGTH", 32)
	CSRFTokenCharset = MustGetString("GOSTARTER_CSRF_TOKEN_CHARSET", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	switch databaseDriver {
	case DatabaseDriverPostgreSQL:
		DatabaseDriver = DatabaseDriverPostgreSQL
	case DatabaseDriverSQLite:
		DatabaseDriver = DatabaseDriverSQLite
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
