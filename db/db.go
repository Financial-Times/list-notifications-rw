package db

import (
	"log"
	"time"

	"github.com/Financial-Times/list-notifications-rw/model"
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
			log.Println("WARNING: There are more MongoDB connections opened than expected! Are you sure this is what you want?")
		}
	}

	return &MongoTX{db.session.Copy()}, nil
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
