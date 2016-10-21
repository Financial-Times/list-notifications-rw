package main

import (
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "list-notifications-rw"
	app.Usage = "R/W for List Notifications"
	app.Version = version()

	app.Run(os.Args)
}

func version() string {
	v := os.Getenv("app_version")
	if v == "" {
		v = "v0.0.0"
	}
	return v
}