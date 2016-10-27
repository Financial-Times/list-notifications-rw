package db

import (
	"gopkg.in/mgo.v2"
	"github.com/Financial-Times/list-notifications-rw/model"
	"time"
)

var maxLimit = 200
var cacheDelay = 10

func (db MongoDB) Open() (TX, error) {
	if db.session == nil {
		session, err := mgo.DialWithTimeout(db.Urls, time.Duration(db.Timeout) * time.Second)
		if err != nil {
			return nil, err
		}
		db.session = session

		maxLimit = db.MaxLimit
		cacheDelay = db.CacheDelay
	}

	return &MongoTX{db.session.Copy()}, nil
}

func (db MongoDB) Limit() int {
	return db.MaxLimit
}

func (tx MongoTX) Ping() error {
	return tx.session.Ping()
}

func (tx MongoTX) Close(){
	tx.session.Close()
}

func (tx MongoTX) WriteNotification(notification *model.InternalNotification) {
	collection := tx.session.DB("upp-store").C("notifications")
	collection.Insert(notification)
}

func (tx MongoTX) ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error) {
	collection := tx.session.DB("upp-store").C("notifications")

	query := generateQuery(offset, since)
	pipe := collection.Pipe(query)

	results := []model.InternalNotification{}

	err := pipe.AllowDiskUse().All(&results)

	if err != nil {
		return nil, err
	}

	return &results, nil
}

type DB interface {
	Open() (TX, error)
	Limit() int // bit hacky, but limit is exposed to resources here
}

type TX interface {
	WriteNotification(notification *model.InternalNotification)
	ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error)
	Ping() error
	Close()
}

type MongoTX struct {
	session *mgo.Session
}

type MongoDB struct {
	Urls string
	Timeout int
	MaxLimit int
	CacheDelay int
	session *mgo.Session
}