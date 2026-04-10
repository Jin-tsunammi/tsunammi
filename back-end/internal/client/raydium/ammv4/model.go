package ammv4

import (
	"math/big"
	raydiumamm "mm/internal/client/raydium/ammv4/ammv4_client"
	openbookv1 "mm/internal/client/raydium/ammv4/openbook/openbook_v1_client"
)

type AMMInfoWithReservers struct {
	ReserveA   *big.Int
	ReserveB   *big.Int
	PoolState  raydiumamm.AmmInfo
	Market     openbookv1.MarketState
	OpenOrders openbookv1.OpenOrders
}
