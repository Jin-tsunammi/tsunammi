package secret

import "time"

type KeySecret struct {
	UserID     uint64
	PublicKey  string
	PrivateKey string
	CreatedAt  time.Time
}

type AccountSecret struct {
	AccountID         uint64
	UserID            uint64
	ExchangeAccountID uint64
	ApiKey            string
	SecretKey         string
	Passphrase        string
}

type value struct {
	Data map[string]any
}
