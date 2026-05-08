package model

import (
	"time"

	"mm/pkg/mtype"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SmartBuybackCampaign struct {
	bun.BaseModel `bun:"table:buyback_campaigns,alias:bc" swaggerignore:"true"`

	ID            uuid.UUID      `bun:"id" json:"id"`
	UserID        uint64         `bun:"user_id" json:"user_id"`
	ProviderID    SwapProviderID `bun:"provider_id" json:"provider_id"`
	PoolID        string         `bun:"pool_id" json:"pool_id"`
	PoolProgramID string         `bun:"pool_program_id" json:"pool_program_id"`
	ProjectID     uint64         `bun:"project_id" json:"project_id"`
	TokenMint     string         `bun:"token_mint" json:"token_mint"`
	Status        BuybackStatus  `bun:"status" json:"status"`
	CreatedAt     time.Time      `bun:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `bun:"updated_at" json:"updated_at"`
}

type SmartBuybackCampaignWithTargets struct {
	bun.BaseModel `bun:"table:buyback_campaigns,alias:bc" swaggerignore:"true"`

	SmartBuybackCampaign

	Targets []SmartBuybackCampaignTarget `bun:"-" json:"targets"`
}

type BuybackCampaignTargetType string

const (
	BuybackCampaignTargetTypeBuy  = "BUY"
	BuybackCampaignTargetTypeSell = "SELL"
)

type SmartBuybackCampaignTarget struct {
	bun.BaseModel `bun:"table:buyback_targets,alias:bt" swaggerignore:"true"`

	ID                         uuid.UUID                 `bun:"id,pk" json:"id"`
	CampaignID                 uuid.UUID                 `bun:"campaign_id" json:"campaign_id"`
	Type                       BuybackCampaignTargetType `bun:"type" json:"type"`
	TargetPrice                mtype.BigRat              `bun:"target_price" json:"target_price" swaggertype:"string"`
	Budget                     mtype.BigRat              `bun:"budget" json:"budget" swaggertype:"string"`
	RemainingBudget            mtype.BigRat              `bun:"remaining_budget" json:"remaining_budget" swaggertype:"string"`
	Slippage                   uint                      `bun:"slippage" json:"slippage"`
	MinTransactionAmount       mtype.BigRat              `bun:"min_transactions_amount" json:"min_transaction_amount" swaggertype:"string"`
	MaxTransactionAmount       mtype.BigRat              `bun:"max_transactions_amount" json:"max_transaction_amount" swaggertype:"string"`
	ParallelTransactionsAmount uint                      `bun:"parallel_transactions_amount" json:"parallel_transactions_amount"`
	MinTimeBetweenTransactions time.Duration             `bun:"min_time_between_transactions" json:"min_time_between_transactions" swaggertype:"integer"`
	MaxTimeBetweenTransactions time.Duration             `bun:"max_time_between_transactions" json:"max_time_between_transactions" swaggertype:"integer"`
	CreatedAt                  time.Time                 `bun:"created_at" json:"created_at"`
	UpdatedAt                  time.Time                 `bun:"updated_at" json:"updated_at"`
	StartAt                    time.Time                 `bun:"start_at" json:"start_at"`
	TransactionSpeed           TransactionSpeed          `bun:"transaction_speed" json:"transaction_speed"`
	UsingJito                  bool                      `bun:"using_jito" json:"using_jito"`
	PriorityFee                mtype.BigRat              `bun:"priority_fee" json:"priority_fee" swaggertype:"string"`
	Status                     BuybackStatus             `bun:"status" json:"status"`
}

type BuybackTransaction struct {
	bun.BaseModel `bun:"table:buyback_transactions,alias:bbtx" swaggerignore:"true"`

	ID              uint64        `bun:"id,pk,autoincrement" json:"id"`
	CampaignID      uuid.UUID     `bun:"campaign_id" json:"campaign_id"`
	TargetID        uuid.UUID     `bun:"target_id" json:"target_id"`
	TransactionHash string        `bun:"transaction_hash" json:"transaction_hash"`
	PoolID          string        `bun:"pool_id" json:"pool_id"`
	TokenMintFrom   string        `bun:"token_mint_from" json:"token_mint_from"`
	TokenMintTo     string        `bun:"token_mint_to" json:"token_mint_to"`
	AddressFrom     string        `bun:"address_from" json:"address_from"`
	AddressTo       string        `bun:"address_to" json:"address_to"`
	AmountTokenFrom mtype.BigRat  `bun:"amount_token_from" json:"amount_token_from" swaggertype:"string"`
	AmountTokenTo   mtype.BigRat  `bun:"amount_token_to" json:"amount_token_to" swaggertype:"string"`
	Status          BuybackStatus `bun:"status" json:"status"`
	Message         string        `bun:"message" json:"message"`
	DebugMessage    *string       `bun:"debug_message,nullzero" json:"debug_message,omitempty"`
	CreatedAt       time.Time     `bun:"created_at" json:"created_at"`
}

type GetSmartBuybackCampaignsResponse struct {
	Page      int                    `json:"page"`
	PageSize  int                    `json:"page_size"`
	Total     int                    `json:"total"`
	Campaigns []SmartBuybackCampaign `json:"campaigns"`
}

type CreateSmartBuybackCampaignRequest struct {
	ProviderID SwapProviderID                    `json:"provider_id" validate:"required"`
	ProjectID  uint64                            `json:"project_id" validate:"required"`
	TokenMint  string                            `json:"token_mint" validate:"required"`
	Targets    []CreateSmartBuybackTargetRequest `json:"targets" validate:"required,dive"`
}

type CreateSmartBuybackTargetRequest struct {
	Type                       BuybackCampaignTargetType `json:"type" validate:"required"`
	TargetPrice                string                    `json:"target_price" validate:"required"`
	Budget                     string                    `json:"budget" validate:"required"`
	PriorityFee                string                    `json:"priority_fee" validate:"required"`
	Slippage                   uint                      `json:"slippage" validate:"required"`
	MinTransactionAmount       string                    `json:"min_transaction_amount" validate:"required"`
	MaxTransactionAmount       string                    `json:"max_transaction_amount" validate:"required"`
	ParallelTransactionsAmount uint                      `json:"parallel_transactions_amount" validate:"required"`
	MinTimeBetweenTransactions time.Duration             `json:"min_time_between_transactions" validate:"required" swaggertype:"integer"`
	MaxTimeBetweenTransactions time.Duration             `json:"max_time_between_transactions" validate:"required" swaggertype:"integer"`
	StartAt                    int64                     `json:"start_at" validate:"required"`
	TransactionSpeed           TransactionSpeed          `json:"transaction_speed"`
	UsingJito                  bool                      `json:"using_jito"`
	CampaignID                 uuid.UUID                 `json:"-" swaggerignore:"true"`
}

type UpdateSmartBuybackTargetRequest struct {
	TargetPrice                string           `json:"target_price" validate:"required"`
	Budget                     string           `json:"budget" validate:"required"`
	PriorityFee                string           `json:"priority_fee" validate:"required"`
	Slippage                   *uint            `json:"slippage" validate:"required"`
	MinTransactionAmount       string           `json:"min_transaction_amount" validate:"required"`
	MaxTransactionAmount       string           `json:"max_transaction_amount" validate:"required"`
	ParallelTransactionsAmount *uint            `json:"parallel_transactions_amount" validate:"required"`
	MinTimeBetweenTransactions *time.Duration   `json:"min_time_between_transactions" validate:"required" swaggertype:"integer"`
	MaxTimeBetweenTransactions *time.Duration   `json:"max_time_between_transactions" validate:"required" swaggertype:"integer"`
	TransactionSpeed           TransactionSpeed `json:"transaction_speed"`
	UsingJito                  *bool            `json:"using_jito"`
	StartAt                    *int64           `json:"start_at"`

	ID uuid.UUID `json:"-" swaggerignore:"true"`
}

type GetAllBuybackCampaignsRequest struct {
	Page   int      `query:"page" example:"1" default:"1"`
	Size   int      `query:"pageSize" example:"20" default:"20"`
	Status []string `query:"status"`
}

func (r *GetAllBuybackCampaignsRequest) Validate() {
	if r.Page <= 0 {
		r.Page = 1
	}

	if r.Size <= 0 {
		r.Size = 20
	}

	r.Size = min(r.Size, 100)
}
