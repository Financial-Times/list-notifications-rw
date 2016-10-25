package db

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func generatePipe(since time.Time, collection *mgo.Collection) *mgo.Pipe {
	till := calculateTill()

	match := bson.M{
		"$match": bson.M{
			"lastModified": bson.M {
				"$gt": since,
				"$lt": till,
			},
		},
	}

	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"uuid": "$uuid",
			},
			"lastModified": bson.M{
				"$max": "$lastModified",
			},
			"uuid": bson.M{
				"$first": "$uuid",
			},
			"title": bson.M{
				"$first": "$title",
			},
			"eventType": bson.M{
				"$first": "$eventType",
			},
			"publishReference": bson.M{
				"$first": "$publishReference",
			},
		},
	}

	pipeline := []bson.M{
		match,
		{
			"$sort": bson.M{
				"lastModified": -1,
				"uuid": 1,
			},
		},
		group,
		{
			"$sort": bson.M{
				"lastModified": 1,
				"_id": 1,
			},
		},
		{"$skip": 0},
		{"$limit": 50},
	}

	pipe := collection.Pipe(pipeline)
	return pipe
}

func calculateTill() time.Time {
	return time.Now().Add(-10 * time.Second)
}