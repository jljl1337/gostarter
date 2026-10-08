package repository

import (
	"context"
)

type PragmaParams struct {
	JournalMode             string `db:"journal_mode"`
	ForeignKeys             int    `db:"foreign_keys"`
	RequireWhere            int    `db:"require_where"`
	BusyTimeout             int    `db:"busy_timeout"`
	MVCCCheckpointThreshold int    `db:"mvcc_checkpoint_threshold"`
	MVCCGCThreshold         int    `db:"mvcc_gc_threshold"`
}

func (q *Queries) Pragma(ctx context.Context, arg PragmaParams) error {
	const pragmaQuery = `
		PRAGMA journal_mode = :journal_mode;
		PRAGMA foreign_keys = :foreign_keys;
		PRAGMA require_where = :require_where;
		PRAGMA busy_timeout = :busy_timeout;
		PRAGMA mvcc_checkpoint_threshold = :mvcc_checkpoint_threshold;
		PRAGMA mvcc_gc_threshold = :mvcc_gc_threshold;
	`

	_, err := q.NamedExecRowsAffectedContext(ctx, pragmaQuery, arg)
	return err
}

func (q *Queries) VacuumInto(ctx context.Context, filePath string) error {
	type VacuumParams struct {
		FilePath string `db:"file_path"`
	}

	const vacuumQuery = `
		VACUUM INTO :file_path;
	`

	_, err := q.NamedExecRowsAffectedContext(ctx, vacuumQuery, VacuumParams{FilePath: filePath})
	return err
}
