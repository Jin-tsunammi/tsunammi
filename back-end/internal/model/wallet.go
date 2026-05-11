package model

import (
	"mm/internal/client/solanarpc"
	"time"

	"github.com/uptrace/bun"
)

type Wallet struct {
	bun.BaseModel `bun:"table:wallets,alias:w" swaggerignore:"true"`

	ID         uint64       `json:"id" db:"id" bun:"id,pk,autoincrement" example:"1234567890"`
	PublicKey  string       `json:"public_key" db:"public_key" bun:"public_key" example:"88888888SOL88888888SOL88888888SOL8888888"`
	PrivateKey string       `json:"private_key" bun:"-" example:"55bjLUoWhf2dFWnGEeLzCBdf5AuuJgWwRh7HghN2LoimnmjWz3qNgmni64x6nM3uTWqRNAET2cwef9pz21Zv4C2S"`
	UserID     uint64       `json:"-" db:"user_id" bun:"user_id"`
	Status     WalletStatus `json:"-" db:"status" bun:"status,notnull"`
	ProjectIDs []uint64     `json:"-" bun:"project_ids,array,scanonly"`

	BalanceToken float64 `json:"-" bun:"-"`
	BalanceSOL   float64 `json:"-" bun:"-"`
	BalanceUSD   float64 `json:"-" bun:"-"`

	CreatedAt time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp" example:"2025-10-01T03:12:20Z"`
}

func (Wallet) TableName() string {
	return "wallets"
}

type WalletWithoutBalance struct {
	ID        uint64 `json:"id" example:"1234567890"`
	PublicKey string `json:"public_key" example:"88888888SOL88888888SOL88888888SOL8888888"`

	CreatedAt time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp" example:"2025-10-01T03:12:20Z"`
}

type WalletResponse struct {
	ID               uint64        `json:"id" example:"1234567890"`
	PublicKey        string        `json:"public_key" example:"88888888SOL88888888SOL88888888SOL8888888"`
	BalanceSOL       float64       `json:"balance_sol" example:"100.294"`
	BalanceUSD       float64       `json:"balance_usd" example:"200294.00"`
	TokensBalanceSOL float64       `json:"tokens_balance_sol" example:"100.294"`
	TokensBalanceUSD float64       `json:"tokens_balance_usd" example:"200294.00"`
	Tokens           []WalletToken `json:"tokens"`
	Rent             float64       `json:"rent" example:"5"`
	CreatedAt        time.Time     `json:"created_at" example:"2025-10-01T03:12:20Z"`
}

type WalletMintResponse struct {
	ID               uint64    `json:"id" example:"1234567890"`
	PublicKey        string    `json:"public_key" example:"88888888SOL88888888SOL88888888SOL8888888"`
	TokensBalanceSOL float64   `json:"tokens_balance_sol" example:"100.294"`
	TokensBalance    float64   `json:"tokens_balance" example:"200294.00"`
	CreatedAt        time.Time `json:"created_at" example:"2025-10-01T03:12:20Z"`
}

type WalletToken struct {
	Mint       string  `json:"token_symbol" example:"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"`
	Balance    float64 `json:"balance" example:"100"`
	BalanceUSD float64 `json:"balance_usd" example:"99.95"`
	BalanceSOL float64 `json:"balance_sol" example:"99.95"`
}

type GenerateWalletsReq struct {
	Count      int      `json:"count" binding:"required,min=1"`
	ProjectIDs []uint64 `json:"project_ids" binding:"required"`
}

type WalletDecrypted struct {
	ID         uint64 `json:"id" db:"id" bun:"id,pk,autoincrement"`
	PublicKey  string `json:"public_key" db:"public_key" bun:"public_key"`
	PrivateKey string `json:"private_key" db:"private_key" bun:"private_key"`
}

type ImportWalletsReq struct {
	PrivateKeys []string `json:"private_keys" binding:"required"`
	ProjectIDs  []uint64 `json:"project_ids" binding:"required"`
}

type MonitorWalletsReq struct {
	WalletIDs []uint64 `json:"wallet_ids" db:"wallet_id" bun:"wallet_id,pk"`
	PageSize  int      `json:"page_size"`
	Before    string   `json:"before"` // base58 encoded transaction signature
}

type MonitorWalletsResp struct {
	WalletID     uint64                        `json:"wallet_id" example:"1234567890"`
	PublicKey    string                        `json:"public_key" example:"88888888SOL88888888SOL88888888SOL8888888"`
	Balance      float64                       `json:"balance" example:"100.294"`
	Transactions []solanarpc.WalletTransaction `json:"transactions"`
}

type PrivateKey struct {
	PrivateKey string `json:"private_key" binding:"required" example:"55bjLUoWhf2dFWnGEeLzCBdf5AuuJgWwRh7HghN2LoimnmjWz3qNgmni64x6nM3uTWqRNAET2cwef9pz21Zv4C2S"`
}

type SolanaVerifyRequest struct {
	PublicAddress string `json:"publicAddress" binding:"required"`
	SignedMessage string `json:"signedMessage" binding:"required"`
}

type TransferReq struct {
	FromWalletID  uint64  `json:"from_wallet_id" validate:"required"`
	ToWalletID    uint64  `json:"to_wallet_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	TipAmount     float64 `json:"tip_amount" validate:"required,gt=0"`
	LamportsPerCU int64   `json:"lamports_per_cu" validate:"required"`
	CU            int64   `json:"cu" validate:"required"`
}
