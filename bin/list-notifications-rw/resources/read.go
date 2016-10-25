package resources

import (
	"net/http"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Sirupsen/logrus"
	"encoding/json"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/Financial-Times/list-notifications-rw/model"
	"time"
)

type msg struct {
	Message string `json:"message"`
}

func ReadNotifications(mapper mapping.NotificationsMapper, db db.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request){
		param := r.URL.Query().Get("since")
		if param == "" {
			logrus.Warn("User didn't provide since date.")
			writeSinceError(w)
			return
		}

		since, err := time.Parse(time.RFC3339Nano, param)
		if err != nil {
			logrus.WithError(err).WithField("since", param).Warn("Failed to parse user provided since date.")
			writeSinceError(w)
			return
		}

		tx, err := db.Open()
		if err != nil {
			logrus.WithError(err).Error("Failed to connect to mongo")
			w.WriteHeader(500)
			return
		}

		notifications, err := tx.ReadNotifications(since)
		if err != nil {
			logrus.WithError(err).Error("Failed to query mongo for notifications!")
			w.WriteHeader(500)
			return
		}

		results := make([]model.PublicNotification, 0)
		for _, n := range *notifications {
			results = append(results, mapper.MapInternalNotificationToPublic(n))
		}

		w.Header().Add("Content-Type", "application/json")

		encoder := json.NewEncoder(w)
		encoder.Encode(results)
	}
}

func writeSinceError(w http.ResponseWriter){
	w.WriteHeader(400)

	m := msg{"A mandatory 'since' query parameter has not been specified. Please supply a since date. For eg., since="+time.Now().UTC().Format(time.RFC3339Nano)+"."}
	encoder := json.NewEncoder(w)
	encoder.Encode(m)

	w.Header().Add("Content-Type", "application/json")
}