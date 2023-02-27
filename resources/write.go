package resources

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/gorilla/mux"
)

type notificationWriter interface {
	WriteNotification(notification *model.InternalNotification) error
}

// WriteNotification will write a new notification for the provided list.
func WriteNotification(dumpRequests bool, mapper mapping.NotificationsMapper, writer notificationWriter, log *logger.UPPLogger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if dumpRequests {
			dumpRequest(r, log)
		}

		decoder := json.NewDecoder(r.Body)
		uuid := mux.Vars(r)["uuid"]
		logEntry := log.WithFields(map[string]any{
			"uuid":           uuid,
			"transaction_id": r.Header.Get("X-Request-Id"),
		})

		notification, err := mapper.MapRequestToInternalNotification(uuid, decoder)
		if err != nil {
			logEntry.WithError(err).Error("Invalid request! See error for details.")
			if err = writeMessage("Invalid Request body.", 400, w); err != nil {
				logEntry.WithError(err).Error("Failed to write message for unsuccessful mapping of notification")
			}
			return
		}

		if err = writer.WriteNotification(notification); err != nil {
			logEntry.WithError(err).Error("Failed to write notification")
			if err = writeMessage("Failed to write notification.", 500, w); err != nil {
				logEntry.WithError(err).Error("Failed to write message for unsuccessful notification write")
			}
			return
		}

		logEntry.Info("Successfully processed a notification for this list.")
		w.WriteHeader(200)
	}
}

func dumpRequest(r *http.Request, log *logger.UPPLogger) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.WithError(err).Warn("Failed to dump request!")
		return
	}
	log.Info(string(dump))
}

type msg struct {
	Message string `json:"message"`
}

func writeMessage(message string, status int, w http.ResponseWriter) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	m := msg{Message: message}
	encoder := json.NewEncoder(w)
	return encoder.Encode(m)
}
