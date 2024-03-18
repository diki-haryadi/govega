package util

import "encoding/json"

func Beautify(data interface{}) string {
	b, _ := json.MarshalIndent(data, "", "    ")
	return string(b)
}
