package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

// Queries is a struct extends [Queryer] and provides methods to execute
// predefined queries.
//
// Always call [Queries.Close] to close the connection when you are done using
// it.
//
// Use this struct when you are using any predefined queries.
type Queries struct {
	Queryer
}

// NewQueriesFromEnv creates a new [Queries] instance with the provided database
// connection. It returns a pointer to a [Queries].
func NewQueriesFromEnv(ctx context.Context, db *sql.DB) (*Queries, error) {
	queryer, err := NewQueryerFromEnv(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create queryer: %w", err)
	}

	return &Queries{
		Queryer: *queryer,
	}, nil
}

// Queryer holds a database connection as [sqlx.Conn] and provides various
// methods to execute queries.
//
// Always call [Queryer.Close] to close the connection when you are done using
// it.
//
// Use this struct when you are not using any predefined queries.
type Queryer struct {
	conn              *sqlx.Conn
	driverName        string
	useBeginConurrent bool
	transactionActive bool
}

// NewQueryerFromEnv calls [NewQueryer] with driver name from the environment
// variable.
//
// Always call [Queryer.Close] to close the connection when you are done using
// it.
func NewQueryerFromEnv(ctx context.Context, db *sql.DB) (*Queryer, error) {
	return NewQueryer(ctx, db, env.DatabaseDriver)
}

// NewQueryer creates a new [Queryer] instance with the provided [sql.DB]
// and driver name. It returns a pointer to a [Queryer].
//
// Always call [Queryer.Close] to close the connection when you are done using
// it.
func NewQueryer(ctx context.Context, db *sql.DB, driverName string) (*Queryer, error) {
	sqlxDB := sqlxDBFromDB(db, driverName)
	sqlxConn, err := sqlxDB.Connx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	return &Queryer{
		conn:              sqlxConn,
		driverName:        driverName,
		useBeginConurrent: false,
		transactionActive: false,
	}, nil
}

func sqlxDBFromDB(db *sql.DB, driverName string) *sqlx.DB {
	sqlxDriverName := driverName
	if sqlxDriverName == "turso" {
		// Swapping the driver name to "sqlite3" for sqlx, as sqlx does not
		// recognize "turso" as a valid driver name.
		sqlxDriverName = "sqlite3"
	}

	return sqlx.NewDb(db, sqlxDriverName)
}

// Close closes the database connection. Make sure there is no active
// transaction before calling this method.
func (q *Queryer) Close() {
	// golangci-lint: disable=errcheck
	q.conn.Close()
}

// BeginTx starts a new transaction. It returns an error if a transaction is
// already active.
func (q *Queryer) BeginTx(ctx context.Context) error {
	if q.transactionActive {
		return fmt.Errorf("transaction already active")
	}

	beginQuery := "BEGIN;"
	if q.useBeginConurrent {
		beginQuery = "BEGIN CONCURRENT;"
	}

	_, err := q.ExecContext(ctx, beginQuery)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	q.transactionActive = true

	return nil
}

// CommitTx commits the current transaction. It returns an error if there
// is no active transaction.
func (q *Queryer) CommitTx(ctx context.Context) error {
	if !q.transactionActive {
		return fmt.Errorf("no active transaction to commit")
	}

	_, err := q.ExecContext(ctx, "COMMIT;")
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	q.transactionActive = false

	return nil
}

// RollbackTx rolls back the current transaction.
//
// If there is no active transaction, it is a no-op.
func (q *Queryer) RollbackTx(ctx context.Context) {
	if !q.transactionActive {
		return
	}

	// golangci-lint: disable=errcheck
	q.ExecContext(ctx, "ROLLBACK;")

	q.transactionActive = false
}

// NamedGetContext calls [Queryer.Named] to bind the named parameters in the
// query string with the provided argument, and then calls
// [Queryer.GetContext] to execute the query and scan the result into the
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

	return q.GetContext(ctx, dest, query, args...)
}

// GetContext is a wrapper around [sqlx.GetContext].
//
// If the query returns no rows, it returns an [sql.ErrNoRows] error. This is
// best used for queries that must return exactly one row, such as aggregation
// queries.
func (q *Queryer) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return sqlx.GetContext(ctx, q.conn, dest, query, args...)
}

// NamedSelectContext calls [Queryer.Named] to bind the named parameters in
// the query string with the provided argument, and then calls
// [Queryer.SelectContext] to execute the query and scan the results into the
// destination slice.
func (q *Queryer) NamedSelectContext(ctx context.Context, dest any, query string, arg any) error {
	query, args, err := q.Named(query, arg)
	if err != nil {
		return fmt.Errorf("failed to bind named query: %w", err)
	}

	return q.SelectContext(ctx, dest, query, args...)
}

// SelectContext is a wrapper around [sqlx.SelectContext].
func (q *Queryer) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return sqlx.SelectContext(ctx, q.conn, dest, query, args...)
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

	return q.ExecRowsAffectedContext(ctx, query, args...)
}

// ExecRowsAffectedContext executes a query and returns the number of rows
// affected as an [int64].
func (q *Queryer) ExecRowsAffectedContext(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := q.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	return result.RowsAffected()
}

// ExecContext executes a query that returns a [sql.Result]
func (q *Queryer) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return q.conn.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns [sql.Rows].
func (q *Queryer) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return q.conn.QueryContext(ctx, query, args...)
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

	return q.conn.Rebind(query), args, nil
}
