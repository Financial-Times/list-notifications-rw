package resources

import (
	"bytes"
	"compress/gzip"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/stretchr/testify/assert"
)

func TestFilterSyntheticTransactions(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, synthTidPrefix+"_a-fake-tid-which-should-be-rejected-if-all-goes-well")

	w := httptest.NewRecorder()
	Filter(next).FilterSyntheticTransactions().Build()(w, req)

	assert.Equal(t, 200, w.Code)
	t.Log("Request was filtered.")
}

func TestAllowsNormalTransactions(t *testing.T) {
	passed := false
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request was forwarded on as expected.")
		passed = true
	}

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283")

	w := httptest.NewRecorder()
	mockDb := new(MockDB)

	Filter(next).FilterSyntheticTransactions().FilterCarouselPublishes(mockDb).Build()(w, req)

	assert.Equal(t, 200, w.Code)
	assert.True(t, passed)
}

func TestAllowsNoTID(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)

	w := httptest.NewRecorder()
	Filter(next).FilterSyntheticTransactions().Build()(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestFailUnzip(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	req, _ := http.NewRequest("PUT", "http://nothing/at/all", strings.NewReader("gibberish"))
	req.Header.Add("Content-Encoding", "gzip")

	w := httptest.NewRecorder()
	Filter(next).Gunzip().Build()(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestUnzipOk(t *testing.T) {
	passed := false
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request was forwarded on as expected.")
		passed = true
	}

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write([]byte("not gibberish"))

	req, _ := http.NewRequest("PUT", "http://nothing/at/all", bytes.NewReader(b.Bytes()))
	req.Header.Add("Content-Encoding", "gzip")

	w := httptest.NewRecorder()
	Filter(next).Gunzip().Build()(w, req)

	assert.Equal(t, 200, w.Code)
	assert.True(t, passed)
}

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

func TestSyntheticCarouselFilter(t *testing.T) {
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
