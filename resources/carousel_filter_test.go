package resources

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCarouselFilter(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	mockClient := new(MockClient)
	mockClient.On("FindNotificationByTransactionID", "tid_123761283").Return(model.InternalNotification{}, nil)

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockClient).Build()(w, req)

	mockClient.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
}

func TestCarouselFilterWithUnconventionalTransactionID(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	mockClient := new(MockClient)
	mockClient.On("FindNotificationByTransactionID", "republish_-10bd337c-66d4-48d9-ab8a-e8441fa2ec98").Return(model.InternalNotification{}, nil)

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "republish_-10bd337c-66d4-48d9-ab8a-e8441fa2ec98_carousel_1493606135")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockClient).Build()(w, req)

	mockClient.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
}

func TestPartialCarouselFilter(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	mockClient := new(MockClient)
	mockClient.On("FindNotificationByTransactionID", "tid_123761283").Return(model.InternalNotification{}, mongo.ErrNoDocuments)
	mockClient.On("FindNotificationByPartialTransactionID", "tid_123761283_carousel").Return(model.InternalNotification{}, nil)

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockClient).Build()(w, req)

	mockClient.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
}

func TestGeneratedCarouselFilter(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	mockClient := new(MockClient)

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890_gentx")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockClient).Build()(w, req)

	mockClient.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
}

func TestNoOriginalPublish(t *testing.T) {
	passed := false
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request was forwarded on as expected.")
		passed = true
	}

	mockClient := new(MockClient)
	mockClient.On("FindNotificationByTransactionID", "tid_123761283").Return(model.InternalNotification{}, mongo.ErrNoDocuments)
	mockClient.On("FindNotificationByPartialTransactionID", "tid_123761283_carousel").Return(model.InternalNotification{}, mongo.ErrNoDocuments)

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockClient).Build()(w, req)

	mockClient.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
	assert.True(t, passed)
}

func TestErrorFindingOriginalPublish(t *testing.T) {
	passed := false
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request was forwarded on as expected.")
		passed = true
	}

	mockClient := new(MockClient)

	mockClient.
		On("FindNotificationByTransactionID", "tid_123761283").
		Return(model.InternalNotification{}, errors.New("blew up finding that pesky original publish"))

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockClient).Build()(w, req)

	mockClient.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
	assert.True(t, passed)
}

func TestErrorFindingPartialCarouselPublish(t *testing.T) {
	passed := false
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request was forwarded on as expected.")
		passed = true
	}

	mockClient := new(MockClient)
	mockClient.On("FindNotificationByTransactionID", "tid_123761283").Return(model.InternalNotification{}, mongo.ErrNoDocuments)
	mockClient.On("FindNotificationByPartialTransactionID", "tid_123761283_carousel").Return(model.InternalNotification{}, errors.New("blew up finding that pesky original publish"))

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockClient).Build()(w, req)

	mockClient.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
	assert.True(t, passed)
}
