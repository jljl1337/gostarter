package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

func NewSQLiteDBFromEnv() (*Database, error) {
	DBPath := filepath.Join(env.Variables.DataDir, env.Variables.LiveDataDir, env.Variables.SQLiteDir, env.Variables.SQLiteLiveFileName)
	return NewSQLiteDB(DBPath, env.Variables.SQLiteBusyTimeoutMs)
}

func NewSQLiteDB(path string, busyTimeout int) (*Database, error) {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directories for SQLite database: %w", err)
	}

	dsn := "file:" + path
	dsn = dsn + "?_journal=WAL"
	dsn = dsn + "&_foreign_keys=true"
	dsn = dsn + "&_txlock=immediate"
	dsn = dsn + "&_busy_timeout=" + strconv.Itoa(busyTimeout)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	return &Database{
		DB:        db,
		VarSyntax: QuestionMark,
	}, nil
}

func BackupSQLiteDBFromEnv(ctx context.Context, srcDB *Database) error {
	backupPath := filepath.Join(env.Variables.DataDir, env.Variables.BackupDataDir, env.Variables.SQLiteDir, env.Variables.SQLiteBackupFileName)
	return BackupSQLiteDB(ctx, srcDB.DB, backupPath)
}

func BackupSQLiteDB(ctx context.Context, srcDB *sql.DB, backupPath string) error {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(backupPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories for backup SQLite database: %w", err)
	}

	// Open destination database
	dstDB, err := sql.Open("sqlite3", backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup database: %w", err)
	}
	defer dstDB.Close()

	// Perform backup
	return backupSQLiteDB(ctx, dstDB, srcDB)
}

func backupSQLiteDB(ctx context.Context, dstDB, srcDB *sql.DB) error {
	// Get raw connections
	destConn, err := dstDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get destination connection: %w", err)
	}
	defer destConn.Close()

	srcConn, err := srcDB.Conn(ctx)
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
