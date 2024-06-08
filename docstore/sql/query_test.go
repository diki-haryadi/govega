package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertQuery(t *testing.T) {
	obj := map[string]interface{}{
		"id":   "1234",
		"name": "sahal",
		"age":  30,
	}

	s := &SQLStore{table: "user"}

	st := s.buildInsertQuery(obj)
	assert.Equal(t, `INSERT INTO "user" ("age", "id", "name") VALUES (30, '1234', 'sahal')`, st)

}

func TestUpdateQuery(t *testing.T) {
	obj := map[string]interface{}{
		"id":   "1234",
		"name": "sahal",
		"age":  30,
	}

	s := &SQLStore{table: "user", idField: "id"}
	st := s.buildUpdateQuery(obj, "1234")
	assert.Equal(t, `UPDATE "user" SET "age"=30,"name"='sahal' WHERE ("id" = '1234')`, st)

}

func TestDeleteQuery(t *testing.T) {
	s := &SQLStore{table: "user", idField: "id"}
	st := s.buildDeleteQuery("1234")
	assert.Equal(t, `DELETE FROM "user" WHERE ("id" = '1234')`, st)
}

func TestGetQuery(t *testing.T) {
	s := &SQLStore{table: "user", idField: "id"}
	st := s.buildGetQuery("1234")
	assert.Equal(t, `SELECT * FROM "user" WHERE ("id" = '1234')`, st)
}

func TestIncrQuery(t *testing.T) {
	s := &SQLStore{table: "user", idField: "id"}
	st := s.buildIncrQuery("1234", "count", 2)
	assert.Equal(t, `UPDATE "user" SET "count"=(2+count) WHERE ("id" = '1234')`, st)
}

func TestGetIncrQuery(t *testing.T) {
	s := &SQLStore{table: "user", idField: "id"}
	st := s.buildGetIncrQuery("1234", "count", 2)
	assert.Equal(t, `UPDATE "user" SET "count"=(2+count) WHERE ("id" = '1234') RETURNING "count"`, st)
}

func TestBulkInsert(t *testing.T) {
	obj := []map[string]interface{}{
		{
			"id":   "1234",
			"name": "sahal",
			"age":  30,
		},
		{
			"id":   "1235",
			"name": "zain",
			"age":  31,
		},
	}

	s := &SQLStore{table: "user"}

	st := s.buildBulkInsertQuery(obj)
	assert.Equal(t, `INSERT INTO "user" ("age", "id", "name") VALUES (30, '1234', 'sahal'), (31, '1235', 'zain')`, st)
}

func TestBulkGet(t *testing.T) {
	s := &SQLStore{table: "user", idField: "id"}
	st := s.buildBulkGetQuery([]interface{}{"1234", "1235"})
	assert.Equal(t, `SELECT * FROM "user" WHERE ("id" IN ('1234', '1235'))`, st)
}
