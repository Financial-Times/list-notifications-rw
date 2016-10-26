package resources

import (
	"github.com/stretchr/testify/mock"
	"github.com/Financial-Times/list-notifications-rw/db"
	"time"
	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/gorilla/mux"
	"net/http"
)

var testMapper = mapping.DefaultMapper{ApiHost: "testing-123.com"}
var testLinkGenerator = mapping.OffsetNextLink{ApiHost: "testing-123.com", CacheDelay: 10, MaxLimit: 200}

func WriteRoute(handler func(w http.ResponseWriter, r *http.Request)) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/lists/notifications/{uuid}", handler)
	return r
}

type MockDB struct {
	mock.Mock
}

type MockTX struct {
	mock.Mock
}

func (m MockDB) Open() (db.TX, error) {
	args := m.Called()
	tx := args.Get(0)
	if tx == nil {
		return nil, args.Error(1)
	}

	return tx.(db.TX), args.Error(1)
}

func (m MockDB) Limit() int {
	args := m.Called()
	return args.Int(0)
}

func (m MockTX) ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error) {
	args := m.Called(offset, since)
	notifications := args.Get(0)
	if notifications == nil {
		return nil, args.Error(1)
	}

	return notifications.(*[]model.InternalNotification), args.Error(1)
}

func (m MockTX) WriteNotification(notification *model.InternalNotification){
	m.Called(notification)
}

func (m MockTX) Close() {
	m.Called()
}