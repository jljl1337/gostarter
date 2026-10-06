package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/jljl1337/gostarter/pkg/core/cron"
	"github.com/jljl1337/gostarter/pkg/core/migration"
	"github.com/jljl1337/gostarter/pkg/core/queue"
	"github.com/jljl1337/gostarter/pkg/core/transport"
	"github.com/jljl1337/gostarter/pkg/shared/crypto"
	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
	"github.com/jljl1337/gostarter/pkg/shared/log"
	"github.com/jljl1337/gostarter/pkg/shared/role"
)

const (
	DefaultUsernameRegex = "^[a-zA-Z0-9_]{3,20}$"
	DefaultPasswordRegex = "^[A-Za-z0-9!@#$%^&*]{8,64}$"
	DefaultRole          = "user"
	DefaultSocketPerm    = os.FileMode(0o666)
	socketProbeTimeout   = 500 * time.Millisecond
)

type Server struct {
	db                      *sql.DB
	runMigrations           bool
	runGostarterMigrations  bool
	appMigrationFS          fs.FS
	scheduler               *cron.Scheduler
	queueManager            *queue.QueueManager
	apiMux                  *http.ServeMux
	mux                     *http.ServeMux
	apiMiddleware           transport.Middleware
	port                    string
	socketPath              string
	socketPerm              os.FileMode
	httpServer              *http.Server
	gracefulShutdownTimeout time.Duration

	idGenerator     func() string
	usernameRegex   *regexp.Regexp
	passwordRegex   *regexp.Regexp
	roleManager     *role.RoleManager
	hashingManager  *crypto.HashingManager
	responseHandler *transport.ResponseHandler
	cookieGenerator *transport.CookieGenerator
}

func NewServer(options ...Option) (*Server, error) {
	hashingManager, err := crypto.NewHashingManagerFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create hashing manager: %w", err)
	}

	socketPerm := DefaultSocketPerm
	if env.SocketPath != "" {
		if socketPerm, err = parseSocketPerm(env.SocketPerm); err != nil {
			return nil, fmt.Errorf("failed to create server: %w", err)
		}
	}

	usernameRegex, err := regexp.Compile(DefaultUsernameRegex)
	if err != nil {
		return nil, fmt.Errorf("failed to compile default username regex: %w", err)
	}

	passwordRegex, err := regexp.Compile(DefaultPasswordRegex)
	if err != nil {
		return nil, fmt.Errorf("failed to compile default password regex: %w", err)
	}

	server := &Server{
		runMigrations:           false,
		runGostarterMigrations:  false,
		port:                    env.Port,
		socketPath:              env.SocketPath,
		socketPerm:              socketPerm,
		apiMux:                  http.NewServeMux(),
		mux:                     http.NewServeMux(),
		gracefulShutdownTimeout: time.Duration(env.GracefulShutdownTimeoutSec) * time.Second,

		idGenerator:     generator.NewULID,
		usernameRegex:   usernameRegex,
		passwordRegex:   passwordRegex,
		hashingManager:  hashingManager,
		responseHandler: transport.NewDefaultResponseHandler(),
		cookieGenerator: transport.NewCookieGeneratorFromEnv(),
		roleManager:     role.NewRoleManager(DefaultRole),
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
	if s.httpServer == nil {
		return fmt.Errorf("http server must be initialized before starting the server")
	}

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

	if s.queueManager != nil {
		log.Info("Starting queue manager")
		if err := s.queueManager.Resume(); err != nil {
			return fmt.Errorf("failed to resume queue manager: %w", err)
		}
	}

	if s.socketPath != "" {
		listener, err := s.listenUnixSocket()
		if err != nil {
			return err
		}
		defer s.removeSocketFile()

		log.Infof("Starting http server on unix socket %s", s.socketPath)
		if err := s.httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start http server: %w", err)
		}

		return nil
	}

	log.Infof("Starting http server on %s", s.httpServer.Addr)
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start http server: %w", err)
	}

	return nil
}

// parseSocketPerm parses an octal file mode, such as "0666", as used by the
// socket permission environment variable.
func parseSocketPerm(perm string) (os.FileMode, error) {
	value, err := strconv.ParseUint(perm, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid socket permission %q, expected an octal file mode such as 0666: %w", perm, err)
	}

	return os.FileMode(value), nil
}

// listenUnixSocket creates the unix socket, making its parent directory when
// missing, so paths under directories like /run work out of the box.
func (s *Server) listenUnixSocket() (net.Listener, error) {
	if dir := filepath.Dir(s.socketPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create socket directory %s: %w", dir, err)
		}
	}

	if err := s.removeStaleSocketFile(); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on unix socket %s: %w", s.socketPath, err)
	}

	if err := os.Chmod(s.socketPath, s.socketPerm); err != nil {
		if closeErr := listener.Close(); closeErr != nil {
			log.Errorf("Failed to close unix socket listener: %v", closeErr)
		}
		s.removeSocketFile()
		return nil, fmt.Errorf("failed to set permission on unix socket %s: %w", s.socketPath, err)
	}

	return listener, nil
}

// removeStaleSocketFile deletes the socket file left behind by a previous
// process, but refuses to take over a socket that is still being served.
func (s *Server) removeStaleSocketFile() error {
	if _, err := os.Stat(s.socketPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to stat unix socket %s: %w", s.socketPath, err)
	}

	if conn, err := net.DialTimeout("unix", s.socketPath, socketProbeTimeout); err == nil {
		if closeErr := conn.Close(); closeErr != nil {
			log.Errorf("Failed to close unix socket probe connection: %v", closeErr)
		}
		return fmt.Errorf("unix socket %s is already in use", s.socketPath)
	}

	if err := os.Remove(s.socketPath); err != nil {
		return fmt.Errorf("failed to remove stale unix socket %s: %w", s.socketPath, err)
	}

	return nil
}

// removeSocketFile unlinks the socket, and must only be called once this
// process has created it, never when another server still owns the path.
func (s *Server) removeSocketFile() {
	if s.socketPath == "" {
		return
	}

	if err := os.Remove(s.socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Errorf("Failed to remove unix socket %s: %v", s.socketPath, err)
	}
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("Stopping server")

	log.Info("Stopping HTTP server")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop HTTP server: %w", err)
	}

	if s.queueManager != nil {
		log.Info("Stopping queue manager")
		if err := s.queueManager.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to stop queue manager: %w", err)
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
