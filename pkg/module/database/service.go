package database

import (
	"context"
	"fmt"

	"github.com/jljl1337/gostarter/pkg/repository"
)

type Service struct {
	dbMap            map[string]*repository.Database
	newDB            func(module string) (*repository.Database, error)
	sqliteConfig     SQLiteConfig
	postgresqlConfig PostgreSQLConfig
}

type SQLiteConfig struct {
	Dir           string
	BusyTimeoutMs int
}

type PostgreSQLConfig struct {
	Username     string
	Password     string
	Host         string
	Port         int
	DatabaseName string
}

func NewService(optionList ...ServiceOption) (*Service, error) {
	service := &Service{
		dbMap: make(map[string]*repository.Database),
	}

	for _, option := range optionList {
		err := option(service)
		if err != nil {
			return nil, fmt.Errorf("failed to apply service option: %w", err)
		}
	}

	if service.newDB == nil {
		return nil, fmt.Errorf("newDB function is not set")
	}

	return service, nil
}

func (s *Service) NewModuleDB(module string) (*repository.Database, error) {
	if module == "" {
		return nil, fmt.Errorf("module name cannot be empty")
	}

	if _, exists := s.dbMap[module]; exists {
		return nil, fmt.Errorf("database for module %s already exists", module)
	}

	db, err := s.newDB(module)
	if err != nil {
		return nil, fmt.Errorf("failed to create database for module %s: %w", module, err)
	}

	s.dbMap[module] = db

	return db, nil
}

func (s *Service) newSQLiteDB(module string) (*repository.Database, error) {
	path := fmt.Sprintf("%s/%s.db", s.sqliteConfig.Dir, module)
	db, err := repository.NewSQLiteDB(path, s.sqliteConfig.BusyTimeoutMs)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQLite database: %w", err)
	}

	return db, nil
}

func (s *Service) newPostgreSQLDB(module string) (*repository.Database, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?search_path=%s",
		s.postgresqlConfig.Username,
		s.postgresqlConfig.Password,
		s.postgresqlConfig.Host,
		s.postgresqlConfig.Port,
		s.postgresqlConfig.DatabaseName,
		module,
	)

	db, err := repository.NewPostgreSQLDB(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL database: %w", err)
	}

	return db, nil
}

func (s *Service) pingAll() error {
	for _, db := range s.dbMap {
		if err := db.Ping(); err != nil {
			return fmt.Errorf("failed to start database service: %w", err)
		}
	}

	return nil
}

func (s *Service) closeAll(ctx context.Context) error {
	for _, db := range s.dbMap {
		if err := db.Close(ctx); err != nil {
			return fmt.Errorf("failed to stop database service: %w", err)
		}
	}

	return nil
}

type ServiceOption func(*Service) error

func WithSQLiteConfig(cfg SQLiteConfig) ServiceOption {
	return func(s *Service) error {
		if cfg.Dir == "" {
			return fmt.Errorf("SQLite directory is not specified")
		}

		if cfg.BusyTimeoutMs <= 0 {
			return fmt.Errorf("SQLite busy timeout must be greater than 0")
		}

		s.sqliteConfig = cfg
		s.newDB = s.newSQLiteDB
		return nil
	}
}

func WithPostgreSQLConfig(cfg PostgreSQLConfig) ServiceOption {
	return func(s *Service) error {
		if cfg.Username == "" || cfg.Password == "" || cfg.Host == "" || cfg.Port == 0 || cfg.DatabaseName == "" {
			return fmt.Errorf("incomplete PostgreSQL configuration")
		}

		s.postgresqlConfig = cfg
		s.newDB = s.newPostgreSQLDB

		return nil
	}
}

func WithNewDBFunc(newDB func(module string) (*repository.Database, error)) ServiceOption {
	return func(s *Service) error {
		if newDB == nil {
			return fmt.Errorf("newDB function cannot be nil")
		}

		s.newDB = newDB
		return nil
	}
}
