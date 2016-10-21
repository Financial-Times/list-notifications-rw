package db

import (
	"gopkg.in/mgo.v2"
)

type DB interface {
	Open(url string)
}

type TX interface {
	Close()
}

type MongoTX struct {
	session *mgo.Session
}

type MongoDB struct {
	urls string
}

func (db MongoDB) Open() (TX, error) {
	session, err := mgo.Dial(db.urls)
	if err != nil {
		return nil, err
	}

	return &MongoTX{session}, nil
}

func (tx MongoTX) Close(){
	tx.session.Close()
}