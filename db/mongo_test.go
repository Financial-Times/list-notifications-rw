package db

import (
	"os"
	"strings"
	"testing"
)

func startMongo(t *testing.T, limit int) *MongoDB {
	if testing.Short() {
		t.Skip("Mongo integration for long tests only.")
	}

	mongoURL := os.Getenv("MONGO_TEST_URL")
	if strings.TrimSpace(mongoURL) == "" {
		t.Fatal("Please set the environment variable MONGO_TEST_URL to run mongo integration tests (e.g. MONGO_TEST_URL=localhost:27017). Alternatively, run `go test -short` to skip them.")
	}

	return &MongoDB{
		Urls:       mongoURL,
		Timeout:    3800,
		MaxLimit:   limit,
		CacheDelay: 10,
	}
}
