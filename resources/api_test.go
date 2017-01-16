package resources

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPI(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://nothing/at/__api", nil)
	w := httptest.NewRecorder()
	API([]byte("hi"))(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "hi", w.Body.String(), "Should be hi!")
	assert.Equal(t, []string{"text/vnd.yaml"}, w.HeaderMap["Content-Type"], "Should be YAML.")
	t.Log("Request was filtered.")
}
