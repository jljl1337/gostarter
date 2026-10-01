package repository

import (
	"context"
)

const getAccounts = `
	SELECT
		*
	FROM
		gs_account
	ORDER BY
		created_at DESC,
		username ASC
`

func (q *Queries) GetAccounts(ctx context.Context) ([]Account, error) {
	items := []Account{}
	err := q.SelectContext(ctx, &items, getAccounts)
	return items, err
}

const getAccountsByID = `
	SELECT
		*
	FROM
		gs_account
	WHERE
		id = :id
`

type GetAccountsByIDParams struct {
	ID string `db:"id"`
}

func (q *Queries) GetAccountsByID(ctx context.Context, id string) ([]Account, error) {
	items := []Account{}
	err := q.NamedSelectContext(ctx, &items, getAccountsByID, GetAccountsByIDParams{ID: id})
	return items, err
}

const updateAccountRole = `
	UPDATE
		gs_account
	SET
		role = :role,
		updated_at = :updated_at
	WHERE
		id = :id
`

type UpdateAccountRoleParams struct {
	Role      string `db:"role"`
	UpdatedAt string `db:"updated_at"`
	ID        string `db:"id"`
}

func (q *Queries) UpdateAccountRole(ctx context.Context, arg UpdateAccountRoleParams) error {
	return q.NamedExecOneRowContext(ctx, updateAccountRole, arg)
}
