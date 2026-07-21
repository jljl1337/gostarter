package repository

import (
	"context"
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
	_, err := q.ExecContext(ctx, createMigrationTable)
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
	err := q.SelectContext(ctx, &items, getAppliedGostarterMigrations)
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
	err := q.SelectContext(ctx, &items, getAppliedAppMigrations)
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
	return q.NamedExecOneRowContext(ctx, insertGostarterMigration, arg)
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
	return q.NamedExecOneRowContext(ctx, insertAppMigration, arg)
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
	return q.NamedExecOneRowContext(ctx, deleteGostarterMigration, DeleteGostarterMigrationParams{ID: id})
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

	return q.NamedExecOneRowContext(ctx, deleteAppMigration, DeleteAppMigrationParams{ID: id})
}
