package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

/*
NewSQLiteDBFromEnv creates a new SQLite database connection using the
environment variables defined in the env package. It returns a pointer to
sqlx.DB and an error if any occurs during the connection process.
*/
func NewSQLiteDBFromEnv() (*sqlx.DB, error) {
	DBPath := filepath.Join(env.DataDir, env.LiveDataDir, env.SQLiteDir, env.LiveSQLiteFileName)
	return NewSQLiteDB(DBPath, env.SQLiteDbBusyTimeout)
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

func BackupSQLiteDBFromEnv(srcDB *sqlx.DB) error {
	backupPath := filepath.Join(env.DataDir, env.BackupDataDir, env.SQLiteDir, env.BackupSQLiteFileName)
	return BackupSQLiteDB(srcDB, backupPath)
}

func BackupSQLiteDB(srcDB *sqlx.DB, backupPath string) error {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(backupPath), os.ModePerm); err != nil {
		return err
	}

	// Open destination database
	dstDB, err := sqlx.Open("sqlite3", backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup database: %w", err)
	}
	defer dstDB.Close()

	// Perform backup
	return backupSQLiteDB(dstDB, srcDB)
}

func backupSQLiteDB(dstDB, srcDB *sqlx.DB) error {
	// Get raw connections
	destConn, err := dstDB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get destination connection: %w", err)
	}
	defer destConn.Close()

	srcConn, err := srcDB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get source connection: %w", err)
	}
	defer srcConn.Close()

	// Perform backup using raw connections
	return destConn.Raw(func(destConn any) error {
		return srcConn.Raw(func(srcConn any) error {
			// Convert to SQLite connections
			destSQLiteConn, ok := destConn.(*sqlite3.SQLiteConn)
			if !ok {
				return fmt.Errorf("can't convert destination connection to SQLiteConn")
			}

			srcSQLiteConn, ok := srcConn.(*sqlite3.SQLiteConn)
			if !ok {
				return fmt.Errorf("can't convert source connection to SQLiteConn")
			}

			// Initialize backup
			b, err := destSQLiteConn.Backup("main", srcSQLiteConn, "main")
			if err != nil {
				return fmt.Errorf("error initializing SQLite backup: %w", err)
			}

			// Perform backup in one step (-1 means copy entire database)
			done, err := b.Step(-1)
			if err != nil {
				return fmt.Errorf("error in stepping backup: %w", err)
			}
			if !done {
				return fmt.Errorf("backup not completed in one step")
			}

			// Finish backup
			if err := b.Finish(); err != nil {
				return fmt.Errorf("error finishing backup: %w", err)
			}

			return nil
		})
	})
}
