package repository

import (
	"github.com/jmoiron/sqlx"
)

/*
NewQueries creates a new Queries instance with the provided database
connection. It returns a pointer to Queries.
*/
func NewQueries(db sqlx.ExtContext) *Queries {
	return &Queries{db: db}
}

/*
Queries is a struct that holds the database connection and provides methods
to execute queries.
*/
type Queries struct {
	db sqlx.ExtContext
}
