package util

import (
	"testing"
	"time"
)

func TestSeqUID(t *testing.T) {
	t.Skip()
	uid := NewUIDSequenceNum()

	dup := make(map[int64]bool)
	count := 0
	for i := 0; i < 1000; i++ {
		id := uid.Generate()
		//fmt.Println(id)
		if dup[id] {
			count++
		}
		dup[id] = true
		time.Sleep(1 * time.Microsecond)
	}

	if count > 0 {
		t.Errorf("total duplicate %v", count)
	}
}

func TestRandUID(t *testing.T) {
	t.Skip()
	uid := NewUIDRandomNum()

	dup := make(map[int64]bool)
	count := 0
	for i := 0; i < 1000; i++ {
		id := uid.Generate()
		//fmt.Println(id)
		if dup[id] {
			count++
		}
		dup[id] = true
	}

	if count > 0 {
		t.Errorf("total duplicate %v", count)
	}
}

func TestRandNodeUID(t *testing.T) {
	t.Skip()
	dup := make(map[int64]bool)
	count := 0
	for i := 0; i < 1000; i++ {
		id := GenerateRandNodeUID()
		//fmt.Println(id)
		if dup[id] {
			count++
		}
		dup[id] = true
	}

	if count > 0 {
		t.Errorf("total duplicate %v", count)
	}
}
