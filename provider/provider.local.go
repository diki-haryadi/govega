package provider

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path"
	"sync"

	"gitlab.com/superman-tech/lib/config"
	"gitlab.com/superman-tech/lib/log"
)

var copyBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

type Local struct {
	Filepath string `json:"filepath" mapstructure:"filepath"`
}

func NewLocalStorage(conf config.Getter) (StorageProvider, error) {
	var local *Local
	if err := conf.Unmarshal(&local); err != nil {
		return nil, err
	}

	err := os.MkdirAll(local.Filepath, os.ModePerm)
	if err != nil {
		log.WithError(err).Errorln(local)
		return nil, err
	}

	return local, nil
}

func (l *Local) Get(ctx context.Context, fullpath string) (io.Reader, error) {
	var (
		data *os.File
		err  error
	)

	fn := path.Join(l.Filepath, fullpath)

	data, err = os.Open(fn)
	if err != nil {
		return nil, err
	}
	// defer data.Close()

	return data, nil
}

func (l *Local) Put(ctx context.Context, fullpath string, f io.Reader) (err error) {
	var ff *os.File
	fn := path.Join(l.Filepath, fullpath)

	switch file := f.(type) {
	case *os.File:
		// If renaming fails we try the normal copying method.
		// Renaming could fail if the files are on different devices.
		ff = file
		if os.Rename(ff.Name(), fn) == nil {
			return nil
		}
	case multipart.File:
		// when f cannot cast to *os.File and f is multipart.sectionReadCloser
		file.Seek(0, 0)
	}

	ff, err = os.Create(fn)
	if err != nil {
		return err
	}
	defer func() {
		e := ff.Close()
		if err == nil {
			err = e
		}
	}()

	_, err = copyZeroAlloc(ff, f)
	return err
}

func copyZeroAlloc(w io.Writer, r io.Reader) (int64, error) {
	vbuf := copyBufPool.Get()
	buf := vbuf.([]byte)
	n, err := io.CopyBuffer(w, r, buf)
	copyBufPool.Put(vbuf)
	return n, err
}
