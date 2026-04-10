package cpmm

import (
	"math/big"
	raydiumcpswap "mm/internal/client/raydium/cpmm/cpmm_client"
)

type PoolStateWithReserve struct {
	ReserveA  *big.Int // Reserve of token we SELL
	ReserveB  *big.Int // Reserve of token we BUY
	PoolState *raydiumcpswap.PoolState
	AmmConfig *raydiumcpswap.AmmConfig
}
