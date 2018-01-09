package resources

import (
	"net/http"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Financial-Times/service-status-go/gtg"
)

const (
	contentType  = "Content-Type"
	plainText    = "text/plain; charset=US-ASCII"
	cacheControl = "Cache-control"
	noCache      = "no-cache"
)

type HealthService struct {
	fthealth.TimedHealthCheck
}

func NewHealthService(db db.DB, appSystemCode string, appName string, appDescription string) *HealthService {
	hcService := &HealthService{}
	hcService.SystemCode = appSystemCode
	hcService.Name = appName
	hcService.Description = appDescription
	hcService.Timeout = 10 * time.Second
	hcService.Checks = getHealthchecks(db)

	return hcService
}

// HealthChecks returns a handler for the standard FT healthchecks
func (service *HealthService) HealthChecksHandler() func(w http.ResponseWriter, r *http.Request) {
	return fthealth.Handler(service)
}

// GTG lightly tests the service and returns an FT standard GTG response
func (service *HealthService) GTG() gtg.Status {
	for _, check := range service.Checks {
		if _, err := check.Checker(); err != nil {
			return gtg.Status{GoodToGo: false, Message: err.Error()}
		}
	}
	return gtg.Status{GoodToGo: true}
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
