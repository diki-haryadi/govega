package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func dial(certFile, keyFile, caCert, username, password, authType string) (*kafka.Dialer, error) {
	dialer := kafka.DefaultDialer

	certOK := certFile != "" && keyFile != "" && caCert != ""
	if certOK {
		keypair, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}

		caCert, err := ioutil.ReadFile(caCert)
		caCertPool := x509.NewCertPool()
		caCertPoolOk := caCertPool.AppendCertsFromPEM(caCert)
		if !caCertPoolOk {
			return nil, fmt.Errorf("failed to parse CA Certificate file : %s", err.Error())
		}

		dialer.TLS = &tls.Config{
			Certificates: []tls.Certificate{keypair},
			RootCAs:      caCertPool,
		}
	}

	authzOK := username != "" && password != ""
	if authzOK {
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
