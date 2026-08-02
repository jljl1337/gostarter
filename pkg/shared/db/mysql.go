package db

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

func NewMySQLDBFromEnv() (*sqlx.DB, error) {
	return NewMySQLDB(env.MySQLURL, env.MySQLMaxLifetimeMin, env.MySQLMaxOpenConns, env.MySQLMaxIdleConns)
}

func NewMySQLDB(url string, maxLifetimeMin, maxOpenConns, maxIdleConns int) (*sqlx.DB, error) {
	if url == "" {
		return nil, fmt.Errorf("MySQL URL is missing")
	}
	db, err := sqlx.Open("mysql", url)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL database: %w", err)
	}

	db.SetConnMaxLifetime(time.Duration(maxLifetimeMin) * time.Minute)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	return db, nil
}
