package service

import (
	"context"

	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/generator"

	"github.com/jljl1337/gostarter/examples/full/internal/repository"
)

func (s *EndpointService) GetAccounts(ctx context.Context) ([]repository.Account, error) {
	queries := repository.NewQueries(s.db)

	accounts, err := queries.GetAccounts(ctx)
	if err != nil {
		return nil, service.NewServiceErrorf(service.ErrCodeInternal, "failed to get accounts: %v", err)
	}

	return accounts, nil
}

type UpdateAccountRoleByIDParams struct {
	Actor     repository.Account
	AccountID string
	NewRole   string
}

func (s *EndpointService) UpdateAccountRoleByID(ctx context.Context, arg UpdateAccountRoleByIDParams) error {
	if !s.roleManager.HasRole(arg.NewRole) {
		return service.NewServiceErrorf(service.ErrCodeUnprocessable, "unknown role %s", arg.NewRole)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	queries := repository.NewQueries(tx)

	account, err := s.getManagedAccountByID(ctx, queries, arg.Actor, arg.AccountID)
	if err != nil {
		return err
	}

	if account.Role == arg.NewRole {
		return service.NewServiceErrorf(service.ErrCodeUnprocessable, "account with ID %s already has role %s", arg.AccountID, arg.NewRole)
	}

	err = queries.UpdateAccountRole(ctx, repository.UpdateAccountRoleParams{
		ID:        arg.AccountID,
		Role:      arg.NewRole,
		UpdatedAt: generator.NowISO8601(),
	})
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to update account role: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to commit transaction: %v", err)
	}

	return nil
}

func (s *EndpointService) DeleteAccountByID(ctx context.Context, actor repository.Account, accountID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	queries := repository.NewQueries(tx)

	// Sessions and notes of the account are removed by the foreign keys
	_, err = s.getManagedAccountByID(ctx, queries, actor, accountID)
	if err != nil {
		return err
	}

	err = queries.DeleteAccount(ctx, accountID)
	if err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to delete account: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return service.NewServiceErrorf(service.ErrCodeInternal, "failed to commit transaction: %v", err)
	}

	return nil
}

// getManagedAccountByID returns the account only if the actor is allowed to
// manage it. Access to the route is already gated by role in the middleware,
// this guards the account being acted upon instead: an account can only be
// managed by a strictly higher role, so an owner cannot touch a peer owner and
// a moderator cannot touch another moderator. Deleting an account is still
// possible for anyone through their own account endpoint.
func (s *EndpointService) getManagedAccountByID(ctx context.Context, queries *repository.Queries, actor repository.Account, accountID string) (repository.Account, error) {
	accounts, err := queries.GetAccountsByID(ctx, accountID)
	if err != nil {
		return repository.Account{}, service.NewServiceErrorf(service.ErrCodeInternal, "failed to get account by ID: %v", err)
	}

	if len(accounts) == 0 {
		return repository.Account{}, service.NewServiceErrorf(service.ErrCodeNotFound, "account with ID %s not found", accountID)
	}

	if len(accounts) > 1 {
		return repository.Account{}, service.NewServiceErrorf(service.ErrCodeInternal, "multiple accounts found with ID %s", accountID)
	}

	account := accounts[0]

	compare, err := s.roleManager.CompareRoles(account.Role, actor.Role)
	if err != nil {
		return repository.Account{}, service.NewServiceErrorf(service.ErrCodeInternal, "failed to compare roles: %v", err)
	}

	if compare >= 0 {
		return repository.Account{}, service.NewServiceErrorf(
			service.ErrCodeForbidden,
			"role %s cannot manage role %s", actor.Role, account.Role,
		)
	}

	return account, nil
}
