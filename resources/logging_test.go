package resources

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDebugLogLevel(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://our.host.name/__log", strings.NewReader(`{"level": "debug"}`))

	w := httptest.NewRecorder()
	UpdateLogLevel()(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
}

func TestInfoLogLevel(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://our.host.name/__log", strings.NewReader(`{"level": "info"}`))

	w := httptest.NewRecorder()
	UpdateLogLevel()(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
}

func TestUnsupportedLogLevel(t *testing.T) {
	expected := logrus.GetLevel()
	req, _ := http.NewRequest("POST", "http://our.host.name/__log", strings.NewReader(`{"level": "warn"}`))

	w := httptest.NewRecorder()
	UpdateLogLevel()(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, expected, logrus.GetLevel())
}

func TestLevelNotJson(t *testing.T) {
	expected := logrus.GetLevel()
	req, _ := http.NewRequest("POST", "http://our.host.name/__log", strings.NewReader(`{"level": "where's my closing quote?}`))

	w := httptest.NewRecorder()
	UpdateLogLevel()(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, expected, logrus.GetLevel())
}
