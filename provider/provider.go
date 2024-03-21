package provider

import (
	"bytes"
	"context"
	"io"
)

type StorageProvider interface {
	Get(ctx context.Context, path string) (io.Reader, error)
	Put(ctx context.Context, path string, object io.Reader) error
}

func getSize(stream io.Reader) (int, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Len(), nil
}
