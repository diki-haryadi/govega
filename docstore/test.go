package docstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/diki-haryadi/govega/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func DriverCRUDTest(d Driver, t *testing.T) {
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

	require.Nil(t, d.Create(ctx, usr))

	var doc User
	require.Nil(t, d.Get(ctx, usr.ID, &doc))
	assert.Equal(t, usr.ID, doc.ID)
	assert.Equal(t, usr.Name, doc.Name)
	assert.Equal(t, usr.Username, doc.Username)
	assert.Equal(t, usr.Age, doc.Age)
	assert.Equal(t, usr.CreatedAt.Unix(), doc.CreatedAt.Unix())

	doc.Age = 36
	require.Nil(t, d.Update(ctx, doc.ID, doc, false))

	var user User
	require.Nil(t, d.Get(ctx, usr.ID, &user))
	assert.Equal(t, doc.Age, user.Age)
	assert.Equal(t, ts.Unix(), user.CreatedAt.Unix())

	nu := &User{
		ID:   user.ID,
		Name: "Sahal Zain",
	}
	require.Nil(t, d.Update(ctx, nu.ID, nu, true))
	require.Nil(t, d.Get(ctx, nu.ID, &user))
	assert.Equal(t, 0, user.Age)

	require.Nil(t, d.UpdateField(ctx, nu.ID, []Field{{Name: "age", Value: 36}}))
	require.Nil(t, d.Get(ctx, nu.ID, &user))
	assert.Equal(t, 36, user.Age)

	require.Nil(t, d.Increment(ctx, nu.ID, "age", 1))
	require.Nil(t, d.Get(ctx, nu.ID, &user))
	assert.Equal(t, 37, user.Age)

	require.Nil(t, d.GetIncrement(ctx, nu.ID, "age", 1, &user))
	assert.Equal(t, 38, user.Age)

	require.Nil(t, d.Delete(ctx, nu.ID))
	require.NotNil(t, d.Get(ctx, nu.ID, &user))

	for i := 0; i < 10; i++ {
		u := &User{
			ID:        fmt.Sprintf("%v", i),
			Name:      "name" + fmt.Sprintf("%v", i),
			Age:       30 + i,
			CreatedAt: time.Now(),
		}
		require.Nil(t, d.Create(ctx, u))
	}

	q := &QueryOpt{
		Filter: []FilterOpt{
			{Field: "name", Ops: constant.EQ, Value: "name1"},
		},
	}

	var out []User

	require.Nil(t, d.Find(ctx, q, &out))
	assert.Equal(t, 1, len(out))

	q = &QueryOpt{
		Filter: []FilterOpt{
			{Field: "age", Ops: constant.GE, Value: 35},
		},
	}

	out = nil
	require.Nil(t, d.Find(ctx, q, &out))
	assert.Equal(t, 5, len(out))

	q = &QueryOpt{
		Filter: []FilterOpt{
			{Field: "age", Ops: constant.GE, Value: 32},
		},
		Limit:    5,
		Skip:     5,
		OrderBy:  "age",
		IsAscend: true,
	}

	out = nil
	require.Nil(t, d.Find(ctx, q, &out))
	assert.Equal(t, 3, len(out))
	assert.Equal(t, 37, out[0].Age)

}

func DriverBulkTest(d Driver, t *testing.T) {
	type User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Username  string    `json:"username"`
		Age       int       `json:"age"`
		CreatedAt time.Time `json:"created_at"`
	}

	ctx := context.Background()

	ins := make([]interface{}, 0)
	ids := make([]interface{}, 0)
	for i := 0; i < 10; i++ {

		u := &User{
			ID:        fmt.Sprintf("BLK-%v", i),
			Name:      "name" + fmt.Sprintf("%v", i),
			Age:       30 + i,
			CreatedAt: time.Now(),
		}
		ins = append(ins, u)
		ids = append(ids, fmt.Sprintf("BLK-%v", i))
	}

	require.Nil(t, d.BulkCreate(ctx, ins))

	var out []*User
	require.Nil(t, d.BulkGet(ctx, ids, &out))

	require.Equal(t, 10, len(out))

	for i := 0; i < 10; i++ {
		assert.Equal(t, ins[i].(*User).ID, out[i].ID)
		assert.Equal(t, ins[i].(*User).Name, out[i].Name)
		assert.Equal(t, ins[i].(*User).Age, out[i].Age)
		assert.Equal(t, ins[i].(*User).CreatedAt.Unix(), out[i].CreatedAt.Unix())
	}
}

func DocstoreTestCRUD(cs *CachedStore, t *testing.T) {
	type User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Username  string    `json:"username"`
		Age       int       `json:"age"`
		CreatedAt time.Time `json:"created_at"`
	}

	ctx := context.Background()

	usr := &User{
		Name:     "sahal",
		Username: "sahalzain",
		Age:      35,
	}

	require.Nil(t, cs.Create(ctx, usr))

	assert.NotEmpty(t, usr.ID)
	assert.Equal(t, time.Now().Unix(), usr.CreatedAt.Unix())

	var doc User
	require.Nil(t, cs.Get(ctx, usr.ID, &doc))
	assert.Equal(t, usr.ID, doc.ID)
	assert.Equal(t, usr.Name, doc.Name)
	assert.Equal(t, usr.Username, doc.Username)
	assert.Equal(t, usr.Age, doc.Age)
	assert.Equal(t, usr.CreatedAt.Unix(), doc.CreatedAt.Unix())

	assert.True(t, cs.cache.Exist(ctx, usr.ID))

	doc.Age = 36
	require.Nil(t, cs.Update(ctx, doc))
	assert.False(t, cs.cache.Exist(ctx, usr.ID))

	var user User
	require.Nil(t, cs.Get(ctx, usr.ID, &user))
	assert.Equal(t, doc.Age, user.Age)
	assert.Equal(t, doc.CreatedAt.Unix(), user.CreatedAt.Unix())

	nu := &User{
		ID:   user.ID,
		Name: "Sahal Zain",
	}
	require.Nil(t, cs.Replace(ctx, nu))
	require.Nil(t, cs.Get(ctx, nu.ID, &user))
	assert.Equal(t, 0, user.Age)

	require.Nil(t, cs.UpdateField(ctx, nu.ID, "age", 36))
	require.Nil(t, cs.Get(ctx, nu.ID, &user))
	assert.Equal(t, 36, user.Age)

	require.Nil(t, cs.Increment(ctx, nu.ID, "age", 2))
	require.Nil(t, cs.Get(ctx, nu.ID, &user))
	assert.Equal(t, 38, user.Age)

	require.Nil(t, cs.Delete(ctx, nu.ID))
	require.NotNil(t, cs.Get(ctx, nu.ID, &user))

	for i := 0; i < 10; i++ {
		u := &User{
			ID:        fmt.Sprintf("%v", i),
			Name:      "name" + fmt.Sprintf("%v", i),
			Age:       30 + i,
			CreatedAt: time.Now(),
		}
		require.Nil(t, cs.Create(ctx, u))
	}

	q := &QueryOpt{
		Filter: []FilterOpt{
			{Field: "name", Ops: constant.EQ, Value: "name1"},
		},
	}

	var out []User

	require.Nil(t, cs.Find(ctx, q, &out))
	assert.Equal(t, 1, len(out))

	q = &QueryOpt{
		Filter: []FilterOpt{
			{Field: "age", Ops: constant.GE, Value: 35},
		},
	}

	out = nil
	require.Nil(t, cs.Find(ctx, q, &out))
	assert.Equal(t, 5, len(out))
}
