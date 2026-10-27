package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Queries struct {
	Queryer
}

func NewQueriesTx(ctx context.Context, db *Database, opts *sql.TxOptions) (*Queries, func(), func() error, error) {
	queryer, rollback, commit, err := NewQueryerTx(ctx, db, opts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create queryer with transaction: %w", err)
	}

	queries := &Queries{
		Queryer: *queryer,
	}

	return queries, rollback, commit, nil
}

func NewQueries(db *Database) (*Queries, error) {
	queryer, err := NewQueryer(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create queryer: %w", err)
	}

	queries := &Queries{
		Queryer: *queryer,
	}

	return queries, nil
}

// Queryer holds database connection(s) and provides basic methods to execute
// queries.
type Queryer struct {
	db sqlx.ExtContext
}

// NewQueryerTx creates a new [Queryer] instance with a transaction. It returns
// a [*Queryer], a rollback function, a commit function, and an error if any.
//
// The rollback function should be called in a defer statement to ensure that
// the transaction is rolled back if an error occurs. The commit function
// should be called to commit the transaction if all operations succeed.
func NewQueryerTx(ctx context.Context, db *Database, opts *sql.TxOptions) (*Queryer, func(), func() error, error) {
	sqlxDB, err := convertDB(db)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to convert database: %w", err)
	}

	tx, err := sqlxDB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	rollback := func() {
		tx.Rollback()
	}
	commit := tx.Commit

	queryer := &Queryer{
		db: tx,
	}

	return queryer, rollback, commit, nil
}

// NewQueryer creates a new Queryer instance with the provided database
// connection. It returns a pointer to Queryer and an error if any.
func NewQueryer(db *Database) (*Queryer, error) {
	sqlxDB, err := convertDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to convert database: %w", err)
	}

	return &Queryer{
		db: sqlxDB,
	}, nil
}

func convertDB(db *Database) (*sqlx.DB, error) {
	var sqlxDriverName string

	switch db.VarSyntax {
	case DollarSign:
		sqlxDriverName = "postgres"
	case QuestionMark:
		sqlxDriverName = "sqlite3"
	case ColonNamed:
		sqlxDriverName = "oci8"
	case AtSign:
		sqlxDriverName = "sqlserver"
	default:
		return nil, fmt.Errorf("unsupported variable syntax: %s", db.VarSyntax)
	}

	return sqlx.NewDb(db.DB, sqlxDriverName), nil
}

// NamedGetContext calls [Queryer.Named] to bind the named parameters in the
// query string with the provided argument, and then calls
// [sqlx.GetContext] to execute the query and scan the result into the
// destination struct.
//
// If the query returns no rows, it returns a [sql.ErrNoRows] error. This
// method is best used for queries that always return exactly one row, such as
// aggregation queries.
func (q *Queryer) NamedGetContext(ctx context.Context, dest any, query string, arg any) error {
	query, args, err := q.Named(query, arg)
	if err != nil {
		return fmt.Errorf("failed to bind named query: %w", err)
	}

	return sqlx.GetContext(ctx, q.db, dest, query, args...)
}

// NamedSelectContext calls [Queryer.Named] to bind the named parameters in
// the query string with the provided argument, and then calls
// [sqlx.SelectContext] to execute the query and scan the results into the
// destination slice.
func (q *Queryer) NamedSelectContext(ctx context.Context, dest any, query string, arg any) error {
	query, args, err := q.Named(query, arg)
	if err != nil {
		return fmt.Errorf("failed to bind named query: %w", err)
	}

	return sqlx.SelectContext(ctx, q.db, dest, query, args...)
}

// NamedExecOneRowContext binds the named parameters in the query string with
// the provided argument using [Queryer.Named], executes the query, and checks
// that exactly one row was affected. If not, it returns an error.
func (q *Queryer) NamedExecOneRowContext(ctx context.Context, query string, arg any) error {
	rows, err := q.NamedExecRowsAffectedContext(ctx, query, arg)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d rows", rows)
	}

	return nil
}

// NamedExecRowsAffectedContext binds the named parameters in the query string
// with the provided argument using [Queryer.Named], executes the query, and
// returns the number of rows affected as an [int64].
func (q *Queryer) NamedExecRowsAffectedContext(ctx context.Context, query string, arg any) (int64, error) {
	query, args, err := q.Named(query, arg)
	if err != nil {
		return 0, fmt.Errorf("failed to bind named query: %w", err)
	}

	result, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	return result.RowsAffected()
}

// Named binds the named parameters in the query string with the provided
// argument, and returns the resulting query string and arguments slice. It also
// calls [sqlx.In] to expand any slice arguments in the query string.
func (q *Queryer) Named(query string, arg any) (string, []any, error) {
	query, args, err := sqlx.Named(query, arg)
	if err != nil {
		return "", nil, fmt.Errorf("failed to bind named query: %w", err)
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return "", nil, fmt.Errorf("failed to bind in query: %w", err)
	}

	return q.db.Rebind(query), args, nil
}

// GetContext is a wrapper around [sqlx.GetContext] without arguments.
//
// If the query returns no rows, it returns an [sql.ErrNoRows] error. This is
// best used for queries that must return exactly one row, such as aggregation
// queries.
func (q *Queryer) GetContext(ctx context.Context, dest any, query string) error {
	return sqlx.GetContext(ctx, q.db, dest, query)
}

// SelectContext is a wrapper around [sqlx.SelectContext] without arguments.
func (q *Queryer) SelectContext(ctx context.Context, dest any, query string) error {
	return sqlx.SelectContext(ctx, q.db, dest, query)
}

// ExecContext executes a query without arguments and returns a [sql.Result]
func (q *Queryer) ExecContext(ctx context.Context, query string) (sql.Result, error) {
	return q.db.ExecContext(ctx, query)
}

// QueryContext executes a query without arguments and returns a [*sql.Rows].
func (q *Queryer) QueryContext(ctx context.Context, query string) (*sql.Rows, error) {
	return q.db.QueryContext(ctx, query)
}
