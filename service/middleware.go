package service

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/env"
	"github.com/jljl1337/gostarter/format"
	"github.com/jljl1337/gostarter/generator"
	"github.com/jljl1337/gostarter/repository"
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
	queries := repository.NewQueries(s.db)

	sessions, err := queries.GetSessionByToken(ctx, sessionToken)

	if err != nil {
		return nil, NewServiceErrorf(ErrCodeInternal, "failed to get session: %v", err)
	}

	if len(sessions) > 1 {
		return nil, NewServiceError(ErrCodeInternal, "multiple sessions found with the same token")
	}

	if len(sessions) < 1 {
		return nil, NewServiceError(ErrCodeUnauthorized, "unauthorized")
	}

	session := sessions[0]

	// Return unauthorized if the session is a pre session
	if session.AccountID == nil {
		return nil, NewServiceError(ErrCodeUnauthorized, "unauthorized")
	}

	// CSRF token does not match
	if CSRFToken != "" && session.CsrfToken != CSRFToken {
		return nil, NewServiceError(ErrCodeUnauthorized, "unauthorized")
	}

	// Session expired
	now := time.Now()
	nowISO8601 := format.TimeToISO8601(now)
	if session.ExpiresAt < nowISO8601 {
		return nil, NewServiceError(ErrCodeUnauthorized, "unauthorized")
	}

	// Only refresh session if remaining lifetime is below threshold
	expiresAt, err := format.ISO8601ToTime(session.ExpiresAt)
	if err != nil {
		return nil, NewServiceErrorf(ErrCodeInternal, "failed to parse session expiration: %v", err)
	}

	remainingLifetimeMin := expiresAt.Sub(now).Minutes()
	if remainingLifetimeMin < float64(env.SessionRefreshThresholdMin) {
		newExpiresAt := generator.MinutesFromNowISO8601(env.SessionLifetimeMin)
		err := queries.UpdateSessionByToken(ctx, repository.UpdateSessionByTokenParams{
			Token:     sessionToken,
			ExpiresAt: newExpiresAt,
			UpdatedAt: nowISO8601,
		})
		if err != nil {
			return nil, NewServiceErrorf(ErrCodeInternal, "failed to refresh session: %v", err)
		}
	}

	// Get account associated with the session
	account, err := queries.GetAccountByID(ctx, *session.AccountID)
	if err != nil {
		return nil, NewServiceErrorf(ErrCodeInternal, "failed to get account by ID: %v", err)
	}

	return &account, nil
}
