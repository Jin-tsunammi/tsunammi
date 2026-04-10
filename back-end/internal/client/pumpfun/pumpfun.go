package pumpfun

import (
	"context"
	"errors"
	"math"
	"math/big"
	"mm/config"
	pump_amm "mm/internal/client/pumpfun/amm"
	pump_bonding "mm/internal/client/pumpfun/bonding"

	poolmath "mm/internal/client/raydium/math"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/pkg/apperrors"
	"mm/pkg/solutil"
	"net/url"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"resty.dev/v3"
)

const (
	poolGlobalConfig         = "ADyA8hdefvWN2dbGGWFotbzWxrAvLW83WG6QCVXvJKqw"
	bondingCurveGlobalConfig = "4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf"
)

type Client struct {
	rpc           solanarpc.SolanaRPC
	restyClient   *resty.Client
	solscanApiKey string
}

func NewClient(rpc solanarpc.SolanaRPC, client *resty.Client, cfg *config.Config) *Client {
	return &Client{
		rpc:           rpc,
		restyClient:   client,
		solscanApiKey: cfg.App.SolscanApiKey,
	}
}

type poolResp struct {
	PoolID             string  `json:"pool_id"`
	ProgramID          string  `json:"program_id"`
	Token1             string  `json:"token_1"`
	Token2             string  `json:"token_2"`
	TokenAccount1      string  `json:"token_account_1"`
	TokenAccount2      string  `json:"token_account_2"`
	TotalTrades24H     int     `json:"total_trades_24h"`
	TotalTradesPrev24H int     `json:"total_trades_prev_24h"`
	TotalVolume24H     float64 `json:"total_volume_24h"`
	TotalVolumePrev24H float64 `json:"total_volume_prev_24h"`
}

type resp struct {
	Success bool       `json:"success"`
	Data    []poolResp `json:"data"`
}

func (p *poolResp) IsPoolTypeOneOf(programID ...solana.PublicKey) bool {
	if programID == nil {
		return true
	}

	for _, id := range programID {
		if p.ProgramID == id.String() {
			return true
		}
	}

	return false
}

func solLikeVariants(mint solana.PublicKey) []solana.PublicKey {
	if !solutil.IsSOLLikeMint(mint) {
		return []solana.PublicKey{mint}
	}

	return []solana.PublicKey{solana.WrappedSol, solutil.NativeSolPubkey}
}

func (c *Client) FindPoolByMints(ctx context.Context, mintA, mintB solana.PublicKey, typeProgramID ...solana.PublicKey) (*model.PoolResponse, error) {
	var (
		r          resp
		result     []poolResp
		queryMintA = mintA
		queryMintB = mintB
	)

outer:
	for _, candidateA := range solLikeVariants(mintA) {
		for _, candidateB := range solLikeVariants(mintB) {
			r = resp{}
			query := url.Values{}
			query.Add("token", candidateA.String())
			query.Add("token", candidateB.String())
			query.Add("sort_by", "tvl")
			for _, p := range typeProgramID {
				query.Add("program", p.String())
			}
			data, err := c.restyClient.R().
				SetContext(ctx).
				SetQueryString(query.Encode()).
				SetResult(&r).
				SetHeader("token", c.solscanApiKey).
				Get("https://pro-api.solscan.io/v2.0/token/markets")
			if err != nil {
				return nil, apperrors.Internal("failed to get pools: ", err)
			}

			if data.IsError() {
				return nil, apperrors.Internal("failed to get data", data.Err)
			}

			filtered := make([]poolResp, 0, len(r.Data))
			for _, pool := range r.Data {
				if pool.IsPoolTypeOneOf(typeProgramID...) {
					filtered = append(filtered, pool)
				}
			}

			if len(filtered) > 0 {
				result = filtered
				queryMintA = candidateA
				queryMintB = candidateB
				break outer
			}
		}
	}

	if len(result) == 0 {
		return nil, apperrors.Internal("pool not found")
	}

	selected := result[0]

	// [ mintA token.Mint, mintB token.Mint, pool global config, bondig curve global config ]
	accs, err := c.rpc.GetMultipleAccounts(ctx,
		queryMintA,
		queryMintB,
		solana.MustPublicKeyFromBase58(poolGlobalConfig),
		solana.MustPublicKeyFromBase58(bondingCurveGlobalConfig))
	if err != nil {
		return nil, err
	}

	aDecimals := solana.SolDecimals
	if !solutil.IsSOLLikeMint(queryMintA) {
		var aParsed token.Mint
		err = aParsed.UnmarshalWithDecoder(bin.NewBinDecoder(accs.Value[0].Data.GetBinary()))
		if err != nil {
			return nil, errors.New("failed to unmarshal mintA")
		}
		aDecimals = aParsed.Decimals
	}

	bDecimals := solana.SolDecimals
	if !solutil.IsSOLLikeMint(queryMintB) {
		var bParsed token.Mint
		err = bParsed.UnmarshalWithDecoder(bin.NewBinDecoder(accs.Value[1].Data.GetBinary()))
		if err != nil {
			return nil, errors.New("failed to unmarshal mintB")
		}
		bDecimals = bParsed.Decimals
	}

	poolResp := &model.PoolResponse{
		ProgramID: selected.ProgramID,
		ID:        selected.PoolID,
		MintA: model.MintInfo{
			Address:  mintA.String(),
			Decimals: int(aDecimals),
		},
		MintB: model.MintInfo{
			Address:  mintB.String(),
			Decimals: int(bDecimals),
		},
	}

	if acc := selected; acc.ProgramID == pump_amm.ProgramID.String() {
		parsedPoolConfig, err := pump_amm.ParseAccount_GlobalConfig(accs.Value[2].Data.GetBinary())
		if err != nil {
			return nil, err
		}

		poolResp.FeeRate = float64(parsedPoolConfig.LpFeeBasisPoints+
			parsedPoolConfig.CoinCreatorFeeBasisPoints+
			parsedPoolConfig.ProtocolFeeBasisPoints) / 10_000
	} else if acc := selected; acc.ProgramID == pump_bonding.ProgramID.String() {
		parsedBondingConfig, err := pump_bonding.ParseAccount_Global(accs.Value[3].Data.GetBinary())
		if err != nil {
			return nil, err
		}

		poolResp.FeeRate = float64(parsedBondingConfig.CreatorFeeBasisPoints+parsedBondingConfig.FeeBasisPoints) / 10_000
	}

	return poolResp, nil
}

type fetchPoolResult struct {
	PoolID              solana.PublicKey
	PoolProgramID       solana.PublicKey
	SourceTokenDecimals uint8
	DestTokenDecimals   uint8
}

func (c *Client) PreparePool(ctx context.Context, srcMint, destMint solana.PublicKey) (*fetchPoolResult, *big.Rat, error) {
	pool, err := c.FindPoolByMints(ctx, srcMint, destMint, pump_amm.ProgramID, pump_bonding.ProgramID)
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

	poolAccount, err := c.rpc.GetAccountInfo(ctx, poolID)
	if err != nil {
		return nil, nil, apperrors.BadRequest("cannot fetch pool account", err)
	}
	if poolAccount == nil || poolAccount.Value == nil {
		return nil, nil, apperrors.BadRequest("cannot fetch pool account", err)
	}

	price, err := calculatePoolPrice(ctx, c.rpc, poolAccount.Value, poolID, srcMint, destMint)
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
		PoolID:              poolID,
		PoolProgramID:       poolProgID,
		SourceTokenDecimals: uint8(sourceTokenDecimals),
		DestTokenDecimals:   uint8(destTokenDecimals),
	}, price, nil
}

func (c *Client) FetchPoolParams(ctx context.Context, poolID solana.PublicKey) (*model.PoolParams, error) {
	pool, err := c.rpc.GetAccountInfo(ctx, poolID)
	if err != nil {
		return nil, err
	}

	var poolParams model.PoolParams

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	switch pool.Value.Owner {
	case pump_amm.ProgramID:
		ammInfo, aErr := pump_amm.ParseAccount_Pool(pool.Value.Data.GetBinary())
		if aErr != nil {
			return nil, aErr
		}

		poolParams = model.PoolParams{
			PoolID:           poolID,
			InputTokenVault:  ammInfo.PoolBaseTokenAccount,
			OutputTokenVault: ammInfo.PoolQuoteTokenAccount,
		}
	case pump_bonding.ProgramID:
		poolParams = model.PoolParams{
			PoolID: poolID,
		}
	}
	return &poolParams, nil
}

func calculatePoolPrice(ctx context.Context, solRPC solanarpc.SolanaRPC, poolAccount *rpc.Account, poolID, inputTokenMint, outputTokenMint solana.PublicKey) (*big.Rat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if poolAccount == nil || poolAccount.Data == nil {
		return nil, errors.New("pool account is nil")
	}

	switch poolAccount.Owner {
	case pump_amm.ProgramID:
		pool, err := pump_amm.ParseAccount_Pool(poolAccount.Data.GetBinary())
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
			return poolmath.ConstantProductCalculatePrice(poolState.ReserveA, poolState.ReserveB, uint64(poolState.BaseMintDecimals), uint64(poolState.QuoteMintDecimals)), nil
		}

		return poolmath.ConstantProductCalculatePrice(poolState.ReserveA, poolState.ReserveB, uint64(poolState.QuoteMintDecimals), uint64(poolState.BaseMintDecimals)), nil
	case pump_bonding.ProgramID:
		poolParams := &model.PoolParams{
			PoolID: poolID,
		}

		poolState, err := pump_bonding.FetchBondingCurveState(ctx, solRPC, poolParams, inputTokenMint, outputTokenMint)
		if err != nil {
			return nil, err
		}

		return poolmath.ConstantProductCalculatePrice(
			poolState.ReserveA,
			poolState.ReserveB,
			uint64(poolState.BaseMintDecimals),
			uint64(poolState.QuoteMintDecimals),
		), nil
	default:
		return nil, errors.New("unknown pool program id")
	}
}
