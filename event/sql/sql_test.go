package sql

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/diki-haryadi/govega/constant"
	"github.com/diki-haryadi/govega/database"
	"github.com/diki-haryadi/govega/event"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTable() *sqlx.DB {
	dbc := database.DBConfig{
		MasterDSN:     "root:password@(localhost:3306)/outbox?parseTime=true",
		SlaveDSN:      "root:password@(localhost:3306)/outbox?parseTime=true",
		RetryInterval: 10,
		MaxIdleConn:   5,
		MaxConn:       10,
	}

	db := database.New(dbc, database.DriverMySQL)
	var schema = `
	CREATE TABLE IF NOT EXISTS outbox (
		id VARCHAR(255) NOT NULL,
		topic VARCHAR(255) NOT NULL,
		message_key VARCHAR(255),
		message_value TEXT,
		created_at TIMESTAMP
	)  ENGINE=INNODB;`

	db.Master.MustExec("DROP TABLE IF EXISTS outbox;")
	db.Master.MustExec(schema)
	return db.Master
}

func TestSQLSender(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := initTable()
	require.NotNil(t, db)
	event.RegisterSender("sql", NewSQLSender)

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "sql",
			Config: map[string]interface{}{
				"driver":     "mysql",
				"table":      "outbox",
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

	var out []SQLOutbox
	err = db.Select(&out, "SELECT * FROM outbox WHERE message_key = ?", key)
	require.Nil(t, err)
	assert.Equal(t, 1, len(out))
}

func TestSQLWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := initTable()
	require.NotNil(t, db)
	event.RegisterWriter("sql", NewSQLWriter)

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "logger",
		},
		Writer: &event.DriverConfig{
			Type: "sql",
			Config: map[string]interface{}{
				"driver": "mysql",
				"table":  "outbox",
				"connection": map[string]interface{}{
					"master_dsn":     "root:password@(localhost:3306)/outbox",
					"slave_dsn":      "root:password@(localhost:3306)/outbox",
					"retry_interval": 10,
					"max_idle":       5,
					"max_con":        10,
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

	var out []SQLOutbox

	time.Sleep(1 * time.Second)
	logStr := buf.String()

	assert.Contains(t, logStr, "key="+key)
	assert.Contains(t, logStr, "data=testdata")
	assert.Contains(t, logStr, "topic=test")

	err = db.Select(&out, "SELECT * FROM outbox WHERE message_key = ?", key)
	require.Nil(t, err)
	assert.Equal(t, 0, len(out))

}

func TestSQLTx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := initTable()
	require.NotNil(t, db)
	event.RegisterSender("sql", NewSQLSender)

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "sql",
			Config: map[string]interface{}{
				"driver":     "mysql",
				"table":      "outbox",
				"connection": db,
			},
		},
	}

	ctx := context.Background()

	em, err := event.New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)
	key := fmt.Sprintf("%v", time.Now().Unix())

	tx, err := db.Begin()
	require.Nil(t, err)
	require.NotNil(t, tx)

	ttx := context.WithValue(ctx, constant.TxKey, tx)

	require.Nil(t, em.Publish(ttx, "test", key, "testdata", nil))

	require.Nil(t, tx.Commit())

	var out []SQLOutbox
	err = db.Select(&out, "SELECT * FROM outbox WHERE message_key = ?", key)
	require.Nil(t, err)
	assert.Equal(t, 1, len(out))

}
