package pricing

import (
	"context"
	"errors"
	"math/big"
	pump_amm "mm/internal/client/pumpfun/amm"
	pump_amm_client "mm/internal/client/pumpfun/amm/amm_client"
	pump_bonding "mm/internal/client/pumpfun/bonding"
	pump_bonding_client "mm/internal/client/pumpfun/bonding/bonding_client"
	"mm/internal/client/raydium/ammv4"
	raydium_amm "mm/internal/client/raydium/ammv4/ammv4_client"
	"mm/internal/client/raydium/cpmm"
	raydium_cp_swap "mm/internal/client/raydium/cpmm/cpmm_client"
	poolmath "mm/internal/client/raydium/math"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/pkg/solutil"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func CalculatePoolPrice(ctx context.Context, solRPC solanarpc.SolanaRPC, poolAccount *rpc.Account, poolID, inputTokenMint, outputTokenMint solana.PublicKey) (*big.Rat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if poolAccount == nil || poolAccount.Data == nil {
		return nil, errors.New("pool account is nil")
	}

	switch poolAccount.Owner {
	case raydium_cp_swap.ProgramID:
		pool, err := raydium_cp_swap.ParseAccount_PoolState(poolAccount.Data.GetBinary())
		if err != nil {
			return nil, err
		}

		poolParams := &model.PoolParams{
			PoolID:           poolID,
			AmmConfig:        pool.AmmConfig,
			InputTokenVault:  pool.Token0Vault,
			OutputTokenVault: pool.Token1Vault,
		}

		poolState, err := cpmm.FetchCPMMPoolState(ctx, solRPC, poolParams)
		if err != nil {
			return nil, err
		}

		if solutil.IsSOLLikeMint(poolState.PoolState.Token0Mint) {
			// Token0=SOL: ReserveA=SOL, ReserveB=token -> SOL/token
			return poolmath.ConstantProductCalculatePrice(poolState.ReserveB, poolState.ReserveA, uint64(poolState.PoolState.Mint1Decimals), uint64(poolState.PoolState.Mint0Decimals)), nil
		}
		// Token0=token: ReserveA=token, ReserveB=SOL -> SOL/token
		return poolmath.ConstantProductCalculatePrice(poolState.ReserveA, poolState.ReserveB, uint64(poolState.PoolState.Mint0Decimals), uint64(poolState.PoolState.Mint1Decimals)), nil
	case raydium_amm.ProgramID:
		pool, err := raydium_amm.UnmarshalAmmInfo(poolAccount.Data.GetBinary())
		if err != nil {
			return nil, err
		}

		poolParams := &model.PoolParams{
			PoolID:           poolID,
			InputTokenVault:  pool.TokenCoin,
			OutputTokenVault: pool.TokenPc,
			OpenOrders:       pool.OpenOrders,
			Market:           pool.Market,
		}

		poolState, err := ammv4.FetchAMMPoolState(ctx, solRPC, poolParams, inputTokenMint, outputTokenMint)
		if err != nil {
			return nil, err
		}

		if poolState.PoolState.CoinMint.Equals(inputTokenMint) {
			// ReserveA=coin=token, ReserveB=pc=SOL -> SOL/token
			return poolmath.ConstantProductCalculatePrice(poolState.ReserveA, poolState.ReserveB, poolState.PoolState.CoinDecimals, poolState.PoolState.PcDecimals), nil
		}
		// ReserveA=pc=SOL, ReserveB=coin=token -> SOL/token
		return poolmath.ConstantProductCalculatePrice(poolState.ReserveB, poolState.ReserveA, poolState.PoolState.CoinDecimals, poolState.PoolState.PcDecimals), nil
	case pump_amm_client.ProgramID:
		pool, err := pump_amm_client.ParseAccount_Pool(poolAccount.Data.GetBinary())
		if err != nil {
			return nil, err
		}

		poolParams := &model.PoolParams{
			PoolID:           poolID,
			InputTokenVault:  pool.PoolBaseTokenAccount,
			OutputTokenVault: pool.PoolQuoteTokenAccount,
		}

		poolState, err := pump_amm.FetchAMMPoolState(ctx, solRPC, poolParams, inputTokenMint, outputTokenMint)
		if err != nil {
			return nil, err
		}

		if poolState.PoolState.BaseMint.Equals(inputTokenMint) && poolState.PoolState.QuoteMint.Equals(outputTokenMint) {
			// ReserveA=base=token, ReserveB=quote=SOL -> SOL/token
			return poolmath.ConstantProductCalculatePrice(
				poolState.ReserveA,
				poolState.ReserveB,
				uint64(poolState.BaseMintDecimals),
				uint64(poolState.QuoteMintDecimals),
			), nil
		}
		// ReserveA=quote=SOL, ReserveB=base=token -> SOL/token
		return poolmath.ConstantProductCalculatePrice(
			poolState.ReserveB,
			poolState.ReserveA,
			uint64(poolState.BaseMintDecimals),
			uint64(poolState.QuoteMintDecimals),
		), nil
	case pump_bonding_client.ProgramID:
		poolParams := &model.PoolParams{PoolID: poolID}

		poolState, err := pump_bonding.FetchBondingCurveState(ctx, solRPC, poolParams, inputTokenMint, outputTokenMint)
		if err != nil {
			return nil, err
		}

		if solutil.IsSOLLikeMint(inputTokenMint) {
			// ReserveA=SOL, ReserveB=token -> SOL/token
			return poolmath.ConstantProductCalculatePrice(
				poolState.ReserveB,
				poolState.ReserveA,
				uint64(poolState.BaseMintDecimals),
				uint64(solana.SolDecimals),
			), nil
		}
		// ReserveA=token, ReserveB=SOL -> SOL/token
		return poolmath.ConstantProductCalculatePrice(
			poolState.ReserveA,
			poolState.ReserveB,
			uint64(poolState.BaseMintDecimals),
			uint64(solana.SolDecimals),
		), nil
	default:
		return nil, errors.New("unknown pool program id")
	}
}
