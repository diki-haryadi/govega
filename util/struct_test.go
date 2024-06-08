package util

import (
	"testing"
	"time"

	"github.com/diki-haryadi/govega/constant"
	"github.com/stretchr/testify/assert"
)

func TestStructField(t *testing.T) {
	type User struct {
		Name     string
		Age      int
		Address  string
		Balance  float64
		IsVerify bool
		Parent   *User
	}

	usr := User{
		Name: "sahal",
		Age:  30,
	}

	n, ok := Lookup("Name", usr)
	assert.True(t, ok)
	assert.Equal(t, usr.Name, n)

	n, ok = Lookup("Age", usr)
	assert.True(t, ok)
	assert.Equal(t, usr.Age, n)

	// zero field
	n, ok = Lookup("Address", usr)
	assert.False(t, ok)
	assert.Equal(t, "", n)

	n, ok = Lookup("Balance", usr)
	assert.False(t, ok)
	// not sure why it cast to int64
	assert.Equal(t, int64(0), n)

	n, ok = Lookup("IsVerify", usr)
	assert.False(t, ok)
	assert.Equal(t, false, n)

	n, ok = Lookup("Parent", usr)
	assert.False(t, ok)
	assert.Nil(t, n)

	// invalid field
	n, ok = Lookup("BirthDate", usr)
	assert.False(t, ok)
	assert.Nil(t, n)

	assert.True(t, FieldExist("Age", usr))
	assert.True(t, FieldExist("Name", usr))
	assert.True(t, FieldExist("Address", usr))
	assert.True(t, FieldExist("Balance", usr))
	assert.True(t, FieldExist("IsVerify", usr))
	assert.True(t, FieldExist("Parent", usr))
	assert.False(t, FieldExist("Location", usr))

}

func TestSetValue(t *testing.T) {
	obj := map[string]interface{}{
		"result": "sahal",
	}

	assert.Nil(t, SetValue(obj, "result", "zain"))
	assert.Equal(t, "zain", obj["result"])

	assert.Nil(t, SetValue(obj, "status", "OK"))
	assert.Equal(t, "OK", obj["status"])

	type User struct {
		Name    string
		Age     int
		Address string
	}

	usr := User{
		Name: "sahal",
		Age:  30,
	}

	assert.Nil(t, SetValue(&usr, "Name", "zain"))
	assert.Equal(t, "zain", usr.Name)

	assert.Nil(t, SetValue(&usr, "Age", 35))
	assert.Equal(t, 35, usr.Age)

	assert.True(t, IsMap(obj))
}

func TestFindField(t *testing.T) {
	type User struct {
		Name    string `json:"name,omitempty" custom:"_name"`
		Age     int    `json:"age"`
		Address string `json:"address"`
	}

	usr := User{
		Name: "sahal",
		Age:  30,
	}

	field, err := FindFieldByTag(usr, "json", "name")
	assert.Nil(t, err)
	assert.Equal(t, "Name", field)

	field, err = FindFieldByTag(&usr, "custom", "_name")
	assert.Nil(t, err)
	assert.Equal(t, "Name", field)

	assert.False(t, IsPointerOfStruct(usr))
	assert.True(t, IsPointerOfStruct(&usr))
}

func TestCompare(t *testing.T) {
	ok, err := CompareValue("sahal", "sahal", constant.EQ)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = CompareValue(100, 50, constant.GT)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = CompareValue(100, 100, constant.GE)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = CompareValue(100, 100, constant.GT)
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = CompareValue(100, 100, constant.EQ)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = CompareValue(100, "100", constant.EQ)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = CompareValue(100, "100", constant.SE)
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = CompareValue(100, "100", constant.NE)
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = CompareValue(100, "100", constant.SN)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = CompareValue(100, 100, constant.SN)
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = CompareValue("Jogja", "Jog.*", constant.RE)
	assert.Nil(t, err)
	assert.True(t, ok)

	ts := time.Now()
	ok, err = CompareValue(ts, ts, constant.EQ)
	assert.Nil(t, err)
	assert.True(t, ok)

	now := time.Now()
	ok, err = CompareValue(now.Add(1000), now, constant.GT)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = CompareValue(now.Add(-1000), now, constant.LT)
	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestMap(t *testing.T) {
	m1 := make(map[string]interface{})
	m2 := make(map[string]string)
	m3 := make(map[interface{}]interface{})

	assert.True(t, IsMap(m1))
	assert.True(t, IsMap(m2))
	assert.True(t, IsMap(m3))
	assert.True(t, IsMapStringInterface(m1))
	assert.True(t, IsMapStringInterface(&m1))
	assert.False(t, IsMapStringInterface(m2))
	assert.False(t, IsMapStringInterface(m3))

}

func TestTime(t *testing.T) {
	ts := time.Now()

	assert.True(t, IsStructOrPointerOf(ts))
	assert.True(t, IsStructOrPointerOf(&ts))
	assert.True(t, IsTime(ts))
	assert.True(t, IsTime(&ts))
}
