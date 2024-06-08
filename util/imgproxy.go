package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

type ImageProxyOpt struct {
	URL       string
	Key       string
	Salt      string
	Resize    string
	Width     int
	Height    int
	Gravity   string
	Extension string
	Enlarge   bool
	ProxyURL  string
}

func NewImageProxy(base, url, key, salt string) *ImageProxyOpt {
	return &ImageProxyOpt{
		URL:       url,
		ProxyURL:  base,
		Key:       key,
		Salt:      salt,
		Resize:    "fit",
		Width:     1024,
		Height:    768,
		Gravity:   "no",
		Enlarge:   false,
		Extension: "jpg",
	}
}

var resizingType = map[string]bool{
	"fit":       true,
	"fill":      true,
	"fill-down": true,
	"force":     true,
	"auto":      true,
}

var gravity = map[string]bool{
	"no":   true,
	"so":   true,
	"ea":   true,
	"we":   true,
	"noea": true,
	"nowe": true,
	"sowe": true,
	"soea": true,
	"ce":   true,
}

var ext = map[string]bool{
	"jpg":  true,
	"png":  true,
	"bmp":  true,
	"webp": true,
	"gif":  true,
	"ico":  true,
	"avif": true,
	"tiff": true,
}

func (i *ImageProxyOpt) validate() error {
	if i.URL == "" {
		return errors.New("empty URL")
	}

	if i.Key == "" {
		return errors.New("missing Key")
	}

	if i.Salt == "" {
		return errors.New("missing salt")
	}

	if i.Resize == "" {
		i.Resize = "auto"
	}

	i.Resize = strings.ToLower(i.Resize)

	if !resizingType[i.Resize] {
		return errors.New("unsupported resize type")
	}

	if i.Width == 0 {
		i.Width = 1024
	}

	if i.Height == 0 {
		i.Height = 768
	}

	if i.Gravity == "" {
		i.Gravity = "no"
	}

	i.Gravity = strings.ToLower(i.Gravity)

	if !gravity[i.Gravity] {
		return errors.New("unsupported gravity")
	}

	if i.Extension == "" {
		i.Extension = "jpg"
	}

	i.Extension = strings.ToLower(i.Extension)

	if !ext[i.Extension] {
		return errors.New("unsupported extension")
	}

	return nil
}

func (i *ImageProxyOpt) GetURL() (string, error) {
	if err := i.validate(); err != nil {
		return "", err
	}

	enc := base64.RawStdEncoding.EncodeToString([]byte(i.URL))
	enc = strings.ReplaceAll(strings.ReplaceAll(enc, "/", "_"), "+", "-")
	enl := "0"
	if i.Enlarge {
		enl = "1"
	}
	path := "/rs:" + i.Resize + ":" + fmt.Sprintf("%v", i.Width) + ":" + fmt.Sprintf("%v", i.Height) + ":" + enl + "/g:" + i.Gravity + "/" + enc + "." + i.Extension
	mac := hmac.New(sha256.New, []byte(i.Key))
	mac.Write([]byte(i.Salt))
	mac.Write([]byte(path))
	hmac := base64.RawStdEncoding.EncodeToString(mac.Sum(nil))
	hmac = strings.ReplaceAll(strings.ReplaceAll(hmac, "/", "_"), "+", "-")
	return i.ProxyURL + "/" + hmac + path, nil
}
