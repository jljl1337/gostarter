package repository

import (
	"context"
)

const createSession = `
	INSERT INTO gs_session (
		id,
		account_id,
		token,
		csrf_token,
		expires_at,
		created_at,
		updated_at
	) VALUES (
		:id,
		:account_id,
		:token,
		:csrf_token,
		:expires_at,
		:created_at,
		:updated_at
	)
`

func (q *Queries) CreateSession(ctx context.Context, arg Session) error {
	return NamedExecOneRowContext(ctx, q.db, createSession, arg)
}

const getSessionByToken = `
	SELECT
		*
	FROM
		gs_session
	WHERE
		token = :token
`

type GetSessionByTokenParams struct {
	Token string `db:"token"`
}

func (q *Queries) GetSessionByToken(ctx context.Context, token string) ([]Session, error) {
	items := []Session{}
	err := NamedSelectContext(ctx, q.db, &items, getSessionByToken, GetSessionByTokenParams{Token: token})
	return items, err
}

const updateSessionByToken = `
	UPDATE
		gs_session
	SET
		expires_at = :expires_at,
		updated_at = :updated_at
	WHERE
		token = :token
`

type UpdateSessionByTokenParams struct {
	ExpiresAt string `db:"expires_at"`
	UpdatedAt string `db:"updated_at"`
	Token     string `db:"token"`
}

func (q *Queries) UpdateSessionByToken(ctx context.Context, arg UpdateSessionByTokenParams) error {
	return NamedExecOneRowContext(ctx, q.db, updateSessionByToken, arg)
}

const updateSessionByAccountID = `
	UPDATE
		gs_session
	SET
		expires_at = :expires_at,
		updated_at = :updated_at
	WHERE
		account_id = :account_id AND
		expires_at > :expires_at
`

type UpdateSessionByAccountIDParams struct {
	ExpiresAt string  `db:"expires_at"`
	UpdatedAt string  `db:"updated_at"`
	AccountID *string `db:"account_id"`
}

func (q *Queries) UpdateSessionByAccountID(ctx context.Context, arg UpdateSessionByAccountIDParams) (int64, error) {
	return NamedExecRowsAffectedContext(ctx, q.db, updateSessionByAccountID, arg)
}

const deleteSessionByExpiresAt = `
	DELETE FROM
		gs_session
	WHERE
		expires_at < :expires_at
`

type DeleteSessionByExpiresAtParams struct {
	ExpiresAt string `db:"expires_at"`
}

func (q *Queries) DeleteSessionByExpiresAt(ctx context.Context, expiresAt string) (int64, error) {
	return NamedExecRowsAffectedContext(ctx, q.db, deleteSessionByExpiresAt, DeleteSessionByExpiresAtParams{ExpiresAt: expiresAt})
}
