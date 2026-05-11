package secret

import (
	"context"
	"errors"
	"mm/internal/model"
)

type AccountStorage struct {
	secretStorage Storage
}

func NewAccountStorage(secretStorage Storage) *AccountStorage {
	return &AccountStorage{secretStorage: secretStorage}
}

func (s *AccountStorage) Get(ctx context.Context, userID, exchangeAccountID, accountID uint64) (*model.Key, error) {
	path := CreateAccountSecretPath(userID, exchangeAccountID, accountID)

	data, err := s.secretStorage.GetSecret(ctx, path)

	if err != nil {
		return nil, err
	}

	apiKey, ok := data["api_key"].(string)

	if !ok {
		return nil, errors.New("api_key not found")
	}

	secretKey, ok := data["secret_key"].(string)
	if !ok {
		return nil, errors.New("secret_key not found")
	}

	passphrase, ok := data["passphrase"].(string)
	if !ok {
		return nil, errors.New("passphrase not found")
	}

	key := &model.Key{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
	}

	return key, nil
}

func (s *AccountStorage) Save(ctx context.Context, secret AccountSecret) error {
	path := CreateAccountSecretPath(secret.UserID, secret.ExchangeAccountID, secret.AccountID)

	data := map[string]interface{}{
		"api_key":    secret.ApiKey,
		"secret_key": secret.SecretKey,
		"passphrase": secret.Passphrase,
	}

	return s.secretStorage.SaveSecret(ctx, path, data)
}

func (s *AccountStorage) Delete(ctx context.Context, userID, exchangeAccountID, accountID uint64) error {
	path := CreateAccountMetedataPath(userID, exchangeAccountID, accountID)
	return s.secretStorage.DeleteSecret(ctx, path)
}
