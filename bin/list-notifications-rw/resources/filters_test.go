package resources

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFilterSyntheticTransactions(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't reach here!")
	}

	req := httptest.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, synthTidPrefix+"_a-fake-tid-which-should-be-rejected-if-all-goes-well")

	w := httptest.NewRecorder()
	FilterSyntheticTransactions(next)(w, req)

	t.Log("Request was filtered.")
}

func TestAllowsNormalTransactions(t *testing.T) {
	passed := false
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request was forwarded on as expected.")
		passed = true
	}

	req := httptest.NewRequest("GET", "http://nothing/at/all", nil)
	req.Header.Add(tidHeader, "tid_123761283")

	w := httptest.NewRecorder()
	FilterSyntheticTransactions(next)(w, req)

	if !passed {
		t.Fatal("Failed to reach next handler!")
	}
}

func TestAllowsNoTID(t *testing.T) {
	passed := false
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request was forwarded on as expected.")
		passed = true
	}

	req := httptest.NewRequest("GET", "http://nothing/at/all", nil)

	w := httptest.NewRecorder()
	FilterSyntheticTransactions(next)(w, req)

	if !passed {
		t.Fatal("Failed to reach next handler!")
	}
}
