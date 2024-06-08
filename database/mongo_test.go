package database

import (
	"testing"
	"time"
)

func TestConnectMongoClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := Client{
		URI:            "<insertYourMongoConnectionURIHere>",
		DB:             "<chooseYourDB>",
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    20 * time.Second,
	}

	db := MongoConnectClient(&c)
	if db.Database.Name() == "" {
		t.Error("Database emtpy")
	} else {
		t.Log("OK", db.Database.Name())
	}

}
