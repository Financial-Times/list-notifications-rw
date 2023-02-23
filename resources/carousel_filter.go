package resources

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

var generatedCarouselTidRegex = regexp.MustCompile(`^(tid_\S+)_carousel_\d{10}_gentx`)
var carouselTidRegex = regexp.MustCompile(`^(.+)_carousel_\d{10}`)

// FilterCarouselPublishes checks whether this is a carousel publish and processes it accordingly
func (f Filters) FilterCarouselPublishes(db db.Database) Filters {
	next := f.next
	f.next = filterCarouselPublishes(db, next, f.log)
	return f
}

func filterCarouselPublishes(db db.Database, next func(w http.ResponseWriter, r *http.Request), log *logger.UPPLogger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(tidHeader)
		uuid := mux.Vars(r)["uuid"]

		if generatedCarouselTidRegex.MatchString(tid) {
			log.WithField("uuid", uuid).WithField("transaction_id", tid).Info("Skipping generated carousel publish.")
			if err := writeMessage("Skipping generated carousel publish.", http.StatusOK, w); err != nil {
				log.WithError(err).Error("Failed to write message")
			}
			return
		}

		if !shouldWriteNotification(uuid, tid, db, log) {
			if err := writeMessage("Skipping carousel publish; the original notification was published successfully.", http.StatusOK, w); err != nil {
				log.WithError(err).Error("Failed to write message")
			}
			return
		}

		next(w, r)
	}
}

func shouldWriteNotification(uuid string, tid string, db db.Database, log *logger.UPPLogger) bool {
	if !carouselTidRegex.MatchString(tid) {
		return true
	}

	log.WithField("uuid", uuid).WithField("transaction_id", tid).Infof("Received carousel notification.")
	originalTid := carouselTidRegex.FindStringSubmatch(tid)[1]

	notification, err := db.FindNotificationByTransactionID(originalTid)
	if err == nil {
		log.WithField("uuid", uuid).WithField("transaction_id", tid).WithField("lastModified", notification.LastModified).Info("Skipping carousel publish; the original notification was published successfully.")
		return false
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		log.WithField("uuid", uuid).WithField("transaction_id", tid).WithError(err).Error("Failed to find original notification for this carousel publish! Writing new notification.")
		return true
	}

	log.WithField("uuid", uuid).WithField("transaction_id", tid).Info("Failed to find notification for original transaction ID, checking for a related carousel transaction.")
	notification, err = db.FindNotificationByPartialTransactionID(originalTid + "_carousel")
	if err == nil {
		log.WithField("uuid", uuid).WithField("transaction_id", tid).WithField("lastModified", notification.LastModified).Info("Skipping carousel publish; the original notification was published successfully.")
		return false
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		log.WithField("uuid", uuid).WithField("transaction_id", tid).WithError(err).Error("Failed to find original notification for this carousel publish! Writing new notification.")
	}
	return true

}
