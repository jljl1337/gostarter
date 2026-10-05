package server

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/jljl1337/gostarter/pkg/core/cron"
	"github.com/jljl1337/gostarter/pkg/core/queue"
	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/core/transport"
	"github.com/jljl1337/gostarter/pkg/shared/crypto"
	"github.com/jljl1337/gostarter/pkg/shared/role"
)

type Option func(*Server) error

func WithDB(db *sql.DB) Option {
	return func(s *Server) error {
		if db == nil {
			return fmt.Errorf("database connection cannot be nil")
		}

		s.db = db
		return nil
	}
}

func WithGostarterMigration() Option {
	return func(s *Server) error {
		s.runMigrations = true
		s.runGostarterMigrations = true
		return nil
	}
}

func WithAppMigrations(appMigrationFS embed.FS) Option {
	return func(s *Server) error {
		s.runMigrations = true
		s.appMigrationFS = appMigrationFS
		return nil
	}
}

func WithCustomIDGenerator(idGenerator func() string) Option {
	return func(s *Server) error {
		s.idGenerator = idGenerator
		return nil
	}
}

func WithCustomUsernameRegex(usernameRegex *regexp.Regexp) Option {
	return func(s *Server) error {
		if usernameRegex == nil {
			return fmt.Errorf("username regex cannot be nil")
		}

		s.usernameRegex = usernameRegex
		return nil
	}
}

func WithCustomPasswordRegex(passwordRegex *regexp.Regexp) Option {
	return func(s *Server) error {
		if passwordRegex == nil {
			return fmt.Errorf("password regex cannot be nil")
		}

		s.passwordRegex = passwordRegex
		return nil
	}
}

func WithCustomHashers(hasherList ...crypto.Hasher) Option {
	return func(s *Server) error {
		hashingManager, err := crypto.NewHashingManager(hasherList...)
		if err != nil {
			return fmt.Errorf("failed to create hashing manager: %w", err)
		}

		return WithCustomHashingManager(hashingManager)(s)
	}
}

func WithCustomHashingManager(hashingManager *crypto.HashingManager) Option {
	return func(s *Server) error {
		s.hashingManager = hashingManager
		return nil
	}
}

func WithCustomResponseHandler(responseHandler *transport.ResponseHandler) Option {
	return func(s *Server) error {
		s.responseHandler = responseHandler
		return nil
	}
}

func WithCustomCookieGenerator(cookieGenerator *transport.CookieGenerator) Option {
	return func(s *Server) error {
		s.cookieGenerator = cookieGenerator
		return nil
	}
}

func WithRoles(roleList ...string) Option {
	return func(s *Server) error {
		roleManager := role.NewRoleManager(roleList...)
		return WithCustomRoleManager(roleManager)(s)
	}
}

func WithCustomRoleManager(roleManager *role.RoleManager) Option {
	return func(s *Server) error {
		s.roleManager = roleManager
		return nil
	}
}

func WithDefaultScheduler(jobList ...cron.Job) Option {
	return func(s *Server) error {
		schedulerService := service.NewSchedulerService(s.db)
		defaultJobList := cron.DefaultSchedulerJobFromEnv(schedulerService)

		return WithScheduler(append(defaultJobList, jobList...)...)(s)
	}
}

func WithQueueManager(queueManager *queue.QueueManager) Option {
	return func(s *Server) error {
		if queueManager == nil {
			return fmt.Errorf("queue manager cannot be nil")
		}

		s.queueManager = queueManager
		return nil
	}
}

func WithScheduler(jobList ...cron.Job) Option {
	return func(s *Server) error {
		scheduler, err := cron.NewScheduler()
		if err != nil {
			return err
		}

		for _, j := range jobList {
			if err := scheduler.AddJob(j); err != nil {
				return err
			}
		}
		s.scheduler = scheduler
		return nil
	}
}

func WithPort(port string) Option {
	return func(s *Server) error {
		s.port = port
		return nil
	}
}

// WithUnixSocket makes the HTTP server listen on the unix domain socket at path
// instead of a TCP port, so the server is only reachable by processes able to
// open that file. Must be applied before WithHttpServer.
func WithUnixSocket(path string) Option {
	return func(s *Server) error {
		s.socketPath = path
		return nil
	}
}

// WithUnixSocketPerm sets the file mode of the created unix socket, allowing
// clients running as another user to connect. Defaults to 0666.
func WithUnixSocketPerm(perm os.FileMode) Option {
	return func(s *Server) error {
		s.socketPerm = perm
		return nil
	}
}

func WithGracefulShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) error {
		s.gracefulShutdownTimeout = timeout
		return nil
	}
}

func WithStaticSite(path string, siteFs fs.FS, subPath string) Option {
	return func(s *Server) error {
		webHandler, err := transport.NewWebHandler(path, siteFs, subPath)
		if err != nil {
			return fmt.Errorf("failed to create web handler: %w", err)
		}
		webHandler.RegisterRoutes(s.mux)

		return nil
	}
}

func WithDefaultMiddleware() Option {
	return func(s *Server) error {
		middlewareService := service.NewMiddlewareService(s.db)
		middlewareProvider := transport.NewMiddlewareProvider(middlewareService, s.responseHandler, s.roleManager)
		return WithMiddleware(middlewareProvider.GetMiddlewareList()...)(s)
	}
}

func WithMiddleware(middlewareList ...transport.Middleware) Option {
	return func(s *Server) error {
		s.apiMiddleware = transport.CreateStack(middlewareList...)
		return nil
	}
}

func WithDefaultApiHandler(handlerList ...transport.Handler) Option {
	return func(s *Server) error {
		endpointService := service.NewEndpointService(service.EndpointServiceConfig{
			DB:             s.db,
			IDGenerator:    s.idGenerator,
			HashingManager: s.hashingManager,
			UsernameRegex:  s.usernameRegex,
			PasswordRegex:  s.passwordRegex,
			RoleManager:    s.roleManager,
		})
		endpointHandler := transport.NewEndpointHandler(endpointService, s.responseHandler, s.cookieGenerator)
		handlerList = append([]transport.Handler{endpointHandler}, handlerList...)
		return WithApiHandler("/api", handlerList...)(s)
	}
}

func WithApiHandler(subpath string, handlerList ...transport.Handler) Option {
	return func(s *Server) error {
		for _, h := range handlerList {
			h.RegisterRoutes(s.apiMux)
		}

		s.mux.Handle(subpath+"/", http.StripPrefix(subpath, s.apiMiddleware(s.apiMux)))
		return nil
	}
}

func WithHttpServer() Option {
	return func(s *Server) error {
		addr := ":" + s.port
		if s.socketPath != "" {
			addr = s.socketPath
		}

		s.httpServer = &http.Server{
			Addr:    addr,
			Handler: s.mux,
		}
		return nil
	}
}
