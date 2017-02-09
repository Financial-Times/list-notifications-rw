package db

import (
	"time"

	"github.com/Financial-Times/list-notifications-rw/model"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

var maxLimit int
var cacheDelay int

var expectedConnections = 1
var connections int

// Open opens a new session to Mongo
func (db *MongoDB) Open() (TX, error) {
	if db.session == nil {
		session, err := mgo.DialWithTimeout(db.Urls, time.Duration(db.Timeout)*time.Millisecond)
		if err != nil {
			return nil, err
		}
		db.session = session
		connections++

		maxLimit = db.MaxLimit
		cacheDelay = db.CacheDelay

		if connections > expectedConnections {
			logrus.Warnf("There are more MongoDB connections opened than expected! Are you sure this is what you want? Open connections: %v, expected %v.", connections, expectedConnections)
		}
	}

	return &MongoTX{db.session.Copy()}, nil
}

func (tx *MongoTX) EnsureIndices() error {
	collection := tx.session.DB("upp-store").C("list-notifications")
	lastModifiedIndex := mgo.Index{
		Name: "last-modified-index",
		Key:  []string{"-lastModified"},
	}
	err := collection.EnsureIndex(lastModifiedIndex)

	if err != nil {
		return err
	}

	publishReferenceIndex := mgo.Index{
		Name: "publish-reference-index",
		Key:  []string{"publishReference"},
	}
	err = collection.EnsureIndex(publishReferenceIndex)

	if err != nil {
		return err
	}

	uuidIndex := mgo.Index{
		Name: "uuid-index",
		Key:  []string{"uuid"},
	}
	return collection.EnsureIndex(uuidIndex)
}

// Close closes the entire database connection
func (db *MongoDB) Close() {
	db.session.Close()
}

// Limit returns the max number of records to return
func (db *MongoDB) Limit() int {
	return db.MaxLimit
}

// Ping returns a mongo ping response
func (tx *MongoTX) Ping() error {
	return tx.session.Ping()
}

// Close closes the transaction
func (tx *MongoTX) Close() {
	tx.session.Close()
}

// WriteNotification inserts a notification into mongo
func (tx *MongoTX) WriteNotification(notification *model.InternalNotification) {
	collection := tx.session.DB("upp-store").C("list-notifications")
	collection.Insert(notification)
}

// ReadNotifications reads notifications from the collection.
func (tx *MongoTX) ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error) {
	collection := tx.session.DB("upp-store").C("list-notifications")

	query := generateQuery(offset, since)
	pipe := collection.Pipe(query)

	results := []model.InternalNotification{}

	err := pipe.AllowDiskUse().All(&results)

	if err != nil {
		return nil, err
	}

	return &results, nil
}

// FindNotification locates one instance of a notification with the given Transaction ID (publishReference)
func (tx *MongoTX) FindNotification(txid string) (*[]model.InternalNotification, bool, error) {
	collection := tx.session.DB("upp-store").C("list-notifications")

	query := findByTxId(txid)

	pipe := collection.Find(query)
	result := []model.InternalNotification{}

	err := pipe.Limit(1).All(&result)

	if err != nil {
		return nil, false, err
	}

	if len(result) == 0 {
		return &result, false, nil
	}

	return &result, true, nil
}

// FindNotificationByPartialTransactionID locates one instance of a notification with the given Transaction ID (publishReference)
func (tx *MongoTX) FindNotificationByPartialTransactionID(txid string) (*[]model.InternalNotification, bool, error) {
	collection := tx.session.DB("upp-store").C("list-notifications")

	query := findByPartialTxId(txid)

	pipe := collection.Find(query)
	result := []model.InternalNotification{}

	err := pipe.Limit(1).All(&result)

	if err != nil {
		return nil, false, err
	}

	if len(result) == 0 {
		return &result, false, nil
	}

	return &result, true, nil
}

// DB contains database functions
type DB interface {
	Open() (TX, error)
	Close()
	Limit() int // bit hacky, but limit is exposed to resources here
}

// TX contains database transaction function
type TX interface {
	WriteNotification(notification *model.InternalNotification)
	ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error)
	FindNotification(txid string) (*[]model.InternalNotification, bool, error)
	FindNotificationByPartialTransactionID(txid string) (*[]model.InternalNotification, bool, error)
	EnsureIndices() error
	Ping() error
	Close()
}

// MongoTX wraps a mongo session
type MongoTX struct {
	session *mgo.Session
}

// MongoDB wraps a mango mongo session
type MongoDB struct {
	Urls       string
	Timeout    int
	MaxLimit   int
	CacheDelay int
	session    *mgo.Session
}
