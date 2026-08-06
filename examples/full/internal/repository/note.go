package repository

import (
	"context"
)

const createNote = `
	INSERT INTO note (
		id,
		account_id,
		body,
		created_at,
		updated_at
	) VALUES (
		:id,
		:account_id,
		:body,
		:created_at,
		:updated_at
	)
`

func (q *Queries) CreateNote(ctx context.Context, arg Note) error {
	return q.NamedExecOneRowContext(ctx, createNote, arg)
}

const getNotesByAccountID = `
	SELECT
		*
	FROM
		note
	WHERE
		account_id = :account_id
	ORDER BY
		updated_at DESC,
		created_at DESC,
		id DESC
`

type GetNotesByAccountIDParams struct {
	AccountID string `db:"account_id"`
}

func (q *Queries) GetNotesByAccountID(ctx context.Context, accountID string) ([]Note, error) {
	items := []Note{}
	err := q.NamedSelectContext(ctx, &items, getNotesByAccountID, GetNotesByAccountIDParams{AccountID: accountID})
	return items, err
}

const getNoteByID = `
	SELECT
		*
	FROM
		note
	WHERE
		id = :id
`

type GetNoteByIDParams struct {
	ID string `db:"id"`
}

func (q *Queries) GetNoteByID(ctx context.Context, id string) ([]Note, error) {
	items := []Note{}
	err := q.NamedSelectContext(ctx, &items, getNoteByID, GetNoteByIDParams{ID: id})
	return items, err
}

const updateNoteBodyByID = `
	UPDATE
		note
	SET
		body = :body,
		updated_at = :updated_at
	WHERE
		id = :id
`

type UpdateNoteBodyByIDParams struct {
	ID        string `db:"id"`
	Body      string `db:"body"`
	UpdatedAt string `db:"updated_at"`
}

func (q *Queries) UpdateNoteBodyByID(ctx context.Context, arg UpdateNoteBodyByIDParams) error {
	return q.NamedExecOneRowContext(ctx, updateNoteBodyByID, arg)
}

const updateNodePositivityByID = `
	UPDATE
		note
	SET
		positivity = :positivity,
		updated_at = :updated_at
	WHERE
		id = :id
`

type UpdateNotePositivityByIDParams struct {
	ID         string `db:"id"`
	Positivity int    `db:"positivity"`
	UpdatedAt  string `db:"updated_at"`
}

func (q *Queries) UpdateNotePositivityByID(ctx context.Context, arg UpdateNotePositivityByIDParams) error {
	return q.NamedExecOneRowContext(ctx, updateNodePositivityByID, arg)
}

const deleteNoteByID = `
	DELETE FROM
		note
	WHERE
		id = :id
`

type DeleteNoteByIDParams struct {
	ID string `db:"id"`
}

func (q *Queries) DeleteNoteByID(ctx context.Context, id string) error {
	return q.NamedExecOneRowContext(ctx, deleteNoteByID, DeleteNoteByIDParams{ID: id})
}

const deleteNotesByUpdatedAt = `
	DELETE FROM
		note
	WHERE
		updated_at < :updated_at
`

type DeleteNotesByUpdatedAtParams struct {
	UpdatedAt string `db:"updated_at"`
}

func (q *Queries) DeleteNotesByUpdatedAt(ctx context.Context, updatedAt string) (int64, error) {
	return q.NamedExecRowsAffectedContext(ctx, deleteNotesByUpdatedAt, DeleteNotesByUpdatedAtParams{UpdatedAt: updatedAt})
}
