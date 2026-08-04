package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/jljl1337/gostarter/pkg/core/cron"
	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/core/transport"
	"github.com/jljl1337/gostarter/pkg/shared/crypto"
	"github.com/jljl1337/gostarter/pkg/shared/validation"
)

type Option func(*Server) error

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

func WithCustomLanguageCodeList(languageCodeList ...string) Option {
	return func(s *Server) error {
		s.languageCodeList = languageCodeList
		return nil
	}
}

func WithCustomUsernameRegex(usernameRegex string) Option {
	return func(s *Server) error {
		s.usernameRegex = usernameRegex
		return nil
	}
}

func WithCustomPasswordRegex(passwordRegex string) Option {
	return func(s *Server) error {
		s.passwordRegex = passwordRegex
		return nil
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

func WithDefaultScheduler(jobList ...cron.Job) Option {
	return func(s *Server) error {
		schedulerService := service.NewSchedulerService(s.db)
		defaultJobList := cron.DefaultSchedulerJobFromEnv(schedulerService)

		return WithScheduler(append(defaultJobList, jobList...)...)(s)
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
		middlewareProvider := transport.NewMiddlewareProvider(middlewareService, s.responseHandler)
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
		validationManager, err := validation.NewValidationManager(s.languageCodeList, s.usernameRegex, s.passwordRegex)
		if err != nil {
			return fmt.Errorf("failed to create validation manager: %w", err)
		}

		endpointService := service.NewEndpointService(s.db, s.idGenerator, s.hashingManager, validationManager)
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
		s.httpServer = &http.Server{
			Addr:    ":" + s.port,
			Handler: s.mux,
		}
		return nil
	}
}
