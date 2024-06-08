package event

import (
	"context"
	"errors"
	"testing"
)

func TestAlwaysCommitStrategy(t *testing.T) {
	tests := []struct {
		name              string
		handler           EventHandler
		expectedCommitted bool
	}{
		{
			name: "successs handle message",
			handler: func(ctx context.Context, message *EventConsumeMessage) error {
				return nil
			},
			expectedCommitted: true,
		},
		{
			name: "failed handle message",
			handler: func(ctx context.Context, message *EventConsumeMessage) error {
				return errors.New("failed")
			},
			expectedCommitted: true,
		},
	}
	for _, tt := range tests {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.name, func(t *testing.T) {
			cm := newTestConsumeMessage(&EventConsumeMessage{})
			AlwaysCommitStrategy(context.Background(), cm, tt.handler)

			if cm.committed != tt.expectedCommitted {
				t.Fatalf("committed [%v] not equals to expected [%v]",
					cm.committed, tt.expectedCommitted)
			}
		})
	}
}

func TestCommitOnSuccessStrategy(t *testing.T) {
	tests := []struct {
		name              string
		handler           EventHandler
		expectedCommitted bool
	}{
		{
			name: "successs handle message",
			handler: func(ctx context.Context, message *EventConsumeMessage) error {
				return nil
			},
			expectedCommitted: true,
		},
		{
			name: "failed handle message",
			handler: func(ctx context.Context, message *EventConsumeMessage) error {
				return errors.New("failed")
			},
			expectedCommitted: false,
		},
	}
	for _, tt := range tests {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.name, func(t *testing.T) {
			cm := newTestConsumeMessage(&EventConsumeMessage{})
			CommitOnSuccessStrategy(context.Background(), cm, tt.handler)

			if cm.committed != tt.expectedCommitted {
				t.Fatalf("committed [%v] not equals to expected [%v]",
					cm.committed, tt.expectedCommitted)
			}
		})
	}
}
