package resources

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDebugLogLevel(t *testing.T) {
	req, _ := http.NewRequest("POST", "/__log", strings.NewReader(`{"level": "debug"}`))
	uppLogger := logger.NewUPPLogger("test", "INFO")

	w := httptest.NewRecorder()
	UpdateLogLevel(uppLogger)(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, log.DebugLevel, uppLogger.GetLevel())
}

func TestInfoLogLevel(t *testing.T) {
	req, _ := http.NewRequest("POST", "/__log", strings.NewReader(`{"level": "info"}`))
	uppLogger := logger.NewUPPLogger("test", "debug")

	w := httptest.NewRecorder()
	UpdateLogLevel(uppLogger)(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, log.InfoLevel, uppLogger.GetLevel())
}

func TestUnsupportedLogLevel(t *testing.T) {
	req, _ := http.NewRequest("POST", "/__log", strings.NewReader(`{"level": "warn"}`))
	uppLogger := logger.NewUPPLogger("test", "debug")
	expected := uppLogger.GetLevel()

	w := httptest.NewRecorder()
	UpdateLogLevel(uppLogger)(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, expected, uppLogger.GetLevel())
}

func TestLevelNotJson(t *testing.T) {
	uppLogger := logger.NewUPPLogger("test", "debug")
	expected := uppLogger.GetLevel()
	req, _ := http.NewRequest("POST", "/__log", strings.NewReader(`{"level": "where's my closing quote?}`))

	w := httptest.NewRecorder()
	UpdateLogLevel(uppLogger)(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, expected, uppLogger.GetLevel())
}
