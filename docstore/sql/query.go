package sql

import (
	"fmt"

	"github.com/diki-haryadi/govega/constant"
	"github.com/diki-haryadi/govega/docstore"
)

func (s *SQLStore) buildInsertQuery(obj map[string]interface{}) string {
	ds := goqu.Dialect(s.driver).Insert(s.table).Rows(goqu.Record(obj))
	stmt, _, _ := ds.ToSQL()
	return stmt
}

func (s *SQLStore) buildBulkInsertQuery(obj []map[string]interface{}) string {
	recs := make([]goqu.Record, 0)
	for _, o := range obj {
		recs = append(recs, goqu.Record(o))
	}
	ds := goqu.Dialect(s.driver).Insert(s.table).Rows(recs)
	stmt, _, _ := ds.ToSQL()
	return stmt
}

func (s *SQLStore) buildUpdateQuery(obj map[string]interface{}, id interface{}) string {
	delete(obj, s.idField)
	ds := goqu.Dialect(s.driver).Update(s.table).Set(goqu.Record(obj)).Where(goqu.Ex{s.idField: id})
	stmt, _, _ := ds.ToSQL()
	return stmt
}

func (s *SQLStore) buildDeleteQuery(id interface{}) string {
	ds := goqu.Dialect(s.driver).Delete(s.table).Where(goqu.Ex{s.idField: id})
	stmt, _, _ := ds.ToSQL()
	return stmt
}

func (s *SQLStore) buildGetQuery(id interface{}) string {
	ds := goqu.Dialect(s.driver).From(s.table).Where(goqu.Ex{s.idField: id})
	stmt, _, _ := ds.ToSQL()
	return stmt
}

func (s *SQLStore) buildBulkGetQuery(ids []interface{}) string {
	ds := goqu.Dialect(s.driver).From(s.table).Where(goqu.Ex{s.idField: ids})
	stmt, _, _ := ds.ToSQL()
	return stmt
}

func (s *SQLStore) buildIncrQuery(id interface{}, key string, value int) string {
	val := goqu.L(fmt.Sprintf(`(%v+%s)`, value, goqu.C(key).GetCol()))

	ds := goqu.Dialect(s.driver).Update(s.table).Set(goqu.Ex{key: val}).Where(goqu.Ex{s.idField: id})
	//ds := goqu.Dialect(s.driver).Update(s.table).Set(goqu.Record{key: value}).Where(goqu.Ex{s.idField: id})
	stmt, _, _ := ds.ToSQL()
	return stmt
}

func (s *SQLStore) buildGetIncrQuery(id interface{}, key string, value int) string {
	val := goqu.L(fmt.Sprintf(`(%v+%s)`, value, goqu.C(key).GetCol()))

	ds := goqu.Dialect(s.driver).Update(s.table).Set(goqu.Ex{key: val}).Where(goqu.Ex{s.idField: id}).Returning(key)
	//ds := goqu.Dialect(s.driver).Update(s.table).Set(goqu.Record{key: value}).Where(goqu.Ex{s.idField: id})
	stmt, _, _ := ds.ToSQL()
	return stmt
}

func (s *SQLStore) buildFindQuery(opt *docstore.QueryOpt) string {
	ds := goqu.Dialect(s.driver).From(s.table)

	filter := make(map[string]interface{})
	for _, f := range opt.Filter {
		switch f.Ops {
		case constant.EQ, constant.SE:
			filter[f.Field] = f.Value
			continue
		case constant.GT:
			filter[f.Field] = goqu.Op{"gt": f.Value}
			continue
		case constant.GE:
			filter[f.Field] = goqu.Op{"gte": f.Value}
			continue
		case constant.LT:
			filter[f.Field] = goqu.Op{"lt": f.Value}
			continue
		case constant.LE:
			filter[f.Field] = goqu.Op{"lte": f.Value}
			continue
		case constant.NE, constant.SN:
			filter[f.Field] = goqu.Op{"neq": f.Value}
			continue
		case constant.IN, constant.EM:
			filter[f.Field] = goqu.Op{"in": f.Value}
			continue
		case constant.RE:
			filter[f.Field] = goqu.Op{"like": f.Value}
			continue
		default:
			filter[f.Field] = f.Value
		}
	}

	ds = ds.Where(goqu.Ex(filter))
	if opt.Limit > 0 {
		ds = ds.Limit(uint(opt.Limit))
		if opt.Page > 0 {
			ds = ds.Offset(uint(opt.Limit * opt.Page))
		}
	}

	if opt.Skip > 0 {
		ds = ds.Offset(uint(opt.Skip))
	}

	if opt.OrderBy != "" {
		col := goqu.C(opt.OrderBy)
		if opt.IsAscend {
			ds = ds.Order(col.Asc())
		} else {
			ds = ds.Order(col.Desc())
		}
	}

	stmt, _, _ := ds.ToSQL()
	return stmt
}
