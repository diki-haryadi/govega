package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImgProxyURL(t *testing.T) {
	res := "http://localhost:8080/oOXB13_3e8ZR1PTQQLcpI6mHACY7qgvf7GjzypnQOqs/rs:fit:1024:768:0/g:no/aHR0cHM6Ly9zaWNlcGF0cmVzaS5zMy5hbWF6b25hd3MuY29tLzAwMTYxMDAvMDAxNjEwMDA0OTgwLmpwZw.jpg"
	opt := &ImageProxyOpt{
		URL:       "https://sicepatresi.s3.amazonaws.com/0016100/001610004980.jpg",
		Key:       "testkey12345",
		Salt:      "testkey12345",
		Resize:    "fit",
		Width:     1024,
		Height:    768,
		Gravity:   "no",
		Enlarge:   false,
		Extension: "jpg",
		ProxyURL:  "http://localhost:8080",
	}

	out, err := opt.GetURL()
	assert.Nil(t, err)
	assert.Equal(t, res, out)

	opt = NewImageProxy("http://localhost:8080", "https://sicepatresi.s3.amazonaws.com/0016100/001610004980.jpg", "testkey12345", "testkey12345")
	out, err = opt.GetURL()
	assert.Nil(t, err)
	assert.Equal(t, res, out)
}
