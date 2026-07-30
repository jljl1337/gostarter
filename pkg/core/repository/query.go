package repository

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
)

/*
Queries is a struct extends [Queryer] and provides methods to execute
predefined queries.

Use this struct when you are using any predefined queries.
*/
type Queries struct {
	Queryer
}

/*
NewQueries creates a new Queries instance with the provided database
connection. It returns a pointer to Queries.
*/
func NewQueries(db sqlx.ExtContext) *Queries {
	return &Queries{
		Queryer: *NewQueryer(db),
	}
}

/*
Queryer is a struct that holds the database connection and provides basic
methods to execute queries.

Use this struct when you are not using any predefined queries.
*/
type Queryer struct {
	db sqlx.ExtContext
}

/*
NewQueryer creates a new Queryer instance with the provided database
connection. It returns a pointer to Queryer.
*/
func NewQueryer(db sqlx.ExtContext) *Queryer {
	return &Queryer{db: db}
}

func (q *Queryer) NamedGetContext(ctx context.Context, dest any, query string, arg any) error {
	return q.selectOneContext(ctx, dest, func(ctx context.Context, results any) error {
		return q.NamedSelectContext(ctx, results, query, arg)
	})
}

func (q *Queryer) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return q.selectOneContext(ctx, dest, func(ctx context.Context, results any) error {
		return q.SelectContext(ctx, results, query, args...)
	})
}

func (q *Queryer) selectOneContext(ctx context.Context, dest any, selectFn func(context.Context, any) error) error {
	// dest must be a non-nil pointer
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Pointer || destVal.IsNil() {
		return fmt.Errorf("dest must be a non-nil pointer")
	}

	// Select into a slice of the same element type as dest to count rows
	sliceType := reflect.SliceOf(destVal.Elem().Type())
	resultsPtr := reflect.New(sliceType)

	if err := selectFn(ctx, resultsPtr.Interface()); err != nil {
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

func (q *Queryer) NamedSelectContext(ctx context.Context, dest any, query string, arg any) error {
	query, args, err := q.db.BindNamed(query, arg)
	if err != nil {
		return err
	}
	return q.SelectContext(ctx, dest, query, args...)
}

func (q *Queryer) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return sqlx.SelectContext(ctx, q.db, dest, query, args...)
}

func (q *Queryer) NamedExecOneRowContext(ctx context.Context, query string, arg any) error {
	rows, err := q.NamedExecRowsAffectedContext(ctx, query, arg)
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d rows", rows)
	}

	return nil
}

func (q *Queryer) NamedExecRowsAffectedContext(ctx context.Context, query string, arg any) (int64, error) {
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

func (q *Queryer) ExecRowsAffectedContext(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := q.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (q *Queryer) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return q.db.ExecContext(ctx, query, args...)
}
