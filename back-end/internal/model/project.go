package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Project struct {
	bun.BaseModel `bun:"table:projects,alias:p" swaggerignore:"true"`

	ID          uint64    `json:"id" db:"id" bun:"id,pk,autoincrement" example:"1234567890"`
	Name        string    `json:"name" db:"name" bun:"name,unique,notnull" example:"My Project"`
	UserID      uint64    `json:"user_id" db:"user_id" bun:"user_id" example:"1234567890"`
	CreatedAt   time.Time `json:"created_at" bun:"created_at,notnull,nullzero,default:current_timestamp" example:"2025-10-01T03:12:20Z"`
	WalletCount uint64    `json:"wallet_count" bun:"-" example:"10"`
}

type ProjectWithWallets struct {
	bun.BaseModel `bun:"table:projects,alias:p" swaggerignore:"true"`

	Project

	BalanceSOL float64 `json:"balance_sol" bun:"-"`

	TotalBalanceSOL float64 `json:"total_balance_sol" bun:"-"`
	TotalBalanceUSD float64 `json:"total_balance_usd" bun:"-"`

	Wallets []Wallet `json:"wallets" bun:"-"`
}

type ProjectWithWalletsWithoutBalance struct {
	Project

	LastSync time.Time              `json:"last_sync" bun:"-"`
	Wallets  []WalletWithoutBalance `json:"wallets" bun:"-"`
}

type ProjectsWithPaginationResponse struct {
	Projects []ProjectWithWalletsResponse `json:"projects"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
	Total    int                          `json:"total"`
}

type ProjectsWithoutBalanceWithPaginationResponse struct {
	Projects []ProjectWithWalletsWithoutBalance `json:"projects"`
	Page     int                                `json:"page"`
	PageSize int                                `json:"page_size"`
	Total    int                                `json:"total"`
}

type ProjectWithWalletsResponse struct {
	ID              uint64    `json:"id" example:"1234567890"`
	Name            string    `json:"name" example:"My Project"`
	UserID          uint64    `json:"user_id" example:"1234567890"`
	WalletCount     int64     `json:"wallet_count" example:"10"`
	LastSync        time.Time `json:"last_sync" example:"2025-10-01T03:12:20Z"`
	BalanceSOL      float64   `json:"balance_sol" example:"1000.294"`
	TotalBalanceSOL float64   `json:"total_balance_sol" example:"1000.294"`
	TotalBalanceUSD float64   `json:"total_balance_usd" example:"2000294.00"`
	RentTotal       float64   `json:"rent_total" example:"5"`
	CreatedAt       time.Time `json:"created_at" example:"2025-10-01T03:12:20Z"`

	Wallets []WalletResponse `json:"wallets"`
}

type ProjectWithMintWalletsResponse struct {
	ID              uint64    `json:"id" example:"1234567890"`
	Name            string    `json:"name" example:"My Project"`
	UserID          uint64    `json:"user_id" example:"1234567890"`
	WalletCount     uint64    `json:"wallet_count" example:"10"`
	LastSync        time.Time `json:"last_sync" example:"2025-10-01T03:12:20Z"`
	TotalBalanceSOL float64   `json:"total_balance_sol" example:"1000.294"`
	TotalBalance    float64   `json:"total_balance" example:"2000294.00"`
	CreatedAt       time.Time `json:"created_at" example:"2025-10-01T03:12:20Z"`

	Wallets []WalletMintResponse `json:"wallets"`
}

type ProjectsWithMintPaginationResponse struct {
	Projects []ProjectWithMintWalletsResponse `json:"projects"`
	Page     int                              `json:"page"`
	PageSize int                              `json:"page_size"`
	Total    int                              `json:"total"`
}

type CreateProjectReq struct {
	Name string `json:"name" validate:"required"`
}

type EditProjectReq struct {
	Name string `json:"name" validate:"required"`
}

type ProjectWallet struct {
	bun.BaseModel `bun:"table:project_wallets,alias:pw" swaggerignore:"true"`

	ProjectID uint64 `json:"project_id" bun:"project_id,notnull"`
	WalletID  uint64 `json:"wallet_id" bun:"wallet_id,notnull"`
}
