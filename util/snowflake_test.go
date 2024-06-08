package util

//base on for detail https://github.com/bwmarrin/snowflake

import (
	"sync"
	"testing"
)

func TestNewPOD(t *testing.T) {

	_, err := NewPOD(&SnowflakeOpts{
		POD: 0,
	})
	if err != nil {
		t.Fatalf("error creating NewPOD, %s", err)
	}

	_, err = NewPOD(&SnowflakeOpts{
		POD: 5000,
	})
	if err == nil {
		t.Fatalf("no error creating NewPOD, %s", err)
	}

}

func TestGenerateDuplicateID(t *testing.T) {
	pod, _ := NewPOD(&SnowflakeOpts{
		POD: 1,
	})

	max := 1000
	intChan := make(chan ID, max)

	wg := sync.WaitGroup{}
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func(ch chan ID) {
			defer wg.Done()
			ch <- pod.Generate()
		}(intChan)
	}

	go func() {
		wg.Wait()
		close(intChan)
	}()

	result := map[ID]bool{}
	for val := range intChan {
		if result[val] {
			t.Fatalf("value %v is duplicate", val)
		}
		result[val] = true
	}

}

func TestInt64(t *testing.T) {
	node, err := NewPOD(&SnowflakeOpts{
		POD: 1,
	})
	if err != nil {
		t.Fatalf("error creating NewPOD, %s", err)
	}

	oID := node.Generate()
	i := oID.Int64()

	pID := ParseInt64(i)
	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	mi := int64(1116766490855473152)
	pID = ParseInt64(mi)
	if pID.Int64() != mi {
		t.Fatalf("pID %v != mi %v", pID.Int64(), mi)
	}

}

func TestString(t *testing.T) {
	node, err := NewPOD(&SnowflakeOpts{
		POD: 1,
	})
	if err != nil {
		t.Fatalf("error creating NewPOD, %s", err)
	}

	oID := node.Generate()
	si := oID.String()

	pID, err := ParseString(si)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := `1116766490855473152`
	_, err = ParseString(ms)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	ms = `1112316766490855473152`
	_, err = ParseString(ms)
	if err == nil {
		t.Fatalf("no error parsing %s", ms)
	}
}

func TestPrintAll(t *testing.T) {
	pod, err := NewPOD(&SnowflakeOpts{
		Epoch: 1626670674000,
		POD:   1,
	})
	if err != nil {
		t.Fatalf("error creating NewPOD, %s", err)
	}

	id := pod.Generate()

	t.Logf("Int64    : %#v", id.Int64())
	t.Logf("String   : %#v", id.String())
	t.Logf("Base2    : %#v", id.Base2())
	t.Logf("Base32   : %#v", id.Base32())
	t.Logf("Base36   : %#v", id.Base36())
	t.Logf("Base58   : %#v", id.Base58())
	t.Logf("Base64   : %#v", id.Base64())
	t.Logf("Bytes    : %#v", id.Bytes())
	t.Logf("IntBytes : %#v", id.IntBytes())
}
