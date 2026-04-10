package crypto_test

import (
	"math/rand"
	"mm/internal/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func TestEncryptor(t *testing.T) {
	encryptor, err := crypto.NewEncryptor([]byte("0123456789abcdef"))
	assert.NoError(t, err, "Failed to create encryptor")

	testString := randomString(100)

	encrypted, err := encryptor.Encrypt(testString)
	assert.NoError(t, err, "Encryption should not return an error")

	decrypted, err := encryptor.Decrypt(encrypted)
	assert.NoError(t, err)

	assert.Equal(t, testString, decrypted, "Decrypted string should match original")
}

func TestNewEncryptor_InvalidKey(t *testing.T) {
	_, err := crypto.NewEncryptor([]byte("short"))
	assert.Error(t, err)
}
