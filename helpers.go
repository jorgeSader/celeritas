package celeritas

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

const (
	randomString = "abccdeffghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_+"
)

// RandomString generates a random string of the specified length.
// It uses characters from the randomString constant and generates the output
// by selecting random indices based on prime numbers.
// The parameter n specifies the desired length of the output string.
// It returns the generated random string.
func (c *Celeritas) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomString)

	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

// CreateDirIfNotExist creates a directory at the specified path if it does not already exist.
// The directory is created with permissions set to 0755 (rwxr-xr-x).
// The parameter path specifies the directory location.
// It returns an error if the directory cannot be created, or nil if successful or if the directory already exists.
func (c *Celeritas) CreateDirIfNotExist(path string) error {
	const mode = 0o755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateFileIfNotExists creates a file at the specified path if it does not already exist.
// The parameter path specifies the file location.
// It returns an error if the file cannot be created, or nil if successful or if the file already exists.
// If the file is created, it is closed before the function returns.
func (c *Celeritas) CreateFileIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				return // Swallow close errors to avoid masking creation errors.
			}
		}(file)
	}
	return nil
}

type Encryption struct {
	Key []byte
}

// Encrypt encrypts the given plaintext string using AES encryption with the Encryption key.
// The parameter text is the plaintext to encrypt.
// It returns the encrypted text as a base64-encoded string and an error if encryption fails.
// The encryption uses CFB mode with a random initialization vector (IV) prepended to the ciphertext.
func (e *Encryption) Encrypt(text string) (string, error) {
	plainText := []byte(text)
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return base64.URLEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts the given base64-encoded ciphertext using AES encryption with the Encryption key.
// The parameter cryptoText is the base64-encoded encrypted text to decrypt.
// It returns the decrypted plaintext string and an error if decryption fails.
// The function assumes the ciphertext includes an initialization vector (IV) as its first aes.BlockSize bytes.
func (e *Encryption) Decrypt(cryptoText string) (string, error) {
	cypherText, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	if len(cypherText) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := cypherText[:aes.BlockSize]
	cypherText = cypherText[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cypherText, cypherText)

	return string(cypherText), nil
}
