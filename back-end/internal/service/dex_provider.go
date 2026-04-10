package service

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
	"mm/pkg/apperrors"

	"github.com/gagliardetto/solana-go"
)

type dexProvider interface {
	ID() model.SwapProviderID
	FindPoolByMints(ctx context.Context, mintA, mintB solana.PublicKey) (*model.PoolResponse, error)
	PreparePool(ctx context.Context, srcMint, destMint solana.PublicKey) (*fetchPoolResult, *big.Rat, error)
	FetchPoolParams(ctx context.Context, poolID solana.PublicKey) (*model.PoolParams, error)
}

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

func (p *raydiumProvider) PreparePool(ctx context.Context, srcMint, destMint solana.PublicKey) (*fetchPoolResult, *big.Rat, error) {
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

	price, err := calculatePoolPrice(ctx, p.rpc, poolAccount.Value, poolID, srcMint, destMint)
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

	return &fetchPoolResult{
		poolID:              poolID,
		poolProgramID:       poolProgID,
		sourceTokenDecimals: uint8(sourceTokenDecimals),
		destTokenDecimals:   uint8(destTokenDecimals),
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

func (p *pumpfunProvider) PreparePool(ctx context.Context, mintA, mintB solana.PublicKey) (*fetchPoolResult, *big.Rat, error) {
	res, price, err := p.client.PreparePool(ctx, mintA, mintB)
	if err != nil {
		return nil, nil, err
	}

	return &fetchPoolResult{
		poolID:              res.PoolID,
		poolProgramID:       res.PoolProgramID,
		sourceTokenDecimals: res.SourceTokenDecimals,
		destTokenDecimals:   res.DestTokenDecimals,
	}, price, nil
}

func (p *pumpfunProvider) FetchPoolParams(ctx context.Context, poolID solana.PublicKey) (*model.PoolParams, error) {
	return p.client.FetchPoolParams(ctx, poolID)
}
