package endpoint

import (
	"context"
	"time"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/format"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
	"github.com/jljl1337/gostarter/pkg/shared/log"
)

type SignUpParams struct {
	Username     string
	Password     string
	LanguageCode string
}

func (s *EndpointService) SignUp(ctx context.Context, arg SignUpParams) error {
	usernameValid := s.validationManager.ValidateUsername(arg.Username)
	if !usernameValid {
		return service.NewServiceError(service.ErrCodeUnprocessable, "invalid username format")
	}

	passwordValid := s.validationManager.ValidatePassword(arg.Password)
	if !passwordValid {
		return service.NewServiceError(service.ErrCodeUnprocessable, "invalid password format")
	}

	languageCodeValid := s.validationManager.ValidateLanguageCode(arg.LanguageCode)
	if !languageCodeValid {
		return service.NewServiceError(service.ErrCodeUnprocessable, "invalid language code")
	}

	queries := repository.NewQueries(s.db)

	users, err := queries.GetAccountByUsername(ctx, arg.Username)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to get user by username: %v", err)
	}

	if len(users) > 1 {
		return service.NewServiceError(service.ErrCodeInternal, "multiple users found with the same username")
	}

	if len(users) > 0 {
		return service.NewServiceError(service.ErrCodeUsernameTaken, "username already exists")
	}

	passwordHash, err := s.hashingManager.HashPassword(arg.Password)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to hash password: %v", err)
	}

	currentTime := generator.NowISO8601()

	// ownerCount, err := queries.GetUserCountByRole(ctx, env.OwnerRole)
	// if err != nil {
	// 	return service.NewServiceErrorf(service.ErrCodeInternal, "failed to get owner user count: %v", err)
	// }

	// role := env.UserRole
	// isVerified := false
	// if ownerCount == 0 {
	// 	role = env.OwnerRole
	// 	isVerified = true
	// }

	if err = queries.CreateAccount(ctx, repository.Account{
		ID:           generator.NewULID(),
		Username:     arg.Username,
		PasswordHash: passwordHash,
		Role:         "user", // TODO
		LanguageCode: arg.LanguageCode,
		CreatedAt:    currentTime,
		UpdatedAt:    currentTime,
	}); err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to create user: %v", err)
	}

	return nil
}

// GetPreSession creates a pre-session with no associated user.
// It returns a non-empty session token and CSRF token.
func (s *EndpointService) GetPreSession(ctx context.Context) (string, string, error) {
	queries := repository.NewQueries(s.db)

	sessionID := generator.NewULID()
	sessionToken := generator.NewToken(env.SessionTokenLength, env.SessionTokenCharset)
	CSRFToken := generator.NewToken(env.CSRFTokenLength, env.CSRFTokenCharset)
	currentTime := generator.NowISO8601()
	expiresAt := generator.MinutesFromNowISO8601(env.PreSessionLifetimeMin)

	if err := queries.CreateSession(ctx, repository.Session{
		ID:        sessionID,
		AccountID: nil,
		Token:     sessionToken,
		CsrfToken: CSRFToken,
		ExpiresAt: expiresAt,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}); err != nil {
		return "", "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to create pre-session: %v", err)
	}

	return sessionToken, CSRFToken, nil
}

type SignInParams struct {
	PreSessionToken     string
	PreSessionCSRFToken string
	Username            string
	Password            string
}

// SignIn authenticates a user and creates a new session.
// It returns non-empty session token and CSRF token if the credentials are valid.
func (s *EndpointService) SignIn(ctx context.Context, arg SignInParams) (string, string, error) {
	if arg.Username == "" {
		return "", "", service.NewServiceError(service.ErrCodeBadRequest, "username must be provided")
	}

	queries := repository.NewQueries(s.db)

	// Validate pre-session
	sessions, err := queries.GetSessionByToken(ctx, arg.PreSessionToken)

	if err != nil {
		return "", "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to get pre-session: %v", err)
	}

	if len(sessions) > 1 {
		return "", "", service.NewServiceError(service.ErrCodeInternal, "multiple sessions found with the same token")
	}

	if len(sessions) < 1 {
		log.Debug("Session not found")
		return "", "", service.NewServiceError(service.ErrCodeUnauthorized, "invalid pre-session")
	}

	session := sessions[0]

	// Check if the session is already associated with a user
	if session.AccountID != nil {
		log.Debug("Session is is not a pre-session")
		return "", "", service.NewServiceError(service.ErrCodeUnauthorized, "invalid pre-session")
	}

	// CSRF token does not match
	if arg.PreSessionCSRFToken != "" && session.CsrfToken != arg.PreSessionCSRFToken {
		log.Debug("CSRF token does not match")
		return "", "", service.NewServiceError(service.ErrCodeUnauthorized, "csrf token does not match")
	}

	// Session expired
	now := time.Now()
	nowISO8601 := format.TimeToISO8601(now)
	if session.ExpiresAt < nowISO8601 {
		return "", "", service.NewServiceError(service.ErrCodeUnauthorized, "pre-session expired")
	}

	// Validate credentials
	accounts, err := queries.GetAccountByUsername(ctx, arg.Username)
	if err != nil {
		return "", "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to get account by username: %v", err)
	}

	if len(accounts) > 1 {
		return "", "", service.NewServiceError(service.ErrCodeInternal, "multiple accounts found with the same username")
	}

	if len(accounts) < 1 {
		log.Debug("Account not found")
		return "", "", service.NewServiceError(service.ErrCodeInvalidCredentials, "invalid credentials")
	}

	account := accounts[0]

	valid, err := s.hashingManager.ComparePassword(account.PasswordHash, arg.Password)
	if err != nil {
		return "", "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to compare password: %v", err)
	}

	if !valid {
		log.Debug("Invalid password")
		return "", "", service.NewServiceError(service.ErrCodeInvalidCredentials, "invalid credentials")
	}

	// Rehash password if needed
	currentTime := generator.NowISO8601()

	needsRehash, err := s.hashingManager.CheckIfNeedsRehash(account.PasswordHash)
	if err != nil {
		return "", "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to check if password needs rehash: %v", err)
	}

	if needsRehash {
		newHash, err := s.hashingManager.HashPassword(arg.Password)
		if err != nil {
			return "", "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to hash password: %v", err)
		}

		err = queries.UpdateAccountPassword(ctx, repository.UpdateAccountPasswordParams{
			PasswordHash: newHash,
			UpdatedAt:    currentTime,
			ID:           account.ID,
		})
		if err != nil {
			return "", "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to update account password hash: %v", err)
		}
	}

	sessionID := generator.NewULID()
	sessionToken := generator.NewToken(env.SessionTokenLength, env.SessionTokenCharset)
	CSRFToken := generator.NewToken(env.CSRFTokenLength, env.CSRFTokenCharset)
	expiresAt := generator.MinutesFromNowISO8601(env.SessionLifetimeMin)

	// Deactivate the pre-session
	err = queries.UpdateSessionByToken(ctx, repository.UpdateSessionByTokenParams{
		Token:     arg.PreSessionToken,
		ExpiresAt: nowISO8601,
		UpdatedAt: nowISO8601,
	})
	if err != nil {
		return "", "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to update pre-session: %v", err)
	}

	// Create a new session associated with the user
	err = queries.CreateSession(ctx, repository.Session{
		ID:        sessionID,
		AccountID: &account.ID,
		Token:     sessionToken,
		CsrfToken: CSRFToken,
		ExpiresAt: expiresAt,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return "", "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to create session: %v", err)
	}

	return sessionToken, CSRFToken, nil
}

func (s *EndpointService) SignOut(ctx context.Context, sessionToken string) error {
	queries := repository.NewQueries(s.db)

	now := generator.NowISO8601()
	err := queries.UpdateSessionByToken(ctx, repository.UpdateSessionByTokenParams{
		Token:     sessionToken,
		ExpiresAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to sign out session: %v", err)
	}

	return nil
}

func (s *EndpointService) SignOutAllSession(ctx context.Context, account repository.Account) error {
	queries := repository.NewQueries(s.db)

	now := generator.NowISO8601()
	rows, err := queries.UpdateSessionByAccountID(ctx, repository.UpdateSessionByAccountIDParams{
		AccountID: &account.ID,
		ExpiresAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to sign out all sessions: %v", err)
	}

	if rows < 1 {
		return service.NewServiceError(service.ErrCodeInternal, "no sessions deleted")
	}

	return nil
}

func (s *EndpointService) CSRFToken(ctx context.Context, sessionToken string) (string, error) {
	queries := repository.NewQueries(s.db)

	sessions, err := queries.GetSessionByToken(ctx, sessionToken)

	if err != nil {
		return "", service.NewServiceErrorf(service.ErrCodeInternal, "failed to get session: %v", err)
	}

	if len(sessions) > 1 {
		return "", service.NewServiceError(service.ErrCodeInternal, "multiple sessions found with the same token")
	}

	if len(sessions) < 1 {
		return "", service.NewServiceError(service.ErrCodeUnauthorized, "invalid session")
	}

	session := sessions[0]

	// Check if pre-session is expired
	if session.ExpiresAt < generator.NowISO8601() {
		return "", service.NewServiceError(service.ErrCodeUnauthorized, "unauthorized")
	}

	return session.CsrfToken, nil
}
