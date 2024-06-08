package mongo

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/diki-haryadi/govega/database"
	"github.com/diki-haryadi/govega/event"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMongoSender(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	event.RegisterSender("mongo", NewMongoSender)

	mconf := database.Client{
		URI:     "mongodb://localhost:27017",
		DB:      "test",
		AppName: "event",
	}
	db := database.MongoConnectClient(&mconf)

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "mongo",
			Config: map[string]interface{}{
				"collection": "outbox",
				"connection": db,
			},
		},
	}

	ctx := context.Background()

	em, err := event.New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	key := fmt.Sprintf("%v", time.Now().Unix())

	require.Nil(t, em.Publish(ctx, "test", key, "testdata", nil))

	var out MongoOutbox
	err = db.Database.Collection("outbox").FindOne(ctx, bson.M{"key": key}).Decode(&out)
	require.Nil(t, err)
	assert.Equal(t, key, out.Key)
	db.Database.Collection("outbox").DeleteMany(ctx, bson.D{})
}

func TestMongoWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	event.RegisterWriter("mongo", NewMongoWriter)

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "logger",
		},
		Writer: &event.DriverConfig{
			Type: "mongo",
			Config: map[string]interface{}{
				"collection": "outbox",
				"connection": map[string]interface{}{
					"uri":      "mongodb://localhost:27017",
					"database": "test",
					"name":     "event",
				},
			},
		},
	}

	ctx := context.Background()

	em, err := event.New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	key := fmt.Sprintf("%v", time.Now().Unix())

	require.Nil(t, em.Publish(ctx, "test", key, "testdata", nil))

	time.Sleep(1 * time.Second)
	logStr := buf.String()

	assert.Contains(t, logStr, "key="+key)
	assert.Contains(t, logStr, "data=testdata")
	assert.Contains(t, logStr, "topic=test")

	mconf := database.Client{
		URI:     "mongodb://localhost:27017",
		DB:      "test",
		AppName: "event",
	}
	db := database.MongoConnectClient(&mconf)
	err = db.Database.Collection("outbox").FindOne(ctx, bson.M{"key": key}).Err()
	require.NotNil(t, err)
	assert.Equal(t, err, mongo.ErrNoDocuments)
	db.Database.Collection("outbox").DeleteMany(ctx, bson.D{})
}
