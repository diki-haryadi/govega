package geoip

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGetUserIP(t *testing.T) {
	t.Run("it should return the user IP", func(t *testing.T) {
		ip := "202.147.207.71"

		header := http.Header{}
		header.Set("CF-Connecting-IP", ip)

		res := GetUserIP(header)

		if res != ip {
			t.Errorf("expected '%s', got '%s'", ip, res)
		}
	})

	t.Run("it should return first index of X-Forwarded-For", func(t *testing.T) {
		ip := "202.147.207.71"
		header := http.Header{}
		header.Set("X-Forwarded-For", fmt.Sprintf("%s,127.0.0.1", ip))

		res := GetUserIP(header)
		if res != ip {
			t.Errorf("expected '%s', got '%s'", ip, res)
		}
	})

	t.Run("it should return ali-cdn-real-ip if multiple header exist", func(t *testing.T) {
		ip := "202.147.207.71"

		header := http.Header{}
		header.Set("ali-cdn-real-ip", ip)
		header.Set("CF-Connecting-IP", ip)

		res := GetUserIP(header)

		if res != ip {
			t.Errorf("expected '%s', got '%s'", ip, res)
		}
	})
}
