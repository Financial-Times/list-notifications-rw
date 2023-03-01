package resources

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

var generatedCarouselTidRegex = regexp.MustCompile(`^(tid_\S+)_carousel_\d{10}_gentx`)
var carouselTidRegex = regexp.MustCompile(`^(.+)_carousel_\d{10}`)

type notificationFinder interface {
	FindNotificationByTransactionID(transactionID string) (model.InternalNotification, error)
	FindNotificationByPartialTransactionID(transactionID string) (model.InternalNotification, error)
}

// FilterCarouselPublishes checks whether this is a carousel publish and processes it accordingly
func (f Filters) FilterCarouselPublishes(finder notificationFinder) Filters {
	next := f.next
	f.next = filterCarouselPublishes(finder, next, f.log)
	return f
}

func filterCarouselPublishes(finder notificationFinder, next func(w http.ResponseWriter, r *http.Request), log *logger.UPPLogger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(tidHeader)
		uuid := mux.Vars(r)["uuid"]

		logEntry := log.WithField("uuid", uuid).WithField("transaction_id", tid)

		if generatedCarouselTidRegex.MatchString(tid) {
			logEntry.Info("Skipping generated carousel publish.")
			if err := writeMessage("Skipping generated carousel publish.", http.StatusOK, w); err != nil {
				logEntry.WithError(err).Error("Failed to write message")
			}
			return
		}

		if !shouldWriteNotification(tid, finder, logEntry) {
			if err := writeMessage("Skipping carousel publish; the original notification was published successfully.", http.StatusOK, w); err != nil {
				logEntry.WithError(err).Error("Failed to write message")
			}
			return
		}

		next(w, r)
	}
}

func shouldWriteNotification(tid string, finder notificationFinder, log *logger.LogEntry) bool {
	if !carouselTidRegex.MatchString(tid) {
		return true
	}

	log.Infof("Received carousel notification.")
	originalTid := carouselTidRegex.FindStringSubmatch(tid)[1]

	notification, err := finder.FindNotificationByTransactionID(originalTid)
	if err == nil {
		log.WithField("lastModified", notification.LastModified).Info("Skipping carousel publish; the original notification was published successfully.")
		return false
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		log.WithError(err).Error("Failed to find original notification for this carousel publish! Writing new notification.")
		return true
	}

	log.Info("Failed to find notification for original transaction ID, checking for a related carousel transaction.")
	notification, err = finder.FindNotificationByPartialTransactionID(originalTid + "_carousel")
	if err == nil {
		log.WithField("lastModified", notification.LastModified).Info("Skipping carousel publish; the original notification was published successfully.")
		return false
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		log.WithError(err).Error("Failed to find original notification for this carousel publish! Writing new notification.")
	}
	return true

}
