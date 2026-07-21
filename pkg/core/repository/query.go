package repository

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
)

/*
NewQueries creates a new Queries instance with the provided database
connection. It returns a pointer to Queries.
*/
func NewQueries(db sqlx.ExtContext) *Queries {
	return &Queries{db: db}
}

/*
Queries is a struct that holds the database connection and provides methods
to execute queries.
*/
type Queries struct {
	db sqlx.ExtContext
}

func (q *Queries) NamedGetContext(ctx context.Context, dest any, query string, arg any) error {
	// dest must be a non-nil pointer
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return fmt.Errorf("dest must be a non-nil pointer")
	}

	// Select into a slice of the same element type as dest to count rows
	sliceType := reflect.SliceOf(destVal.Elem().Type())
	resultsPtr := reflect.New(sliceType)

	err := q.NamedSelectContext(ctx, resultsPtr.Interface(), query, arg)
	if err != nil {
		return err
	}

	results := resultsPtr.Elem()
	if results.Len() != 1 {
		return fmt.Errorf("expected to select 1 row, selected %d rows", results.Len())
	}

	// Copy the single result into dest
	destVal.Elem().Set(results.Index(0))
	return nil
}

func (q *Queries) GetContext(ctx context.Context, dest any, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, q.db, dest, query, args...)
}

func (q *Queries) NamedSelectContext(ctx context.Context, dest any, query string, arg any) error {
	query, args, err := q.db.BindNamed(query, arg)
	if err != nil {
		return err
	}
	return q.SelectContext(ctx, dest, query, args...)
}

func (q *Queries) SelectContext(ctx context.Context, dest any, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, q.db, dest, query, args...)
}

func (q *Queries) NamedExecOneRowContext(ctx context.Context, query string, arg any) error {
	rows, err := q.NamedExecRowsAffectedContext(ctx, query, arg)
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d rows", rows)
	}

	return nil
}

func (q *Queries) NamedExecRowsAffectedContext(ctx context.Context, query string, arg any) (int64, error) {
	query, args, err := q.db.BindNamed(query, arg)
	if err != nil {
		return 0, err
	}
	result, err := q.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (q *Queries) ExecRowsAffectedContext(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := q.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (q *Queries) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return q.db.ExecContext(ctx, query, args...)
}
