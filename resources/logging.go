package resources

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// UpdateLogLevel changes the logrus log level dynamically.
func UpdateLogLevel() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		level := struct {
			Level string `json:"level"`
		}{}

		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&level)

		if err != nil {
			writeMessage("Please specify one of [debug, info]", 400, w)
			return
		}

		switch level.Level {
		case "debug":
			log.SetLevel(log.DebugLevel)
			log.Debug("Log level updated to debug.")
			break
		case "info":
			log.SetLevel(log.InfoLevel)
			log.Info("Log level updated to info.")
			break
		default:
			writeMessage("Please specify one of [debug, info]", 400, w)
			return
		}

		writeMessage("Log level changed to "+level.Level, 200, w)
	}
}
