package docstore

import (
	"context"
	"errors"
)

type Field struct {
	Name  string
	Value interface{}
}

type Driver interface {
	//GetID(doc interface{}) (interface{}, error)
	Create(ctx context.Context, doc interface{}) error
	Update(ctx context.Context, id, doc interface{}, replace bool) error
	UpdateField(ctx context.Context, id interface{}, fields []Field) error
	Increment(ctx context.Context, id interface{}, key string, value int) error
	GetIncrement(ctx context.Context, id interface{}, key string, value int, doc interface{}) error
	Delete(ctx context.Context, id interface{}) error
	Get(ctx context.Context, id interface{}, doc interface{}) error
	Find(ctx context.Context, query *QueryOpt, docs interface{}) error
	BulkCreate(ctx context.Context, docs []interface{}) error
	BulkGet(ctx context.Context, ids []interface{}, docs interface{}) error
	Migrate(ctx context.Context, config interface{}) error
}

type DriverFactory func(config *Config) (Driver, error)

var drivers = map[string]DriverFactory{
	"memory": MemoryStoreFactory,
}

func RegisterDriver(name string, fn DriverFactory) {
	drivers[name] = fn
}

func GetDriver(config *Config) (Driver, error) {
	fn, ok := drivers[config.Driver]
	if !ok {
		return nil, errors.New("[docstore] driver not supported")
	}

	return fn(config)
}
