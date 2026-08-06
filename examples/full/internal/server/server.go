package server

import (
	"github.com/jljl1337/gostarter/pkg/core/queue"
	"github.com/jljl1337/gostarter/pkg/core/server"
	gsTransport "github.com/jljl1337/gostarter/pkg/core/transport"
	"github.com/jljl1337/gostarter/pkg/shared/db"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
	"github.com/jljl1337/gostarter/pkg/shared/log"

	"github.com/jljl1337/gostarter/examples/full/internal/cron"
	"github.com/jljl1337/gostarter/examples/full/internal/env"
	"github.com/jljl1337/gostarter/examples/full/internal/service"
	"github.com/jljl1337/gostarter/examples/full/internal/sql"
	"github.com/jljl1337/gostarter/examples/full/internal/transport"
	"github.com/jljl1337/gostarter/examples/full/web"
)

func MustNewServer(envFile string) *server.Server {
	env.MustSetConstants(envFile)

	err := log.SetCustomLoggerFromEnv()
	if err != nil {
		panic(err)
	}

	db, err := db.NewDBFromEnv()
	if err != nil {
		panic(err)
	}

	schedulerService := service.NewSchedulerService(db)
	job := cron.DeleteNotesJob(schedulerService)

	queueService := service.NewQueueService(db)
	queueManager := queue.NewDefaultQueueManager(db, queueService.GetQueueLanes()...)

	responseHandler := gsTransport.NewDefaultResponseHandler()
	service := service.NewEndpointService(db, generator.NewULID, queueManager)
	handler := transport.NewEndpointHandler(service, responseHandler)

	s, err := server.NewServer(
		server.WithDB(db),
		server.WithGostarterMigration(),
		server.WithAppMigrations(sql.MigrationDir),
		server.WithCustomLanguageCodeList("en-US", "fr-FR"),
		server.WithDefaultScheduler(job),
		server.WithQueueManager(queueManager),
		server.WithStaticSite("/", web.SiteDir, "site"),
		server.WithDefaultMiddleware(),
		server.WithDefaultApiHandler(handler),
		server.WithHttpServer(),
	)
	if err != nil {
		panic(err)
	}

	return s
}
