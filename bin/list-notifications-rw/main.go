package main

import (
	"github.com/urfave/cli"
	"os"
	"gopkg.in/urfave/cli.v1/altsrc"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Financial-Times/list-notifications-rw/mapping"
)

func main() {
	app := cli.NewApp()
	app.Name = "list-notifications-rw"
	app.Usage = "R/W for List Notifications"

	flags := []cli.Flag {
		altsrc.NewStringFlag(cli.StringFlag{
			Name: "db",
			Usage: "MongoDB database connection string (i.e. comma separated list of ip:port)",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name: "apiHost",
			Usage: "Api host to use for read responses.",
			Value: "test.api.ft.com",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name: "db-connect-timeout",
			Usage: "Timeout in seconds for the initial database connection.",
			Value: 60,
		}),
		cli.StringFlag{
			Name: "config",
			Value: "./config.yml",
			Usage: "Path to the YAML config file.",
		},
	}

	app.Version = version()
	app.Before = altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("config"))
	app.Flags = flags

	app.Action = func(ctx *cli.Context){
		mongo := db.MongoDB{Urls: ctx.String("db"), Timeout: ctx.Int("db-connect-timeout")}
		mapper := mapping.DefaultMapper{ApiHost: ctx.String("apiHost")}

		server(mapper, mongo)
	}

	app.Run(os.Args)
}

func version() string {
	v := os.Getenv("app_version")
	if v == "" {
		v = "v0.0.0"
	}
	return v
}