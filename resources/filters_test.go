package resources

import (
	"net/http"
	"net/http/httptest"
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
	FilterSyntheticTransactions(next)(w, req)

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
	FilterSyntheticTransactions(next)(w, req)

	assert.Equal(t, 200, w.Code)
	if !passed {
		t.Fatal("Failed to reach next handler!")
	}
}

func TestAllowsNoTID(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	req, _ := http.NewRequest("GET", "http://nothing/at/all", nil)

	w := httptest.NewRecorder()
	FilterSyntheticTransactions(next)(w, req)

	assert.Equal(t, 400, w.Code)
}
