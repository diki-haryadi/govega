package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/diki-haryadi/govega/constant"
	"github.com/diki-haryadi/govega/database"
	"github.com/diki-haryadi/govega/docstore"
	"github.com/diki-haryadi/govega/util"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"go.opentelemetry.io/otel"
)

type SQLStore struct {
	db      *sqlx.DB
	idField string
	table   string
	driver  string
}

type QueryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
}

func init() {
	docstore.RegisterDriver(database.DriverMySQL, SQLStoreFactory)
	docstore.RegisterDriver(database.DriverPostgres, SQLStoreFactory)
}

func SQLStoreFactory(config *docstore.Config) (docstore.Driver, error) {
	return NewSQLStore(config)
}

func NewSQLStore(config *docstore.Config) (*SQLStore, error) {
	switch con := config.Connection.(type) {
	case *sqlx.DB:
		return NewSQLstore(con, config.IDField, config.Collection, config.Driver), nil
	case database.DBConfig:
		db := database.New(con, config.Driver)
		return NewSQLstore(db.Master, config.IDField, config.Collection, config.Driver), nil
	case *database.DBConfig:
		db := database.New(*con, config.Driver)
		return NewSQLstore(db.Master, config.IDField, config.Collection, config.Driver), nil
	case map[string]interface{}:
		var dc database.DBConfig
		if err := util.DecodeJSON(con, &dc); err != nil {
			return nil, err
		}
		db := database.New(dc, config.Driver)
		return NewSQLstore(db.Master, config.IDField, config.Collection, config.Driver), nil
	default:
		return nil, errors.New("[docstore/sql] unsupported connection type")
	}
}

func NewSQLstore(db *sqlx.DB, idField, table, driver string) *SQLStore {
	db.Mapper = reflectx.NewMapper("json")
	return &SQLStore{
		db:      db,
		idField: idField,
		table:   table,
		driver:  driver,
	}
}

func (s *SQLStore) getID(doc interface{}) (interface{}, error) {
	idf, err := util.FindFieldByTag(doc, "json", s.idField)
	if err != nil {
		return nil, err
	}

	id, ok := util.Lookup(idf, doc)

	if !ok {
		return nil, errors.New("[docstore/sql] missing document ID")
	}

	return id, nil
}

func (s *SQLStore) Create(ctx context.Context, doc interface{}) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.create")
	defer span.End()
	id, err := s.getID(doc)
	if err != nil {
		return err
	}

	gs := s.buildGetQuery(id)
	ok, err := s.recordExists(gs)
	if err != nil {
		return err
	}

	if ok {
		return errors.New("[docstore/sql] document ID is already exist")
	}

	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, d); err != nil {
		return err
	}

	for _, v := range d {
		if util.IsTime(v) {
			continue
		}
		if util.IsStructOrPointerOf(v) {
			return errors.New("[docstore/sql] unsupported data type")
		}
	}

	ex, err := getExecutor(ctx, s.db)
	if err != nil {
		return err
	}
	stmt := s.buildInsertQuery(d)

	if _, err = ex.ExecContext(ctx, stmt); err != nil {
		return err
	}

	return nil
}

func (s *SQLStore) Update(ctx context.Context, id, doc interface{}, replace bool) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.update")
	defer span.End()
	ex, err := getExecutor(ctx, s.db)
	if err != nil {
		return err
	}

	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, d); err != nil {
		return err
	}

	if replace {
		ds := s.buildDeleteQuery(id)
		if _, err := ex.ExecContext(ctx, ds); err != nil {
			return err
		}

		cs := s.buildInsertQuery(d)
		res, err := ex.ExecContext(ctx, cs)
		if c, _ := res.RowsAffected(); c == 0 {
			return docstore.NotFound
		}
		return err
	}

	us := s.buildUpdateQuery(d, id)
	res, err := ex.ExecContext(ctx, us)
	if c, _ := res.RowsAffected(); c == 0 {
		return docstore.NotFound
	}
	return err
}

func (s *SQLStore) UpdateField(ctx context.Context, id interface{}, fields []docstore.Field) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.update_field")
	defer span.End()
	ex, err := getExecutor(ctx, s.db)
	if err != nil {
		return err
	}

	d := make(map[string]interface{})
	for _, f := range fields {
		d[f.Name] = f.Value
	}

	us := s.buildUpdateQuery(d, id)
	res, err := ex.ExecContext(ctx, us)
	if c, _ := res.RowsAffected(); c == 0 {
		return docstore.NotFound
	}
	return err
}

func (s *SQLStore) Increment(ctx context.Context, id interface{}, key string, value int) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.increment")
	defer span.End()
	ex, err := getExecutor(ctx, s.db)
	if err != nil {
		return err
	}

	is := s.buildIncrQuery(id, key, value)
	res, err := ex.ExecContext(ctx, is)
	if c, _ := res.RowsAffected(); c == 0 {
		return docstore.NotFound
	}
	return err
}

func (s *SQLStore) GetIncrement(ctx context.Context, id interface{}, key string, value int, doc interface{}) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.getincrement")
	defer span.End()
	ex, err := getExecutor(ctx, s.db)
	if err != nil {
		return err
	}

	is := s.buildGetIncrQuery(id, key, value)
	if is == "" {
		return errors.New("this operation is not supported by driver")
	}
	if err := ex.GetContext(ctx, doc, is); err != nil {
		if err == sql.ErrNoRows {
			return docstore.NotFound
		}
		return err
	}
	return nil
}

func (s *SQLStore) Delete(ctx context.Context, id interface{}) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.delete")
	defer span.End()
	ex, err := getExecutor(ctx, s.db)
	if err != nil {
		return err
	}
	ds := s.buildDeleteQuery(id)
	_, err = ex.ExecContext(ctx, ds)
	return err
}

func (s *SQLStore) Get(ctx context.Context, id interface{}, doc interface{}) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.get")
	defer span.End()
	ex, err := getExecutor(ctx, s.db)
	if err != nil {
		return err
	}

	gs := s.buildGetQuery(id)

	if err := ex.GetContext(ctx, doc, gs); err != nil {
		if err == sql.ErrNoRows {
			return docstore.NotFound
		}
		return err
	}
	return nil

}

func (s *SQLStore) Find(ctx context.Context, query *docstore.QueryOpt, docs interface{}) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.find")
	defer span.End()
	fs := s.buildFindQuery(query)
	return s.db.SelectContext(ctx, docs, fs)
}

func (s *SQLStore) BulkCreate(ctx context.Context, docs []interface{}) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.bulk_create")
	defer span.End()
	ins := make([]map[string]interface{}, 0)

	for _, doc := range docs {
		d := make(map[string]interface{})
		if err := util.DecodeJSON(doc, d); err != nil {
			return err
		}
		for _, v := range d {
			if util.IsTime(v) {
				continue
			}
			if util.IsStructOrPointerOf(v) {
				return errors.New("[docstore/sql] unsupported data type")
			}
		}
		ins = append(ins, d)
	}

	ex, err := getExecutor(ctx, s.db)
	if err != nil {
		return err
	}
	stmt := s.buildBulkInsertQuery(ins)

	if _, err = ex.ExecContext(ctx, stmt); err != nil {
		return err
	}

	return nil
}

func (s *SQLStore) BulkGet(ctx context.Context, ids []interface{}, docs interface{}) error {
	tr := otel.Tracer("docstore/sql")
	ctx, span := tr.Start(ctx, "docstore.bulk_get")
	defer span.End()
	ex, err := getExecutor(ctx, s.db)
	if err != nil {
		return err
	}

	gs := s.buildBulkGetQuery(ids)

	return ex.SelectContext(ctx, docs, gs)
}

func (s *SQLStore) recordExists(query string) (bool, error) {
	var exists bool
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err := s.db.QueryRow(query).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

func (s *SQLStore) Migrate(ctx context.Context, config interface{}) error {
	schema, ok := config.(string)
	if !ok {
		return errors.New("[docstore/sql] database schema should be string")
	}

	s.db.MustExec(fmt.Sprintf(`DROP TABLE IF EXISTS %s;`, s.table))
	s.db.MustExec(schema)
	return nil
}

func getExecutor(ctx context.Context, db *sqlx.DB) (QueryExecutor, error) {
	tx, ok := ctx.Value(constant.TxKey).(*sqlx.Tx)
	if ok {
		return tx, nil
	}
	return db, nil
}
