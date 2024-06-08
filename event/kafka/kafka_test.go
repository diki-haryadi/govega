package kafka

import (
	"context"
	"testing"

	"github.com/diki-haryadi/govega/event"
	"github.com/stretchr/testify/require"
)

func TestPublish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	event.RegisterSender("kafka", NewKafkaSender)

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "kafka",
			Config: map[string]interface{}{
				"brokers": []string{"localhost:9092"},
			},
		},
	}

	ctx := context.Background()

	em, err := event.New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	require.Nil(t, em.Publish(ctx, "test", "t123", "testdata", nil))
}
