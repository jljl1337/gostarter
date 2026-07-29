package service

import (
	"context"

	"github.com/jljl1337/gostarter/pkg/core/repository"
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
		return NewServiceError(ErrCodeUnprocessable, "invalid new username format")
	}

	if arg.Account.Username == arg.NewUsername {
		return NewServiceError(ErrCodeUnprocessable, "new username must be different from the old username")
	}

	queries := repository.NewQueries(s.db)

	// Check if new username is the same as the old one or already taken
	accounts, err := queries.GetAccountByUsername(ctx, arg.NewUsername)
	if err != nil {
		return NewServiceErrorf(ErrCodeInternal, "failed to get account: %v", err)
	}

	if len(accounts) > 1 {
		return NewServiceError(ErrCodeInternal, "multiple accounts found with the same ID")
	}

	if len(accounts) == 1 {
		return NewServiceError(ErrCodeUsernameTaken, "username already taken")
	}

	err = queries.UpdateAccountUsername(ctx, repository.UpdateAccountUsernameParams{
		ID:        arg.Account.ID,
		Username:  arg.NewUsername,
		UpdatedAt: generator.NowISO8601(),
	})
	if err != nil {
		return NewServiceErrorf(ErrCodeInternal, "failed to update username: %v", err)
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
		return NewServiceError(ErrCodeUnprocessable, "invalid new password format")
	}

	if arg.OldPassword == arg.NewPassword {
		return NewServiceError(ErrCodeUnprocessable, "new password must be different from the old password")
	}

	queries := repository.NewQueries(s.db)

	valid, err := s.hashingManager.ComparePassword(arg.OldPassword, arg.Account.PasswordHash)
	if err != nil {
		return NewServiceErrorf(ErrCodeInternal, "failed to compare passwords: %v", err)
	}
	if !valid {
		return NewServiceError(ErrCodeUnprocessable, "old password is incorrect")
	}

	// Update password hash
	passwordHash, err := s.hashingManager.HashPassword(arg.NewPassword)
	if err != nil {
		return NewServiceErrorf(ErrCodeInternal, "failed to hash password: %v", err)
	}

	err = queries.UpdateAccountPassword(ctx, repository.UpdateAccountPasswordParams{
		PasswordHash: passwordHash,
		UpdatedAt:    generator.NowISO8601(),
		ID:           arg.Account.ID,
	})
	if err != nil {
		return NewServiceErrorf(ErrCodeInternal, "failed to update password: %v", err)
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
		return NewServiceError(ErrCodeUnprocessable, "invalid language code")
	}

	if arg.Account.LanguageCode == arg.LanguageCode {
		return NewServiceError(ErrCodeUnprocessable, "new language code must be different from the old language code")
	}

	queries := repository.NewQueries(s.db)

	err := queries.UpdateAccountLanguage(ctx, repository.UpdateAccountLanguageParams{
		ID:           arg.Account.ID,
		LanguageCode: arg.LanguageCode,
		UpdatedAt:    generator.NowISO8601(),
	})
	if err != nil {
		return NewServiceErrorf(ErrCodeInternal, "failed to update language: %v", err)
	}

	return nil
}

func (s *EndpointService) DeleteAccountByID(ctx context.Context, account repository.Account) error {
	// Delete user record
	queries := repository.NewQueries(s.db)

	err := queries.DeleteAccount(ctx, account.ID)
	if err != nil {
		return NewServiceErrorf(ErrCodeInternal, "failed to delete account: %v", err)
	}

	return nil
}
