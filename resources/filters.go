package resources

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/Financial-Times/go-logger/v2"
)

const synthTidPrefix = "SYNTHETIC-REQ-MON"

const tidHeader = "X-Request-Id"

// Filters contains the composed chain
type Filters struct {
	next func(w http.ResponseWriter, r *http.Request)
	log  *logger.UPPLogger
}

// Filter creates a new composable filter.
func Filter(next func(w http.ResponseWriter, r *http.Request), log *logger.UPPLogger) Filters {
	return Filters{next: next, log: log}
}

// FilterSyntheticTransactions will filter out incoming requests if they have a synthetic prefix.
func (f Filters) FilterSyntheticTransactions() Filters {
	next := f.next
	f.next = filterSyntheticTransactions(next, f.log)
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

func filterSyntheticTransactions(next func(w http.ResponseWriter, r *http.Request), log *logger.UPPLogger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(tidHeader)
		if tid == "" {
			log.WithField("transaction_id", tid).Infof("Rejecting notification; it has no transaction id.")
			writeMessage("Rejecting notification; it has no transaction id.", 400, w)
			return
		}

		if strings.HasPrefix(strings.ToUpper(tid), synthTidPrefix) {
			log.WithField("transaction_id", tid).Infof("Rejecting notification; it has a synthetic transaction id.")
			writeMessage("Rejecting notification; it has a synthetic transaction id.", 200, w)
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
				writeMessage(err.Error(), http.StatusBadRequest, w)
				return
			}

			r.Body = unzipped
		}
		next(w, r)
	}
}
