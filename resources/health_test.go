package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/stretchr/testify/assert"
)

func TestHealthy(t *testing.T) {
	mockClient := new(MockClient)

	mockClient.On("Ping").Return(nil)
	mockClient.On("EnsureIndexes").Return(nil)

	req, _ := http.NewRequest("GET", "http://nothing/__health", nil)
	w := httptest.NewRecorder()

	hs := NewHealthService(mockClient, "app-system-code", "app-name", "Description of app")
	hs.HealthChecksHandler()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	health := fthealth.HealthResult{}
	err := json.NewDecoder(w.Body).Decode(&health)
	if err != nil {
		t.Fatal("Should return valid json!")
	}

	assert.NotEmpty(t, health.Name, "Should have a non-empty name")
	assert.NotEmpty(t, health.Description, "Should have a non-empty description")
	assert.NotEmpty(t, health.SchemaVersion, "Should have a non-empty schema version")
	assert.True(t, health.Ok, "Expect it's ok")

	assert.Len(t, health.Checks, 2, "Only one health check currently")
	check := health.Checks[0]

	assert.NotEmpty(t, check.Name, "Should have a non-empty name")
	assert.NotEmpty(t, check.PanicGuide, "Should have a non-empty panic guide")
	assert.Equal(t, uint8(1), check.Severity, "Severity 1")
	assert.NotEmpty(t, check.BusinessImpact, "Should have a non-empty business impact")
	assert.NotEmpty(t, check.TechnicalSummary, "Should have a non-empty technical summary")

	assert.True(t, check.Ok, "Expect it's ok")

	mockClient.AssertExpectations(t)
}

func TestUnhealthyBecauseOfPing(t *testing.T) {
	mockClient := new(MockClient)

	mockClient.On("Ping").Return(errors.New("we ain't looking too good"))
	mockClient.On("EnsureIndexes").Return(nil)

	req, _ := http.NewRequest("GET", "http://nothing/__health", nil)
	w := httptest.NewRecorder()

	hs := NewHealthService(mockClient, "app-system-code", "app-name", "Description of app")
	hs.HealthChecksHandler()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	health := fthealth.HealthResult{}
	err := json.NewDecoder(w.Body).Decode(&health)
	if err != nil {
		t.Fatal("Should return valid json!")
	}

	assert.NotEmpty(t, health.Name, "Should have a non-empty name")
	assert.NotEmpty(t, health.Description, "Should have a non-empty description")
	assert.NotEmpty(t, health.SchemaVersion, "Should have a non-empty schema version")
	assert.False(t, health.Ok, "Expect it's ok")

	assert.Len(t, health.Checks, 2, "Only one health check currently")
	check := health.Checks[0]

	assert.NotEmpty(t, check.Name, "Should have a non-empty name")
	assert.NotEmpty(t, check.PanicGuide, "Should have a non-empty panic guide")
	assert.Equal(t, uint8(1), check.Severity, "Severity 1")
	assert.NotEmpty(t, check.BusinessImpact, "Should have a non-empty business impact")
	assert.NotEmpty(t, check.TechnicalSummary, "Should have a non-empty technical summary")

	assert.False(t, check.Ok, "Expect it's not ok")

	mockClient.AssertExpectations(t)
}

func TestUnhealthyBecauseOfEnsureIndexes(t *testing.T) {
	mockClient := new(MockClient)

	mockClient.On("Ping").Return(nil)
	mockClient.On("EnsureIndexes").Return(errors.New("we ain't looking too good"))

	req, _ := http.NewRequest("GET", "http://nothing/__health", nil)
	w := httptest.NewRecorder()

	hs := NewHealthService(mockClient, "app-system-code", "app-name", "Description of app")
	hs.HealthChecksHandler()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	health := fthealth.HealthResult{}
	err := json.NewDecoder(w.Body).Decode(&health)
	if err != nil {
		t.Fatal("Should return valid json!")
	}

	assert.NotEmpty(t, health.Name, "Should have a non-empty name")
	assert.NotEmpty(t, health.Description, "Should have a non-empty description")
	assert.NotEmpty(t, health.SchemaVersion, "Should have a non-empty schema version")
	assert.True(t, health.Ok, "Expect it's ok")

	assert.Len(t, health.Checks, 2, "Only one health check currently")
	check := health.Checks[1]

	assert.NotEmpty(t, check.Name, "Should have a non-empty name")
	assert.NotEmpty(t, check.PanicGuide, "Should have a non-empty panic guide")
	assert.Equal(t, uint8(2), check.Severity, "Severity 2")
	assert.NotEmpty(t, check.BusinessImpact, "Should have a non-empty business impact")
	assert.NotEmpty(t, check.TechnicalSummary, "Should have a non-empty technical summary")

	assert.True(t, check.Ok, "Expect it's not ok")

	mockClient.AssertExpectations(t)
}

func TestWorkingGTG(t *testing.T) {
	mockClient := new(MockClient)

	mockClient.On("Ping").Return(nil)
	mockClient.On("EnsureIndexes").Return(nil)

	req, _ := http.NewRequest("GET", "http://nothing/at/__gtg", nil)
	w := httptest.NewRecorder()

	hs := NewHealthService(mockClient, "app-system-code", "app-name", "Description of app")
	status.NewGoodToGoHandler(hs.GTG)(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}

func TestFailingGTG(t *testing.T) {
	mockClient := new(MockClient)

	mockClient.On("Ping").Return(errors.New("omg we are not gtg"))

	req, _ := http.NewRequest("GET", "http://nothing/at/__gtg", nil)
	w := httptest.NewRecorder()

	hs := NewHealthService(mockClient, "app-system-code", "app-name", "Description of app")
	status.NewGoodToGoHandler(hs.GTG)(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	mockClient.AssertExpectations(t)
}

func TestFailingDBGTG(t *testing.T) {
	mockClient := new(MockClient)

	mockClient.On("Ping").Return(errors.New("omg we are not gtg"))

	req, _ := http.NewRequest("GET", "http://nothing/at/__gtg", nil)
	w := httptest.NewRecorder()

	hs := NewHealthService(mockClient, "app-system-code", "app-name", "Description of app")
	status.NewGoodToGoHandler(hs.GTG)(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	mockClient.AssertExpectations(t)
}
