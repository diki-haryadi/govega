package provider

import (
	"context"
	"errors"
	"io"

	"github.com/diki-haryadi/govega/log"
)

type InactiveProvider struct {
	ProviderName string `json:"provider_name"`
}

func NewInactiveProvider(providerName string) StorageProvider {
	return &InactiveProvider{
		ProviderName: providerName,
	}
}

func (ip *InactiveProvider) Get(ctx context.Context, path string) (io.Reader, error) {
	log.Errorf("Error Inactive Storage Provider: %s", ip.ProviderName)
	return nil, errors.New("inactive storage provider")
}

func (ip *InactiveProvider) Put(ctx context.Context, path string, object io.Reader) error {
	log.Errorf("Error Inactive Storage Provider: %s", ip.ProviderName)
	return errors.New("inactive storage provider")
}
