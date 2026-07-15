package db

import (
	"os"
	"path/filepath"

	"github.com/jljl1337/gostarter/env"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

/*
NewSQLiteDBFromEnv creates a new SQLite database connection using the
environment variables defined in the env package. It returns a pointer to
sqlx.DB and an error if any occurs during the connection process.
*/
func NewSQLiteDBFromEnv() (*sqlx.DB, error) {
	return NewSQLiteDB(env.SQLiteDbPath, env.SQLiteDbBusyTimeout)
}

/*
NewSQLiteDB creates a new SQLite database connection using the provided
path and busy timeout. It returns a pointer to sqlx.DB and an error if any
occurs during the connection process.
*/
func NewSQLiteDB(path, busyTimeout string) (*sqlx.DB, error) {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}

	dsn := "file:" + path
	dsn = dsn + "?_journal=WAL"
	dsn = dsn + "&_foreign_keys=true"
	dsn = dsn + "&_busy_timeout=" + busyTimeout
	return sqlx.Open("sqlite3", dsn)
}
