package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

/*
NewSQLiteDBFromEnv creates a new SQLite database connection using the
environment variables defined in the env package. It returns a pointer to
sql.DB and an error if any occurs during the connection process.
*/
func NewSQLiteDBFromEnv() (*sql.DB, error) {
	DBPath := filepath.Join(env.DataDir, env.LiveDataDir, env.SQLiteDir, env.SQLiteLiveFileName)
	return NewSQLiteDB(DBPath, env.SQLiteDbBusyTimeout)
}

/*
NewSQLiteDB creates a new SQLite database connection using the provided
path and busy timeout. It returns a pointer to sql.DB and an error if any
occurs during the connection process.
*/
func NewSQLiteDB(path, busyTimeout string) (*sql.DB, error) {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}

	dsn := "file:" + path
	dsn = dsn + "?_journal=WAL"
	dsn = dsn + "&_foreign_keys=true"
	dsn = dsn + "&_txlock=immediate"
	dsn = dsn + "&_busy_timeout=" + busyTimeout
	return sql.Open("sqlite3", dsn)
}

func BackupSQLiteDBFromEnv(srcDB *sql.DB) error {
	backupPath := filepath.Join(env.DataDir, env.BackupDataDir, env.SQLiteDir, env.SQLiteBackupFileName)
	return BackupSQLiteDB(srcDB, backupPath)
}

func BackupSQLiteDB(srcDB *sql.DB, backupPath string) error {
	// TODO
	return nil
}
