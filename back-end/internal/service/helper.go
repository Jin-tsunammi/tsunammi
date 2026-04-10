package service

import (
	"context"
	cryptoRand "crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand/v2"
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
	"mm/pkg/apperrors"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func fetchWalletsBalanceWithTotal(
	ctx context.Context,
	wallets []model.Wallet,
	solanaRPC solanarpc.SolanaRPC,
	rate float64,
) (totalBalanceSOL float64, totalBalanceUSD float64, err error) {
	publicKeys := make([]solana.PublicKey, 0, len(wallets))

	totalBalanceSOL = 0.0
	totalBalanceUSD = 0.0

	for i := 0; i < len(wallets); i++ {
		wallet := &wallets[i]

		publicKey, err := solana.PublicKeyFromBase58(wallet.PublicKey)

		if err != nil {
			return 0.0, 0.0, apperrors.Internal("failed to get public key", err)
		}

		publicKeys = append(publicKeys, publicKey)
	}

	balances, err := solanaRPC.GetMulltipyWalletBalance(ctx, publicKeys)

	if err != nil {
		return 0.0, 0.0, apperrors.Internal("failed to get wallet balance", err)
	}

	for i := 0; i < len(wallets); i++ {
		wallet := &wallets[i]

		wallet.BalanceUSD = balances[i] * rate
		wallet.BalanceSOL = balances[i]

		totalBalanceSOL += wallet.BalanceSOL
		totalBalanceUSD += wallet.BalanceUSD
	}

	return totalBalanceSOL, totalBalanceUSD, nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint")
}

func generateRandomNum() (uint64, error) {
	randomNum, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(math.MaxUint32))
	if err != nil {
		return 0, apperrors.Internal("failed to generate num", err)
	}

	return randomNum.Uint64(), nil
}

func generateVerificationCode() (string, error) {
	num, err := generateRandomNum()
	if err != nil {
		return "", err
	}

	code := int(num)%900000 + 100000

	return fmt.Sprintf("%06d", code), nil
}

func percentageToBasicPoints(percentage float64) uint64 {
	return uint64(math.Round(percentage * 100))
}

func basicPointToBasicPoints(point uint64) float64 {
	return float64(point) / 100
}

func projectToProjectResponse(project *model.ProjectWithWallets) *model.ProjectWithWalletsResponse {

	wallets := make([]model.WalletResponse, 0, len(project.Wallets))

	for i := 0; i < len(project.Wallets); i++ {
		wallet := &project.Wallets[i]
		wallets = append(wallets, model.WalletResponse{
			ID:         wallet.ID,
			PublicKey:  wallet.PublicKey,
			BalanceSOL: wallet.BalanceSOL,
			BalanceUSD: wallet.BalanceUSD,
			CreatedAt:  wallet.CreatedAt,
		})
	}

	return &model.ProjectWithWalletsResponse{
		ID:              project.ID,
		Name:            project.Name,
		UserID:          project.UserID,
		Wallets:         wallets,
		BalanceSOL:      project.BalanceSOL,
		TotalBalanceSOL: project.TotalBalanceSOL,
		TotalBalanceUSD: project.TotalBalanceUSD,
		LastSync:        time.Now(),
		CreatedAt:       project.CreatedAt,
		WalletCount:     int64(len(wallets)),
	}
}

func CalculateGoalPrice(price *big.Rat, percentChange float64, isUp bool) *big.Rat {
	percentage := new(big.Rat).SetFloat64(percentChange / 100.0)
	one := new(big.Rat).SetInt64(1)
	multiplier := new(big.Rat)

	if isUp {
		multiplier.Add(one, percentage)
	} else {
		multiplier.Sub(one, percentage)
	}

	goal := new(big.Rat).Mul(price, multiplier)
	return goal
}

func calculatePoolPrice(ctx context.Context, solRPC solanarpc.SolanaRPC, poolAccount *rpc.Account, poolID, inputTokenMint, outputTokenMint solana.PublicKey) (*big.Rat, error) {
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

		var price *big.Rat

		if poolState.PoolState.Token0Mint.Equals(inputTokenMint) && poolState.PoolState.Token1Mint.Equals(outputTokenMint) {
			price = poolmath.ConstantProductCalculatePrice(poolState.ReserveA, poolState.ReserveB, uint64(poolState.PoolState.Mint0Decimals), uint64(poolState.PoolState.Mint1Decimals))
		} else {
			price = poolmath.ConstantProductCalculatePrice(poolState.ReserveB, poolState.ReserveA, uint64(poolState.PoolState.Mint1Decimals), uint64(poolState.PoolState.Mint0Decimals))
		}
		return price, nil
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

		var price *big.Rat

		if poolState.PoolState.CoinMint.Equals(outputTokenMint) && poolState.PoolState.PcMint.Equals(inputTokenMint) {
			price = poolmath.ConstantProductCalculatePrice(poolState.ReserveA, poolState.ReserveB, poolState.PoolState.CoinDecimals, poolState.PoolState.PcDecimals)
		} else {
			price = poolmath.ConstantProductCalculatePrice(poolState.ReserveB, poolState.ReserveA, poolState.PoolState.PcDecimals, poolState.PoolState.CoinDecimals)
		}
		return price, nil
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
			return poolmath.ConstantProductCalculatePrice(
				poolState.ReserveA,
				poolState.ReserveB,
				uint64(poolState.BaseMintDecimals),
				uint64(poolState.QuoteMintDecimals),
			), nil
		}

		return poolmath.ConstantProductCalculatePrice(
			poolState.ReserveA,
			poolState.ReserveB,
			uint64(poolState.QuoteMintDecimals),
			uint64(poolState.BaseMintDecimals),
		), nil
	case pump_bonding_client.ProgramID:
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

func generateRandomFloat(min, max float64) float64 {
	res := min + rand.Float64()*(max-min)
	return math.Round(res*1000) / 1000
}
