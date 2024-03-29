package resources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/Financial-Times/list-notifications-rw/model"
)

type notificationReader interface {
	ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error)
	GetLimit() int
}

// ReadNotifications reads notifications from the backing db
func ReadNotifications(mapper mapping.NotificationsMapper, nextLink mapping.NextLinkGenerator, reader notificationReader, maxSinceInterval int, log *logger.UPPLogger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		param := r.URL.Query().Get("since")
		if param == "" {
			log.Info("User didn't provide since date.")
			writeMessage(sinceMessage(), 400, w)
			return
		}

		since, err := time.Parse(time.RFC3339Nano, param)
		if err != nil {
			log.WithError(err).WithField("since", param).Info("Failed to parse user provided since date.")
			writeMessage(sinceMessage(), 400, w)
			return
		}
		if since.Before(time.Now().UTC().AddDate(0, 0, -maxSinceInterval)) {
			log.Infof("User provided since date before query cap date, since= [%v].", since.Format(time.RFC3339Nano))
			writeMessage(fmt.Sprintf("Since date must be within the last %d days.", maxSinceInterval), 400, w)
			return
		}

		offset, err := getOffset(r)

		if err != nil {
			log.WithError(err).Info("User provided offset is not an integer!")
			writeMessage("Please specify an integer offset.", 400, w)
			return
		}

		notifications, err := reader.ReadNotifications(offset, since)
		if err != nil {
			log.WithError(err).Error("Failed to query database for notifications!")
			writeMessage("Failed to retrieve list notifications due to internal server error", 500, w)
			return
		}

		results := make([]model.PublicNotification, 0)
		for i, n := range *notifications {
			if i >= reader.GetLimit() {
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
		if err = encoder.Encode(page); err != nil {
			log.WithError(err).Error("Failed to encode page")
		}
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
