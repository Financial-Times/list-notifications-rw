package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/Financial-Times/api-endpoint"
	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/Financial-Times/list-notifications-rw/resources"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
)

const (
	appDescription = "Provides Notifications for UPP Lists"
)

func main() {
	app := cli.App("list-notifications-rw", appDescription)

	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "upp-list-notifications-rw",
		Desc:   "System Code of the application",
		EnvVar: "APP_SYSTEM_CODE",
	})

	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  "list-notifications-rw",
		Desc:   "Application name",
		EnvVar: "APP_NAME",
	})

	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})

	dumpRequests := app.Bool(cli.BoolOpt{
		Name:   "dump-requests",
		Desc:   "Logs every write request in full HTTP/1.1 spec.",
		Value:  false,
		EnvVar: "DUMP_REQUESTS",
	})

	apiHost := app.String(cli.StringOpt{
		Name:   "api-host",
		Desc:   "Api host to use for read responses.",
		Value:  "api.ft.com",
		EnvVar: "API_HOST",
	})

	mongoConnectionTimeout := app.Int(cli.IntOpt{
		Name:   "db-connect-timeout",
		Desc:   "Timeout in milliseconds for the initial database connection.",
		Value:  3000,
		EnvVar: "DB_CONNECTION_TIMEOUT",
	})

	maxSinceInterval := app.Int(cli.IntOpt{
		Name:   "max-since-interval",
		Desc:   "The maximum time interval clients are allowed to query for notifications in days.",
		Value:  90,
		EnvVar: "MAX_SINCE_INTERVAL",
	})

	cacheMaxAge := app.Int(cli.IntOpt{
		Name:   "cache-max-age",
		Desc:   "The max age for content records in varnish in seconds.",
		Value:  10,
		EnvVar: "CACHE_TTL",
	})

	limit := app.Int(cli.IntOpt{
		Name:   "limit",
		Desc:   "The max number of results for a notifications query.",
		Value:  200,
		EnvVar: "NOTIFICATIONS_LIMIT",
	})

	mongoAddress := app.String(cli.StringOpt{
		Name:   "db",
		Desc:   "MongoDB database connection string (i.e. comma separated list of ip:port)",
		Value:  "localhost:27017",
		EnvVar: "MONGO_ADDRESSES",
	})

	apiYml := app.String(cli.StringOpt{
		Name:   "api-yml",
		Value:  "./api.yml",
		Desc:   "Location of the API Swagger YML file.",
		EnvVar: "API_YML",
	})

	logLevel := app.String(cli.StringOpt{
		Name:   "logLevel",
		Value:  "INFO",
		Desc:   "Logging level (DEBUG, INFO, WARN, ERROR)",
		EnvVar: "LOG_LEVEL",
	})

	mongoDatabase := app.String(cli.StringOpt{
		Name:   "mongoDatabase",
		Value:  "upp-store",
		Desc:   "Mongo database to read from",
		EnvVar: "MONGO_DATABASE",
	})

	mongoCollection := app.String(cli.StringOpt{
		Name:   "mongoCollection",
		Value:  "content",
		Desc:   "Mongo collection to read from",
		EnvVar: "MONGO_COLLECTION",
	})

	log := logger.NewUPPLogger(*appName, *logLevel)

	app.Action = func() {
		log.Infof("System code: %s, App Name: %s, Port: %s", *appSystemCode, *appName, *port)

		timeout := time.Duration(*mongoConnectionTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		log.Info("Initialising MongoDB.")
		client, err := db.NewClient(ctx, *mongoAddress, *mongoDatabase, *mongoCollection, *cacheMaxAge, *limit, log)
		if err != nil {
			log.WithError(err).Error("Failed to create database client")
			return
		}

		log.Info("Ensuring Mongo indices are setup...")
		err = client.EnsureIndexes()
		if err != nil {
			log.WithError(err).Warn("Failed to ensure database indices!")
		}
		log.Info("Finished ensuring indices.")

		defer func(client db.Database) {
			if err := client.Close(); err != nil {
				log.WithError(err).Error("Failed to close connection to DB")
			}
		}(client)

		mapper := mapping.DefaultMapper{ApiHost: *apiHost}

		nextLink := mapping.OffsetNextLink{
			ApiHost:    *apiHost,
			CacheDelay: *cacheMaxAge,
			MaxLimit:   *limit,
		}

		healthService := resources.NewHealthService(client, *appSystemCode, *appName, appDescription)

		startService(apiYml, *port, *maxSinceInterval, *dumpRequests, healthService, mapper, nextLink, client, log)
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Error("Failed to run app")
		return
	}
}

func startService(
	apiYml *string,
	port string,
	maxSinceInterval int,
	dumpRequests bool,
	healthService *resources.HealthService,
	mapper mapping.NotificationsMapper,
	nextLink mapping.NextLinkGenerator,
	db db.Database,
	log *logger.UPPLogger,
) {
	r := mux.NewRouter()

	var monitoringRouter http.Handler = r

	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log, monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	if apiYml != nil {
		apiEndpoint, err := api.NewAPIEndpointForFile(*apiYml)
		if err != nil {
			log.WithError(err).WithField("file", *apiYml).Warn("Failed to serve the API Endpoint for this service. Please validate the OpenAPI YML and the file location")
		} else {
			r.Handle(api.DefaultPath, apiEndpoint)
		}
	}

	r.HandleFunc("/lists/notifications", resources.ReadNotifications(mapper, nextLink, db, maxSinceInterval, log))

	write := resources.Filter(resources.WriteNotification(dumpRequests, mapper, db, log), log).FilterSyntheticTransactions().FilterCarouselPublishes(db).Gunzip().Build()
	r.HandleFunc("/lists/{uuid}", write).Methods("PUT")

	r.HandleFunc("/__health", healthService.HealthChecksHandler())

	r.HandleFunc("/__log", resources.UpdateLogLevel(log)).Methods("POST")

	r.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(healthService.GTG))

	r.HandleFunc(status.PingPath, status.PingHandler)
	r.HandleFunc(status.PingPathDW, status.PingHandler)

	r.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
	r.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)

	addr := ":" + port
	server := &http.Server{
		Handler: monitoringRouter,
		Addr:    addr,

		WriteTimeout: 60 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info("Starting server on " + addr)
	if err := server.ListenAndServe(); err != nil {
		log.WithError(err).Error("Failed to start server")
		return
	}
}
