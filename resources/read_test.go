package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/stretchr/testify/assert"
)

func TestReadNotifications(t *testing.T) {
	log := logger.NewUPPLogger("test", "debug")
	mockSince, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.99999Z")

	req, _ := http.NewRequest("GET", "http://nothing/at/all?since=2006-01-02T15:04:05.99999Z", nil)
	w := httptest.NewRecorder()

	mockClient := new(MockClient)

	mockClient.On("GetLimit").Return(1)

	changeDate := time.Now()
	mockNotifications := []model.InternalNotification{
		{
			UUID:             "uuid",
			Title:            "title",
			LastModified:     changeDate,
			EventType:        "UPDATE",
			PublishReference: "tid_blah-blah-blah",
		},
		{
			UUID:             "uuid2",
			Title:            "title",
			LastModified:     changeDate,
			EventType:        "UPDATE",
			PublishReference: "tid_blah-blah-blah",
		},
	}

	mockClient.On("ReadNotifications", 0, mockSince).Return(&mockNotifications, nil)

	ReadNotifications(testMapper, testLinkGenerator, mockClient, 10000, log)(w, req)

	assert.Equal(t, 200, w.Code, "Everything should be OK but we didn't return 200!")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Everything should be OK but we didn't return json!")

	decoder := json.NewDecoder(w.Body)
	page := model.PublicNotificationPage{}
	decoder.Decode(&page)

	assert.Equal(t, "http://testing-123.com/at/all?since=2006-01-02T15:04:05.99999Z", page.RequestURL)

	results := page.Notifications

	// TODO: Mock the mapper?
	assert.Len(t, results, 1, "Data should contain one item!")
	assert.Equal(t, "http://testing-123.com/things/uuid", results[0].ID)
	assert.Equal(t, "http://testing-123.com/lists/uuid", results[0].APIURL)
	assert.Equal(t, "http://www.ft.com/thing/ThingChangeType/UPDATE", results[0].Type)
	assert.Equal(t, "title", results[0].Title)
	assert.Equal(t, changeDate.UTC(), results[0].LastModified)
	assert.Equal(t, "tid_blah-blah-blah", results[0].PublishReference)

	mockClient.AssertExpectations(t)

	t.Log("Read request worked as expected.")
}

func TestReadNoNotifications(t *testing.T) {
	log := logger.NewUPPLogger("test", "debug")
	mockSince, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.99999Z")

	req, _ := http.NewRequest("GET", "http://nothing/at/all?since=2006-01-02T15:04:05.99999Z", nil)
	w := httptest.NewRecorder()

	mockClient := new(MockClient)

	var mockNotifications []model.InternalNotification

	mockClient.On("ReadNotifications", 0, mockSince).Return(&mockNotifications, nil)

	ReadNotifications(testMapper, testLinkGenerator, mockClient, 10000, log)(w, req)

	assert.Equal(t, 200, w.Code, "Everything should be OK but we didn't return 200!")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Everything should be OK but we didn't return json!")

	decoder := json.NewDecoder(w.Body)
	page := model.PublicNotificationPage{}
	decoder.Decode(&page)

	results := page.Notifications

	assert.Len(t, results, 0)
	mockClient.AssertExpectations(t)
}

func Test400NoSinceDate(t *testing.T) {
	log := logger.NewUPPLogger("test", "debug")
	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	w := httptest.NewRecorder()

	mockClient := new(MockClient)
	ReadNotifications(testMapper, testLinkGenerator, mockClient, 10000, log)(w, req)

	assert.Equal(t, 400, w.Code, "No since date, should be 400!")
	assert.True(t, strings.Contains(w.Body.String(), "{\"message\":\"A mandatory 'since' query parameter has not been specified. Please supply a since date. For eg., since="), "Did not receive expected error message for missing since date")

	t.Log("Recorded 400 response as expected.")
	mockClient.AssertNotCalled(t, "Open")
}

func Test400JunkSinceDate(t *testing.T) {
	log := logger.NewUPPLogger("test", "debug")
	req, _ := http.NewRequest("GET", "http://nothing/at/all?since=some-garbage-date", nil)
	w := httptest.NewRecorder()

	mockClient := new(MockClient)
	ReadNotifications(testMapper, testLinkGenerator, mockClient, 10000, log)(w, req)

	assert.Equal(t, 400, w.Code, "The since date was garbage! Should be 400!")
	assert.True(t, strings.Contains(w.Body.String(), "{\"message\":\"A mandatory 'since' query parameter has not been specified. Please supply a since date. For eg., since="), "Did not receive expected error message junk since date")

	t.Log("Recorded 400 response as expected.")
}

func Test400SinceDateTooEarly(t *testing.T) {
	log := logger.NewUPPLogger("test", "debug")
	req, _ := http.NewRequest("GET", "http://nothing/at/all?since=2006-01-02T15:04:05.999Z", nil)
	w := httptest.NewRecorder()

	mockClient := new(MockClient)
	ReadNotifications(testMapper, testLinkGenerator, mockClient, 90, log)(w, req)

	assert.Equal(t, 400, w.Code, "Since date too early, should be 400!")
	assert.Equal(t, "{\"message\":\"Since date must be within the last 90 days.\"}\n", w.Body.String(), "Did not receive correct error message for since date before max time interval")

	t.Log("Recorded 400 response as expected.")
}

func TestFailedDatabaseOnRead(t *testing.T) {
	log := logger.NewUPPLogger("test", "debug")
	req, _ := http.NewRequest("GET", "http://nothing/at/all?since=2006-01-02T15:04:05.999Z", nil)
	w := httptest.NewRecorder()
	mockSince, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999Z")

	mockClient := new(MockClient)
	mockClient.On("ReadNotifications", 0, mockSince).Return(nil, errors.New("I broke soz"))

	ReadNotifications(testMapper, testLinkGenerator, mockClient, 10000, log)(w, req)

	assert.Equal(t, 500, w.Code, "Mongo was broken but we didn't return 500!")
	assert.Equal(t, "{\"message\":\"Failed to retrieve list notifications due to internal server error\"}\n", w.Body.String(), "Did not receive expected error message Mongo database read fail")

	mockClient.AssertExpectations(t)
	t.Log("Recorded 500 response as expected, and since date was accepted.")
}

func TestInvalidOffset(t *testing.T) {
	log := logger.NewUPPLogger("test", "debug")
	req, _ := http.NewRequest("GET", "http://nothing/at/all?since=2006-01-02T15:04:05.99999Z&offset=i-am-soooo-wrong", nil)
	w := httptest.NewRecorder()

	mockClient := new(MockClient)
	ReadNotifications(testMapper, testLinkGenerator, mockClient, 10000, log)(w, req)

	assert.Equal(t, 400, w.Code, "Offset was invalid but we didn't 400!")
	assert.Equal(t, "{\"message\":\"Please specify an integer offset.\"}\n", w.Body.String(), "Did not receive expected  error message for invalid offset")

	t.Log("Recorded 400 response as expected.")
}

func TestFailedToQueryAndOffset(t *testing.T) {
	log := logger.NewUPPLogger("test", "debug")
	mockSince, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.99999Z")

	req, _ := http.NewRequest("GET", "http://nothing/at/all?since=2006-01-02T15:04:05.99999Z&offset=100", nil)
	w := httptest.NewRecorder()

	mockClient := new(MockClient)

	mockClient.On("ReadNotifications", 100, mockSince).Return(nil, errors.New("I broke again soz"))

	ReadNotifications(testMapper, testLinkGenerator, mockClient, 10000, log)(w, req)

	assert.Equal(t, 500, w.Code, "Mongo failed to query but we didn't return 500!")
	assert.Equal(t, "{\"message\":\"Failed to retrieve list notifications due to internal server error\"}\n", w.Body.String(), "Did not receive expected error message Mongo database read fail")

	mockClient.AssertExpectations(t)
	t.Log("Recorded 500 response as expected, and since date was accepted.")
}
