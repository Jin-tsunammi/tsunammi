package raydium

import (
	"errors"
	"math/big"
	"mm/internal/model"

	"github.com/gagliardetto/solana-go"
)

var PriceIsAlreadyReachedError = errors.New("price is already reached")

type SwapParams struct {
	UserWallet           solana.PublicKey
	PoolID               solana.PublicKey
	AmmConfig            solana.PublicKey
	Token0Vault          solana.PublicKey
	Token1Vault          solana.PublicKey
	UserSourceToken      solana.PublicKey
	UserDestToken        solana.PublicKey
	InputTokenMint       solana.PublicKey
	OutputTokenMint      solana.PublicKey
	AmountIn             *big.Int
	MinAmountOut         *big.Int
	ObservationState     solana.PublicKey
	InputTokenProgramID  solana.PublicKey // For Token/Token-2022
	OutputTokenProgramID solana.PublicKey // For Token/Token-2022
	PoolParams           *model.PoolParams
}

type TWAPConfig struct {
	MinTransactionsAmount         uint64
	MaxTransactionsAmount         uint64
	SlippageBPS                   uint64
	ComputeUnitLimit              uint32
	ComputeUnitPriceMicroLamports uint64
}

type findPoolByMintsData struct {
	Count       int                  `json:"count"`
	Pools       []model.PoolResponse `json:"data"`
	HasNextPage bool                 `json:"hasNextPage"`
}
type FindPoolByMintsResponse struct {
	Id      string              `json:"id"`
	Success bool                `json:"success"`
	Data    findPoolByMintsData `json:"data"`
}

type CustomError interface {
	Code() int
	Name() string
	Error() string
}
