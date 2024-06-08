package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeStruct(t *testing.T) {
	type User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Username  string    `json:"username"`
		Age       int       `json:"age"`
		CreatedAt time.Time `json:"created_at"`
	}

	usr := User{
		ID:        "1233",
		Name:      "sahal",
		CreatedAt: time.Now(),
	}

	out := make(map[string]interface{})

	require.Nil(t, DecodeJSON(usr, out))
	assert.Equal(t, usr.CreatedAt, out["created_at"])

	var user User
	require.Nil(t, DecodeJSON(out, &user))
	assert.Equal(t, usr.CreatedAt, user.CreatedAt)
}

func TestDecodeString(t *testing.T) {
	v, err := DecodeString("sahal")
	assert.Nil(t, err)
	assert.Equal(t, "sahal", v)

	v, err = DecodeString("35")
	assert.Nil(t, err)
	assert.Equal(t, int64(35), v)

	v, err = DecodeString("true")
	assert.Nil(t, err)
	assert.True(t, v.(bool))

	v, err = DecodeString("1.01")
	assert.Nil(t, err)
	assert.Equal(t, 1.01, v)
}
