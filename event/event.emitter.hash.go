package event

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

func hash(m interface{}) (string, error) {
	mb, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	k := sha256.Sum256(mb)

	return string(base64.StdEncoding.EncodeToString(k[:])), nil
}
