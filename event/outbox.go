package event

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// OutboxRecord outbox model
type OutboxRecord struct {
	ID        string    `json:"id,omitempty" mapstructure:"id"`
	Topic     string    `json:"topic,omitempty" mapstructure:"topic"`
	Key       string    `json:"key,omitempty" mapstructure:"key"`
	Value     string    `json:"value,omitempty" mapstructure:"value"`
	CreatedAt time.Time `json:"created_at,omitempty" mapstructure:"created_at"`
}

// Hash calculate request hash
func (o *OutboxRecord) Hash() []byte {
	val := fmt.Sprintf("%v", o)
	h := sha256.Sum256([]byte(val))
	return h[:]
}

// GenerateID generate record ID
func (o *OutboxRecord) GenerateID() *OutboxRecord {
	h := o.Hash()
	o.ID = base64.StdEncoding.EncodeToString(h[:])
	o.CreatedAt = time.Now()
	return o
}

func OutboxFromMessage(msg *EventMessage) (*OutboxRecord, error) {
	mb, err := msg.ToBytes()
	if err != nil {
		return nil, err
	}
	return (&OutboxRecord{
		Topic: msg.Topic,
		Key:   msg.Key,
		Value: string(mb),
	}).GenerateID(), nil
}
