package secret

import (
	"context"
	"errors"
)

type KeyStorage struct {
	secretStorage Storage
}

func NewKeyStorage(secretStorage Storage) *KeyStorage {
	return &KeyStorage{secretStorage: secretStorage}
}

func (s *KeyStorage) Get(ctx context.Context, userID uint64, publicKey string) (string, error) {
	path := CreateWalletSecretPath(userID, publicKey)
	data, err := s.secretStorage.GetSecret(ctx, path)

	if err != nil {
		return "", err
	}

	privateKey, ok := data["privateKey"].(string)

	if !ok {
		return "", errors.New("failed to get private key")
	}

	return privateKey, nil
}

func (s *KeyStorage) Save(ctx context.Context, secret KeySecret) error {
	path := CreateWalletSecretPath(secret.UserID, secret.PublicKey)
	data := map[string]interface{}{
		"privateKey": secret.PrivateKey,
		"createdAt":  secret.CreatedAt,
	}

	return s.secretStorage.SaveSecret(ctx, path, data)
}

func (s *KeyStorage) Delete(ctx context.Context, userID uint64, publicKey string) error {
	path := CreateWalletSecretMetadataPath(userID, publicKey)
	return s.secretStorage.DeleteSecret(ctx, path)
}
