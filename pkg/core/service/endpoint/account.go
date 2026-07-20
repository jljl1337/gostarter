package endpoint

import (
	"context"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
)

type UpdateUsernameByIDParams struct {
	Account     repository.Account
	NewUsername string
}

func (s *EndpointService) UpdateUsernameByID(ctx context.Context, arg UpdateUsernameByIDParams) error {
	// Validate new username
	newUsernameValid := s.validationManager.ValidateUsername(arg.NewUsername)
	if !newUsernameValid {
		return service.NewServiceError(service.ErrCodeUnprocessable, "invalid new username format")
	}

	if arg.Account.Username == arg.NewUsername {
		return service.NewServiceError(service.ErrCodeUnprocessable, "new username must be different from the old username")
	}

	queries := repository.NewQueries(s.db)

	// Check if new username is the same as the old one or already taken
	accounts, err := queries.GetAccountByUsername(ctx, arg.NewUsername)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to get account: %v", err)
	}

	if len(accounts) > 1 {
		return service.NewServiceError(service.ErrCodeInternal, "multiple accounts found with the same ID")
	}

	if len(accounts) == 1 {
		return service.NewServiceError(service.ErrCodeUsernameTaken, "username already taken")
	}

	err = queries.UpdateAccountUsername(ctx, repository.UpdateAccountUsernameParams{
		ID:        arg.Account.ID,
		Username:  arg.NewUsername,
		UpdatedAt: generator.NowISO8601(),
	})
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to update username: %v", err)
	}

	return nil
}

type UpdatePasswordByIDParams struct {
	Account     repository.Account
	OldPassword string
	NewPassword string
}

func (s *EndpointService) UpdatePasswordByID(ctx context.Context, arg UpdatePasswordByIDParams) error {
	newPasswordValid := s.validationManager.ValidatePassword(arg.NewPassword)
	if !newPasswordValid {
		return service.NewServiceError(service.ErrCodeUnprocessable, "invalid new password format")
	}

	if arg.OldPassword == arg.NewPassword {
		return service.NewServiceError(service.ErrCodeUnprocessable, "new password must be different from the old password")
	}

	queries := repository.NewQueries(s.db)

	valid, err := s.hashingManager.ComparePassword(arg.OldPassword, arg.Account.PasswordHash)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to compare passwords: %v", err)
	}
	if !valid {
		return service.NewServiceError(service.ErrCodeUnprocessable, "old password is incorrect")
	}

	// Update password hash
	passwordHash, err := s.hashingManager.HashPassword(arg.NewPassword)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to hash password: %v", err)
	}

	err = queries.UpdateAccountPassword(ctx, repository.UpdateAccountPasswordParams{
		PasswordHash: passwordHash,
		UpdatedAt:    generator.NowISO8601(),
		ID:           arg.Account.ID,
	})
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to update password: %v", err)
	}

	return nil
}

type UpdateLanguageByIDParams struct {
	Account      repository.Account
	LanguageCode string
}

func (s *EndpointService) UpdateLanguageByID(ctx context.Context, arg UpdateLanguageByIDParams) error {
	languageCodeValid := s.validationManager.ValidateLanguageCode(arg.LanguageCode)
	if !languageCodeValid {
		return service.NewServiceError(service.ErrCodeUnprocessable, "invalid language code")
	}

	if arg.Account.LanguageCode == arg.LanguageCode {
		return service.NewServiceError(service.ErrCodeUnprocessable, "new language code must be different from the old language code")
	}

	queries := repository.NewQueries(s.db)

	err := queries.UpdateAccountLanguage(ctx, repository.UpdateAccountLanguageParams{
		ID:           arg.Account.ID,
		LanguageCode: arg.LanguageCode,
		UpdatedAt:    generator.NowISO8601(),
	})
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to update language: %v", err)
	}

	return nil
}

func (s *EndpointService) DeleteAccountByID(ctx context.Context, account repository.Account) error {
	// Delete user record
	queries := repository.NewQueries(s.db)

	err := queries.DeleteAccount(ctx, account.ID)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to delete account: %v", err)
	}

	return nil
}
