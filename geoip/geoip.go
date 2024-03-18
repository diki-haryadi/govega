package geoip

import (
	"github.com/oschwald/geoip2-golang"
	"net"
	"net/http"
	"strings"
)

// GetIPCountry returns the country ISO Code of the given IP
//
// geoIPCountryPath is the path to the GeoLite2-Country.mmdb
// ip is the IP address
//
// returns the country ISO Code
func GetIPCountry(geoIPCountryPath string, ip net.IP) (string, error) {
	db, err := geoip2.Open(geoIPCountryPath)
	if err != nil {
		return "", err
	}
	defer func(db *geoip2.Reader) {
		err := db.Close()
		if err != nil {
			return
		}
	}(db)

	record, err := db.Country(ip)
	if err != nil {
		return "", err
	}

	return record.Country.IsoCode, nil
}

// GetUserIP returns the user IP
//
// header is the HTTP request header
func GetUserIP(header http.Header) string {
	headerList := [...]string{
		"ali-cdn-real-ip",
		"CF-Connecting-IP",
		"X-Forwarded-For",
		"X-Forwarded",
		"X-Original-Forwarded-For",
		"Forwarded-For",
		"Forwarded",
		"True-Client-Ip",
		"X-Client-IP",
		"Fastly-Client-Ip",
		"X-Real-IP",
	}

	var result string
	for _, h := range headerList {
		if header.Get(h) == "" {
			continue
		}

		result = header.Get(h)
		if h == "X-Forwarded-For" {
			result = strings.Split(result, ",")[0]
			return result
		}

		return result
	}

	return result
}
