package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PumpfunPrepareCreateTxRequest struct {
	Name            string            `json:"name"`
	Ticker          string            `json:"ticker"`
	Description     string            `json:"description"`
	BuyInSol        float64           `json:"buy_in_sol"`
	Mayhem          bool              `json:"mayhem"`
	CashbackRewards bool              `json:"cashback_rewards"`
	WalletBuys      []WalletBuyConfig `json:"wallet_buys"`

	Twitter  string `json:"twitter"`
	Discord  string `json:"discord"`
	Website  string `json:"website"`
	Telegram string `json:"telegram"`

	OwnerPublicKey string `json:"owner"`

	Logo *multipart.FileHeader `json:"-" swaggerignore:"true"`
}

type Metadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Image       string `json:"image"`
	ShowName    bool   `json:"showName"`
	CreatedOn   string `json:"createdOn"`
	Twitter     string `json:"twitter"`
	Telegram    string `json:"telegram"`
	Discord     string `json:"discord"`
	Website     string `json:"website"`
}

type PumpfunPrepareCreateTxResponse struct {
	ID                uuid.UUID `json:"id"`
	CreateTransaction string    `json:"create_transaction"` // base64 encoded, partially signed (mint only); contains Jito tip
	BuyTransaction    string    `json:"buy_transaction"`    // base64 encoded, unsigned; empty when owner buy is zero
	MintPubkey        string    `json:"mint_pubkey"`
}

type PumpfunLaunchStatus string

const (
	PumpfunLaunchStatusPending PumpfunLaunchStatus = "PENDING"
	PumpfunLaunchStatusSuccess PumpfunLaunchStatus = "SUCCESS"
	PumpfunLaunchStatusFailed  PumpfunLaunchStatus = "FAILED"
)

type PumpfunLaunch struct {
	bun.BaseModel `bun:"table:pumpfun_launches"`

	ID           uuid.UUID           `bun:"id,pk"`
	UserID       uint64              `bun:"user_id"`
	CreateTx     []byte              `bun:"create_tx"`
	BuyTx        []byte              `bun:"buy_tx,nullzero"`
	WalletBuyTxs []string            `bun:"wallet_buy_txs,array"`
	MintPubkey   string              `bun:"mint_pubkey"`
	Signer       string              `bun:"signer"`
	Status       PumpfunLaunchStatus `bun:"status"`
	CreatedAt    time.Time           `bun:"created_at"`
	UpdatedAt    time.Time           `bun:"updated_at"`
	ExpiresAt    time.Time           `bun:"expires_at"`
}

type WalletBuyConfig struct {
	WalletID  uint64  `json:"wallet_id"`
	AmountSol float64 `json:"amount_sol"`
}

type PumpfunEstimateCreateRequest struct {
	OwnerPublicKey string            `json:"owner"`
	BuyInSol       float64           `json:"buy_in_sol"`
	WalletBuys     []WalletBuyConfig `json:"wallet_buys"`
}

type PumpfunEstimateCreateResponse struct {
	TotalTokensOut       float64 `json:"total_tokens_out"`
	CreationFeeSOL       float64 `json:"creation_fee_sol"`
	JitoTipSOL           float64 `json:"jito_tip_sol"`
	PriorityFeeSOL       float64 `json:"priority_fee_sol"`
	PumpfunCommissionSOL float64 `json:"pumpfun_commission_sol"`
}

type PumpfunProcessCreateRequest struct {
	TxID           uuid.UUID `json:"tx_id"`
	MintPubkey     string    `json:"mint_pubkey"`
	SignedCreateTx string    `json:"signed_create_tx"` // base64 encoded full signed tx
	SignedBuyTx    string    `json:"signed_buy_tx"`    // base64 encoded, owner-signed; required when buy_transaction is not empty
}

type PumpfunProcessCreateResponse struct {
	MintPubkey string              `json:"mint_pubkey"`
	Status     PumpfunLaunchStatus `json:"status"`
	Signatures []string            `json:"signatures"`
}
