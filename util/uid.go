package util

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/diki-haryadi/govega/constant"
)

var nodeID int64
var maxSeq = int64(math.Pow(2, constant.UIDSequenceBits)) - 1
var maxNode = int64(math.Pow(2, constant.UIDNodeBits)) - 1
var seq int64
var mux = sync.Mutex{}

type NodeIDGenerator func() int64
type SequenceGenerator func() int64

type UID struct {
	NodeGen NodeIDGenerator
	SeqGen  SequenceGenerator
}

func EncodeUID(uid int64) string {
	buf := make([]byte, 9)
	binary.PutVarint(buf, uid)
	return base64.StdEncoding.EncodeToString(buf[:8])
}

// GenerateRandNodeUID generate UID from random NodeID and random sequence number
func GenerateRandNodeUID() int64 {
	return (&UID{
		NodeGen: GetRandomNodeID,
		SeqGen:  GetRandomNumber,
	}).Generate()
}

// GenerateRandUID generate UID with NodeID from network interface and random sequence number
func GenerateRandUID() int64 {
	return NewUIDRandomNum().Generate()
}

// GenerateSeqUID generate UID with NodeID from network interface and incremental sequence number
func GenerateSeqUID() int64 {
	return NewUIDSequenceNum().Generate()
}

func NewUIDRandomNum() *UID {
	return &UID{
		NodeGen: GetNodeIDFromMac,
		SeqGen:  GetRandomNumber,
	}
}

func NewUIDSequenceNum() *UID {
	return &UID{
		NodeGen: GetNodeIDFromMac,
		SeqGen:  GetSequenceNumber,
	}
}

func (u *UID) Generate() int64 {
	return CalculateUID(u.NodeGen(), u.SeqGen())
}

func CalculateUID(node, sequence int64) int64 {
	t := time.Now().UnixNano()
	nid := node << constant.UIDSequenceBits
	id := t<<constant.UIDNodeBits + constant.UIDSequenceBits
	return id | nid | sequence
}

func GetSequenceNumber() int64 {
	mux.Lock()
	defer mux.Unlock()
	seq++
	if seq > maxSeq {
		seq = 1
	}
	//fmt.Println(seq)
	return seq
}

func GetRandomNumber() int64 {
	rand.Seed(time.Now().UnixNano())
	return int64(rand.Intn(int(maxSeq)))
}

func GetRandomNodeID() int64 {
	rand.Seed(time.Now().UnixNano())
	return int64(rand.Intn(int(maxNode)))
}

func GetNodeIDFromMac() int64 {
	if nodeID > 0 {
		return nodeID
	}

	madr, err := getMacAddr()
	if err != nil {
		return -1
	}

	iadr, err := binary.ReadVarint(bytes.NewBuffer(madr))
	if err != nil {
		return -1
	}

	nodeID = maxNode & iadr
	return nodeID
}

func getMacAddr() ([]byte, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			return []byte(ifa.HardwareAddr), nil
		}
	}
	return nil, errors.New("no network interface or failed to read hardware address")
}
