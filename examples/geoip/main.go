package main

import (
	"bitbucket.org/rctiplus/vegapunk/geoip"
	"fmt"
	"net"
	"net/http"
)

func main() {
	// parse request header
	header := http.Header{}
	header.Set("X-Real-IP", "127.0.0.1")
	header.Set("X-Forwarded-For", "202.147.207.71,202.147.207.70")
	ipStr := geoip.GetUserIP(header)

	fmt.Println(ipStr)

	// get country by ip
	geoIPCountryPath := "GeoLite2-Country.mmdb"
	ip := net.ParseIP(ipStr)

	country, err := geoip.GetIPCountry(geoIPCountryPath, ip)
	if err != nil {
		_ = fmt.Errorf("error get IP Country: %v", err)
		return
	}

	fmt.Println(country)
}
