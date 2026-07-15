package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

const createMigrationTable = `
	CREATE TABLE IF NOT EXISTS migration (
		id TEXT PRIMARY KEY,
		up_statement TEXT NOT NULL,
		down_statement TEXT NOT NULL,
		executed_at TEXT NOT NULL
	);
`

func (q *Queries) CreateMigrationTable(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, createMigrationTable)
	return err
}

const getAppliedMigrations = `
	SELECT
		*
	FROM
		migration
	ORDER BY
		id ASC;
`

func (q *Queries) GetAppliedMigrations(ctx context.Context) ([]Migration, error) {
	items := []Migration{}
	err := sqlx.SelectContext(ctx, q.db, &items, getAppliedMigrations)
	return items, err
}

const insertMigration = `
	INSERT INTO
		migration (
			id,
			up_statement,
			down_statement,
			executed_at
		) VALUES (
			:id,
			:up_statement,
			:down_statement,
			:executed_at
		);
`

func (q *Queries) InsertMigration(ctx context.Context, arg Migration) error {
	return NamedExecOneRowContext(ctx, q.db, insertMigration, arg)
}

const deleteMigration = `
	DELETE FROM
		migration
	WHERE
		id = :id;
`

type DeleteMigrationParams struct {
	ID string `db:"id"`
}

func (q *Queries) DeleteMigration(ctx context.Context, id string) error {
	return NamedExecOneRowContext(ctx, q.db, deleteMigration, DeleteMigrationParams{ID: id})
}
