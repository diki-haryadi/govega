package mongo

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/diki-haryadi/govega/database"
	"github.com/diki-haryadi/govega/docstore"
	"github.com/diki-haryadi/govega/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	store      *mongo.Collection
	idField    string
	collection string
}

func init() {
	docstore.RegisterDriver("mongo", MongoStoreFactory)
}

func MongoStoreFactory(config *docstore.Config) (docstore.Driver, error) {
	return NewMongoStore(config)
}

func NewMongoStore(config *docstore.Config) (*MongoStore, error) {
	switch con := config.Connection.(type) {
	case *database.Database:
		return NewMongostore(con, config.Collection, config.IDField)
	case *database.Client:
		db := database.MongoConnectClient(con)
		return NewMongostore(db, config.Collection, config.IDField)
	case *mongo.Client:
		db := &database.Database{Database: con.Database(config.Database)}
		return NewMongostore(db, config.Collection, config.IDField)
	case map[string]interface{}:
		var mc database.Client
		if err := util.DecodeJSON(config.Connection, &mc); err != nil {
			return nil, err
		}
		db := database.MongoConnectClient(&mc)
		return NewMongostore(db, config.Collection, config.IDField)
	default:
		return nil, errors.New("[docstore/mongo] unsupported connection type")
	}
}

func NewMongostore(db *database.Database, collection, idField string) (*MongoStore, error) {

	return &MongoStore{
		store:      db.Database.Collection(collection),
		idField:    idField,
		collection: collection,
	}, nil
}

func (m *MongoStore) getID(doc interface{}) (interface{}, error) {
	idf, err := util.FindFieldByTag(doc, "json", m.idField)
	if err != nil {
		return nil, err
	}

	id, ok := util.Lookup(idf, doc)

	if !ok {
		return nil, errors.New("[docstore/mongo] missing document ID")
	}

	return id, nil
}

func (m *MongoStore) Create(ctx context.Context, doc interface{}) error {

	id, err := m.getID(doc)
	if err != nil {
		return err
	}

	if m.exist(ctx, id) {
		return errors.New("[docstore/mongo] document already exist")
	}

	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, d); err != nil {
		return err
	}
	convertTime(d)

	_, err = m.store.InsertOne(ctx, d)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoStore) Update(ctx context.Context, id, doc interface{}, replace bool) error {

	if replace {
		_, err := m.store.ReplaceOne(ctx, bson.D{{Key: m.idField, Value: id}}, doc)
		return err
	}

	out := make(map[string]interface{})
	if err := util.DecodeJSON(doc, out); err != nil {
		return err
	}

	fields := bson.D{}
	for k, v := range out {
		fields = append(fields, bson.E{Key: k, Value: v})
	}

	update := bson.D{{Key: "$set", Value: fields}}

	res, err := m.store.UpdateOne(ctx, bson.D{{Key: m.idField, Value: id}}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return docstore.NotFound
	}

	return nil
}

func (m *MongoStore) UpdateField(ctx context.Context, id interface{}, fields []docstore.Field) error {
	fs := bson.D{}
	for _, v := range fields {
		fs = append(fs, bson.E{Key: v.Name, Value: v.Value})
	}

	update := bson.D{{Key: "$set", Value: fs}}

	res, err := m.store.UpdateOne(ctx, bson.D{{Key: m.idField, Value: id}}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return docstore.NotFound
	}
	return err
}

func (m *MongoStore) Increment(ctx context.Context, id interface{}, key string, value int) error {
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: key, Value: value}}}}
	upsert := true
	res, err := m.store.UpdateOne(ctx, bson.D{{Key: m.idField, Value: id}}, update, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return docstore.NotFound
	}
	return err
}

func (m *MongoStore) GetIncrement(ctx context.Context, id interface{}, key string, value int, doc interface{}) error {
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: key, Value: value}}}}
	rp := options.After
	upsert := true
	res := m.store.FindOneAndUpdate(ctx, bson.D{{Key: m.idField, Value: id}}, update, &options.FindOneAndUpdateOptions{ReturnDocument: &rp, Upsert: &upsert})
	return res.Decode(doc)
}

func (m *MongoStore) Delete(ctx context.Context, id interface{}) error {
	_, err := m.store.DeleteOne(ctx, bson.D{{Key: m.idField, Value: id}})
	return err
}

func (m *MongoStore) Get(ctx context.Context, id interface{}, doc interface{}) error {
	out := make(map[string]interface{})
	if err := m.store.FindOne(ctx, bson.D{{Key: m.idField, Value: id}}).Decode(&out); err != nil {
		if err == mongo.ErrNoDocuments {
			return docstore.NotFound
		}
		return err
	}
	out[m.idField] = id

	return util.DecodeJSON(out, doc)
}

func (m *MongoStore) exist(ctx context.Context, id interface{}) bool {
	out := make(map[string]interface{})
	if err := m.store.FindOne(ctx, bson.D{{Key: m.idField, Value: id}}).Decode(&out); err != nil {
		return false
	}
	return true
}

func (m *MongoStore) Find(ctx context.Context, query *docstore.QueryOpt, docs interface{}) error {

	f, opt := toMongoFilter(query)
	res, err := m.store.Find(ctx, f, opt)
	if err != nil {
		return err
	}

	var out []map[string]interface{}

	if err := res.All(ctx, &out); err != nil {
		return err
	}

	return util.DecodeJSON(out, docs)
}

func (m *MongoStore) BulkCreate(ctx context.Context, docs []interface{}) error {
	ins := make([]interface{}, 0)
	for _, doc := range docs {
		d := make(map[string]interface{})
		if err := util.DecodeJSON(doc, d); err != nil {
			return err
		}
		convertTime(d)
		ins = append(ins, d)
	}

	_, err := m.store.InsertMany(ctx, ins)
	return err
}

func (m *MongoStore) BulkGet(ctx context.Context, ids []interface{}, docs interface{}) error {

	res, err := m.store.Find(ctx, bson.D{{Key: m.idField, Value: bson.M{"$in": ids}}})
	if err != nil {
		return err
	}

	var out []map[string]interface{}

	if err := res.All(ctx, &out); err != nil {
		return err
	}

	return util.DecodeJSON(out, docs)
}

func (m *MongoStore) Migrate(ctx context.Context, config interface{}) error {
	db := m.store.Database()
	cols, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}
	for _, c := range cols {
		if c == m.collection {
			return nil
		}
	}
	return db.CreateCollection(ctx, m.collection)
}

func convertTime(obj map[string]interface{}) {
	for k, v := range obj {
		if reflect.TypeOf(v) == reflect.TypeOf(time.Time{}) {
			obj[k] = primitive.NewDateTimeFromTime(v.(time.Time))
		}
	}
}
