package resources

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// UpdateLogLevel changes the logrus log level dynamically.
func UpdateLogLevel() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		level := mux.Vars(r)["level"]

		switch level {
		case "debug":
			logrus.SetLevel(logrus.DebugLevel)
			break
		case "info":
			logrus.SetLevel(logrus.InfoLevel)
			break
		case "warn":
			logrus.SetLevel(logrus.WarnLevel)
			break
		default:
			writeError("Please specify one of [debug, info, warn]", 400, w)
			return
		}
		writeMessage(200, "Log level changed to "+level, w)
	}
}
