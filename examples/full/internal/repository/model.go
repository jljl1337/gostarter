package repository

import (
	"github.com/jljl1337/gostarter/pkg/core/repository"
)

// Account is the core account model, aliased so the application queries can be
// written against the same type as the predefined core queries.
type Account = repository.Account

type Note struct {
	ID         string `json:"id" db:"id"`
	AccountID  string `json:"accountID" db:"account_id"`
	Body       string `json:"body" db:"body"`
	Positivity int    `json:"positivity" db:"positivity"`
	CreatedAt  string `json:"createdAt" db:"created_at"`
	UpdatedAt  string `json:"updatedAt" db:"updated_at"`
}
