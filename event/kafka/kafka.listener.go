package kafka

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"

	"bitbucket.org/rctiplus/vegapunk/event"
)

type (
	KafkaListener struct {
		Brokers       []string `json:"brokers" mapstructure:"brokers"`
		KeyFile       string   `json:"key_file" mapstructure:"key_file"`
		CertFile      string   `json:"cert_file" mapstructure:"cert_file"`
		CACertificate string   `json:"ca_cert" mapstructure:"ca_cert"`
		AuthType      string   `json:"auth_type" mapstructure:"auth_type"`
		Username      string   `json:"username" mapstructure:"username"`
		Password      string   `json:"password" mapstructure:"password"`

		QueueCapacity          int    `json:"queue_capacity" mapstructure:"queue_capacity"`
		MinBytes               int    `json:"min_bytes" mapstructure:"min_bytes"`
		MaxBytes               int    `json:"max_bytes" mapstructure:"max_bytes"`
		MaxWait                string `json:"max_wait" mapstructure:"max_wait"`
		MaxAttempts            int    `json:"max_attempts" mapstructure:"max_attempts"`
		ReadlagInterval        string `json:"read_lag_interval" mapstructure:"read_lag_interval"`
		HeartbeatInterval      string `json:"heartbeat_interval" mapstructure:"heartbeat_interval"`
		CommitInterval         string `json:"commit_interval" mapstructure:"commit_interval"`
		PartitionWatchInterval string `json:"partition_watch_interval" mapstructure:"partition_watch_interval"`
		WatchPartitionChanges  bool   `json:"watch_partition_changes" mapstructure:"watch_partition_changes"`
		SessionTimeout         string `json:"session_timeout" mapstructure:"session_timeout"`
		RebalanceTimeout       string `json:"rebalance_timeout" mapstructure:"rebalance_timeout"`
		JoinGroupBackoff       string `json:"join_group_backoff" mapstructure:"join_group_backoff"`
		RetentionTime          string `json:"retention_time" mapstructure:"retention_time"`
		StartOffset            int64  `json:"start_offset" mapstructure:"start_offset"`
		ReadBackoffMin         string `json:"read_backoff_min" mapstructure:"read_backoff_min"`
		ReadBackoffMax         string `json:"read_backoff_max" mapstructure:"read_backoff_max"`
	}
)

func NewKafkaListener(_ context.Context, config interface{}) (event.Listener, error) {
	var kafListener KafkaListener
	if err := mapstructure.Decode(config, &kafListener); err != nil {
		return nil, fmt.Errorf("[kafka/listener] failed to decode config: %w", err)
	}

	return &kafListener, nil
}

func (k *KafkaListener) Listen(ctx context.Context, topic, group string) (event.Iterator, error) {
	return newKafkaIterator(k, topic, group)
}
