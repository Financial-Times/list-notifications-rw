package db

import (
	"gopkg.in/mgo.v2"
	"github.com/Financial-Times/list-notifications-rw/model"
	"time"
)

type DB interface {
	Open() (TX, error)
}

type TX interface {
	WriteNotification(notification *model.InternalNotification)
	ReadNotifications(since time.Time) (*[]model.InternalNotification, error)
	Close()
}

type MongoTX struct {
	session *mgo.Session
}

type MongoDB struct {
	Urls string
	Timeout int
	session *mgo.Session
}

func (db MongoDB) Open() (TX, error) {
	if db.session == nil {
		session, err := mgo.DialWithTimeout(db.Urls, time.Duration(db.Timeout) * time.Second)
		if err != nil {
			return nil, err
		}
		db.session = session
	}

	return &MongoTX{db.session.Copy()}, nil
}

func (tx MongoTX) Close(){
	tx.session.Close()
}

func (tx MongoTX) WriteNotification(notification *model.InternalNotification) {
	collection := tx.session.DB("upp-store").C("notifications")
	collection.Insert(notification)
}

func (tx MongoTX) ReadNotifications(since time.Time) (*[]model.InternalNotification, error) {
	collection := tx.session.DB("upp-store").C("notifications")

	query := generateQuery(since, collection)
	find := collection.Find(query)

	results := []model.InternalNotification{}
	err := find.All(&results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}