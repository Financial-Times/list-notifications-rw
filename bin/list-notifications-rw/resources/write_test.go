package resources

import (
	"testing"
	"net/http/httptest"
	"strings"
	"errors"
	"encoding/json"
)

var mockWriteBody = `{"uuid":"ef863741-709a-4062-a8f1-987c44db1db5","title":"Unlocking Yield Top Stories","concept":{"uuid":"3095386b-bb12-37af-bb7b-b84390937caf","prefLabel":"Investing 2.0: Unlocking Yield"},"listType":"SpecialReports","items":[{"uuid":"2b3c6398-7f3f-11e6-8e50-8ec15fb462f4"},{"uuid":"0de7bf4c-8c08-11e6-8aa5-f79f5696c731"},{"uuid":"6c9109fc-8b9c-11e6-8cb7-e7ada1d123b1"},{"uuid":"f3e173f2-8ae7-11e6-8aa5-f79f5696c731"},{"uuid":"5c94a898-8952-11e6-8aa5-f79f5696c731"}],"publishReference":"tid_uvo7bcngao","lastModified":"2016-10-20T17:08:37.668Z"}`;

func TestWriteNotification(t *testing.T) {
	req := httptest.NewRequest("PUT", "http://our.host.name/lists/notifications/ef863741-709a-4062-a8f1-987c44db1db5", strings.NewReader(mockWriteBody))

	w := httptest.NewRecorder()

	mockDb := new(MockDB)
	mockTx := new(MockTX)

	decoder := json.NewDecoder(strings.NewReader(mockWriteBody))
	expectedNotification, _ := testMapper.MapRequestToInternalNotification("ef863741-709a-4062-a8f1-987c44db1db5", decoder)

	mockDb.On("Open").Return(mockTx, nil)

	mockTx.On("Close").Return()
	mockTx.On("WriteNotification", expectedNotification).Return()

	r := WriteRoute(WriteNotification(testMapper, mockDb))
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatal("Expected a 200 response!")
	}

	mockDb.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestNotJson(t *testing.T) {
	req := httptest.NewRequest("PUT", "http://our.host.name/lists/notifications/ef863741-709a-4062-a8f1-987c44db1db5", nil)

	w := httptest.NewRecorder()

	mockDb := new(MockDB)

	r := WriteRoute(WriteNotification(testMapper, mockDb))
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatal("Expected a 400 response!")
	}

	mockDb.AssertNotCalled(t, "Open")
}

func TestNoUUID(t *testing.T) {
	req := httptest.NewRequest("PUT", "http://our.host.name/lists/notifications/ef863741-709a-4062-a8f1-987c44db1db5", strings.NewReader(`{"uuid":""}`))

	w := httptest.NewRecorder()

	mockDb := new(MockDB)

	r := WriteRoute(WriteNotification(testMapper, mockDb))
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatal("Expected a 400 response!")
	}

	mockDb.AssertNotCalled(t, "Open")
}

func TestInvalidUUID(t *testing.T) {
	req := httptest.NewRequest("PUT", "http://our.host.name/lists/notifications/ef863741-709a-4062-a8f1-987c44db1db5", strings.NewReader(`{"uuid":"i am a bit invalid"}`))

	w := httptest.NewRecorder()

	mockDb := new(MockDB)

	r := WriteRoute(WriteNotification(testMapper, mockDb))
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatal("Expected a 400 response!")
	}

	mockDb.AssertNotCalled(t, "Open")
}

func TestUUIDDoesNotMatch(t *testing.T) {
	req := httptest.NewRequest("PUT", "http://our.host.name/lists/notifications/ef863741-709a-4062-a8f1-987c44db1db5", strings.NewReader(`{"uuid":"cee15258-6762-4fc5-8f57-0b5ca4c3aa20"}`))

	w := httptest.NewRecorder()

	mockDb := new(MockDB)

	r := WriteRoute(WriteNotification(testMapper, mockDb))
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatal("Expected a 400 response!")
	}

	mockDb.AssertNotCalled(t, "Open")
}

func TestInvalidUUIDInPath(t *testing.T) {
	req := httptest.NewRequest("PUT", "http://our.host.name/lists/notifications/uuid", strings.NewReader(`{"uuid":"cee15258-6762-4fc5-8f57-0b5ca4c3aa20"}`))

	w := httptest.NewRecorder()

	mockDb := new(MockDB)

	r := WriteRoute(WriteNotification(testMapper, mockDb))
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatal("Expected a 400 response!")
	}

	mockDb.AssertNotCalled(t, "Open")
}

func TestFailedDatabaseOnWrite(t *testing.T) {
	req := httptest.NewRequest("PUT", "http://our.host.name/lists/notifications/ef863741-709a-4062-a8f1-987c44db1db5", strings.NewReader(mockWriteBody))

	w := httptest.NewRecorder()

	mockDb := new(MockDB)
	mockDb.On("Open").Return(nil, errors.New("No writes for u"))

	r := WriteRoute(WriteNotification(testMapper, mockDb))
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Fatal("Expected a 500 response!")
	}

	mockDb.AssertExpectations(t)
}