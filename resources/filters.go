package resources

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
)

const synthTidPrefix = "SYNTHETIC-REQ-MON"
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

// Build returns the final chained handler
func (f Filters) Build() func(w http.ResponseWriter, r *http.Request) {
	return f.next
}

func filterSyntheticTransactions(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(tidHeader)
		if tid == "" {
			logrus.WithField("tid", tid).Infof("Rejecting notification; it has no transaction id.")
			writeError("Rejecting notification; it has no transaction id.", 400, w)
			return
		}

		if strings.HasPrefix(strings.ToUpper(tid), synthTidPrefix) {
			logrus.WithField("tid", tid).Infof("Rejecting notification; it has a synthetic transaction id.")
			writeError("Rejecting notification; it has a synthetic transaction id.", 200, w)
			return
		}

		next(w, r)
	}
}

func gunzip(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			unzipped, err := gzip.NewReader(r.Body)
			if err != nil {
				writeError(err.Error(), http.StatusBadRequest, w)
				return
			}

			r.Body = unzipped
		}
		next(w, r)
	}
}
