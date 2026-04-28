package dex

import (
	"context"
	"math"
	"math/big"

	"mm/internal/client/pumpfun"
	pumpAMM "mm/internal/client/pumpfun/amm"
	pump_bonding "mm/internal/client/pumpfun/bonding/bonding_client"
	"mm/internal/client/raydium"
	raydiumamm "mm/internal/client/raydium/ammv4/ammv4_client"
	raydiumcpswap "mm/internal/client/raydium/cpmm/cpmm_client"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/pricing"
	"mm/pkg/apperrors"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type raydiumProvider struct {
	client *raydium.Client
	rpc    solanarpc.SolanaRPC
}

func (p *raydiumProvider) ID() model.SwapProviderID {
	return model.SwapProviderRaydium
}

func (p *raydiumProvider) FindPoolByMints(ctx context.Context, mintA, mintB solana.PublicKey) (*model.PoolResponse, error) {
	return p.client.FindPoolByMints(ctx, mintA, mintB, raydiumcpswap.ProgramID, raydiumamm.ProgramID)
}

func (p *raydiumProvider) PreparePool(ctx context.Context, srcMint, destMint solana.PublicKey) (*model.DexPreparedPool, *big.Rat, error) {
	pool, err := p.FindPoolByMints(ctx, srcMint, destMint)
	if err != nil {
		return nil, nil, apperrors.BadRequest("cannot find pool", err)
	}

	poolID, err := solana.PublicKeyFromBase58(pool.ID)
	if err != nil {
		return nil, nil, apperrors.BadRequest("invalid pool id", err)
	}
	poolProgID, err := solana.PublicKeyFromBase58(pool.ProgramID)
	if err != nil {
		return nil, nil, apperrors.BadRequest("invalid pool program id", err)
	}

	poolAccount, err := p.rpc.GetAccountInfo(ctx, poolID)
	if err != nil {
		return nil, nil, apperrors.BadRequest("cannot fetch pool account", err)
	}
	if poolAccount == nil || poolAccount.Value == nil {
		return nil, nil, apperrors.BadRequest("cannot fetch pool account", err)
	}

	price, err := pricing.CalculatePoolPrice(ctx, p.rpc, poolAccount.Value, poolID, srcMint, destMint)
	if err != nil {
		return nil, nil, apperrors.BadRequest("cannot fetch price", err)
	}

	var sourceTokenDecimals, destTokenDecimals int
	if srcMint.String() == pool.MintA.Address && destMint.String() == pool.MintB.Address {
		sourceTokenDecimals = pool.MintA.Decimals
		destTokenDecimals = pool.MintB.Decimals
	} else {
		sourceTokenDecimals = pool.MintB.Decimals
		destTokenDecimals = pool.MintA.Decimals
	}

	if sourceTokenDecimals < math.MinInt8 || sourceTokenDecimals > math.MaxInt8 || destTokenDecimals < math.MinInt8 || destTokenDecimals > math.MaxInt8 {
		return nil, nil, apperrors.BadRequest("invalid source token decimals")
	}

	return &model.DexPreparedPool{
		PoolID:              poolID,
		PoolProgramID:       poolProgID,
		SourceTokenDecimals: uint8(sourceTokenDecimals),
		DestTokenDecimals:   uint8(destTokenDecimals),
	}, price, nil
}

func (p *raydiumProvider) FetchPoolParams(ctx context.Context, poolID solana.PublicKey) (*model.PoolParams, error) {
	pool, err := p.rpc.GetAccountInfo(ctx, poolID)
	if err != nil {
		return nil, err
	}

	return fetchRaydiumPoolParams(ctx, pool, poolID)
}

type pumpfunProvider struct {
	client *pumpfun.Client
}

func (p *pumpfunProvider) ID() model.SwapProviderID {
	return model.SwapProviderPumpfun
}

func (p *pumpfunProvider) FindPoolByMints(ctx context.Context, mintA, mintB solana.PublicKey) (*model.PoolResponse, error) {
	return p.client.FindPoolByMints(ctx, mintA, mintB, pumpAMM.ProgramID, pump_bonding.ProgramID)
}

func (p *pumpfunProvider) PreparePool(ctx context.Context, mintA, mintB solana.PublicKey) (*model.DexPreparedPool, *big.Rat, error) {
	res, price, err := p.client.PreparePool(ctx, mintA, mintB)
	if err != nil {
		return nil, nil, err
	}

	return &model.DexPreparedPool{
		PoolID:              res.PoolID,
		PoolProgramID:       res.PoolProgramID,
		SourceTokenDecimals: res.SourceTokenDecimals,
		DestTokenDecimals:   res.DestTokenDecimals,
	}, price, nil
}

func (p *pumpfunProvider) FetchPoolParams(ctx context.Context, poolID solana.PublicKey) (*model.PoolParams, error) {
	return p.client.FetchPoolParams(ctx, poolID)
}

func NewDexProviders(
	raydiumClient *raydium.Client,
	pumpfunClient *pumpfun.Client,
	rpc solanarpc.SolanaRPC,
) map[model.SwapProviderID]model.DexProvider {
	return map[model.SwapProviderID]model.DexProvider{
		model.SwapProviderRaydium: &raydiumProvider{client: raydiumClient, rpc: rpc},
		model.SwapProviderPumpfun: &pumpfunProvider{client: pumpfunClient},
	}
}

func fetchRaydiumPoolParams(ctx context.Context, pool *rpc.GetAccountInfoResult, poolID solana.PublicKey) (*model.PoolParams, error) {
	var poolParams model.PoolParams

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	switch pool.Value.Owner {
	case raydiumamm.ProgramID:
		ammInfo, aErr := raydiumamm.UnmarshalAmmInfo(pool.GetBinary())
		if aErr != nil {
			return nil, aErr
		}

		poolParams = model.PoolParams{
			PoolID:           poolID,
			InputTokenVault:  ammInfo.TokenCoin,
			OutputTokenVault: ammInfo.TokenPc,
			OpenOrders:       ammInfo.OpenOrders,
			Market:           ammInfo.Market,
		}
	case raydiumcpswap.ProgramID:
		cpmmInfo, cErr := raydiumcpswap.ParseAccount_PoolState(pool.GetBinary())
		if cErr != nil {
			return nil, cErr
		}
		poolParams = model.PoolParams{
			PoolID:           poolID,
			InputTokenVault:  cpmmInfo.Token0Vault,
			OutputTokenVault: cpmmInfo.Token1Vault,
			AmmConfig:        cpmmInfo.AmmConfig,
		}
	}

	return &poolParams, nil
}
