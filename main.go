package main

import (
	"net/http"
	"os"
	"time"

	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/Financial-Times/list-notifications-rw/resources"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	metrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
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

	mongoConnection := app.String(cli.StringOpt{
		Name:   "db",
		Desc:   "MongoDB database connection string (i.e. comma separated list of ip:port)",
		Value:  "localhost:27017",
		EnvVar: "MONGO_ADDRESSES",
	})

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	log.Infof("[Startup] %v is starting", *appSystemCode)

	app.Action = func() {
		log.Infof("System code: %s, App Name: %s, Port: %s", *appSystemCode, *appName, *port)

		log.Info("Initialising MongoDB.")
		mongo := &db.MongoDB{
			Urls:       *mongoConnection,
			Timeout:    *mongoConnectionTimeout,
			MaxLimit:   *limit,
			CacheDelay: *cacheMaxAge,
		}

		defer mongo.Close()

		log.Info("Opening initial connection to Mongo.")
		tx, err := mongo.Open()
		if err != nil {
			log.WithError(err).Error("Failed to connect to Mongo!")
			return
		}

		log.Info("Ensuring Mongo indices are setup...")
		err = tx.EnsureIndices()
		if err != nil {
			log.WithError(err).Error("Failed to ensure mongo indices!")
			return
		}
		log.Info("Finished ensuring indices.")

		tx.Close()

		mapper := mapping.DefaultMapper{ApiHost: *apiHost}

		nextLink := mapping.OffsetNextLink{
			ApiHost:    *apiHost,
			CacheDelay: *cacheMaxAge,
			MaxLimit:   *limit,
		}

		healthService := resources.NewHealthService(mongo, *appSystemCode, *appName, appDescription)

		server(*port, *maxSinceInterval, *dumpRequests, healthService, mapper, nextLink, mongo)
	}

	app.Run(os.Args)
}

func server(port string, maxSinceInterval int, dumpRequests bool, healthService *resources.HealthService, mapper mapping.NotificationsMapper, nextLink mapping.NextLinkGenerator, db db.DB) {
	r := mux.NewRouter()

	var monitoringRouter http.Handler = r

	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	r.HandleFunc("/lists/notifications", resources.ReadNotifications(mapper, nextLink, db, maxSinceInterval))

	write := resources.Filter(resources.WriteNotification(dumpRequests, mapper, db)).FilterSyntheticTransactions().FilterCarouselPublishes(db).Gunzip().Build()
	r.HandleFunc("/lists/{uuid}", write).Methods("PUT")

	r.HandleFunc("/__health", healthService.HealthChecks())

	r.HandleFunc("/__log", resources.UpdateLogLevel()).Methods("POST")

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
	server.ListenAndServe()
}
