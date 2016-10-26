package db

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func generatePipe(offset int, since time.Time, collection *mgo.Collection) *mgo.Pipe {
	match := getMatch(offset, since)

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
		{"$skip": offset},
		{"$limit": maxLimit + 1},
	}

	pipe := collection.Pipe(pipeline)
	return pipe
}

func getMatch(offset int, since time.Time) bson.M {
	shifted := shiftSince(since)
	till := calculateTill()

	if offset > 0 {
		return bson.M{
			"$match": bson.M{
				"lastModified": bson.M {
					"$gte": shifted,
					"$lte": till,
				},
			},
		}
	}

	return bson.M{
		"$match": bson.M{
			"lastModified": bson.M {
				"$gt": shifted,
				"$lt": till,
			},
		},
	}
}

func shiftSince(since time.Time) time.Time {
	return since.Add(time.Duration(-1 * cacheDelay) * time.Second)
}

func calculateTill() time.Time {
	return time.Now().UTC().Add(time.Duration(-1 * cacheDelay) * time.Second)
}