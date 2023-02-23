package db

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func TestShiftSince(t *testing.T) {
	now := time.Now()

	shifted := shiftSince(42, now)
	assert.Equal(t, now.Add(-42*time.Second), shifted, "Should've been shifted by 42s!")
}

func TestCalculateTillDate(t *testing.T) {
	now := time.Now().UTC()

	till := calculateTill(48, now)
	assert.Equal(t, now.Add(-48*time.Second), till, "Should've been shifted by 48s!")
}

func TestGetMatchNoOffset(t *testing.T) {
	since := time.Now().UTC()

	match := getMatch(10, 0, since)

	data, err := json.Marshal(match)
	assert.NoError(t, err)
	regex := regexp.MustCompile(`\{"\$match":\{"lastModified":\{"\$gt":".*","\$lt":".*"}}}`)
	assert.True(t, regex.MatchString(string(data)), "Query json should match!")
}

func TestGetMatchWithOffset(t *testing.T) {
	since, err := time.Parse(time.RFC3339Nano, "2016-10-26T16:15:09.46Z")
	assert.NoError(t, err)

	match := getMatch(10, 60, since)

	data, err := json.Marshal(match)
	assert.NoError(t, err)
	regex := regexp.MustCompile(`\{"\$match":\{"lastModified":\{"\$gte":".*","\$lte":".*"}}}`)
	assert.True(t, regex.MatchString(string(data)), "Query json should match!")
}

func TestQuery(t *testing.T) {
	since, err := time.Parse(time.RFC3339Nano, "2016-10-26T16:15:09.46Z")
	assert.NoError(t, err)
	log := logger.NewUPPLogger("test", "debug")

	query := generateQuery(10, 50, 102, since)

	regex := regexp.MustCompile(`\[\{"\$match":\{"lastModified":\{"\$gte":".*","\$lte":".*"}}},\{"\$sort":\{"lastModified":-1}},\{"\$group":\{"_id":"\$uuid","eventType":\{"\$first":"\$eventType"},"lastModified":\{"\$first":"\$lastModified"},"publishReference":\{"\$first":"\$publishReference"},"title":\{"\$first":"\$title"},"uuid":\{"\$first":"\$uuid"}}},\{"\$sort":\{"lastModified":1,"uuid":1}},\{"\$skip":50},\{"\$limit":103}]`)
	data, err := json.Marshal(query)
	assert.NoError(t, err)
	assert.True(t, regex.MatchString(string(data)), "Query json should match!")
}

func TestFindNotificationQuery(t *testing.T) {
	query := findByTransactionID("tid_i-am-a-tid")

	data, err := json.Marshal(query)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `{"publishReference":"tid_i-am-a-tid"}`, "Query json should match!")
}

func TestFindNotificationQueryByPartialTXID(t *testing.T) {
	query := findByPartialTransactionID("tid_i-am-a-tid")

	data, err := json.Marshal(query)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `{"publishReference":{"$regex":"^tid_i-am-a-tid"}}`)
}
