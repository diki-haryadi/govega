package docstore

import (
	"context"
	"errors"
	"sort"
	"sync"

	"dario.cat/mergo"
	"github.com/diki-haryadi/govega/constant"
	"github.com/diki-haryadi/govega/util"
)

//var memstore = make(map[string]*MemoryStore)

type MemoryStore struct {
	storage map[interface{}]map[string]interface{}
	idField string
	mux     *sync.Mutex
}

func MemoryStoreFactory(config *Config) (Driver, error) {
	return NewMemoryStore(config.Collection, config.IDField), nil
}

func NewMemoryStore(name, idField string) *MemoryStore {

	m := &MemoryStore{
		storage: make(map[interface{}]map[string]interface{}),
		idField: idField,
		mux:     &sync.Mutex{},
	}

	return m
}

func (m *MemoryStore) getID(doc interface{}) (interface{}, error) {
	idf, err := util.FindFieldByTag(doc, "json", m.idField)
	if err != nil {
		return nil, err
	}

	id, ok := util.Lookup(idf, doc)

	if !ok {
		return nil, errors.New("[docstore/memory] missing document ID")
	}

	return id, nil
}

func (m *MemoryStore) Create(ctx context.Context, doc interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	id, err := m.getID(doc)
	if err != nil {
		return err
	}

	if _, ok := m.storage[id]; ok {
		return errors.New("[docstore/memory] document ID is already exist")
	}

	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, d); err != nil {
		return err
	}

	m.storage[id] = d
	return nil
}

func (m *MemoryStore) Update(ctx context.Context, id, doc interface{}, replace bool) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, ok := m.storage[id]; !ok {
		return NotFound
	}
	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, d); err != nil {
		return err
	}

	if replace {
		m.storage[id] = d
		return nil
	}

	cd := m.storage[id]

	if err := mergo.MergeWithOverwrite(&cd, d); err != nil {
		return err
	}

	m.storage[id] = cd

	return nil
}

func (m *MemoryStore) UpdateField(ctx context.Context, id interface{}, fields []Field) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	d, ok := m.storage[id]
	if !ok {
		return NotFound
	}

	for _, f := range fields {
		//fn, _ := util.FindFieldByTag(d, "json", f.Name)
		if err := util.SetValue(d, f.Name, f.Value); err != nil {
			return err
		}
	}

	m.storage[id] = d
	return nil
}

func (m *MemoryStore) Increment(ctx context.Context, id interface{}, key string, value int) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	d, ok := m.storage[id]
	if !ok {
		m.storage[id] = map[string]interface{}{key: value}
		return nil
	}

	field, ok := d[key]
	if !ok {
		return errors.New("[docstore/memory] field not found")
	}

	switch v := field.(type) {
	case int:
		d[key] = v + value
	case int32:
		d[key] = v + int32(value)
	case int64:
		d[key] = v + int64(value)
	case float64:
		d[key] = v + float64(value)
	case float32:
		d[key] = v + float32(value)
	case uint:
		d[key] = v + uint(value)
	case uint32:
		d[key] = v + uint32(value)
	case uint64:
		d[key] = v + uint64(value)
	default:
		return errors.New("[docstore/memory] destination type is not a number")
	}

	m.storage[id] = d
	return nil
}

func (m *MemoryStore) GetIncrement(ctx context.Context, id interface{}, key string, value int, doc interface{}) error {
	if err := m.Increment(ctx, id, key, value); err != nil {
		return err
	}
	return m.Get(ctx, id, doc)
}

func (m *MemoryStore) Delete(ctx context.Context, id interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	delete(m.storage, id)
	return nil
}

func (m *MemoryStore) Get(ctx context.Context, id interface{}, doc interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	d, ok := m.storage[id]
	if !ok {
		return NotFound
	}

	if err := util.DecodeJSON(d, doc); err != nil {
		return err
	}

	return nil
}

func (m *MemoryStore) Find(ctx context.Context, query *QueryOpt, docs interface{}) error {
	// Only support = operation
	m.mux.Lock()
	defer m.mux.Unlock()

	out := make([]interface{}, 0)
	for _, d := range m.storage {
		match := false
		for _, f := range query.Filter {
			if !util.Assert(f.Field, d, f.Value, f.Ops) {
				match = false
				break
			}
			match = true
		}
		if match {
			out = append(out, d)
		}
	}

	if query.OrderBy != "" {
		sort.Slice(out, func(i, j int) bool {
			val, ok := util.Lookup(query.OrderBy, out[j])
			if ok {
				return util.Assert(query.OrderBy, out[i], val, constant.GT)
			}
			return false
		})
		if query.IsAscend {
			util.Reverse(out)
		}
	}

	if query.Page > 0 && query.Limit > 0 {
		query.Skip = query.Page * query.Limit
	}

	if query.Skip > 0 && len(out) > query.Skip {
		out = out[query.Skip:]
	}

	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}

	if err := util.DecodeJSON(out, docs); err != nil {
		return err
	}
	return nil

}

func (m *MemoryStore) BulkCreate(ctx context.Context, docs []interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	for _, doc := range docs {
		id, err := m.getID(doc)
		if err != nil {
			return err
		}

		if _, ok := m.storage[id]; ok {
			return errors.New("[docstore/memory] document ID is already exist")
		}

		d := make(map[string]interface{})
		if err := util.DecodeJSON(doc, d); err != nil {
			return err
		}

		m.storage[id] = d
	}

	return nil
}

func (m *MemoryStore) BulkGet(ctx context.Context, ids []interface{}, docs interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	out := make([]map[string]interface{}, 0)
	for _, id := range ids {
		d, ok := m.storage[id]
		if !ok {
			return NotFound
		}
		out = append(out, d)
	}

	if err := util.DecodeJSON(out, docs); err != nil {
		return err
	}

	return nil
}

func (m *MemoryStore) Migrate(ctx context.Context, config interface{}) error {
	return nil
}
