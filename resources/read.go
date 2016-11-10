package resources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/Sirupsen/logrus"
)

type msg struct {
	Message string `json:"message"`
}

// ReadNotifications reads notifications from the backing db
func ReadNotifications(mapper mapping.NotificationsMapper, nextLink mapping.NextLinkGenerator, db db.DB, maxSinceInterval int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		param := r.URL.Query().Get("since")
		if param == "" {
			logrus.Info("User didn't provide since date.")
			writeMessage(400, sinceMessage(), w)
			return
		}

		since, err := time.Parse(time.RFC3339Nano, param)
		if err != nil {
			logrus.WithError(err).WithField("since", param).Info("Failed to parse user provided since date.")
			writeMessage(400, sinceMessage(), w)
			return
		}
		if since.Before(time.Now().UTC().AddDate(0, 0, -maxSinceInterval)) {
			logrus.Infof("User provided since date before query cap date, since= [%v].", since.Format(time.RFC3339Nano))
			writeMessage(400, fmt.Sprintf("Since date must be within the last %d days.", maxSinceInterval), w)
			return
		}

		offset, err := getOffset(r)

		if err != nil {
			logrus.WithError(err).Info("User provided offset is not an integer!")
			writeMessage(400, "Please specify an integer offset.", w)
			return
		}

		tx, err := db.Open()
		if err != nil {
			logrus.WithError(err).Error("Failed to connect to mongo")
			writeMessage(500, "Failed to retrieve list notifications due to internal server error", w)
			return
		}

		defer tx.Close()

		notifications, err := tx.ReadNotifications(offset, since)
		if err != nil {
			logrus.WithError(err).Error("Failed to query mongo for notifications!")
			writeMessage(500, "Failed to retrieve list notifications due to internal server error", w)
			return
		}

		results := []model.PublicNotification{}
		for i, n := range *notifications {
			if i >= db.Limit() {
				break
			}
			results = append(results, mapper.MapInternalNotificationToPublic(n))
		}

		page := model.PublicNotificationPage{
			Links: []model.Link{
				nextLink.NextLink(since, offset, *notifications),
			},
			Notifications: results,
			RequestURL:    nextLink.ProcessRequestLink(r.URL).String(),
		}

		w.Header().Add("Content-Type", "application/json")

		encoder := json.NewEncoder(w)
		encoder.Encode(page)
	}
}

func getOffset(r *http.Request) (offset int, err error) {
	offset = 0

	offsetParam := r.URL.Query().Get("offset")
	if offsetParam != "" {
		offset, err = strconv.Atoi(offsetParam)
	}

	return offset, err
}

func sinceMessage() string {
	return fmt.Sprintf("A mandatory 'since' query parameter has not been specified. Please supply a since date. For eg., since=%s .", time.Now().UTC().AddDate(0, 0, -1).Format(time.RFC3339Nano))
}

func writeMessage(status int, message string, w http.ResponseWriter) {
	w.WriteHeader(status)

	m := msg{message}
	encoder := json.NewEncoder(w)
	encoder.Encode(m)

	w.Header().Add("Content-Type", "application/json")
}
