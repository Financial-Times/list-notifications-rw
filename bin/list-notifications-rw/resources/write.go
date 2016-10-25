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
				Error("Failed to parse json request! See error for details.")
			w.WriteHeader(400)
			return
		}

		tx, err := db.Open()
		if err != nil {
			logrus.WithError(err).Error("Failed to connect to mongo")
			w.WriteHeader(500)
			return
		}

		defer tx.Close()

		tx.WriteNotification(notification)

		w.WriteHeader(200)
	}
}