package resources

import (
	"net/http"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Financial-Times/service-status-go/gtg"
)

type HealthService struct {
	fthealth.TimedHealthCheck
}

func NewHealthService(db db.Database, appSystemCode string, appName string, appDescription string) *HealthService {
	hcService := &HealthService{}
	hcService.SystemCode = appSystemCode
	hcService.Name = appName
	hcService.Description = appDescription
	hcService.Timeout = 10 * time.Second
	hcService.Checks = getHealthChecks(db)

	return hcService
}

// HealthChecksHandler HealthChecks returns a handler for the standard FT health checks
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

func getHealthChecks(db db.Database) []fthealth.Check {
	return []fthealth.Check{
		{
			Name:             "CheckConnectivityToListsDatabase",
			BusinessImpact:   "Notifications for list changes will not be available to API consumers (NextFT).",
			TechnicalSummary: "The service is unable to connect to MongoDB. Notifications cannot be written to or read from the store.",
			Severity:         1,
			PanicGuide:       "https://runbooks.ftops.tech/upp-list-notifications-rw",
			Checker:          pingMongo(db),
		},
		{
			Name:           "Page Notifications RW - Search indexes are created",
			BusinessImpact: "Some API consumers may experience slow performance for content requests",
			TechnicalSummary: "The application indexes for the DocumentDB instance may not be up-to-date (indexing may be in progress). " +
				"This will result in degraded performance from the content platform and affect a variety of products.",
			PanicGuide: "https://runbooks.ftops.tech/upp-list-notifications-rw",
			Severity:   2,
			Checker:    ensureIndexes(db),
		},
	}
}

func pingMongo(db db.Database) func() (string, error) {
	return func() (string, error) {
		return "", db.Ping()
	}
}

func ensureIndexes(db db.Database) func() (string, error) {
	return func() (string, error) {
		if err := db.EnsureIndexes(); err != nil {
			return "DocumentDB indexes may not be up-to-date", err
		}
		return "DocumentDB indexes are updated", nil
	}
}
