package migration

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jljl1337/gostarter/pkg/repository"
)

type Migration struct {
	ID            string `json:"id" db:"id"`
	UpStatement   string `json:"upStatement" db:"up_statement"`
	DownStatement string `json:"downStatement" db:"down_statement"`
	ExecutedAt    string `json:"executedAt" db:"executed_at"`
}

type NewRepositoryTx func(context.Context, *repository.Database, *sql.TxOptions) (Repository, repository.RollbackFunc, repository.CommitFunc, error)

type Repository interface {
	CreateMigrationTable(context.Context) error
	GetAppliedMigrations(context.Context) ([]Migration, error)
	RunStatement(context.Context, string) error
	InsertMigration(context.Context, Migration) error
	DeleteMigrationByID(context.Context, string) error
}

type defaultRepository struct {
	queryer *repository.Queryer
}

func newDefaultRepositoryTx(ctx context.Context, db *repository.Database, opts *sql.TxOptions) (Repository, repository.RollbackFunc, repository.CommitFunc, error) {
	queryer, rollback, commit, err := repository.NewQueryerTx(ctx, db, opts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create queryer with transaction: %w", err)
	}

	repository := &defaultRepository{
		queryer: queryer,
	}

	return repository, rollback, commit, nil
}

func (r *defaultRepository) CreateMigrationTable(ctx context.Context) error {
	const statement = `
		CREATE TABLE IF NOT EXISTS gs_migration (
			id TEXT PRIMARY KEY,
			up_statement TEXT NOT NULL,
			down_statement TEXT NOT NULL,
			executed_at TEXT NOT NULL
		);
	`

	_, err := r.queryer.ExecContext(ctx, statement)
	return err
}

func (r *defaultRepository) GetAppliedMigrations(ctx context.Context) ([]Migration, error) {
	const statement = `
		SELECT
			id,
			up_statement,
			down_statement,
			executed_at
		FROM
			gs_migration
		ORDER BY
			id ASC;
	`

	items := []Migration{}
	err := r.queryer.SelectContext(ctx, &items, statement)

	return items, err
}

func (r *defaultRepository) RunStatement(ctx context.Context, statement string) error {
	_, err := r.queryer.ExecContext(ctx, statement)
	return err
}

func (r *defaultRepository) InsertMigration(ctx context.Context, migration Migration) error {
	const statement = `
		INSERT INTO gs_migration (
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

	return r.queryer.NamedExecOneRowContext(ctx, statement, migration)
}

func (r *defaultRepository) DeleteMigrationByID(ctx context.Context, id string) error {
	const statement = `
		DELETE FROM
			gs_migration
		WHERE
			id = :id;
	`

	type deleteMigrationByIDParams struct {
		ID string `db:"id"`
	}

	return r.queryer.NamedExecOneRowContext(ctx, statement, deleteMigrationByIDParams{ID: id})
}
