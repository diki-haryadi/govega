package docstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/diki-haryadi/govega/cache"
	"github.com/diki-haryadi/govega/log"
	"github.com/diki-haryadi/govega/util"
)

const (
	defaultID         = "id"
	defaultTimestamp  = "created_at"
	defaultExpiration = 3600 * 24
)

type Config struct {
	Database        string      `json:"db,omitempty"`
	Collection      string      `json:"collection,omitempty"`
	CacheURL        string      `json:"cache_url,omitempty"`
	CacheExpiration int         `json:"cache_expiration,omitempty"`
	IDField         string      `json:"id_field,omitempty"`
	TimestampField  string      `json:"timestamp_field,omitempty"`
	Driver          string      `json:"driver,omitempty"`
	Connection      interface{} `json:"connection,omitempty"`
}

type CachedStore struct {
	*Config
	cache   cache.Cache
	storage Driver
}

func New(config *Config) (*CachedStore, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	dv, err := GetDriver(config)
	if err != nil {
		return nil, err
	}

	cache, err := cache.New(config.CacheURL)
	if err != nil {
		return nil, err
	}

	return NewDocstore(dv, cache, config), nil
}

func NewDocstore(storage Driver, cache cache.Cache, config *Config) *CachedStore {
	return &CachedStore{
		Config:  config,
		cache:   cache,
		storage: storage,
	}
}

func (c *Config) validate() error {

	if c.Database == "" {
		return errors.New("[docstore] missing database param")
	}

	if c.Collection == "" {
		return errors.New("[docstore] missing collection param")
	}

	if c.CacheURL == "" {
		return errors.New("[docstore] missing cache_url param")
	}

	if c.Driver == "" {
		return errors.New("[docstore] missing driver param")
	}

	if c.Connection == nil {
		return errors.New("[docstore] missing connection param")
	}

	if c.IDField == "" {
		c.IDField = defaultID
	}

	if c.TimestampField == "" {
		c.TimestampField = defaultTimestamp
	}

	if c.CacheExpiration == 0 {
		c.CacheExpiration = defaultExpiration
	}

	return nil
}

func (s *CachedStore) getID(doc interface{}) (interface{}, error) {
	idf, err := util.FindFieldByTag(doc, "json", s.IDField)
	if err != nil {
		return nil, err
	}

	id, ok := util.Lookup(idf, doc)

	if !ok {
		return nil, errors.New("[docstore] missing document ID")
	}

	return id, nil
}

func (s *CachedStore) setID(doc interface{}, idField string) error {
	idf := idField

	if idf == "" {
		idfl, err := util.FindFieldByTag(doc, "json", s.IDField)
		if err != nil {
			return err
		}
		idf = idfl
	}

	_, ok := util.Lookup(idf, doc)

	if !ok {
		ft, err := util.FindFieldTypeByTag(doc, "json", s.IDField)
		if err != nil {
			return err
		}
		switch ft {
		case reflect.TypeOf(""):
			uid := util.Hash64(doc)
			return util.SetValue(doc, idf, uid)
		case reflect.TypeOf(int64(0)):
			uid := util.GenerateRandUID()
			return util.SetValue(doc, idf, uid)
		case reflect.TypeOf(int(0)):
			uid := util.GenerateRandUID()
			return util.SetValue(doc, idf, int(uid))
		case reflect.TypeOf(uint64(0)):
			uid := util.GenerateRandUID()
			return util.SetValue(doc, idf, uint64(uid))
		case reflect.TypeOf(uint(0)):
			uid := util.GenerateRandUID()
			return util.SetValue(doc, idf, uint(uid))
		default:
			return errors.New("[docstore] unsupported ID type")
		}
	}

	return nil
}

func (s *CachedStore) Create(ctx context.Context, doc interface{}) error {
	if err := s.setID(doc, ""); err != nil {
		return err
	}
	tsf, err := util.FindFieldByTag(doc, "json", s.TimestampField)
	if err != nil {
		return err
	}

	_, ok := util.Lookup(tsf, doc)
	if !ok {
		if err := util.SetValue(doc, tsf, time.Now()); err != nil {
			return err
		}
	}

	return s.storage.Create(ctx, doc)
}

func (s *CachedStore) update(ctx context.Context, doc interface{}, replace bool) error {
	id, err := s.getID(doc)
	if err != nil {
		return err
	}

	if err := s.cache.Delete(ctx, fmt.Sprintf("%v", id)); err != nil {
		log.WithError(err).Error("error deleting cache ")
	}

	return s.storage.Update(ctx, id, doc, replace)
}

func (s *CachedStore) Update(ctx context.Context, doc interface{}) error {
	return s.update(ctx, doc, false)
}

func (s *CachedStore) UpdateField(ctx context.Context, id interface{}, key string, value interface{}) error {
	if err := s.cache.Delete(ctx, fmt.Sprintf("%v", id)); err != nil {
		log.WithError(err).Error("error deleting cache ")
	}

	return s.storage.UpdateField(ctx, id, []Field{{Name: key, Value: value}})
}

func (s *CachedStore) Increment(ctx context.Context, id interface{}, fieldName string, value int) error {
	if err := s.cache.Delete(ctx, fmt.Sprintf("%v", id)); err != nil {
		log.WithError(err).Error("error deleting cache ")
	}
	return s.storage.Increment(ctx, id, fieldName, value)
}

func (s *CachedStore) Replace(ctx context.Context, doc interface{}) error {
	return s.update(ctx, doc, true)
}

func (s *CachedStore) Get(ctx context.Context, id, doc interface{}) error {

	if !util.IsPointerOfStruct(doc) && !util.IsMap(doc) {
		return errors.New("[docstore] docs should be a pointer of struct or map")
	}

	if s.cache.Exist(ctx, fmt.Sprintf("%v", id)) {
		if err := s.cache.GetObject(ctx, fmt.Sprintf("%v", id), doc); err == nil {
			return nil
		}
	}

	if err := s.storage.Get(ctx, id, doc); err != nil {
		return err
	}

	return s.cache.Set(ctx, fmt.Sprintf("%v", id), doc, s.CacheExpiration)
}

func (s *CachedStore) Delete(ctx context.Context, id interface{}) error {
	if err := s.cache.Delete(ctx, fmt.Sprintf("%v", id)); err != nil {
		log.WithError(err).Error("error deleting cache ")
	}

	return s.storage.Delete(ctx, id)
}

func (s *CachedStore) Find(ctx context.Context, query *QueryOpt, docs interface{}) error {

	if !util.IsPointerOfSlice(docs) {
		return errors.New("[docstore] docs should be a pointer of slice")
	}

	return s.storage.Find(ctx, query, docs)
}

func (s *CachedStore) BulkCreate(ctx context.Context, docs interface{}) error {
	if !util.IsSlice(docs) {
		return errors.New("[docstore] documents should be a slice")
	}

	rdocs := reflect.ValueOf(docs)
	ins := make([]interface{}, rdocs.Len())
	for i := 0; i < rdocs.Len(); i++ {
		ins[i] = rdocs.Index(i).Interface()
	}

	tsf := ""
	idf := ""

	for i, d := range ins {

		if idf == "" {
			idfl, err := util.FindFieldByTag(d, "json", s.IDField)
			if err != nil {
				return err
			}
			idf = idfl
		}

		if err := s.setID(d, idf); err != nil {
			return err
		}

		if tsf == "" {
			tf, err := util.FindFieldByTag(d, "json", s.TimestampField)
			if err != nil {
				return err
			}
			tsf = tf
		}

		_, ok := util.Lookup(tsf, d)
		if !ok {
			if err := util.SetValue(d, tsf, time.Now()); err != nil {
				return err
			}
		}
		ins[i] = d
	}

	return s.storage.BulkCreate(ctx, ins)
}

func (s *CachedStore) BulkGet(ctx context.Context, ids, docs interface{}) error {
	if !util.IsSlice(ids) {
		return errors.New("[docstore] IDs should be a slice")
	}

	if !util.IsPointerOfSlice(docs) {
		return errors.New("[docstore] docs should be a pointer of slice")
	}

	rids := reflect.ValueOf(ids)
	ins := make([]interface{}, rids.Len())
	for i := 0; i < rids.Len(); i++ {
		ins[i] = rids.Index(i).Interface()
	}

	return s.storage.BulkGet(ctx, ins, docs)
}

func (s *CachedStore) Migrate(ctx context.Context, config interface{}) error {
	return s.storage.Migrate(ctx, config)
}

func (s *CachedStore) GetCache() cache.Cache {
	return s.cache
}
