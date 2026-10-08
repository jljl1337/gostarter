package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "turso.tech/database/tursogo"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

type TursoConfig struct {
	JournalMode             string
	ForeignKeys             bool
	RequireWhere            bool
	BusyTimeout             int
	MVCCCheckpointThreshold int
	MVCCGCThreshold         int
}

/*
NewTursoDBFromEnv creates a new Turso database connection using the
environment variables defined in the env package. It returns a pointer to
sql.DB and an error if any occurs during the connection process.
*/
func NewTursoDBFromEnv() (*sql.DB, error) {
	DBPath := filepath.Join(env.DataDir, env.LiveDataDir, env.TursoDir, env.TursoLiveFileName)

	return NewTursoDB(DBPath, TursoConfig{
		JournalMode:             env.TursoJournalMode,
		ForeignKeys:             env.TursoForeignKeys,
		RequireWhere:            env.TursoRequireWhere,
		BusyTimeout:             env.TursoBusyTimeout,
		MVCCCheckpointThreshold: env.TursoMVCCCheckpointThreshold,
		MVCCGCThreshold:         env.TursoMVCCGCThreshold,
	})
}

/*
NewTursoDB creates a new Turso database connection using the provided
path and config. It returns a [*sql.DB].
*/
func NewTursoDB(path string, config TursoConfig) (*sql.DB, error) {
	if config.JournalMode != "wal" && config.JournalMode != "mvcc" {
		return nil, fmt.Errorf("invalid journal mode: %s", config.JournalMode)
	}

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create parent directories for Turso database: %w", err)
	}

	db, err := sql.Open("turso", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Turso database: %w", err)
	}

	pragmaStmt := fmt.Sprintf("PRAGMA journal_mode=%s;", config.JournalMode)
	if config.ForeignKeys {
		pragmaStmt += "PRAGMA foreign_keys = ON;"
	} else {
		pragmaStmt += "PRAGMA foreign_keys = OFF;"
	}
	if config.RequireWhere {
		pragmaStmt += "PRAGMA require_where = 1;"
	} else {
		pragmaStmt += "PRAGMA require_where = 0;"
	}
	pragmaStmt += fmt.Sprintf("PRAGMA busy_timeout = %d;", config.BusyTimeout)
	pragmaStmt += fmt.Sprintf("PRAGMA mvcc_checkpoint_threshold = %d;", config.MVCCCheckpointThreshold)
	pragmaStmt += fmt.Sprintf("PRAGMA mvcc_gc_threshold = %d;", config.MVCCGCThreshold)

	_, err = db.Exec(pragmaStmt)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run PRAGMA statement: %w", err)
	}

	return db, nil
}

func BackupTursoDBFromEnv(ctx context.Context, srcDB *sql.DB) error {
	backupPath := filepath.Join(env.DataDir, env.BackupDataDir, env.TursoDir, env.TursoBackupFileName)
	return BackupTursoDB(ctx, srcDB, backupPath)
}

func BackupTursoDB(ctx context.Context, srcDB *sql.DB, backupPath string) error {
	// TODO
	return nil
}
