package resources

import (
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/list-notifications-rw/db"
)

const (
	contentType      = "Content-Type"
	plainText        = "text/plain; charset=US-ASCII"
	cacheControl     = "Cache-control"
	noCache          = "no-cache"
)

// Health returns a handler for the standard FT healthchecks
func Health(db db.DB) func(w http.ResponseWriter, r *http.Request) {
	return fthealth.Handler("list-notifications-rw", "Notifies clients of updates to UPP Lists.", getHealthchecks(db)[0])
}

// GTG returns a handler for a standard GTG endpoint.
func GTG(db db.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, plainText)
		w.Header().Set(cacheControl, noCache)
		_, err := pingMongo(db)()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getHealthchecks(db db.DB) []fthealth.Check {
	return []fthealth.Check{
		{
			Name:             "CheckConnectivityToListsDatabase",
			BusinessImpact:   "Notifications for list changes will not be available to API consumers (NextFT).",
			TechnicalSummary: "The service is unable to connect to MongoDB. Notifications cannot be written to or read from the store.",
			Severity:         1,
			PanicGuide:       "https://dewey.ft.com/upp-list-notifications-rw.html",
			Checker:          pingMongo(db),
		},
	}
}

func pingMongo(db db.DB) func() (string, error) {
	return func() (string, error) {
		tx, err := db.Open()
		if err != nil {
			return "", err
		}

		defer tx.Close()

		return "", tx.Ping()
	}
}
