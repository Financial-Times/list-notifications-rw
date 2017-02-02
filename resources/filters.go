package resources

import (
	"compress/gzip"
	"net/http"
	"regexp"
	"strings"

	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Sirupsen/logrus"
)

const synthTidPrefix = "SYNTHETIC-REQ-MON"

var generatedCarouselTidRegex = regexp.MustCompile(`^(tid_[\S]+)_carousel_[\d]{10}_gentx`)
var carouselTidRegex = regexp.MustCompile(`^(tid_[\S]+)_carousel_[\d]{10}`)

const tidHeader = "X-Request-Id"

// Filters contains the composed chain
type Filters struct {
	next func(w http.ResponseWriter, r *http.Request)
}

// Filter creates a new composable filter.
func Filter(next func(w http.ResponseWriter, r *http.Request)) Filters {
	return Filters{next}
}

// FilterSyntheticTransactions will filter out incoming requests if they have a synthetic prefix.
func (f Filters) FilterSyntheticTransactions() Filters {
	next := f.next
	f.next = filterSyntheticTransactions(next)
	return f
}

// Gunzip pre-processes the request body if it's gzipped.
func (f Filters) Gunzip() Filters {
	next := f.next
	f.next = gunzip(next)
	return f
}

// FilterCarouselPublishes checks whether this is a carousel publish and processes it accordingly
func (f Filters) FilterCarouselPublishes(db db.DB) Filters {
	next := f.next
	f.next = filterCarouselPublishes(db, next)
	return f
}

// Build returns the final chained handler
func (f Filters) Build() func(w http.ResponseWriter, r *http.Request) {
	return f.next
}

func filterSyntheticTransactions(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(tidHeader)
		if tid == "" {
			logrus.WithField("transaction_id", tid).Infof("Rejecting notification; it has no transaction id.")
			writeMessage("Rejecting notification; it has no transaction id.", 400, w)
			return
		}

		if strings.HasPrefix(strings.ToUpper(tid), synthTidPrefix) {
			logrus.WithField("transaction_id", tid).Infof("Rejecting notification; it has a synthetic transaction id.")
			writeMessage("Rejecting notification; it has a synthetic transaction id.", 200, w)
			return
		}

		next(w, r)
	}
}

func filterCarouselPublishes(db db.DB, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(tidHeader)

		if generatedCarouselTidRegex.MatchString(tid) {
			logrus.WithField("transaction_id", tid).Info("Skipping generated carousel publish.")
			writeMessage("Skipping generated carousel publish.", 200, w)
			return
		}

		if !carouselTidRegex.MatchString(tid) {
			next(w, r)
			return
		}

		logrus.WithField("transaction_id", tid).Infof("Received carousel notification.")
		originalTid := carouselTidRegex.FindStringSubmatch(tid)[1]

		tx, err := db.Open()
		if err != nil {
			logrus.WithError(err).Error("Failed to connect to mongo")
			writeMessage("An internal server error prevented processing of your request.", 500, w)
			return
		}

		defer tx.Close()

		notifications, err := tx.FindNotification(originalTid)
		if err != nil {
			logrus.WithField("transaction_id", tid).WithError(err).Error("Failed to find original notification for this carousel publish! Writing new notification.")
			next(w, r)
			return
		}

		if notifications == nil || len(*notifications) == 0 {
			next(w, r)
			return
		}

		logrus.WithField("transaction_id", tid).WithField("lastModified", (*notifications)[0].LastModified).Info("Skipping carousel publish; the original notification was published successfully.")
		writeMessage("Skipping carousel publish; the original notification was published successfully.", 200, w)
	}
}

func gunzip(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			unzipped, err := gzip.NewReader(r.Body)
			if err != nil {
				writeMessage(err.Error(), http.StatusBadRequest, w)
				return
			}

			r.Body = unzipped
		}
		next(w, r)
	}
}
