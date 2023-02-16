package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// integration tests runs with local mongo instance
func NewMockClient(address, _, _, database, collection string, cacheDelay, maxLimit int, log *logger.UPPLogger) (*Client, error) {
	address = fmt.Sprintf("mongodb://%s", address)
	opts := options.Client().ApplyURI(address)

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:     client,
		database:   database,
		collection: collection,
		cacheDelay: cacheDelay,
		maxLimit:   maxLimit,
		log:        log,
	}, nil
}

func TestOpenPingAndConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Database integration for long tests only.")
	}

	mongoURL := os.Getenv("MONGO_TEST_URL")
	if strings.TrimSpace(mongoURL) == "" {
		t.Fatal("Please set the environment variable MONGO_TEST_URL to run mongo integration tests (e.g. MONGO_TEST_URL=localhost:27017). Alternatively, run `go test -short` to skip them.")
	}

	database := "upp-store"
	collection := "testing"
	username := "testUser"
	password := "testPass"
	cacheDelay := 10
	maxLimit := 200

	log := logger.NewUPPLogger("test", "PANIC")

	client, err := NewMockClient(mongoURL, username, password, database, collection, cacheDelay, maxLimit, log)
	require.NoError(t, err)

	assert.NoError(t, client.Ping(), "We should not error pinging mongo!")

	assert.Equal(t, maxLimit, client.GetLimit(), "Limit should be set to %s", maxLimit)
}

func TestReadWriteFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Database integration for long tests only.")
	}

	mongoURL := os.Getenv("MONGO_TEST_URL")
	if strings.TrimSpace(mongoURL) == "" {
		t.Fatal("Please set the environment variable MONGO_TEST_URL to run mongo integration tests (e.g. MONGO_TEST_URL=localhost:27017). Alternatively, run `go test -short` to skip them.")
	}

	exampleTime := time.Date(2017, 02, 02, 12, 51, 0, 0, time.UTC)
	database := "upp-store"
	collection := "testing"
	username := "testUser"
	password := "testPass"
	cacheDelay := 10
	maxLimit := 200

	log := logger.NewUPPLogger("test", "PANIC")

	client, err := NewMockClient(mongoURL, username, password, database, collection, cacheDelay, maxLimit, log)
	require.NoError(t, err)

	notification := model.InternalNotification{
		Title:            "The Art of the Deal: Donald Z Trump",
		UUID:             "my-new-uuid",
		PublishReference: "tid_faketxid",
		LastModified:     exampleTime,
		EventType:        "http://www.ft.com/thing/ThingChangeType/UPDATE",
	}
	require.NoError(t, client.WriteNotification(&notification))

	notifications, err := client.ReadNotifications(0, exampleTime)
	require.NoError(t, err, "Should not error")
	assert.NotNil(t, notifications, "Should not be nil")

	require.Len(t, *notifications, 1, "Should be one notification")
	assert.Equal(t, (*notifications)[0].Title, "The Art of the Deal: Donald Z Trump", "Should be DJTs book")
	assert.Equal(t, (*notifications)[0].UUID, "my-new-uuid", "UUID should match")
	assert.Equal(t, (*notifications)[0].PublishReference, "tid_faketxid", "TXID should match")
	assert.Equal(t, (*notifications)[0].EventType, "http://www.ft.com/thing/ThingChangeType/UPDATE", "EventType should match")
	assert.Equal(t, (*notifications)[0].LastModified, exampleTime, "Time should match")

	notification, err = client.FindNotificationByTransactionID("tid_faketxid")
	require.NoError(t, err, "Should not error")
	assert.NotNil(t, notification.UUID != "", "Should not be empty string")
	assert.Equal(t, notification.PublishReference, "tid_faketxid", "Transaction ID should match")

	notification, err = client.FindNotificationByPartialTransactionID("tid_fake")
	require.NoError(t, err, "Should not error")
	assert.NotNil(t, notification.UUID != "", "Should not be empty string")
	assert.Equal(t, notification.PublishReference, "tid_faketxid", "Transaction ID should match")
}

func TestNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Database integration for long tests only.")
	}

	mongoURL := os.Getenv("MONGO_TEST_URL")
	if strings.TrimSpace(mongoURL) == "" {
		t.Fatal("Please set the environment variable MONGO_TEST_URL to run mongo integration tests (e.g. MONGO_TEST_URL=localhost:27017). Alternatively, run `go test -short` to skip them.")
	}

	database := "upp-store"
	collection := "testing"
	username := "testUser"
	password := "testPass"
	cacheDelay := 10
	maxLimit := 200

	log := logger.NewUPPLogger("test", "PANIC")

	client, err := NewMockClient(mongoURL, username, password, database, collection, cacheDelay, maxLimit, log)
	require.NoError(t, err)

	_, err = client.FindNotificationByTransactionID("tid_i-dont-exist")
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)

	_, err = client.FindNotificationByPartialTransactionID("tid_i-dont-exist")
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
}
