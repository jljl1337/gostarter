package gostarter

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/jljl1337/gostarter/pkg/core/cron"
	"github.com/jljl1337/gostarter/pkg/core/http/common"
	"github.com/jljl1337/gostarter/pkg/core/http/handler"
	"github.com/jljl1337/gostarter/pkg/core/http/handler/endpoint"
	"github.com/jljl1337/gostarter/pkg/core/http/handler/web"
	"github.com/jljl1337/gostarter/pkg/core/http/middleware"
	"github.com/jljl1337/gostarter/pkg/core/service"
	serviceCron "github.com/jljl1337/gostarter/pkg/core/service/cron"
	serviceEndpoint "github.com/jljl1337/gostarter/pkg/core/service/endpoint"
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

func WithCustomHashingManager(hashingManager *crypto.HashingManager) Option {
	return func(s *Server) error {
		s.hashingManager = hashingManager
		return nil
	}
}

func WithCustomValidationManager(validationManager *validation.ValidationManager) Option {
	return func(s *Server) error {
		s.validationManager = validationManager
		return nil
	}
}

func WithCustomResponseHandler(responseHandler *common.ResponseHandler) Option {
	return func(s *Server) error {
		s.responseHandler = responseHandler
		return nil
	}
}

func WithCustomCookieGenerator(cookieGenerator *common.CookieGenerator) Option {
	return func(s *Server) error {
		s.cookieGenerator = cookieGenerator
		return nil
	}
}

func WithDefaultScheduler(jobList ...cron.Job) Option {
	return func(s *Server) error {
		schedulerService := serviceCron.NewSchedulerService(s.db)
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

func WithStaticSite(path string, siteFs fs.FS) Option {
	return func(s *Server) error {
		webHandler := web.NewWebHandler(path, siteFs)
		webHandler.RegisterRoutes(s.mux)

		return nil
	}
}

func WithDefaultMiddleware() Option {
	return func(s *Server) error {
		middlewareService := service.NewMiddlewareService(s.db)
		middlewareProvider := middleware.NewMiddlewareProvider(middlewareService, s.responseHandler)
		return WithMiddleware(middlewareProvider.GetMiddlewareList()...)(s)
	}
}

func WithMiddleware(middlewareList ...middleware.Middleware) Option {
	return func(s *Server) error {
		s.apiMiddleware = middleware.CreateStack(middlewareList...)
		return nil
	}
}

func WithDefaultApiHandler(handlerList ...handler.Handler) Option {
	return func(s *Server) error {
		endpointService := serviceEndpoint.NewEndpointService(s.db, s.idGenerator, s.hashingManager, s.validationManager)
		endpointHandler := endpoint.NewEndpointHandler(endpointService, s.responseHandler, s.cookieGenerator)
		handlerList = append([]handler.Handler{endpointHandler}, handlerList...)
		return WithApiHandler("/api", handlerList...)(s)
	}
}

func WithApiHandler(subpath string, handlerList ...handler.Handler) Option {
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
