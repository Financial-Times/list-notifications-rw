package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/Financial-Times/list-notifications-rw/resources"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"
	"gopkg.in/urfave/cli.v1/altsrc"
)

func main() {
	app := cli.NewApp()
	app.Name = "list-notifications-rw"
	app.Usage = "R/W for List Notifications"

	flags := []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "db",
			Usage:  "MongoDB database connection string (i.e. comma separated list of ip:port)",
			EnvVar: "MONGO_ADDRESSES",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:   "limit",
			Usage:  "The max number of results for a notifications query.",
			Value:  200,
			EnvVar: "NOTIFICATIONS_LIMIT",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:   "port",
			Usage:  "The port number to run on.",
			Value:  8080,
			EnvVar: "PORT",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:   "cache-max-age",
			Usage:  "The max age for content records in varnish in seconds.",
			Value:  10,
			EnvVar: "CACHE_TTL",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:   "max-since-interval",
			Usage:  "The maximum time interval clients are allowed to query for notifications in days.",
			Value:  90,
			EnvVar: "MAX_SINCE_INTERVAL",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "api-host",
			Usage:  "Api host to use for read responses.",
			Value:  "api.ft.com",
			EnvVar: "API_HOST",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:   "db-connect-timeout",
			Usage:  "Timeout in milliseconds for the initial database connection.",
			Value:  3800,
			EnvVar: "DB_CONNECTION_TIMEOUT",
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:   "dump-requests",
			Usage:  "Logs every write request in full HTTP/1.1 spec.",
			EnvVar: "DUMP_REQUESTS",
		}),
		cli.StringFlag{
			Name:  "config",
			Value: "./config.yml",
			Usage: "Path to the YAML config file.",
		},
	}

	app.Version = version()
	app.Before = altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("config"))
	app.Flags = flags

	app.Action = func(ctx *cli.Context) {
		mongo := &db.MongoDB{
			Urls:       ctx.String("db"),
			Timeout:    ctx.Int("db-connect-timeout"),
			MaxLimit:   ctx.Int("limit"),
			CacheDelay: ctx.Int("cache-max-age"),
		}

		defer mongo.Close()

		mapper := mapping.DefaultMapper{ApiHost: ctx.String("api-host")}

		nextLink := mapping.OffsetNextLink{
			ApiHost:    ctx.String("api-host"),
			CacheDelay: ctx.Int("cache-max-age"),
			MaxLimit:   ctx.Int("limit"),
		}

		server(ctx.Int("port"), ctx.Int("max-since-interval"), ctx.Bool("dump-requests"), mapper, nextLink, mongo)
	}

	app.Run(os.Args)
}

func server(port int, maxSinceInterval int, dumpRequests bool, mapper mapping.NotificationsMapper, nextLink mapping.NextLinkGenerator, db db.DB) {
	r := mux.NewRouter()

	r.HandleFunc("/lists/notifications", resources.ReadNotifications(mapper, nextLink, db, maxSinceInterval))

	gunzipped := resources.UnzipGzip(resources.WriteNotification(dumpRequests, mapper, db))
	r.HandleFunc("/lists/{uuid}", resources.FilterSyntheticTransactions(gunzipped)).Methods("PUT")

	r.HandleFunc("/__health", resources.Health(db))
	r.HandleFunc(status.GTGPath, resources.GTG(db))

	r.HandleFunc(status.PingPath, status.PingHandler)
	r.HandleFunc(status.PingPathDW, status.PingHandler)

	r.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
	r.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)

	addr := ":" + strconv.Itoa(port)
	server := &http.Server{
		Handler: r,
		Addr:    addr,

		WriteTimeout: 60 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logrus.Info("Starting server on " + addr)
	server.ListenAndServe()
}

func version() string {
	v := os.Getenv("app_version") // set in service file
	if v == "" {
		v = "v0.0.0"
	}
	return v
}
