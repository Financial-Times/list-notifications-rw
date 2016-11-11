package resources

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
)

const synthTidPrefix = "SYNTHETIC-REQ-MON"
const tidHeader = "X-Request-Id"

// FilterSyntheticTransactions will filter out incoming requests if they have a synthetic prefix.
func FilterSyntheticTransactions(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
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

// UnzipGzip pre-processes the request body if it's gzipped.
func UnzipGzip(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
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
