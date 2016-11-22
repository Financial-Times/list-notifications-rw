package resources

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDebugLogLevel(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://our.host.name/__log/debug", nil)

	w := httptest.NewRecorder()
	r := LogRoute(UpdateLogLevel())
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
}

func TestInfoLogLevel(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://our.host.name/__log/info", nil)

	w := httptest.NewRecorder()
	r := LogRoute(UpdateLogLevel())
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
}

func TestWarnLogLevel(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://our.host.name/__log/warn", nil)

	w := httptest.NewRecorder()
	r := LogRoute(UpdateLogLevel())
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, logrus.WarnLevel, logrus.GetLevel())
}

func TestUnsupportedLogLevel(t *testing.T) {
	expected := logrus.GetLevel()
	req, _ := http.NewRequest("GET", "http://our.host.name/__log/error", nil)

	w := httptest.NewRecorder()
	r := LogRoute(UpdateLogLevel())
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, expected, logrus.GetLevel())
}
