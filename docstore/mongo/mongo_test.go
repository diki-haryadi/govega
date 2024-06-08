package mongo

import (
	"context"
	"testing"

	_ "github.com/diki-haryadi/govega/cache/mem"
	"github.com/diki-haryadi/govega/database"
	"github.com/diki-haryadi/govega/docstore"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func initDriver(t *testing.T) *database.Database {
	client := &database.Client{
		URI:     "mongodb://localhost:27017",
		DB:      "test",
		AppName: "test",
	}

	ctx := context.Background()

	db := database.MongoConnectClient(client)
	require.NotNil(t, db)

	col := db.Database.Collection("docstore")
	col.DeleteMany(ctx, bson.D{})
	return db
}

func TestMongoStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := initDriver(t)

	ms, err := NewMongostore(db, "docstore", "id")
	ms.Migrate(context.Background(), nil)

	require.Nil(t, err)
	docstore.DriverCRUDTest(ms, t)
	docstore.DriverBulkTest(ms, t)
}

func TestDocstore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	docstore.RegisterDriver("mongo", MongoStoreFactory)

	db := initDriver(t)
	config := &docstore.Config{
		Database:   "test",
		Collection: "docstore",
		IDField:    "id",
		Driver:     "mongo",
		Connection: db,
		CacheURL:   "mem://ms",
	}

	cs, err := docstore.New(config)
	require.Nil(t, err)
	require.NotNil(t, cs)

	cs.Migrate(context.Background(), nil)

	docstore.DocstoreTestCRUD(cs, t)
}
