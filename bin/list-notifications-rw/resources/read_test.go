package resources

import (
	"testing"
	"net/http/httptest"
	"time"
	"github.com/Financial-Times/list-notifications-rw/model"
	"errors"
	"encoding/json"
	"github.com/stretchr/testify/assert"
)

func TestReadNotifications(t *testing.T) {
	mockSince, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.99999Z")

	req := httptest.NewRequest("GET", "http://nothing/at/all?since=2006-01-02T15:04:05.99999Z", nil)

	w := httptest.NewRecorder()

	mockDb := new(MockDB)
	mockTx := new(MockTX)

	mockDb.On("Open").Return(mockTx, nil)

	changeDate := time.Now()
	mockNotifications := []model.InternalNotification{
		{
			UUID: "uuid",
			Title: "title",
			LastModified: changeDate,
			EventType: "UPDATE",
			PublishReference: "tid_blah-blah-blah",
		},
	}

	mockTx.On("Close").Return()
	mockTx.On("ReadNotifications", mockSince).Return(&mockNotifications, nil)

	ReadNotifications(testMapper, mockDb)(w, req)

	if w.Code != 200 {
		t.Fatal("Everything should be OK but we didn't return 200!")
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Fatal("Everything should be OK but we didn't return json!")
	}

	decoder := json.NewDecoder(w.Body)
	results := []model.PublicNotification{}
	decoder.Decode(&results)

	// TODO: Mock the mapper?
	assert.Len(t, results, 1, "Data should contain one item!")
	assert.Equal(t, "http://testing-123.com/things/uuid", results[0].Id)
	assert.Equal(t, "http://testing-123.com/lists/uuid", results[0].ApiUrl)
	assert.Equal(t, "http://www.ft.com/thing/ThingChangeType/UPDATE", results[0].Type)
	assert.Equal(t, "title", results[0].Title)
	assert.Equal(t, changeDate.UTC(), results[0].LastModified)
	assert.Equal(t, "tid_blah-blah-blah", results[0].PublishReference)

	mockDb.AssertExpectations(t)
	mockTx.AssertExpectations(t)

	t.Log("Read request worked as expected.")
}

func Test400NoSinceDate(t *testing.T) {
	req := httptest.NewRequest("GET", "http://nothing/at/all", nil)

	w := httptest.NewRecorder()

	mockDb := new(MockDB)
	ReadNotifications(testMapper, mockDb)(w, req)

	if w.Code != 400 {
		t.Fatal("No since date! Should be 400!")
	}

	t.Log("Recorded 400 response as expected.")
	mockDb.AssertNotCalled(t, "Open")
}

func Test400JunkSinceDate(t *testing.T) {
	req := httptest.NewRequest("GET", "http://nothing/at/all?since=some-garbage-date", nil)

	w := httptest.NewRecorder()

	mockDb := new(MockDB)
	ReadNotifications(testMapper, mockDb)(w, req)

	if w.Code != 400 {
		t.Fatal("The since date was garbage! Should be 400!")
	}

	mockDb.AssertNotCalled(t, "Open")
	t.Log("Recorded 400 response as expected.")
}

func TestFailedDatabaseOnRead(t *testing.T) {
	req := httptest.NewRequest("GET", "http://nothing/at/all?since=2006-01-02T15:04:05.999Z", nil)

	w := httptest.NewRecorder()

	mockDb := new(MockDB)
	mockDb.On("Open").Return(nil, errors.New("I broke soz"))

	ReadNotifications(testMapper, mockDb)(w, req)

	if w.Code != 500 {
		t.Fatal("Mongo was broken but we didn't return 500!")
	}

	mockDb.AssertExpectations(t)
	t.Log("Recorded 500 response as expected, and since date was accepted.")
}

func TestFailedToQuery(t *testing.T) {
	mockSince, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.99999Z")

	req := httptest.NewRequest("GET", "http://nothing/at/all?since=2006-01-02T15:04:05.99999Z", nil)

	w := httptest.NewRecorder()

	mockDb := new(MockDB)
	mockTx := new(MockTX)

	mockTx.On("Close").Return()
	mockTx.On("ReadNotifications", mockSince).Return(nil, errors.New("I broke again soz"))
	mockDb.On("Open").Return(mockTx, nil)

	ReadNotifications(testMapper, mockDb)(w, req)

	if w.Code != 500 {
		t.Fatal("Mongo failed to query but we didn't return 500!")
	}

	mockDb.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	t.Log("Recorded 500 response as expected, and since date was accepted.")
}