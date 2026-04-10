package model

import (
	"context"
	"mm/pkg/mtype"
	"time"

	"github.com/uptrace/bun"
)

type DepositStatus string

const (
	DepositPending          DepositStatus = "PENDING"
	DepositCompleted        DepositStatus = "COMPLETED"
	DepositFailed           DepositStatus = "FAILED"
	DepositAwaitingApproval DepositStatus = "AWAITING_APPROVAL"
)

// DepositOrder represents a deposit order consisting of one or multiple deposits.
type DepositOrder struct {
	bun.BaseModel `bun:"table:deposit_orders,alias:wo" json:"-" swaggerignore:"true"`

	ID        uint64        `bun:"id,pk,autoincrement" json:"id"`
	AccountID uint64        `bun:"account_id,notnull" json:"account_id"`
	MinAmount float64       `bun:"min_amount,type:decimal(30,10),notnull" json:"min_amount"`
	MaxAmount float64       `bun:"max_amount,type:decimal(30,10),notnull" json:"max_amount"`
	WalletIDs []uint64      `bun:"wallet_ids,array,notnull" json:"wallet_ids"`
	Status    DepositStatus `bun:"status,notnull" json:"status"`
	ProjectID uint64        `bun:"project_id,notnull" json:"project_id"`
	CreatedAt time.Time     `bun:"created_at,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time     `bun:"updated_at,default:current_timestamp" json:"updated_at"`

	Deposits []Deposit `bun:"rel:has-many,join:id=deposit_order_id" json:"deposits"`
}

func (o *DepositOrder) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	o.UpdatedAt = time.Now()
	return nil
}

// Deposit represents a single deposit entry for a deposit order.
type Deposit struct {
	bun.BaseModel `bun:"table:deposits,alias:d" json:"-" swaggerignore:"true"`

	ID             uint64        `bun:"id,pk,autoincrement" json:"id"`
	DepositOrderID uint64        `bun:"deposit_order_id,notnull" json:"order_id"`
	ExternalID     string        `bun:"external_id,notnull" json:"external_id"`
	WalletID       uint64        `bun:"wallet_id" json:"wallet_id,omitempty"`
	Status         DepositStatus `bun:"status,notnull" json:"status"`
	TransactionID  string        `bun:"transaction_id,notnull" json:"transaction_id"`
	Amount         mtype.BigRat  `bun:"amount,type:decimal(42,22),notnull" json:"amount" swaggertype:"string"`
	CreatedAt      time.Time     `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time     `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

func (w *Deposit) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	w.UpdatedAt = time.Now()
	return nil
}

type DepositSolanaReq struct {
	ProjectID uint64  `json:"project_id"`
	Quantity  int     `json:"quantity" validate:"required,gt=0"`
	MinAmount float64 `json:"min_amount" validate:"required,gt=0"`
	MaxAmount float64 `json:"max_amount" validate:"required,gt=0"`
	AccountID uint64  `json:"account_id" validate:"required"`
}

type DepositResponse struct {
	OrderID uint64        `json:"order_id"`
	Status  DepositStatus `json:"status"`
	Fee     float64       `json:"fee"`
	Amount  float64       `json:"amount"`
}

type DepositProcessResponse struct {
	Status  DepositStatus `json:"status"`
	OrderID uint64        `json:"order_id"`
}

type PaginationDepositHistoryResponse struct {
	Deposits []DepositHistoryResponse `json:"deposits"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
	Total    int                      `json:"total"`
}

type DepositHistoryResponse struct {
	ID          uint64        `json:"id" bun:"deposit_order_id"`
	ProjectID   uint64        `json:"project_id" bun:"project_id"`
	ProjectName string        `json:"project_name" bun:"project_name"`
	TotalSumSOL float64       `json:"total_sum_sol" bun:""`
	Status      DepositStatus `json:"status" bun:"deposit_order_status"`
	CreatedAt   string        `json:"created_at" bun:"deposit_order_created_at"`

	Transactions []Transaction `json:"transactions" bun:"inline"`
}

type Transaction struct {
	PublicKey     string        `json:"public_key" bun:"wallet_public_key"`
	TransactionID string        `json:"transaction_id" bun:"transaction_hash"`
	Status        DepositStatus `json:"status" bun:"transaction_status"`
	SumSOL        float64       `json:"sum_sol" bun:"deposit_order_amount"`
	BalanceSOL    float64       `json:"balance_sol" bun:"-"`
}
