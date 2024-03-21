// To the extent possible under law, the Yawning Angel has waived all copyright
// and related or neighboring rights to orhttp_example, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package main

import (
	// Things needed by the actual interface.
	"golang.org/x/net/proxy"
	"net/http"
	"net/url"

	// Things needed by the example code.
	"fmt"
	"io/ioutil"
	"os"
)

func fatalf(fmtStr string, args interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args)
	os.Exit(-1)
}

func main() {
	// Create a transport that uses Tor Browser's SocksPort.  If
	// talking to a system tor, this may be an AF_UNIX socket, or
	// 127.0.0.1:9050 instead.
	tbProxyURL, err := url.Parse("socks5://127.0.0.1:9150")
	if err != nil {
		fatalf("Failed to parse proxy URL: %v\n", err)
	}

	// Get a proxy Dialer that will create the connection on our
	// behalf via the SOCKS5 proxy.  Specify the authentication
	// and re-create the dialer/transport/client if tor's
	// IsolateSOCKSAuth is needed.
	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		fatalf("Failed to obtain proxy dialer: %v\n", err)
	}

	// Make a http.Transport that uses the proxy dialer, and a
	// http.Client that uses the transport.
	tbTransport := &http.Transport{Dial: tbDialer.Dial}
	client := &http.Client{Transport: tbTransport}

	// Example: Fetch something.  Real code will probably want to use
	// client.Do() so they can change the User-Agent.
	resp, err := client.Get("http://check.torproject.org")
	if err != nil {
		fatalf("Failed to issue GET request: %v\n", err)
	}
	defer resp.Body.Close()

	fmt.Printf("GET returned: %v\n", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fatalf("Failed to read the body: %v\n", err)
	}
	fmt.Printf("----- Body -----\n%s\n----- Body -----", body)
}
