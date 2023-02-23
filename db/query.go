package db

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func findByTransactionID(transactionID string) bson.M {
	return bson.M{"publishReference": transactionID}
}

func findByPartialTransactionID(transactionID string) bson.M {
	return bson.M{"publishReference": bson.M{"$regex": "^" + transactionID}}
}

func generateQuery(delay, offset, maxLimit int, since time.Time) []bson.M {
	match := getMatch(delay, offset, since)

	pipeline := []bson.M{
		match, // get all records that exist between the start and end dates
		{
			"$sort": bson.M{
				"lastModified": -1,
			},
		}, // sort most recent notifications first
		{
			"$group": bson.M{
				"_id": "$uuid", // group all notifications together by uuid...
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
				"lastModified": bson.M{
					"$first": "$lastModified",
				},
			},
		}, //... and create one notification based on the most recent fields (the "first" notification's fields)
		{
			"$sort": bson.M{
				"lastModified": 1,
				"uuid":         1,
			},
		}, // sort by oldest first, and to ensure strict ordering, also sort by uuid when lastModified dates match
		{"$skip": offset},
		{"$limit": maxLimit + 1},
	}

	j, err := json.Marshal(pipeline)
	if err == nil { // Use /__log/debug endpoint to see the full query.
		log.WithField("query", string(j)).Debug("Full query.")
	}

	return pipeline
}

func getMatch(delay, offset int, since time.Time) bson.M {
	shifted := shiftSince(delay, since)
	till := calculateTill(delay, time.Now().UTC())

	if offset > 0 {
		return bson.M{
			"$match": bson.M{
				"lastModified": bson.M{
					"$gte": shifted,
					"$lte": till,
				},
			},
		}
	}

	return bson.M{
		"$match": bson.M{
			"lastModified": bson.M{
				"$gt": shifted,
				"$lt": till,
			},
		},
	}
}

func shiftSince(cacheDelay int, since time.Time) time.Time {
	return since.Add(time.Duration(-1*cacheDelay) * time.Second)
}

func calculateTill(cacheDelay int, base time.Time) time.Time {
	return base.Add(time.Duration(-1*cacheDelay) * time.Second)
}
