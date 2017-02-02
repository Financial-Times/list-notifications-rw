package resources

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/stretchr/testify/assert"
)

func TestCarouselFilter(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	mockDb := new(MockDB)
	mockTx := new(MockTX)

	mockDb.On("Open").Return(mockTx, nil)
	mockTx.On("Close").Return()

	expectedNotification := []model.InternalNotification{{}}
	mockTx.On("FindNotification", "tid_123761283").Return(&expectedNotification, nil)

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockDb).Build()(w, req)

	mockDb.AssertExpectations(t)
	mockTx.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
}

func TestGeneratedCarouselFilter(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	mockDb := new(MockDB)

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890_gentx")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockDb).Build()(w, req)

	mockDb.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
}

func TestNoOriginalPublish(t *testing.T) {
	passed := false
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request was forwarded on as expected.")
		passed = true
	}

	mockDb := new(MockDB)
	mockTx := new(MockTX)

	mockDb.On("Open").Return(mockTx, nil)
	mockTx.On("Close").Return()

	mockTx.On("FindNotification", "tid_123761283").Return(nil, nil)

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockDb).Build()(w, req)

	mockDb.AssertExpectations(t)
	mockTx.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
	assert.True(t, passed)
}

func TestErrorFindingOriginalPublish(t *testing.T) {
	passed := false
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request was forwarded on as expected.")
		passed = true
	}

	mockDb := new(MockDB)
	mockTx := new(MockTX)

	mockDb.On("Open").Return(mockTx, nil)
	mockTx.On("Close").Return()

	mockTx.On("FindNotification", "tid_123761283").Return(nil, errors.New("blew up finding that pesky original publish"))

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockDb).Build()(w, req)

	mockDb.AssertExpectations(t)
	mockTx.AssertExpectations(t)

	assert.Equal(t, 200, w.Code)
	assert.True(t, passed)
}

func TestErrorOpeningMongoConnection(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	mockDb := new(MockDB)
	mockTx := new(MockTX)

	mockDb.On("Open").Return(nil, errors.New("blew up connecting to mongo"))

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283_carousel_1234567890")

	w := httptest.NewRecorder()
	Filter(next).FilterCarouselPublishes(mockDb).Build()(w, req)

	mockDb.AssertExpectations(t)
	mockTx.AssertExpectations(t)

	assert.Equal(t, 500, w.Code)
}
