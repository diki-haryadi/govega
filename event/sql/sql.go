package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/diki-haryadi/govega/constant"
	"github.com/diki-haryadi/govega/database"
	"github.com/diki-haryadi/govega/event"
	"github.com/diki-haryadi/govega/util"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
)

type SQLSender struct {
	Driver     string      `json:"driver" mapstructure:"driver"`
	Connection interface{} `json:"connection" mapstructure:"connection"`
	Table      string      `json:"table" mapstructure:"table"`
	db         *sqlx.DB
}

type SQLOutbox struct {
	ID        string    `db:"id"`
	Topic     string    `db:"topic"`
	Key       string    `db:"message_key"`
	Value     string    `db:"message_value"`
	CreatedAt time.Time `db:"created_at"`
}

func init() {
	event.RegisterSender("sql", NewSQLSender)
	event.RegisterWriter("sql", NewSQLWriter)
}

func NewSQLSender(ctx context.Context, config interface{}) (event.Sender, error) {
	return NewSQLOutbox(ctx, config)
}

func NewSQLWriter(ctx context.Context, config interface{}) (event.Writer, error) {
	return NewSQLOutbox(ctx, config)
}

func NewSQLOutbox(ctx context.Context, config interface{}) (*SQLSender, error) {
	var ss SQLSender
	if err := mapstructure.Decode(config, &ss); err != nil {
		return nil, err
	}

	if ss.Connection == nil {
		return nil, errors.New("[event/sql] missing connection param")
	}

	if ss.Table == "" {
		return nil, errors.New("[event/sql] missing table param")
	}

	switch con := ss.Connection.(type) {
	case *sqlx.DB:
		ss.db = con
		return &ss, nil
	case *database.DBConfig:
		db := database.New(*con, ss.Driver)
		ss.db = db.Master
		return &ss, nil
	case database.DBConfig:
		db := database.New(con, ss.Driver)
		ss.db = db.Master
		return &ss, nil
	case map[string]interface{}:
		var conf database.DBConfig
		if err := util.DecodeJSON(con, &conf); err != nil {
			return nil, err
		}
		db := database.New(conf, ss.Driver)
		ss.db = db.Master
		return &ss, nil
	default:
		return nil, errors.New("[event/sql] unsupported connection type")
	}
}

func (s *SQLSender) Send(ctx context.Context, message *event.EventMessage) error {
	var tx *sql.Tx
	ownTx := false
	//ck := NewSQLTxContext(s.ContextKey)
	tx, ok := ctx.Value(constant.TxKey).(*sql.Tx)
	if !ok {
		t, err := s.db.Begin()
		if err != nil {
			return err
		}
		tx = t
		ownTx = true
	}

	outbox, err := event.OutboxFromMessage(message)
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf("INSERT INTO %s (id, topic, message_key, message_value, created_at) VALUES (?, ?, ?, ?, ?)", s.Table)

	_, err = tx.Exec(stmt, outbox.ID, outbox.Topic, outbox.Key, outbox.Value, outbox.CreatedAt)
	if err != nil {
		return err
	}

	if ownTx {
		return tx.Commit()
	}
	return nil
}

func (s *SQLSender) Delete(ctx context.Context, message *event.EventMessage) error {
	outbox, err := event.OutboxFromMessage(message)
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf("DELETE FROM %s WHERE id = ?", s.Table)

	_, err = s.db.Exec(stmt, outbox.ID)
	if err != nil {
		return err
	}
	return err
}
