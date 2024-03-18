package httprq

import (
	"crypto/tls"
	"net/http"
)

var (
	requestManager RequestManager
)

func init() {
	SetRequestManager(&RequestClient{
		Client: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		}},
		Timeout: 5,
	})
}

func SetRequestManager(rc RequestManager) {
	requestManager = rc
}

func GetRequestManager() RequestManager {
	return requestManager
}
