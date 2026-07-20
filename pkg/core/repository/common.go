package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

/*
NamedExecOneRowContext executes a named query and expects exactly one row to
be affected. If the number of affected rows is not equal to one, it returns
an error.
*/
func NamedExecOneRowContext(ctx context.Context, q sqlx.ExtContext, query string, arg any) error {
	rows, err := NamedExecRowsAffectedContext(ctx, q, query, arg)
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d rows", rows)
	}

	return nil
}

/*
NamedExecRowsAffectedContext executes a named query and returns the number of
rows affected. It returns an error if the query execution fails.
*/
func NamedExecRowsAffectedContext(ctx context.Context, q sqlx.ExtContext, query string, arg any) (int64, error) {
	query, args, err := q.BindNamed(query, arg)
	if err != nil {
		return 0, err
	}
	return ExecRowsAffectedContext(ctx, q, query, args...)
}

/*
NamedGetContext executes a named query and expects exactly one row to be returned.
If the number of returned rows is not equal to one, it returns an error.
*/
func NamedGetContext[DestStruct any](ctx context.Context, q sqlx.ExtContext, dest *DestStruct, query string, arg any) error {
	// Select into a slice to count rows
	results := []DestStruct{}
	err := NamedSelectContext(ctx, q, &results, query, arg)
	if err != nil {
		return err
	}

	if len(results) != 1 {
		return fmt.Errorf("expected to select 1 row, selected %d rows", len(results))
	}

	// Copy the single result into dest
	*dest = results[0]

	return nil
}

/*
NamedSelectContext executes a named query and populates the dest slice with
the results. It returns an error if the query execution fails.
*/
func NamedSelectContext(ctx context.Context, q sqlx.ExtContext, dest any, query string, arg any) error {
	query, args, err := q.BindNamed(query, arg)
	if err != nil {
		return err
	}
	return sqlx.SelectContext(ctx, q, dest, query, args...)
}

/*
ExecRowsAffectedContext executes a query and returns the number of rows
affected. It returns an error if the query execution fails.
*/
func ExecRowsAffectedContext(ctx context.Context, q sqlx.ExtContext, query string, args ...any) (int64, error) {
	result, err := q.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
