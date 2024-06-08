package zklock

import (
	"log"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var (
	zkConn  *zk.Conn
	hosts   []string
	timeout time.Duration
)

func getZkConn() *zk.Conn {
	return zkConn
}

func reConnectZk() error {
	return Connect(hosts, timeout)
}

// Connect conncet to ZK cluster
func Connect(_hosts []string, zkTimeOut time.Duration) error {
	var err error
	timeout = zkTimeOut
	hosts = _hosts
	zkConn, _, err = zk.Connect(hosts, timeout)
	if err != nil {
		return err
	}
	return nil
}

// PConnect connect to ZK cluster
func PConnect(_hosts []string, zkTimeOut time.Duration) error {
	var err error
	timeout = zkTimeOut
	hosts = _hosts
RECONNECT:
	zkConn = nil
	zkConn, _, err = zk.Connect(hosts, timeout)
	if err != nil {
		time.Sleep(3 * time.Second)
		log.Println("EstablishZkConn  ", err.Error())
		goto RECONNECT
	}
	return err
}

func createPath(path string) {
	getZkConn().Create(path, []byte(""), int32(0), zk.WorldACL(zk.PermAll))
}

// Close close connection
func Close() {
	zkConn.Close()
}
