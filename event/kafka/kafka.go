package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/diki-haryadi/govega/event"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.opentelemetry.io/otel/attribute"
)

const (
	KafkaPartitionKey     = attribute.Key("messaging.kafka.partition")
	KafkaConsumerGroupKey = attribute.Key("messaging.kafka.consumer_group")
)

func init() {
	event.RegisterSender("kafka", NewKafkaSender)
	event.RegisterListener("kafka", NewKafkaListener)
}

func dial(certFile, keyFile, caCert, username, password, authType string) (*kafka.Dialer, error) {
	dialer := kafka.DefaultDialer
	if certFile != "" && keyFile != "" && caCert != "" {
		keypair, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}

		caCert, err := ioutil.ReadFile(caCert)

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, fmt.Errorf("failed to parse CA Certificate file: %s", err)
		}
		dialer.TLS = &tls.Config{
			Certificates: []tls.Certificate{keypair},
			RootCAs:      caCertPool,
		}
	}

	if username != "" && password != "" {
		switch authType {
		case "scram":
			mechanism, err := scram.Mechanism(scram.SHA512, username, password)
			if err != nil {
				return nil, err
			}
			dialer.SASLMechanism = mechanism
		default:
			dialer.SASLMechanism = plain.Mechanism{
				Username: username,
				Password: password,
			}
		}

	}

	return dialer, nil
}
