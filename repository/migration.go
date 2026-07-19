package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

const createMigrationTable = `
	CREATE TABLE IF NOT EXISTS gs_gostarter_migration (
		id TEXT PRIMARY KEY,
		up_statement TEXT NOT NULL,
		down_statement TEXT NOT NULL,
		executed_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS gs_app_migration (
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

const getAppliedGostarterMigrations = `
	SELECT
		*
	FROM
		gs_gostarter_migration
	ORDER BY
		id ASC;
`

func (q *Queries) GetAppliedGostarterMigrations(ctx context.Context) ([]Migration, error) {
	items := []Migration{}
	err := sqlx.SelectContext(ctx, q.db, &items, getAppliedGostarterMigrations)
	return items, err
}

const getAppliedAppMigrations = `
	SELECT
		*
	FROM
		gs_app_migration
	ORDER BY
		id ASC;
`

func (q *Queries) GetAppliedAppMigrations(ctx context.Context) ([]Migration, error) {
	items := []Migration{}
	err := sqlx.SelectContext(ctx, q.db, &items, getAppliedAppMigrations)
	return items, err
}

const insertGostarterMigration = `
	INSERT INTO
		gs_gostarter_migration (
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

func (q *Queries) InsertGostarterMigration(ctx context.Context, arg Migration) error {
	return NamedExecOneRowContext(ctx, q.db, insertGostarterMigration, arg)
}

const insertAppMigration = `
	INSERT INTO
		gs_app_migration (
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

func (q *Queries) InsertAppMigration(ctx context.Context, arg Migration) error {
	return NamedExecOneRowContext(ctx, q.db, insertAppMigration, arg)
}

const deleteGostarterMigration = `
	DELETE FROM
		gs_gostarter_migration
	WHERE
		id = :id;
`

type DeleteGostarterMigrationParams struct {
	ID string `db:"id"`
}

func (q *Queries) DeleteGostarterMigration(ctx context.Context, id string) error {
	return NamedExecOneRowContext(ctx, q.db, deleteGostarterMigration, DeleteGostarterMigrationParams{ID: id})
}

const deleteAppMigration = `
	DELETE FROM
		gs_app_migration
	WHERE
		id = :id;
`

type DeleteAppMigrationParams struct {
	ID string `db:"id"`
}

func (q *Queries) DeleteAppMigration(ctx context.Context, id string) error {

	return NamedExecOneRowContext(ctx, q.db, deleteAppMigration, DeleteAppMigrationParams{ID: id})
}
