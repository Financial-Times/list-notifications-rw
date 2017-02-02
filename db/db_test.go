package db

import (
	"testing"
	"time"

	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/stretchr/testify/assert"
)

func TestOpenPingAndConfig(t *testing.T) {
	mongo := startMongo(t, 200)
	defer mongo.Close()

	tx, err := mongo.Open()

	assert.NoError(t, err, "We should not error opening a connection to a fresh mongo instance!")
	assert.NotNil(t, mongo.session, "Sesssion should not be nil!")
	assert.NoError(t, tx.Ping(), "We should not error pinging mongo!")

	assert.Equal(t, 200, mongo.Limit(), "Limit should be set to 200")

	tx.Close()
	assert.Panics(t, func() {
		tx.(*MongoTX).session.Ping()
	}, "Ping should panic since the session is closed")

	mongo.Close()
	assert.Panics(t, func() {
		mongo.session.Ping()
	}, "Ping should panic since the session is closed")
}

func TestEnsureIndexes(t *testing.T) {
	mongo := startMongo(t, 200)
	defer mongo.Close()

	tx, err := mongo.Open()

	assert.NoError(t, err, "Should not error opening a connection to a fresh mongo instance!")
	assert.NotNil(t, mongo.session, "Sesssion should not be nil!")

	err = tx.EnsureIndices()
	assert.NoError(t, err, "Should not error creating indices!")

	indices, err := mongo.session.DB("upp-store").C("list-notifications").Indexes()
	assert.NoError(t, err, "Should not throw error getting indices")

	count := 0
	for _, index := range indices {
		switch index.Name {
		case "uuid-index":
			assert.Equal(t, []string{"uuid"}, index.Key, "There should be a uuid index!")
			count++
		case "publish-reference-index":
			assert.Equal(t, []string{"publishReference"}, index.Key, "There should be a publishReference index!")
			count++
		case "last-modified-index":
			assert.Equal(t, []string{"-lastModified"}, index.Key, "There should be a descending lastModified index!")
			count++
		}
	}

	assert.Equal(t, 3, count, "We should have checked 3 indices!")
}

func TestReadWriteFind(t *testing.T) {
	mongo := startMongo(t, 200)
	defer mongo.Close()

	tx, err := mongo.Open()

	assert.NoError(t, err, "Should not error opening a connection to a fresh mongo instance!")
	assert.NotNil(t, mongo.session, "Sesssion should not be nil!")

	notification := model.InternalNotification{
		Title:            "The Art of the Deal: Donald Z Trump",
		UUID:             "my-new-uuid",
		PublishReference: "tid_faketxid",
		LastModified:     time.Date(2017, 02, 02, 12, 51, 0, 0, time.Local),
		EventType:        "http://www.ft.com/thing/ThingChangeType/UPDATE",
	}
	tx.WriteNotification(&notification)

	notifications, err := tx.ReadNotifications(0, time.Date(2017, 02, 02, 12, 50, 0, 0, time.Local))
	assert.NoError(t, err, "Should not error")
	assert.NotNil(t, notifications, "Should not be nil")

	assert.Len(t, *notifications, 1, "Should be one notification")
	assert.Equal(t, (*notifications)[0].Title, "The Art of the Deal: Donald Z Trump", "Should be DJTs book")
	assert.Equal(t, (*notifications)[0].UUID, "my-new-uuid", "UUID should match")
	assert.Equal(t, (*notifications)[0].PublishReference, "tid_faketxid", "TXID should match")
	assert.Equal(t, (*notifications)[0].EventType, "http://www.ft.com/thing/ThingChangeType/UPDATE", "EventType should match")
	assert.Equal(t, (*notifications)[0].LastModified, time.Date(2017, 02, 02, 12, 51, 0, 0, time.Local), "Time should match")

	notifications, err = tx.FindNotification("tid_faketxid")
	assert.NoError(t, err, "Should not error")
	assert.NotNil(t, notifications, "Should not be nil")

	assert.Len(t, *notifications, 1, "Should be one notification")
	assert.Equal(t, (*notifications)[0].PublishReference, "tid_faketxid", "TXID should match")
}
