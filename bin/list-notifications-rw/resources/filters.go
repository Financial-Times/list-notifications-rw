package resources

import (
	"net/http"
	"strings"
	"github.com/Sirupsen/logrus"
)

const synth_tid_prefix = "SYNTHETIC-REQ-MON"
const tid_header = "X-Request-Id"

func FilterSyntheticTransactions(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request){
		tid := r.Header.Get(tid_header)

		if strings.HasPrefix(strings.ToUpper(tid), synth_tid_prefix) {
			logrus.WithField("tid", tid).Infof("Rejecting notification; it has a synthetic transaction id.")
			w.WriteHeader(200)
			return
		}

		next(w, r)
	}
}
