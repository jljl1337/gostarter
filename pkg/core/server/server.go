package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/core/cron"
	"github.com/jljl1337/gostarter/pkg/core/migration"
	"github.com/jljl1337/gostarter/pkg/core/transport"
	"github.com/jljl1337/gostarter/pkg/shared/crypto"
	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
	"github.com/jljl1337/gostarter/pkg/shared/log"
)

const (
	DefaultLanguageCode  = "en-US"
	DefaultUsernameRegex = "^[a-zA-Z0-9_]{3,20}$"
	DefaultPasswordRegex = "^[A-Za-z0-9!@#$%^&*]{8,64}$"
)

type Server struct {
	db                      *sqlx.DB
	runMigrations           bool
	runGostarterMigrations  bool
	appMigrationFS          embed.FS
	scheduler               *cron.Scheduler
	apiMux                  *http.ServeMux
	mux                     *http.ServeMux
	apiMiddleware           transport.Middleware
	port                    string
	gracefulShutdownTimeout time.Duration
	httpServer              *http.Server

	idGenerator      func() string
	languageCodeList []string
	usernameRegex    string
	passwordRegex    string
	hashingManager   *crypto.HashingManager
	responseHandler  *transport.ResponseHandler
	cookieGenerator  *transport.CookieGenerator
}

func NewServer(options ...Option) (*Server, error) {
	hashingManager, err := crypto.NewHashingManagerFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create hashing manager: %w", err)
	}

	server := &Server{
		runMigrations:           false,
		runGostarterMigrations:  false,
		port:                    env.Port,
		gracefulShutdownTimeout: time.Duration(env.GracefulShutdownTimeoutSec) * time.Second,
		apiMux:                  http.NewServeMux(),
		mux:                     http.NewServeMux(),

		idGenerator:      generator.NewULID,
		languageCodeList: []string{DefaultLanguageCode},
		usernameRegex:    DefaultUsernameRegex,
		passwordRegex:    DefaultPasswordRegex,
		hashingManager:   hashingManager,
		responseHandler:  transport.NewDefaultResponseHandler(),
		cookieGenerator:  transport.NewCookieGeneratorFromEnv(),
	}

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("failed to apply server option: %w", err)
		}
	}

	return server, nil
}

func (s *Server) StartWithGracefulShutdown() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startErrCh := make(chan error, 1)
	go func() {
		startErrCh <- s.Start()
	}()

	select {
	case err := <-startErrCh:
		if err != nil {
			log.Errorf("Server stopped with error: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.gracefulShutdownTimeout)
		defer cancel()

		if err := s.Stop(shutdownCtx); err != nil {
			log.Errorf("Failed to stop server gracefully: %v", err)
		}

		if err := <-startErrCh; err != nil {
			log.Errorf("Server stopped with error: %v", err)
		}
	}
}

func (s *Server) Start() error {
	log.Info("Starting server")

	if s.db != nil {
		log.Info("Testing database connection")
		if err := s.db.Ping(); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
		log.Info("Database connection successful")
	}

	if s.runMigrations {
		log.Info("Running migrations")
		if err := migration.Migrate(s.db, s.runGostarterMigrations, s.appMigrationFS); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	if s.scheduler != nil {
		log.Info("Starting scheduler")
		s.scheduler.Start()
	}

	if s.httpServer != nil {
		log.Infof("Starting http server on %s", s.httpServer.Addr)
		err := s.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start http server: %w", err)
		}
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("Stopping server")

	if s.httpServer != nil {
		log.Info("Stopping HTTP server")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to stop HTTP server: %w", err)
		}
	}

	if s.scheduler != nil {
		log.Info("Stopping scheduler")
		if err := s.scheduler.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to stop scheduler: %w", err)
		}
	}

	if s.db != nil {
		log.Info("Closing database connection")
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	log.Info("Server stopped successfully")

	return nil
}
