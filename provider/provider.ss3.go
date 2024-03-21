package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gitlab.com/superman-tech/lib/config"
	"gitlab.com/superman-tech/lib/file"
)

type SS3 struct {
	Endpoint     string `json:"endpoint"`
	AccessKey    string `json:"access_key" mapstructure:"access_key"`
	SecretKey    string `json:"secret_key" mapstructure:"secret_key"`
	SSL          bool   `json:"ssl"`
	Bucket       string `json:"bucket"`
	BucketRegion string `json:"bucket_region" mapstructure:"bucket_region"`
	Secret       string `json:"secret"`
	encryptor    file.Encryptor
	client       *minio.Client
}

func NewSS3Storage(conf config.Getter) (StorageProvider, error) {
	var storage SS3
	if err := conf.Unmarshal(&storage); err != nil {
		return nil, err
	}

	s3Client, err := minio.New(storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(storage.AccessKey, storage.SecretKey, ""),
		Secure: storage.SSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	ok, err := s3Client.BucketExists(ctx, storage.Bucket)
	if err != nil {
		return nil, err
	}

	if !ok {
		reg := storage.BucketRegion
		if reg == "" {
			reg = "us-east-1"
		}
		if err := s3Client.MakeBucket(ctx, storage.Bucket, minio.MakeBucketOptions{Region: reg}); err != nil {
			return nil, err
		}
	}
	storage.client = s3Client
	storage.encryptor = file.NewAESEncryptor(storage.Secret)

	return &storage, nil
}

func (s *SS3) Get(ctx context.Context, path string) (io.Reader, error) {
	out, err := s.client.GetObject(ctx, s.Bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		if err := s.encryptor.Decrypt(out, pw); err != nil {
			fmt.Println(err)
		}
	}()

	return pr, nil
}

func (s *SS3) Put(ctx context.Context, path string, f io.Reader) error {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		if err := s.encryptor.Encrypt(f, pw); err != nil {
			fmt.Println(err)
		}
	}()

	var buf bytes.Buffer
	f2 := io.TeeReader(pr, &buf)
	size, err := getSize(f2)
	if err != nil {
		return err
	}

	info, _ := s.client.StatObject(ctx, s.Bucket, path, minio.StatObjectOptions{})
	if info.Size > 0 {
		return errors.New("object already exist")
	}

	opt := minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	}

	_, err = s.client.PutObject(ctx, s.Bucket, path, &buf, int64(size), opt)
	return err
}
