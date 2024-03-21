package file

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

const (
	aesOverhead      = 28
	defaultChunkSize = 32
)

type AESEncryptor struct {
	Secret    string
	ChunkSize int
}

type Encryptor interface {
	Encrypt(plain io.Reader, cipher io.Writer) error
	Decrypt(cipher io.Reader, plain io.Writer) error
}

func NewAESEncryptor(secret string) Encryptor {
	return &AESEncryptor{
		Secret:    secret,
		ChunkSize: defaultChunkSize,
	}
}

func (a *AESEncryptor) Encrypt(data io.Reader, out io.Writer) error {
	chunk := make([]byte, a.ChunkSize)
	for {
		n, rerr := data.Read(chunk)

		if n > 0 {
			if n != a.ChunkSize {
				chunk = chunk[:n]
			}

			o, err := encrypt(chunk, a.Secret)
			if err != nil {
				return err
			}

			if _, err := out.Write(o); err != nil {
				return err
			}
		}

		if rerr == io.EOF {
			break
		}

	}

	return nil
}

func (a *AESEncryptor) Decrypt(data io.Reader, out io.Writer) error {
	chunk := make([]byte, a.ChunkSize+aesOverhead)
	for {
		n, rerr := data.Read(chunk)
		if n > 0 {

			if n != a.ChunkSize+aesOverhead {
				chunk = chunk[:n]
			}

			o, err := decrypt(chunk, a.Secret)
			if err != nil {
				return err
			}

			if _, err := out.Write(o); err != nil {
				return err
			}
		}

		if rerr == io.EOF {
			break
		}

	}

	return nil
}

func createHash(key string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func encrypt(data []byte, passphrase string) ([]byte, error) {
	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func decrypt(data []byte, passphrase string) ([]byte, error) {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
