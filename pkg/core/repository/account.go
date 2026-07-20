package repository

import (
	"context"
)

const createAccount = `
	INSERT INTO gs_account (
		id,
		username,
		email,
		password_hash,
		role,
		language_code,
		is_verified,
		created_at,
		updated_at
	) VALUES (
		:id,
		:username,
		:email,
		:password_hash,
		:role,
		:language_code,
		:is_verified,
		:created_at,
		:updated_at
	)
`

func (q *Queries) CreateAccount(ctx context.Context, arg Account) error {
	return NamedExecOneRowContext(ctx, q.db, createAccount, arg)
}

const getAccountCountByRole = `
	SELECT
		COUNT(*) AS count
	FROM
		gs_account
	WHERE
		role = :role
`

type GetAccountCountByRoleParams struct {
	Role string `db:"role"`
}

func (q *Queries) GetAccountCountByRole(ctx context.Context, role string) (int, error) {
	var count int
	err := NamedGetContext(ctx, q.db, &count, getAccountCountByRole, GetAccountCountByRoleParams{Role: role})
	return count, err
}

const getAccountByID = `
	SELECT
		*
	FROM
		gs_account
	WHERE
		id = :id
`

type GetAccountByIDParams struct {
	ID string `db:"id"`
}

func (q *Queries) GetAccountByID(ctx context.Context, id string) (Account, error) {
	account := Account{}
	err := NamedGetContext(ctx, q.db, &account, getAccountByID, GetAccountByIDParams{ID: id})
	return account, err
}

const getAccountByUsername = `
	SELECT
		*
	FROM
		gs_account
	WHERE
		username = :username
`

type GetAccountByUsernameParams struct {
	Username string `db:"username"`
}

func (q *Queries) GetAccountByUsername(ctx context.Context, username string) ([]Account, error) {
	items := []Account{}
	err := NamedSelectContext(ctx, q.db, &items, getAccountByUsername, GetAccountByUsernameParams{Username: username})
	return items, err
}

const updateAccountPassword = `
	UPDATE
		gs_account
	SET
		password_hash = :password_hash,
		updated_at = :updated_at
	WHERE
		id = :id
`

type UpdateAccountPasswordParams struct {
	PasswordHash string `db:"password_hash"`
	UpdatedAt    string `db:"updated_at"`
	ID           string `db:"id"`
}

func (q *Queries) UpdateAccountPassword(ctx context.Context, arg UpdateAccountPasswordParams) error {
	return NamedExecOneRowContext(ctx, q.db, updateAccountPassword, arg)
}

const updateAccountUsername = `
	UPDATE
		gs_account
	SET
		username = :username,
		updated_at = :updated_at
	WHERE
		id = :id
`

type UpdateAccountUsernameParams struct {
	Username  string `db:"username"`
	UpdatedAt string `db:"updated_at"`
	ID        string `db:"id"`
}

func (q *Queries) UpdateAccountUsername(ctx context.Context, arg UpdateAccountUsernameParams) error {
	return NamedExecOneRowContext(ctx, q.db, updateAccountUsername, arg)
}

const updateAccountLanguage = `
	UPDATE
		gs_account
	SET
		language_code = :language_code,
		updated_at = :updated_at
	WHERE
		id = :id
`

type UpdateAccountLanguageParams struct {
	LanguageCode string `db:"language_code"`
	UpdatedAt    string `db:"updated_at"`
	ID           string `db:"id"`
}

func (q *Queries) UpdateAccountLanguage(ctx context.Context, arg UpdateAccountLanguageParams) error {
	return NamedExecOneRowContext(ctx, q.db, updateAccountLanguage, arg)
}

const deleteAccount = `
	DELETE FROM
		gs_account
	WHERE
		id = :id
`

type DeleteAccountParams struct {
	ID string `db:"id"`
}

func (q *Queries) DeleteAccount(ctx context.Context, id string) error {
	return NamedExecOneRowContext(ctx, q.db, deleteAccount, DeleteAccountParams{ID: id})
}
