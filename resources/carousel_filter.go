package resources

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

var generatedCarouselTidRegex = regexp.MustCompile(`^(tid_\S+)_carousel_\d{10}_gentx`)
var carouselTidRegex = regexp.MustCompile(`^(.+)_carousel_\d{10}`)

// FilterCarouselPublishes checks whether this is a carousel publish and processes it accordingly
func (f Filters) FilterCarouselPublishes(db db.Database) Filters {
	next := f.next
	f.next = filterCarouselPublishes(db, next)
	return f
}

func filterCarouselPublishes(db db.Database, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(tidHeader)
		uuid := mux.Vars(r)["uuid"]

		if generatedCarouselTidRegex.MatchString(tid) {
			log.WithField("uuid", uuid).WithField("transaction_id", tid).Info("Skipping generated carousel publish.")
			writeMessage("Skipping generated carousel publish.", http.StatusOK, w)
			return
		}

		writeNotification, err := shouldWriteNotification(uuid, tid, db)
		if err != nil {
			writeMessage("An internal server error prevented processing of your request.", http.StatusInternalServerError, w)
			return
		}

		if !writeNotification {
			writeMessage("Skipping carousel publish; the original notification was published successfully.", http.StatusOK, w)
			return
		}

		next(w, r)
	}
}

func shouldWriteNotification(uuid string, tid string, db db.Database) (bool, error) {
	if !carouselTidRegex.MatchString(tid) {
		return true, nil
	}

	log.WithField("uuid", uuid).WithField("transaction_id", tid).Infof("Received carousel notification.")
	originalTid := carouselTidRegex.FindStringSubmatch(tid)[1]

	notification, err := db.FindNotificationByTransactionID(originalTid)
	if err == nil {
		log.WithField("uuid", uuid).WithField("transaction_id", tid).WithField("lastModified", notification.LastModified).Info("Skipping carousel publish; the original notification was published successfully.")
		return false, nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {

		log.WithField("uuid", uuid).WithField("transaction_id", tid).Info("Failed to find notification for original transaction ID, checking for a related carousel transaction.")
		notification, err = db.FindNotificationByPartialTransactionID(originalTid + "_carousel")
		if err == nil {
			log.WithField("uuid", uuid).WithField("transaction_id", tid).WithField("lastModified", notification.LastModified).Info("Skipping carousel publish; the original notification was published successfully.")
			return false, nil
		}

		if errors.Is(err, mongo.ErrNoDocuments) {
			return true, nil
		}

		log.WithField("uuid", uuid).WithField("transaction_id", tid).WithError(err).Error("Failed to find original notification for this carousel publish! Writing new notification.")
		return true, nil
	}

	log.WithField("uuid", uuid).WithField("transaction_id", tid).WithError(err).Error("Failed to find original notification for this carousel publish! Writing new notification.")
	return true, nil
}
