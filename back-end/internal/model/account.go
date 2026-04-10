package model

import (
	"mm/internal/crypto"
	"time"

	"github.com/uptrace/bun"
)

type AddExchangeAccountReq struct {
	Name       string `json:"name" binding:"required"`
	ExchangeID int64  `json:"exchange_id" binding:"required"`
	ApiKey     string `json:"api_key" binding:"required"`
	SecretKey  string `json:"secret_key" binding:"required"`
	Passphrase string `json:"passphrase" binding:"required"`
}

type AccountResponse struct {
	ID               uint64    `json:"id" example:"1234567890"`
	Name             string    `json:"name" example:"My Account"`
	Exchange         string    `json:"exchange" example:"Binance"`
	AccountId        uint64    `json:"account_id" example:"1234567890"`
	CreatedAt        time.Time `json:"created_at" example:"2025-10-01T03:12:20Z"`
	ApiName          string    `json:"api_name" example:"kucoin"`
	WithdrawLimit    int       `json:"withdraw_limit" example:"25"`
	TotalDepositsSOL float64   `json:"total_deposits_sol" example:"1000.294"`
	TotalDepositsUSD float64   `json:"total_deposits_usd" example:"2000294.00"`
	Status           Status    `json:"status" example:"active"`
}

type AccountsWithPaginationResponse struct {
	Accounts []AccountResponse `json:"accounts"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int               `json:"total"`
}

type Account struct {
	bun.BaseModel `bun:"table:accounts,alias:a" swaggerignore:"true"`

	ID                uint64  `json:"id" bun:"id,pk,autoincrement" example:"1234567890"`
	Name              string  `json:"name" bun:"name" example:"My Account"`
	UserID            uint64  `json:"user_id" bun:"user_id" example:"1234567890"`
	ExchangeID        int64   `json:"exchange_id" bun:"exchange_id" example:"1"`
	ExchangeName      string  `json:"exchange_name" bun:"-" example:"Binance"`
	ExchangeAccountId uint64  `json:"exchange_account_id" bun:"exchange_account_id" example:"1234567890"`
	ExchangeApiName   string  `json:"api_name" bun:"api_name" example:"kucoin"`
	Status            Status  `json:"status" bun:"status" example:"active"`
	DepositBalance    float64 `json:"-" bun:"-" example:"1000.294"`
	WithdrawLimit     int     `json:"withdraw_limit" bun:"withdraw_limit" example:"25"`

	Key       `json:",inline" bun:"-"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero" example:"2025-10-01T03:12:20Z"`
}

func (_ Account) TableName() string {
	return "accounts"
}

type Key struct {
	ApiKey     string `json:"api_key" bun:"api_key" example:"ZXc4Vb7N6mK5jH4gF3dS2aQ1wE9rT8yU"`
	SecretKey  string `json:"secret_key" bun:"secret_key" example:"9f7e6d5c4b3a2910ffeeddccbbaa99887766554433221100aabbccddeeff0011"`
	Passphrase string `json:"passphrase" bun:"passphrase" example:"dev_passphrase_!@#"`
}

func (k *Key) Decrypt(c crypto.Encryptor) {
	apiKey, _ := c.Decrypt(k.ApiKey)
	secretKey, _ := c.Decrypt(k.SecretKey)
	passphrase, _ := c.Decrypt(k.Passphrase)
	k.ApiKey = apiKey
	k.SecretKey = secretKey
	k.Passphrase = passphrase
}

func (k *Key) Encrypt(c crypto.Encryptor) {
	k.ApiKey, _ = c.Encrypt(k.ApiKey)
	k.SecretKey, _ = c.Encrypt(k.SecretKey)
	k.Passphrase, _ = c.Encrypt(k.Passphrase)
}
