package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"gitlab.com/superman-tech/lib/config"
)

type GCP struct {
	ProjectID  string `json:"project_id" mapstructure:"project_id"`
	BucketName string `json:"bucket_name" mapstructure:"bucket_name"`
	Host       string `json:"host" mapstructure:"host"`
	ApiKey     string `json:"api_key" mapstructure:"api_key"`
	client     *storage.Client
}

func NewGCPStorage(cfg config.Getter) (StorageProvider, error) {
	var storeGcp GCP
	if err := cfg.Unmarshal(&storeGcp); err != nil {
		return nil, errors.New("[provider/gcp] failed unmarshal config")
	}

	var (
		cl  *storage.Client
		err error
	)

	if storeGcp.Host != "" {
		cl, err = storage.NewClient(
			context.Background(),
			option.WithEndpoint(storeGcp.Host),
			option.WithAPIKey(storeGcp.ApiKey),
		)
		if err != nil {
			log.Fatalf("[provider/gcp] failed to create client gcp, err : %+v\n", err)
		}
	} else {
		cl, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalf("[provider/gcp] failed to create client gcp, err : %+v\n", err)
		}
	}

	ctx := context.Background()
	bucket := cl.Bucket(storeGcp.BucketName)

	exists, err := bucket.Attrs(ctx)
	if err != nil && err != storage.ErrBucketNotExist {
		return nil, errors.New("[provider/gcp] no exists bucket")
	}

	// create bucket if doesn't exist
	if exists == nil && err == storage.ErrBucketNotExist {
		err = bucket.Create(
			ctx,
			storeGcp.ProjectID,
			&storage.BucketAttrs{
				Name: storeGcp.BucketName,
			},
		)
		if err != nil {
			fmt.Println("[provider/gcp] error creating bucket", err)
			return nil, err
		}

	}

	storeGcp.client = cl
	return &storeGcp, nil
}

func (gcp *GCP) Get(ctx context.Context, path string) (io.Reader, error) {
	reader, err := gcp.client.Bucket(gcp.BucketName).Object(path).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (gcp *GCP) Put(ctx context.Context, path string, object io.Reader) error {
	attrs, err := gcp.client.Bucket(gcp.BucketName).Object(path).Attrs(ctx)
	if err != nil && err != storage.ErrObjectNotExist {
		return errors.New("[provider/gcp] error while getting object attr")
	}

	if attrs != nil {
		return errors.New("[provider/gcp] object already exist")
	}

	wc := gcp.client.Bucket(gcp.BucketName).Object(path).NewWriter(ctx)
	if _, err := io.Copy(wc, object); err != nil {
		return fmt.Errorf("io.Copy: %+v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %+v", err)
	}

	return nil
}
