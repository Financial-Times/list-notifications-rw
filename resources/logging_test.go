package resources

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDebugLogLevel(t *testing.T) {
	req, _ := http.NewRequest("POST", "/__log", strings.NewReader(`{"level": "debug"}`))

	w := httptest.NewRecorder()
	UpdateLogLevel()(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, log.DebugLevel, log.GetLevel())
}

func TestInfoLogLevel(t *testing.T) {
	req, _ := http.NewRequest("POST", "/__log", strings.NewReader(`{"level": "info"}`))

	w := httptest.NewRecorder()
	UpdateLogLevel()(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, log.InfoLevel, log.GetLevel())
}

func TestUnsupportedLogLevel(t *testing.T) {
	expected := log.GetLevel()
	req, _ := http.NewRequest("POST", "/__log", strings.NewReader(`{"level": "warn"}`))

	w := httptest.NewRecorder()
	UpdateLogLevel()(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, expected, log.GetLevel())
}

func TestLevelNotJson(t *testing.T) {
	expected := log.GetLevel()
	req, _ := http.NewRequest("POST", "/__log", strings.NewReader(`{"level": "where's my closing quote?}`))

	w := httptest.NewRecorder()
	UpdateLogLevel()(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, expected, log.GetLevel())
}
