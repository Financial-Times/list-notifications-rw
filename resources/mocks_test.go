package resources

import (
	"net/http"
	"time"

	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
)

var testMapper = mapping.DefaultMapper{ApiHost: "testing-123.com"}
var testLinkGenerator = mapping.OffsetNextLink{ApiHost: "testing-123.com", CacheDelay: 10, MaxLimit: 200}

func WriteRoute(handler func(w http.ResponseWriter, r *http.Request)) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/lists/notifications/{uuid}", handler)
	return r
}

type MockClient struct {
	mock.Mock
}

func (m *MockClient) GetLimit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockClient) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClient) ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error) {
	args := m.Called(offset, since)
	notifications := args.Get(0)
	if notifications == nil {
		return nil, args.Error(1)
	}

	return notifications.(*[]model.InternalNotification), args.Error(1)
}

func (m *MockClient) FindNotificationByTransactionID(transactionID string) (model.InternalNotification, error) {
	args := m.Called(transactionID)
	notifications := args.Get(0)
	if notifications == nil {
		return model.InternalNotification{}, args.Error(1)
	}

	return notifications.(model.InternalNotification), args.Error(1)
}

func (m *MockClient) FindNotificationByPartialTransactionID(transactionID string) (model.InternalNotification, error) {
	args := m.Called(transactionID)
	notifications := args.Get(0)
	if notifications == nil {
		return model.InternalNotification{}, args.Error(1)
	}

	return notifications.(model.InternalNotification), args.Error(1)
}

func (m *MockClient) WriteNotification(notification *model.InternalNotification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockClient) EnsureIndexes() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
