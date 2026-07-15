package env

import "net/http"

const (
	DatabaseDriverSQLite             = "sqlite"
	DatabaseDriverPostgreSQL         = "postgresql"
	PasswordHashingAlgorithmArgon2id = "argon2id"
	PasswordHashingAlgorithmBcrypt   = "bcrypt"
)

var (
	// Internal variables

	ConstantsSet bool = false

	// Plain variables from environment variables

	PostgresURL                string
	SQLiteDbPath               string
	SQLiteDbBusyTimeout        string
	SQLiteBackupDbPath         string
	SQLiteBackupCronSchedule   string
	SessionCleanupCronSchedule string
	LogLevel                   int
	LogHealthCheck             bool
	Port                       string
	CORSOrigins                string
	PasswordBcryptCost         int
	SessionCookieName          string
	SessionCookieHttpOnly      bool
	SessionCookieSecure        bool
	SessionTokenLength         int
	SessionTokenCharset        string
	SessionLifetimeMin         int
	SessionRefreshThresholdMin int
	PreSessionLifetimeMin      int
	CSRFTokenLength            int
	CSRFTokenCharset           string
	PageSizeMax                int
	PageSizeDefault            int

	// Derived variables

	DatabaseDriver            string
	SessionCookieSameSiteMode http.SameSite
)

func MustSetConstants() {
	MustLoadOptionalEnvFile()

	databaseDriver := MustGetString("GOSTARTER_DATABASE_DRIVER", "sqlite")
	PostgresURL = MustGetString("GOSTARTER_POSTGRES_URL", "")
	SQLiteDbPath = MustGetString("GOSTARTER_SQLITE_DB_PATH", "data/live/db/live.db")
	SQLiteDbBusyTimeout = MustGetString("GOSTARTER_SQLITE_BUSY_TIMEOUT", "30000")
	SQLiteBackupDbPath = MustGetString("GOSTARTER_SQLITE_BACKUP_DB_PATH", "data/backup/db/backup.db")
	SQLiteBackupCronSchedule = MustGetString("GOSTARTER_SQLITE_BACKUP_CRON_SCHEDULE", "0 0 * * *")
	SessionCleanupCronSchedule = MustGetString("GOSTARTER_SESSION_CLEANUP_CRON_SCHEDULE", "0 0 * * 0")
	LogLevel = MustGetInt("GOSTARTER_LOG_LEVEL", 0)
	LogHealthCheck = MustGetBool("GOSTARTER_LOG_HEALTH_CHECK", false)
	Port = MustGetString("GOSTARTER_PORT", "3000")
	CORSOrigins = MustGetString("GOSTARTER_CORS_ORIGINS", "*")
	PasswordBcryptCost = MustGetInt("GOSTARTER_PASSWORD_BCRYPT_COST", 12)
	SessionCookieName = MustGetString("GOSTARTER_SESSION_COOKIE_NAME", "issho_session_token")
	SessionCookieHttpOnly = MustGetBool("GOSTARTER_SESSION_COOKIE_HTTP_ONLY", true)
	SessionCookieSecure = MustGetBool("GOSTARTER_SESSION_COOKIE_SECURE", false)
	sessionCookieSameSite := MustGetString("GOSTARTER_SESSION_COOKIE_SAME_SITE_MODE", "lax")
	SessionTokenLength = MustGetInt("GOSTARTER_SESSION_TOKEN_LENGTH", 32)
	SessionTokenCharset = MustGetString("GOSTARTER_SESSION_TOKEN_CHARSET", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	SessionLifetimeMin = MustGetInt("GOSTARTER_SESSION_LIFETIME_MIN", 60*24*7)
	SessionRefreshThresholdMin = MustGetInt("GOSTARTER_SESSION_REFRESH_THRESHOLD_MIN", 60*24)
	PreSessionLifetimeMin = MustGetInt("GOSTARTER_PRE_SESSION_LIFETIME_MIN", 15)
	CSRFTokenLength = MustGetInt("GOSTARTER_CSRF_TOKEN_LENGTH", 32)
	CSRFTokenCharset = MustGetString("GOSTARTER_CSRF_TOKEN_CHARSET", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	PageSizeMax = MustGetInt("GOSTARTER_PAGE_SIZE_MAX", 100)
	PageSizeDefault = MustGetInt("GOSTARTER_PAGE_SIZE_DEFAULT", 10)

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
		SessionCookieSameSiteMode = http.SameSiteLaxMode
	case "strict":
		SessionCookieSameSiteMode = http.SameSiteStrictMode
	case "none":
		SessionCookieSameSiteMode = http.SameSiteNoneMode
	default:
		SessionCookieSameSiteMode = http.SameSiteNoneMode
	}

	ConstantsSet = true
}
