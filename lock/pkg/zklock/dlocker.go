package zklock

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

const defaultInterval = 100

var ErrAcquireLock = errors.New("lock is already acquired by another client")

// Dlocker distributed locker
type Dlocker struct {
	lockerPath string
	prefix     string
	basePath   string
	timeout    time.Duration
	innerLock  *sync.Mutex
	interval   int
}

// NewLocker create instance of ZK lock
func NewLocker(path string, timeout time.Duration) (*Dlocker, error) {

	isExsit, _, err := getZkConn().Exists(path)
	if err != nil {
		return nil, err
	}

	if !isExsit {
		//log.Println("create the znode:" + path)
		if _, err := getZkConn().Create(path, []byte(""), int32(0), zk.WorldACL(zk.PermAll)); err != nil {
			return nil, err
		}
	}
	return &Dlocker{
		basePath:  path,
		prefix:    "lock-",
		timeout:   timeout,
		innerLock: &sync.Mutex{},
		interval:  defaultInterval,
	}, nil
}

func (d *Dlocker) createZnodePath() (string, error) {
	path := d.basePath + "/" + d.prefix
	//save the create unixTime into znode
	nowUnixTime := time.Now().Unix()
	nowUnixTimeBytes := []byte(strconv.FormatInt(nowUnixTime, 10))
	return getZkConn().Create(path, nowUnixTimeBytes, zk.FlagSequence|zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
}

// get the path of minimum serial number znode from sequential children
func (d *Dlocker) getMinZnodePath() (string, error) {
	children, err := d.getPathChildren()
	if err != nil {
		return "", err
	}
	minSNum := getMinSerialNumber(children, d.prefix)
	minZnodePath := d.basePath + "/" + children[minSNum]
	return minZnodePath, nil
}

// get the children of basePath znode
func (d *Dlocker) getPathChildren() ([]string, error) {
	children, _, err := getZkConn().Children(d.basePath)
	return children, err
}

// get the last znode of created znode
func (d *Dlocker) getLastZnodePath() string {
	return getLastNodeName(d.lockerPath,
		d.basePath, d.prefix)
}

// TryLock try lock and return error immediately if fail
func (d *Dlocker) TryLock() error {
	d.innerLock.Lock()
	defer d.innerLock.Unlock()
	//create a znode for the locker path
	var err error
	d.lockerPath, err = d.createZnodePath()
	if err != nil {
		return err
	}

	//get the znode which get the lock
	minZnodePath, err := d.getMinZnodePath()
	if err != nil {
		return err
	}

	if minZnodePath == d.lockerPath {
		// if the created node is the minimum znode, getLock success
		return nil
	}
	return ErrAcquireLock
}

// Lock just list mutex.Lock()
func (d *Dlocker) Lock() error {
	max := int(d.timeout) / d.interval
	for i := 0; i < max; i++ {
		s, err := d.lock()
		if err == nil {
			if s {
				return nil
			}
			return errors.New("Timeout reached")
		}
		time.Sleep(time.Duration(d.interval) * time.Millisecond)
	}
	return errors.New("Max retry reached")
}

// Unlock just list mutex.Unlock(), return false when zookeeper connection error or locker timeout
func (d *Dlocker) Unlock() error {
	return d.unlock()
}

func (d *Dlocker) lock() (bool, error) {
	defer func() {
		e := recover()
		if e == zk.ErrConnectionClosed {
			//try reconnect the zk server
			log.Println("connection closed, reconnect to the zk server")
			if err := reConnectZk(); err != nil {
				log.Println("Error reconnecting ", err)
			}
		}
	}()
	d.innerLock.Lock()
	defer d.innerLock.Unlock()
	//create a znode for the locker path
	var err error
	d.lockerPath, err = d.createZnodePath()
	if err != nil {
		return false, err
	}

	//get the znode which get the lock
	minZnodePath, err := d.getMinZnodePath()
	if err != nil {
		return false, err
	}

	if minZnodePath == d.lockerPath {
		// if the created node is the minimum znode, getLock success
		return true, nil
	}

	// if the created znode is not the minimum znode,
	// listen for the last znode delete notification
	lastNodeName := d.getLastZnodePath()
	watchPath := d.basePath + "/" + lastNodeName
	isExist, _, watch, err := getZkConn().ExistsW(watchPath)
	if err != nil {
		return false, err
	}
	if isExist {
		select {
		//get lastNode been deleted event
		case event := <-watch:
			if event.Type == zk.EventNodeDeleted {
				//check out the lockerPath existence
				isExist, _, err = getZkConn().Exists(d.lockerPath)
				if err != nil {
					return false, err
				}
				if isExist {
					//checkout the minZnodePath is equal to the lockerPath
					minZnodePath, err := d.getMinZnodePath()
					if err != nil {
						return false, err
					}
					if minZnodePath == d.lockerPath {
						return true, nil
					}
				}
			}
		//time out
		case <-time.After(d.timeout):
			// if timeout, delete the timeout znode
			children, err := d.getPathChildren()
			if err != nil {
				return false, err
			}
			for _, child := range children {
				data, _, err := getZkConn().Get(d.basePath + "/" + child)
				if err != nil {
					continue
				}
				if checkOutTimeOut(data, d.timeout) {
					err := getZkConn().Delete(d.basePath+"/"+child, 0)
					if err == nil {
						log.Println("timeout delete:", d.basePath+"/"+child)
					}
				}
			}
			return false, nil
		}
	} else {
		// recheck the min znode
		// the last znode may be deleted too fast to let the next znode cannot listen to it deletion
		minZnodePath, err := d.getMinZnodePath()
		if err != nil {
			return false, err
		}
		if minZnodePath == d.lockerPath {
			return true, nil
		}
	}

	return false, nil
}

func (d *Dlocker) unlock() error {
	defer func() {
		e := recover()
		if e == zk.ErrConnectionClosed {
			//try reconnect the zk server
			log.Println("connection closed, reconnect to the zk server")
			reConnectZk()
		}
	}()
	err := getZkConn().Delete(d.lockerPath, 0)
	if err == zk.ErrNoNode {
		return err
	}

	return nil
}
