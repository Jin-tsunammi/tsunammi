package swaptxlog

import (
	"context"
	"errors"
	"strings"
	"time"

	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/internal/swaperror"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Params struct {
	PoolID          string
	TokenMintFrom   string
	TokenMintTo     string
	AddressFrom     string
	AddressTo       string
	TransactionHash string
}

func Resolve(err error) (string, string, *string) {
	status := "Failed"
	message := "Network Error"

	if err == nil {
		status = "Pending"
		message = "Transaction was sent"
		return status, message, nil
	}

	debug := err.Error()
	debugMessage := &debug

	if errors.Is(err, swaperror.ErrSimulationError) {
		message = "Simulation Error"
	}
	if errors.Is(err, swaperror.ErrSlippageExceeded) {
		message = "Slippage Exceeded"
	}
	if errors.Is(err, swaperror.ErrInsufficientFunds) || strings.Contains(strings.ToLower(err.Error()), "budget exceeded") {
		message = "Insufficient Funds/Budget Exceeded"
	}
	if errors.Is(err, swaperror.ErrCustomProgramError) {
		message = "Custom Program Error"
	}
	if errors.Is(err, swaperror.ErrRateLimit) {
		message = "Rate Limit"
	}
	if errors.Is(err, swaperror.ErrGatewayTimeout) {
		message = "Gateway Timeout"
	}
	if errors.Is(err, swaperror.ErrComputeBudgetExceeded) {
		message = "Compute Budget Exceeded"
	}
	if errors.Is(err, swaperror.ErrBundleRejected) {
		message = "Bundle Rejected"
	}

	return status, message, debugMessage
}

func LogSwapTransaction(
	ctx context.Context,
	err error,
	campaignID uuid.UUID,
	params Params,
	transactionRepository *repository.SwapTransactionRepository,
	logger *zap.Logger,
) error {
	status, message, debugMessage := Resolve(err)

	swapTx := &model.SwapTransaction{
		CampaignID:      campaignID,
		TransactionHash: params.TransactionHash,
		PoolID:          params.PoolID,
		TokenMintFrom:   params.TokenMintFrom,
		TokenMintTo:     params.TokenMintTo,
		AddressFrom:     params.AddressFrom,
		AddressTo:       params.AddressTo,
		Status:          status,
		Message:         message,
		DebugMessage:    debugMessage,
		CreatedAt:       time.Now(),
	}

	logger.Error("Error in log func", zap.Error(err))
	logger.Error("Swap transaction", zap.Any("swap_tx", swapTx))
	logger.Error("Status", zap.String("status", status))
	logger.Error("Message", zap.String("message", message))

	if dbErr := transactionRepository.Create(ctx, swapTx); dbErr != nil {
		logger.Error("Failed to save swap transaction", zap.Error(dbErr))
	}

	return nil
}
