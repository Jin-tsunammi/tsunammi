package poolmath

import (
	"fmt"
	"math/big"
)

const (
	// basisPointDenominator represents 100% in standard basis points (10,000 bps = 100%).
	basisPointDenominator uint64 = 10000

	// feeDenominator represents 100% in high-precision fee units (1,000,000 units = 100%).
	// It is used for fees denominated in hundredths of a bip (10^-6).
	feeDenominator uint64 = 1000000

	// feeUnitsPerBasisPoint is the conversion factor between standard BPS and high-precision units.
	// 1 bp = 100 high-precision units.
	feeUnitsPerBasisPoint = feeDenominator / basisPointDenominator
)

// ConstantProductCalculateMaxAmountForSlippage The feeBps1Of100, denominated in hundredths of a bip (10^-6)
func ConstantProductCalculateMaxAmountForSlippage(reserveIn *big.Int, maxSlippageBps uint64, feeBps1Of100 uint64) *big.Int {

	// Denominate slippage in hundredths of a bip (10^-6)
	slippageHP := maxSlippageBps * feeUnitsPerBasisPoint

	if slippageHP >= feeDenominator || feeBps1Of100 >= feeDenominator {
		return big.NewInt(0)
	}

	gamma := new(big.Int).SetUint64(feeDenominator - feeBps1Of100)

	invSlippage := new(big.Int).SetUint64(feeDenominator - slippageHP)

	// Denom = Gamma * InvSlippage
	denominator := new(big.Int).Mul(gamma, invSlippage)

	numerator := new(big.Int).Mul(reserveIn, new(big.Int).SetUint64(slippageHP))
	numerator.Mul(numerator, new(big.Int).SetUint64(feeDenominator))

	return new(big.Int).Div(numerator, denominator)
}

// ConstantProductApplyFee feeHighPrecision denominated in hundredths of a bip (10^-6).
func ConstantProductApplyFee(amount *big.Int, feeHighPrecision uint64) (*big.Int, error) {

	if amount == nil {
		return nil, fmt.Errorf("amount is nil")
	}
	if amount.Sign() < 0 {
		return nil, fmt.Errorf("amount must be non-negative")
	}
	if feeHighPrecision > feeDenominator {
		return nil, fmt.Errorf("fee cannot exceed 100%% (%d units)", feeDenominator)
	}

	multi := new(big.Int).SetUint64(feeDenominator - feeHighPrecision)

	out := new(big.Int).Mul(amount, multi)
	out.Div(out, new(big.Int).SetUint64(feeDenominator))

	return out, nil
}

// ConstantProductTokenAmountOut feeHighPrecision denominated in hundredths of a bip (10^-6).
func ConstantProductTokenAmountOut(amountIn *big.Int, reserveA, reserveB *big.Int, feeHighPrecision uint64) (*big.Int, error) {
	if amountIn == nil {
		return nil, fmt.Errorf("amountIn is nil")
	}
	if reserveA == nil || reserveB == nil {
		return nil, fmt.Errorf("pool reserves cannot be nil")
	}
	if amountIn.Sign() < 0 {
		return nil, fmt.Errorf("amountIn cannot be negative")
	}
	if reserveA.Sign() < 0 || reserveB.Sign() < 0 {
		return nil, fmt.Errorf("pool reserves cannot be negative")
	}

	if amountIn.Sign() == 0 {
		return big.NewInt(0), nil
	}

	amountInAfterFee, err := ConstantProductApplyFee(amountIn, feeHighPrecision)
	if err != nil {
		return nil, err
	}

	// Constant Product formula (x * y = k):
	// dy = (dx_eff * y) / (x + dx_eff)
	// Where:
	// dx_eff = amountInAfterFee
	// x = ReserveA (input reserve)
	// y = ReserveB (output reserve)

	// denom = ReserveA + amountInAfterFee
	denom := new(big.Int).Add(reserveA, amountInAfterFee)

	if denom.Sign() <= 0 {
		return nil, fmt.Errorf("invalid AMM state: denom <= 0")
	}

	// numer = amountInAfterFee * ReserveB
	numer := new(big.Int).Mul(amountInAfterFee, reserveB)

	// out = numer / denom
	out := new(big.Int).Div(numer, denom)

	return out, nil
}

func ConstantProductFloorWithSlippage(amount *big.Int, slippageBps uint64) (*big.Int, error) {
	if amount == nil {
		return nil, fmt.Errorf("amount is nil")
	}
	if amount.Sign() < 0 {
		return nil, fmt.Errorf("amount cannot be negative")
	}
	if slippageBps > basisPointDenominator {
		return nil, fmt.Errorf("slippage cannot exceed 100%% (10000 bps)")
	}

	// multiplier = (basisPointDenominator - slippage) / basisPointDenominator
	multiplier := new(big.Rat).SetFrac(
		new(big.Int).SetUint64(basisPointDenominator-slippageBps),
		new(big.Int).SetUint64(basisPointDenominator),
	)

	r := new(big.Rat).Mul(new(big.Rat).SetInt(amount), multiplier)

	out := new(big.Int)
	out.Div(r.Num(), r.Denom()) // floor

	return out, nil
}

func ConstantProductCalculatePrice(reserveA, reserveB *big.Int, decimalsTokenA, decimalsTokenB uint64) *big.Rat {
	lamportsInTokenA := new(big.Int).Exp(big.NewInt(10), new(big.Int).SetUint64(decimalsTokenA), nil)
	lamportsInTokenB := new(big.Int).Exp(big.NewInt(10), new(big.Int).SetUint64(decimalsTokenB), nil)

	reserveAToken := new(big.Rat).SetFrac(reserveA, lamportsInTokenA)
	reserveBToken := new(big.Rat).SetFrac(reserveB, lamportsInTokenB)

	price := new(big.Rat).Quo(reserveBToken, reserveAToken)

	return price
}
