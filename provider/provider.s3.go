package provider

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gitlab.com/superman-tech/lib/config"
)

type S3 struct {
	Endpoint     string `json:"endpoint"`
	AccessKey    string `json:"access_key" mapstructure:"access_key"`
	SecretKey    string `json:"secret_key" mapstructure:"secret_key"`
	SSL          bool   `json:"ssl"`
	Bucket       string `json:"bucket"`
	BucketRegion string `json:"bucket_region" mapstructure:"bucket_region"`
	client       *minio.Client
}

func NewS3Storage(conf config.Getter) (StorageProvider, error) {
	var storage S3
	if err := conf.Unmarshal(&storage); err != nil {
		return nil, err
	}

	s3Client, err := minio.New(storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(storage.AccessKey, storage.SecretKey, ""),
		Secure: storage.SSL,
	})
	if err != nil {
		//fmt.Println("error connection to s3 ", err)
		return nil, err
	}

	ctx := context.Background()

	ok, err := s3Client.BucketExists(ctx, storage.Bucket)
	if err != nil {
		//fmt.Println("error check if bucket exist ", err)
		return nil, err
	}

	if !ok {
		reg := storage.BucketRegion
		if reg == "" {
			reg = "us-east-1"
		}
		if err := s3Client.MakeBucket(ctx, storage.Bucket, minio.MakeBucketOptions{Region: reg}); err != nil {
			//fmt.Println("error creating bucket", err)
			return nil, err
		}
	}
	storage.client = s3Client

	return &storage, nil
}

func (s *S3) Get(ctx context.Context, path string) (io.Reader, error) {
	out, err := s.client.GetObject(ctx, s.Bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *S3) Put(ctx context.Context, path string, f io.Reader) error {

	info, _ := s.client.StatObject(ctx, s.Bucket, path, minio.StatObjectOptions{})
	if info.Size > 0 {
		return errors.New("object already exist")
	}

	var buf bytes.Buffer
	f2 := io.TeeReader(f, &buf)
	size, err := getSize(f2)
	if err != nil {
		return err
	}

	opt := minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	}

	_, err = s.client.PutObject(ctx, s.Bucket, path, &buf, int64(size), opt)
	return err
}
