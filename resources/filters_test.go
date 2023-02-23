package resources

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	mockClient := new(MockClient)

	Filter(next).FilterSyntheticTransactions().FilterCarouselPublishes(mockClient).Build()(w, req)

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
