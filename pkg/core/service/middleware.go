package service

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/format"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
)

type MiddlewareService struct {
	db *sqlx.DB
}

func NewMiddlewareService(db *sqlx.DB) *MiddlewareService {
	return &MiddlewareService{
		db: db,
	}
}

func (s *MiddlewareService) GetSessionAccountAndRefreshSession(ctx context.Context, sessionToken, CSRFToken string) (*repository.Account, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, NewServiceErrorf(ErrCodeInternal, "failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	queries := repository.NewQueries(tx)

	sessions, err := queries.GetSessionByToken(ctx, sessionToken)

	if err != nil {
		return nil, NewServiceErrorf(ErrCodeInternal, "failed to get session: %w", err)
	}

	if len(sessions) > 1 {
		return nil, NewServiceErrorf(ErrCodeInternal, "multiple sessions found with the same token")
	}

	if len(sessions) < 1 {
		return nil, NewServiceErrorf(ErrCodeUnauthorized, "unauthorized")
	}

	session := sessions[0]

	// Return unauthorized if the session is a pre session
	if session.AccountID == nil {
		return nil, NewServiceErrorf(ErrCodeUnauthorized, "unauthorized")
	}

	// CSRF token does not match
	if CSRFToken != "" && session.CsrfToken != CSRFToken {
		return nil, NewServiceErrorf(ErrCodeUnauthorized, "unauthorized")
	}

	// Session expired
	now := time.Now()
	nowISO8601 := format.TimeToISO8601(now)
	if session.ExpiresAt < nowISO8601 {
		return nil, NewServiceErrorf(ErrCodeUnauthorized, "unauthorized")
	}

	// Only refresh session if remaining lifetime is below threshold
	expiresAt, err := format.ISO8601ToTime(session.ExpiresAt)
	if err != nil {
		return nil, NewServiceErrorf(ErrCodeInternal, "failed to parse session expiration: %w", err)
	}

	remainingLifetimeMin := expiresAt.Sub(now).Minutes()
	if remainingLifetimeMin < float64(env.Variables.SessionRefreshThresholdMin) {
		newExpiresAt := generator.MinutesFromNowISO8601(env.Variables.SessionLifetimeMin)
		err := queries.UpdateSessionByToken(ctx, repository.UpdateSessionByTokenParams{
			Token:     sessionToken,
			ExpiresAt: newExpiresAt,
			UpdatedAt: nowISO8601,
		})
		if err != nil {
			return nil, NewServiceErrorf(ErrCodeInternal, "failed to refresh session: %w", err)
		}
	}

	// Get account associated with the session
	account, err := queries.GetAccountByID(ctx, *session.AccountID)
	if err != nil {
		return nil, NewServiceErrorf(ErrCodeInternal, "failed to get account by ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, NewServiceErrorf(ErrCodeInternal, "failed to commit transaction: %w", err)
	}

	return &account, nil
}
