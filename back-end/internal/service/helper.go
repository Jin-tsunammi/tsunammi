package service

import (
	"context"
	cryptoRand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand/v2"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/pkg/apperrors"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
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

func generateRandomFloat(min, max float64) float64 {
	res := min + rand.Float64()*(max-min)
	return math.Round(res*1000) / 1000
}
