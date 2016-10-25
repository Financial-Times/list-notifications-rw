package db

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func generateQuery(since time.Time, collection *mgo.Collection) *mgo.Query {
	match := bson.M{
		"changeDate": bson.M {
			"$gt": since,
		},
	}

	query := collection.Find(match)

	query.Sort("changeDate", "_id")
	query.Limit(50)
	query.Skip(0)

	return query
}