package ammv4

import (
	"math/big"
	openbookv1 "mm/internal/client/raydium/ammv4/openbook/openbook_v1_client"
	"mm/pkg/apperrors"
)

// Ported from https://github.com/raydium-io/raydium-amm/blob/3b087ade40da365b2dd1df5e8baf77a3b97245d4/program/src/math.rs#L243
// In the future, we can add parsing of the event queue to get the exact vault amounts,
// but since the error at this stage is very small (around 0.001%), we can safely ignore it
func ammCalcExactVaultInSerum(openOrders openbookv1.OpenOrders) (pcTotal uint64, coinTotal uint64) {
	pcTotal = openOrders.NativePcTotal
	coinTotal = openOrders.NativeCoinTotal

	return pcTotal, coinTotal
}

// Ported from https://github.com/raydium-io/raydium-amm/blob/3b087ade40da365b2dd1df5e8baf77a3b97245d4/program/src/math.rs#L293
func ammCalcTotalWithoutTakePNL(ammInfo *AMMInfoWithReservers) (coinTotalWithoutPNL *big.Int, pcTotalWithoutPNL *big.Int, err error) {
	pcTotalInSerum, coinTotalInSerum := ammCalcExactVaultInSerum(ammInfo.OpenOrders)

	coinTotalWithoutPNL = new(big.Int).Add(ammInfo.ReserveA, new(big.Int).SetUint64(coinTotalInSerum))
	pcTotalWithoutPNL = new(big.Int).Add(ammInfo.ReserveB, new(big.Int).SetUint64(pcTotalInSerum))

	coinTotalWithoutPNL.Sub(
		coinTotalWithoutPNL,
		new(big.Int).SetUint64(ammInfo.PoolState.OutPut.NeedTakePnlCoin),
	)

	pcTotalWithoutPNL.Sub(
		pcTotalWithoutPNL,
		new(big.Int).SetUint64(ammInfo.PoolState.OutPut.NeedTakePnlPc),
	)

	if coinTotalWithoutPNL.Sign() < 0 || pcTotalWithoutPNL.Sign() < 0 {
		return nil, nil, apperrors.Internal("failed to calculate total of coins without PNL")
	}

	if coinTotalWithoutPNL.BitLen() > 64 || pcTotalWithoutPNL.BitLen() > 64 {
		return nil, nil, apperrors.Internal("value too large for uint64")
	}

	return coinTotalWithoutPNL, pcTotalWithoutPNL, nil
}
