package provider

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path"

	"github.com/diki-haryadi/govega/config"
	"github.com/diki-haryadi/govega/file"
	"github.com/diki-haryadi/govega/log"
)

type SLocal struct {
	Filepath  string `json:"filepath" mapstructure:"filepath"`
	Secret    string `json:"secret"`
	encryptor file.Encryptor
}

func NewSLocalStorage(conf config.Getter) (StorageProvider, error) {
	var slocal *SLocal
	if err := conf.Unmarshal(&slocal); err != nil {
		return nil, err
	}

	err := os.MkdirAll(slocal.Filepath, os.ModePerm)
	if err != nil {
		log.WithError(err).Errorln(slocal)
		return nil, err
	}

	slocal.encryptor = file.NewAESEncryptor(slocal.Secret)

	return slocal, nil
}

func (l *SLocal) Get(ctx context.Context, fullpath string) (io.Reader, error) {
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

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		if err := l.encryptor.Decrypt(data, pw); err != nil {
			log.WithError(err).Errorln("Error Decrypting")
		}
	}()

	return pr, nil
}

func (l *SLocal) Put(ctx context.Context, fullpath string, f io.Reader) (err error) {
	var (
		ff *os.File
		r  io.Reader
	)
	pr, pw := io.Pipe()
	fn := path.Join(l.Filepath, fullpath)

	switch file := f.(type) {
	case *os.File:
		log.Infoln("os.File")
		// If renaming fails we try the normal copying method.
		// Renaming could fail if the files are on different devices.
		ff = file
		ff.Seek(0, 0)
		r = ff
	case multipart.File:
		log.Infoln("multipart.File")
		// when f cannot cast to *os.File and f is multipart.sectionReadCloser
		file.Seek(0, 0)
		r = file
	default:
		log.Infoln("default")
		r = file
	}

	go func() {
		defer pw.Close()
		if err := l.encryptor.Encrypt(r, pw); err != nil {
			log.WithError(err).Errorln("Error Ecrypting")
		}
	}()

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

	_, err = copyZeroAlloc(ff, pr)
	if err != nil {
		log.WithError(err).Errorln("[Put] Error when save the file")
		return err
	}

	return nil
}
