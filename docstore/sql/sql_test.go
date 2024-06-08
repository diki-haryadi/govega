package sql

import (
	"context"
	"testing"
	"time"

	_ "github.com/diki-haryadi/govega/cache/mem"
	"github.com/diki-haryadi/govega/database"
	"github.com/diki-haryadi/govega/docstore"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initDB() *sqlx.DB {
	dbconf := database.DBConfig{
		MasterDSN:     "root:password@(localhost:3306)/outbox?parseTime=true",
		SlaveDSN:      "root:password@(localhost:3306)/outbox?parseTime=true",
		RetryInterval: 5,
		MaxIdleConn:   10,
		MaxConn:       5,
	}

	db := database.New(dbconf, "mysql")

	return db.Master
}

func initStore() *SQLStore {
	db := initDB()
	var schema = `
	CREATE TABLE IF NOT EXISTS user (
		id VARCHAR(255) NOT NULL,
		name VARCHAR(255),
		username VARCHAR(255),
		age INT,
		created_at DATETIME
	)  ENGINE=INNODB;`

	//db.Master.MustExec(`DROP TABLE IF EXISTS user;`)
	//db.Master.MustExec(schema)
	store := NewSQLstore(db, "id", "user", "mysql")
	store.Migrate(context.Background(), schema)
	return store
}

func TestSQLStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store := initStore()
	docstore.DriverCRUDTest(store, t)
	docstore.DriverBulkTest(store, t)
}

func TestIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store := initStore()
	type User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Username  string    `json:"username"`
		Age       int       `json:"age"`
		CreatedAt time.Time `json:"created_at"`
	}

	ctx := context.Background()
	ts := time.Now()
	usr := &User{
		ID:        "1234",
		Name:      "sahal",
		Username:  "sahalzain",
		Age:       35,
		CreatedAt: ts,
	}

	require.Nil(t, store.Create(ctx, usr))

	var doc User
	require.Nil(t, store.Get(ctx, usr.ID, &doc))
	assert.Equal(t, usr.ID, doc.ID)
	assert.Equal(t, usr.Name, doc.Name)
	assert.Equal(t, usr.Username, doc.Username)
	assert.Equal(t, usr.Age, doc.Age)
	assert.Equal(t, usr.CreatedAt.Unix(), doc.CreatedAt.Unix())

	require.Nil(t, store.Increment(ctx, usr.ID, "age", 2))

	var user User
	require.Nil(t, store.Get(ctx, usr.ID, &user))
	assert.Equal(t, 37, user.Age)
}

func TestDocstore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	docstore.RegisterDriver(database.DriverMySQL, SQLStoreFactory)
	db := initDB()
	var schema = `
	CREATE TABLE IF NOT EXISTS user (
		id VARCHAR(255) NOT NULL,
		name VARCHAR(255),
		username VARCHAR(255),
		age INT,
		created_at DATETIME
	)  ENGINE=INNODB;`

	config := &docstore.Config{
		Database:   "outbox",
		Collection: "user",
		IDField:    "id",
		Driver:     "mysql",
		Connection: db,
		CacheURL:   "mem://ms",
	}

	cs, err := docstore.New(config)
	require.Nil(t, err)
	require.NotNil(t, cs)
	cs.Migrate(context.Background(), schema)

	docstore.DocstoreTestCRUD(cs, t)
}
