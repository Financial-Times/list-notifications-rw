package resources

import (
	"net/http"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Sirupsen/logrus"
	"encoding/json"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/gorilla/mux"
)


func WriteNotification(mapper mapping.NotificationsMapper, db db.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request){
		decoder := json.NewDecoder(r.Body)
		uuid := mux.Vars(r)["uuid"]

		notification, err := mapper.MapRequestToInternalNotification(uuid, decoder)
		if err != nil {
			logrus.WithError(err).
				WithField("uuid", uuid).
				WithField("tid", r.Header.Get("X-Request-Id")).
				Error("Invalid request! See error for details.")

			writeError("Invalid Request body.", 400, w)
			return
		}

		tx, err := db.Open()
		if err != nil {
			logrus.WithError(err).Error("Failed to connect to mongo")
			writeError("An internal server error prevented processing of your request.", 500, w)
			return
		}

		defer tx.Close()

		tx.WriteNotification(notification)

		w.WriteHeader(200)
	}
}

func writeError(message string, status int, w http.ResponseWriter){
	w.WriteHeader(status)

	m := msg{Message: message}
	encoder := json.NewEncoder(w)
	encoder.Encode(m)

	w.Header().Add("Content-Type", "application/json")
}