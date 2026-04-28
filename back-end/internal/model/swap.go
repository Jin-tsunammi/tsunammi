package model

import (
	"math/big"
	"mm/pkg/mtype"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	StatusInUse             = "in_use"
	StatusDone              = "done"
	StatusBudgetDone        = "budget_done"
	StatusInsufficientFunds = "insufficient_funds"
	StatusStop              = "stop"
	StatusError             = "error"
	TargetUpTaskType        = "up"
	TargetDownTaskType      = "down"
)

type TransactionSpeed string

const (
	Default TransactionSpeed = "default"
	Fast    TransactionSpeed = "fast"
	Extra   TransactionSpeed = "extra"
)

type PriorityFee string

const (
	PriorityFeeMin    PriorityFee = "Min"
	PriorityFeeLow    PriorityFee = "Low"
	PriorityFeeMedium PriorityFee = "Medium"
	PriorityFeeHigh   PriorityFee = "High"
)

type SwapProviderID uint8

const (
	SwapProviderRaydium SwapProviderID = 1
	SwapProviderPumpfun SwapProviderID = 2
)

type SwapCampaign struct {
	bun.BaseModel `bun:"table:swap_campaigns,alias:swapc" swaggerignore:"true"`

	ID                         uuid.UUID        `bun:"id,pk" json:"id"`
	UserID                     uint64           `bun:"user_id" json:"-"`
	CampaignTypeID             uint64           `bun:"type_id" json:"-"`
	CampaignType               *CampaignType    `bun:"rel:belongs-to,join:type_id=id" json:"type"`
	ProjectID                  uint64           `bun:"project_id" json:"project_id"`
	PoolID                     string           `bun:"pool_id" json:"pool_id" swaggertype:"string"`
	ProviderID                 uint64           `bun:"provider_id" json:"provider_id"`
	TokenMintFrom              string           `bun:"token_mint_from" json:"token_mint_from" swaggertype:"string"`
	TokenMintTo                string           `bun:"token_mint_to" json:"token_mint_to" swaggertype:"string"`
	Budget                     float64          `bun:"budget" json:"budget"`
	SlippageBPS                uint64           `bun:"slippage" json:"slippage_bps"`
	GoalPercentChange          float64          `bun:"-" json:"goal_percent_change"`
	GoalBPSChange              uint64           `bun:"goal_bps_change" json:"-"`
	StartedPrice               mtype.BigRat     `bun:"started_price" json:"started_price" swaggertype:"string"`
	GoalPrice                  mtype.BigRat     `bun:"goal_price" json:"goal_price" swaggertype:"string"`
	CurrentPrice               mtype.BigRat     `bun:"-" json:"current_price" swaggertype:"string"`
	Status                     string           `bun:"status" json:"status"`
	ParallelTransactionsAmount int              `bun:"parallel_transactions_amount" json:"parallel_transactions_amount"`
	MinTransactionsBudget      float64          `bun:"min_transactions_budget" json:"min_transactions_budget"`
	MaxTransactionsBudget      float64          `bun:"max_transactions_budget" json:"max_transactions_budget"`
	MinTimeBetweenTransactions time.Duration    `bun:"min_time_between_transactions" json:"min_time_between_transactions" swaggertype:"integer"`
	MaxTimeBetweenTransactions time.Duration    `bun:"max_time_between_transactions" json:"max_time_between_transactions" swaggertype:"integer"`
	CreatedAt                  time.Time        `bun:"created_at" json:"created_at"`
	UpdatedAt                  time.Time        `bun:"updated_at" json:"updated_at"`
	TransactionSpeed           TransactionSpeed `bun:"transaction_speed" json:"transaction_speed"`
	UsingJito                  bool             `bun:"using_jito" json:"using_jito"`
	PriorityFee                float64          `bun:"priority_fee" json:"priority_fee"`
}

type CampaignsWithPaginationResponse struct {
	Campaigns []SwapCampaign `json:"campaigns"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
	Total     int            `json:"total"`
}

type CampaignType struct {
	bun.BaseModel `bun:"table:swap_campaign_types,alias:swapct" swaggerignore:"true"`

	ID   uint64 `bun:"id,pk,autoincrement" json:"-"`
	Name string `bun:"name" json:"name"`
}

type EstimatePullRequest struct {
	SourceTokenMint  solana.PublicKey `json:"source_token_mint" validate:"required" swaggertype:"string"`
	DestTokenMint    solana.PublicKey `json:"dest_token_mint" validate:"required" swaggertype:"string"`
	ProjectID        uint64           `json:"project_id" validate:"required"`
	Budget           float64          `json:"budget" validate:"required,gt=0"`
	Slippage         float64          `json:"slippage" validate:"required,gt=0.1,lt=100"`
	TransactionSpeed TransactionSpeed `json:"transaction_speed" validate:"required,oneof=default fast extra"`

	ProviderID SwapProviderID `json:"-" swaggerignore:"true"`
}

type TargetPullUpRequest struct {
	DestTokenMint              solana.PublicKey `json:"dest_token_mint" swaggertype:"string"`
	BudgetPercent              float64          `json:"budget_percent"`
	ProjectID                  uint64           `json:"project_id"`
	Budget                     float64          `json:"budget"`
	Slippage                   float64          `json:"slippage"`
	GoalPercentageChange       float64          `json:"goal_percentage_change"`
	ParallelTransactionsAmount int              `json:"parallel_transactions_amount"`
	MinTransactionsBudget      float64          `json:"min_transactions_budget"`
	MaxTransactionsBudget      float64          `json:"max_transactions_budget"`
	MinTimeBetweenTransactions time.Duration    `json:"min_time_between_transactions" swaggertype:"integer"`
	MaxTimeBetweenTransactions time.Duration    `json:"max_time_between_transactions" swaggertype:"integer"`
	TransactionSpeed           TransactionSpeed `json:"transaction_speed"`
	UsingJito                  bool             `json:"using_jito"`
	PriorityFee                float64          `json:"priority_fee" validate:"gte=0"`

	ProviderID SwapProviderID `json:"-" swaggerignore:"true"`
}

type TargetPullDownRequest struct {
	SourceTokenMint            solana.PublicKey `json:"source_token_mint" swaggertype:"string"`
	BudgetPercent              float64          `json:"budget_percent"`
	ProjectID                  uint64           `json:"project_id"`
	Budget                     float64          `json:"budget"`
	Slippage                   float64          `json:"slippage"`
	GoalPercentageChange       float64          `json:"goal_percentage_change"`
	ParallelTransactionsAmount int              `json:"parallel_transactions_amount"`
	MinTransactionsBudget      float64          `json:"min_transactions_budget"`
	MaxTransactionsBudget      float64          `json:"max_transactions_budget"`
	MinTimeBetweenTransactions time.Duration    `json:"min_time_between_transactions" swaggertype:"integer"`
	MaxTimeBetweenTransactions time.Duration    `json:"max_time_between_transactions" swaggertype:"integer"`
	TransactionSpeed           TransactionSpeed `json:"transaction_speed"`
	UsingJito                  bool             `json:"using_jito"`
	PriorityFee                float64          `json:"priority_fee" validate:"gte=0"`

	ProviderID SwapProviderID `json:"-" swaggerignore:"true"`
}

type CampaignRequest struct {
	Budget                     *float64          `json:"budget" validate:"omitempty,gt=0"`
	Slippage                   *float64          `json:"slippage" validate:"omitempty,gte=0,lte=100"`
	GoalPercentageChange       *float64          `json:"goal_percentage_change" validate:"omitempty,gte=0"`
	ParallelTransactionsAmount *int              `json:"parallel_transactions_amount" validate:"omitempty,gt=0"`
	MinTransactionsAmount      *float64          `json:"min_transactions_amount" validate:"omitempty,gt=0"`
	MaxTransactionsAmount      *float64          `json:"max_transactions_amount" validate:"omitempty,gt=0"`
	MinTimeBetweenTransactions *time.Duration    `json:"min_time_between_transactions" validate:"omitempty,min=0" swaggertype:"integer"`
	MaxTimeBetweenTransactions *time.Duration    `json:"max_time_between_transactions" validate:"omitempty,min=0" swaggertype:"integer"`
	TransactionSpeed           *TransactionSpeed `json:"transaction_speed" validate:"omitempty,oneof=default fast extra"`
}

type SwapTransactionsWithPaginationResponse struct {
	Transactions []SwapTransaction `json:"transactions"`
	Page         int               `json:"page"`
	PageSize     int               `json:"page_size"`
	Total        int               `json:"total"`
}

type SwapTransaction struct {
	bun.BaseModel `bun:"table:swap_transactions,alias:swapt" swaggerignore:"true"`

	ID              uint64       `bun:"id,pk,autoincrement" json:"id"`
	CampaignID      uuid.UUID    `bun:"campaign_id" json:"campaign_id"`
	TransactionHash string       `bun:"transaction_hash" json:"transaction_hash"`
	PoolID          string       `bun:"pool_id" json:"pool_id"`
	TokenMintFrom   string       `bun:"token_mint_from" json:"token_mint_from"`
	TokenMintTo     string       `bun:"token_mint_to" json:"token_mint_to"`
	AddressFrom     string       `bun:"address_from" json:"address_from"`
	AddressTo       string       `bun:"address_to" json:"address_to"`
	AmountTokenFrom mtype.BigRat `bun:"amount_token_from" json:"amount_token_from" swaggertype:"string"`
	AmountTokenTo   mtype.BigRat `bun:"amount_token_to" json:"amount_token_to" swaggertype:"string"`
	Status          string       `bun:"status" json:"status"`
	Message         string       `bun:"message" json:"message"`
	DebugMessage    *string      `bun:"debug_message,nullzero" json:"debug_message,omitempty"`
	CreatedAt       time.Time    `bun:"created_at" json:"created_at"`
}

type AsyncSwapTask struct {
	SwapCampaignID        uuid.UUID
	GoalPrice             *big.Rat
	MinTransactionsAmount float64
	MaxTransactionsAmount float64
	Slippage              uint64
	PoolID                solana.PublicKey
	PoolProgramID         solana.PublicKey
	SourceTokenMint       solana.PublicKey
	DestTokenMint         solana.PublicKey
	SourceAddress         solana.PublicKey
	DestAddress           solana.PublicKey
	PoolParams            *PoolParams
	SourceTokenDecimals   uint8
	DestTokenDecimals     uint8
	PrivateKey            solana.PrivateKey
	TaskType              string
	TransactionSpeed      TransactionSpeed
	ATAKeyCreated         bool
	UsingJito             bool
	PriorityFeeMLP        uint64 // microlamports/cu
}

type PoolParams struct {
	PoolID           solana.PublicKey
	InputTokenVault  solana.PublicKey
	OutputTokenVault solana.PublicKey
	AmmConfig        solana.PublicKey
	Market           solana.PublicKey
	OpenOrders       solana.PublicKey
}

type TargetPullResponse struct {
	CampaignID uuid.UUID `json:"campaign_id"`
}

type PriorityFees struct {
	Low    float64 `json:"low"`
	Medium float64 `json:"medium"`
	High   float64 `json:"high"`
}
type TargetPullEstimateResponse struct {
	BudgetSOL    float64      `json:"budget_sol"`
	TipSOL       float64      `json:"tip_sol"`
	RentSOl      float64      `json:"rent_sol"`
	PriorityFees PriorityFees `json:"priority_fees"`
}

type CampaignSummaryWithPagination struct {
	CampaignSummary []CampaignSummary `json:"campaign_summary"`
	Page            int               `json:"page"`
	PageSize        int               `json:"page_size"`
	Total           int               `json:"total"`
}

type CampaignSummary struct {
	bun.BaseModel `bun:"table:swap_campaigns" json:"-" swaggerignore:"true"`

	CampaignID     uuid.UUID `bun:"campaign_id" json:"campaign_id"`
	TokenMintFrom  string    `bun:"token_mint_from" json:"token_mint_from"`
	TokenMintTo    string    `bun:"token_mint_to" json:"token_mint_to"`
	Status         string    `bun:"status" json:"status"`
	TotalBudgetSOL float64   `bun:"budget" json:"budget"`
	SpendBudgetSOL float64   `bun:"spent_budget" json:"spent_budget"`
	ProjectID      uint64    `bun:"project_id" json:"project_id"`
	TypeName       string    `bun:"type_name" json:"type_name"`
}

type TipFloorSOL struct {
	Default float64 `json:"default" example:"0.000005"`
	Fast    float64 `json:"fast" example:"0.00001"`
	Extra   float64 `json:"extra" example:"0.0002"`
}
