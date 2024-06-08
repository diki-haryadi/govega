package event

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SafeBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}
func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}
func (b *SafeBuffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func TestDirectEmitter(t *testing.T) {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
	}
	ctx := context.Background()

	em, err := New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	require.Nil(t, em.Publish(ctx, "test", "t123", "testdata", nil))
	logStr := buf.String()

	assert.Contains(t, logStr, "key=t123")
	assert.Contains(t, logStr, "data=testdata")
	assert.Contains(t, logStr, "topic=test")

}

func TestHybridEmitter(t *testing.T) {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
		Writer: &DriverConfig{
			Type: "logger",
		},
	}
	ctx := context.Background()

	em, err := New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	buf := SafeBuffer{
		b: bytes.Buffer{},
		m: sync.Mutex{},
	}
	logrus.SetOutput(&buf)

	require.Nil(t, em.Publish(ctx, "test", "t123", "testdata", nil))
	time.Sleep(1 * time.Second)

	logStr := buf.String()

	assert.Contains(t, logStr, "key=t123")
	assert.Contains(t, logStr, "data=testdata")
	assert.Contains(t, logStr, "topic=test")
	assert.Contains(t, logStr, "message succesfully sent")

}

func TestMetadata(t *testing.T) {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
		EventConfig: &EventConfig{
			Metadata: map[string]map[string]interface{}{
				MetaDefault: {
					"foo": "bar",
				},
			},
		},
	}
	ctx := context.Background()

	em, err := New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	require.Nil(t, em.Publish(ctx, "test", "t123", "testdata", map[string]interface{}{
		"baz": "qux",
	}))
	logStr := buf.String()

	assert.Contains(t, logStr, "topic=test")
	assert.Contains(t, logStr, "foo:bar")
	assert.Contains(t, logStr, "baz:qux")

}
