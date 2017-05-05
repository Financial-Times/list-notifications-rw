package resources

import (
	"net/http"
	"regexp"

	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var generatedCarouselTidRegex = regexp.MustCompile(`^(tid_[\S]+)_carousel_[\d]{10}_gentx`)
var carouselTidRegex = regexp.MustCompile(`^(.+)_carousel_[\d]{10}`)

// FilterCarouselPublishes checks whether this is a carousel publish and processes it accordingly
func (f Filters) FilterCarouselPublishes(db db.DB) Filters {
	next := f.next
	f.next = filterCarouselPublishes(db, next)
	return f
}

func filterCarouselPublishes(db db.DB, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(tidHeader)
		uuid := mux.Vars(r)["uuid"]

		if generatedCarouselTidRegex.MatchString(tid) {
			logrus.WithField("uuid", uuid).WithField("transaction_id", tid).Info("Skipping generated carousel publish.")
			writeMessage("Skipping generated carousel publish.", 200, w)
			return
		}

		writeNotification, err := shouldWriteNotification(uuid, tid, db)
		if err != nil {
			writeMessage("An internal server error prevented processing of your request.", 500, w)
			return
		}

		if !writeNotification {
			writeMessage("Skipping carousel publish; the original notification was published successfully.", 200, w)
			return
		}

		next(w, r)
	}
}

func shouldWriteNotification(uuid string, tid string, db db.DB) (bool, error) {
	if !carouselTidRegex.MatchString(tid) {
		return true, nil
	}

	logrus.WithField("uuid", uuid).WithField("transaction_id", tid).Infof("Received carousel notification.")
	originalTid := carouselTidRegex.FindStringSubmatch(tid)[1]

	tx, err := db.Open()
	if err != nil {
		logrus.WithField("uuid", uuid).WithField("transaction_id", tid).WithError(err).Error("Failed to connect to mongo")
		return false, err
	}

	defer tx.Close()

	notifications, found, err := tx.FindNotification(originalTid)
	if err != nil {
		logrus.WithField("uuid", uuid).WithField("transaction_id", tid).WithError(err).Error("Failed to find original notification for this carousel publish! Writing new notification.")
		return true, nil
	}

	if found {
		logrus.WithField("uuid", uuid).WithField("transaction_id", tid).WithField("lastModified", (*notifications)[0].LastModified).Info("Skipping carousel publish; the original notification was published successfully.")
		return false, nil
	}

	logrus.WithField("uuid", uuid).WithField("transaction_id", tid).Info("Failed to find notification for original transaction ID, checking for a related carousel transacation.")
	notifications, found, err = tx.FindNotificationByPartialTransactionID(originalTid + "_carousel")
	if err != nil {
		logrus.WithField("uuid", uuid).WithField("transaction_id", tid).WithError(err).Error("Failed to find original notification for this carousel publish! Writing new notification.")
		return true, nil
	}

	if !found {
		return true, nil
	}

	logrus.WithField("uuid", uuid).WithField("transaction_id", tid).WithField("lastModified", (*notifications)[0].LastModified).Info("Skipping carousel publish; the original notification was published successfully.")
	return false, nil
}
