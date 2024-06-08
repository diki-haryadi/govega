package util

import (
	"testing"

	"github.com/diki-haryadi/govega/constant"
	"github.com/stretchr/testify/assert"
)

func TestChildLookup(t *testing.T) {
	obj := map[string]interface{}{
		"result": map[string]interface{}{
			"name": "sahal",
		},
	}

	name, ok := Lookup("result.name", obj)
	assert.True(t, ok)
	assert.Equal(t, "sahal", name)

	assert.True(t, FieldExist("result.name", obj))
}

func TestStructLookup(t *testing.T) {
	type User struct {
		Name    string
		Age     int
		Address string
	}

	usr := User{
		Name: "sahal",
		Age:  30,
	}

	name, ok := Lookup("Name", usr)
	assert.True(t, ok)
	assert.Equal(t, "sahal", name)

	name, ok = Lookup("Name", &usr)
	assert.True(t, ok)
	assert.Equal(t, "sahal", name)
}

func TestSimpleLookup(t *testing.T) {
	obj := map[string]interface{}{
		"result": "sahal",
		"num":    -1.23232,
	}

	name, ok := Lookup("result", obj)
	assert.True(t, ok)
	assert.Equal(t, "sahal", name)

	num, ok := Lookup("num", obj)
	assert.True(t, ok)
	assert.Equal(t, -1.23232, num)
}

func TestSliceLookup(t *testing.T) {
	obj := map[string]interface{}{
		"result": map[string]interface{}{
			"name":  "sahal",
			"roles": []string{"user", "member", "admin"},
			"data": []interface{}{
				map[string]interface{}{
					"age": 37,
				},
				map[string]interface{}{
					"location": "Jogja",
				},
			},
		},
	}

	o, ok := Lookup("result.roles.0", obj)
	assert.True(t, ok)
	assert.Equal(t, "user", o)

	o, ok = Lookup("result.data.0.age", obj)
	assert.True(t, ok)
	assert.Equal(t, 37, o)
}

func TestMatch(t *testing.T) {
	obj := map[string]interface{}{
		"result": map[string]interface{}{
			"name":  "sahal",
			"roles": []string{"user", "member", "admin"},
			"data": []interface{}{
				map[string]interface{}{
					"age": 37,
				},
				map[string]interface{}{
					"location": "Jogja",
				},
			},
		},
	}

	assert.True(t, match("result.name", obj, "sahal"))
	assert.True(t, match("result.roles._", obj, "user"))
	assert.True(t, match("result.data._.location", obj, "Jogja"))
	assert.True(t, match("result.data._.location", obj, "Jog.*"))
	assert.True(t, match("result.data._.age", obj, 37))

}

func TestAssert(t *testing.T) {
	obj := map[string]interface{}{
		"result": map[string]interface{}{
			"name":      "sahal",
			"user_name": "sahalzain",
			"roles":     []string{"user", "member", "admin"},
			"data": []interface{}{
				map[string]interface{}{
					"age": 37,
				},
				map[string]interface{}{
					"location": "Jogja",
				},
			},
		},
	}

	assert.True(t, Assert("result.name", obj, "sahal", constant.EQ))
	assert.True(t, Assert("result.user_name", obj, "sahalzain", constant.EQ))
	assert.True(t, Assert("result.roles._", obj, "user", constant.EQ))
	assert.True(t, Assert("result.data._.location", obj, "Jogja", constant.EQ))
	assert.True(t, Assert("result.data._.location", obj, "Jog.*", constant.RE))
	assert.True(t, Assert("result.data._.age", obj, 37, constant.EQ))

}
