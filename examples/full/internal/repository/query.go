package repository

import (
	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/core/repository"
)

type Queries struct {
	repository.Queries
}

func NewQueries(db sqlx.ExtContext) *Queries {
	return &Queries{
		Queries: *repository.NewQueries(db),
	}
}
