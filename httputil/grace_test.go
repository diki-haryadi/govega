package httputil

import (
	"log"
	"net/http"
	"time"
)

func ExampleServe() {
	timeout := 1 * time.Second
	http.HandleFunc("/foo/bar", foobarHandler)
	log.Fatal(Serve(":9000", nil, timeout, timeout, timeout))
}

func foobarHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("foobar"))
}
