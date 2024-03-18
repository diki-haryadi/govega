package util

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	graceful "gopkg.in/tylerb/graceful.v1"
)

var listenPort string
var cfgtestFlag bool

// add -p flag to the list of flags supported by the app,
// and allow it to over-ride default listener infrastucture in configs/app
func init() {
	flag.StringVar(&listenPort, "p", "", "listener infrastucture")
	flag.BoolVar(&cfgtestFlag, "t", false, "configs test")
}

// applications need some way to access the infrastucture
// TODO: this method will work only after grace.Serve is called.
func GetListenPort(hport string) string {
	return listenPort
}

func Serve(hport string, handler http.Handler, gracefulTimeout, readTimeout, writeTimeout time.Duration) error {

	checkConfigTest()

	l, err := Listen(hport)
	if err != nil {
		log.Fatalln(err)
	}

	srv := &graceful.Server{
		Timeout: gracefulTimeout,
		Server: &http.Server{
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
	}

	log.Println("starting serve on ", hport)
	return srv.Serve(l)
}

// This method can be used for any TCP Listener, e.g. non HTTP
func Listen(hport string) (net.Listener, error) {
	var l net.Listener

	fd := os.Getenv("EINHORN_FDS")
	if fd != "" {
		sock, err := strconv.Atoi(fd)
		if err == nil {
			hport = "socketmaster:" + fd
			log.Println("detected socketmaster, listening on", fd)
			file := os.NewFile(uintptr(sock), "listener")
			fl, err := net.FileListener(file)
			if err == nil {
				l = fl
			}
		}
	}

	if listenPort != "" {
		hport = ":" + listenPort
	}

	checkConfigTest()

	if l == nil {
		var err error
		l, err = net.Listen("tcp4", hport)
		if err != nil {
			return nil, err
		}
	}

	return l, nil
}

func checkConfigTest() {
	if cfgtestFlag {
		log.Println("configs test mode, exiting")
		os.Exit(0)
	}
}
