package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mm/config"
)

//go:generate mockgen -source=encryptor.go -destination=mocks/encryptor_mock.go -package=cryptoMocks

// EncryptorInterface defines the interface for encryption operations
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(encodedCiphertext string) (string, error)
}

// Encryptor handles encryption and decryption of sensitive data
type encryptor struct {
	key []byte
}

func NewWalletEncryptor(cfg *config.Config) (Encryptor, error) {
	return NewEncryptor([]byte(cfg.Crypto.WalletEncryptionKey))
}

func NewAccountEncryptor(cfg *config.Config) (Encryptor, error) {
	return NewEncryptor([]byte(cfg.Crypto.AccountEncryptionKey))
}

// NewEncryptor creates a new encryptor with the provided key
// Key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
func NewEncryptor(key []byte) (Encryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: %d bytes. Key must be 16, 24, or 32 bytes", len(key))
	}
	return &encryptor{key: key}, nil
}

// Encrypt encrypts plaintext and returns hex encoded ciphertext
func (e *encryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts hex encoded ciphertext and returns plaintext string
func (e *encryptor) Decrypt(encodedCiphertext string) (string, error) {
	ciphertext, err := hex.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
