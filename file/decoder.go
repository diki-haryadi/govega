package file

import (
	"context"
	"errors"
	"mime/multipart"
)

var (
	ErrMultipartFileHeader = errors.New("not multipart file header")
)

type Multipart struct {
}

type MultipartDecoder interface {
	Decode(ctx context.Context, object interface{}) (multipart.File, error)
}

func NewMultipartDecoder() MultipartDecoder {
	return &Multipart{}
}

func (m *Multipart) Decode(ctx context.Context, object interface{}) (multipart.File, error) {
	var (
		f   multipart.File
		err error
	)

	fh, ok := object.(*multipart.FileHeader)
	if !ok {
		return nil, ErrMultipartFileHeader
	}

	f, err = fh.Open()
	if err != nil {
		return nil, err
	}

	return f, nil
}
