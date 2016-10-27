package resources

import (
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

		if strings.HasPrefix(strings.ToUpper(tid), synthTidPrefix) {
			logrus.WithField("tid", tid).Infof("Rejecting notification; it has a synthetic transaction id.")
			w.WriteHeader(200)
			return
		}

		next(w, r)
	}
}
