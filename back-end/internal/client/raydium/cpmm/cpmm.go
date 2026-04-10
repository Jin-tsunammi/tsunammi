package cpmm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"mm/config"
	"mm/internal/client/jito"
	"mm/internal/client/raydium"
	raydiumcpswap "mm/internal/client/raydium/cpmm/cpmm_client"
	poolmath "mm/internal/client/raydium/math"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/internal/swaperror"
	"mm/internal/swaptxlog"
	"mm/pkg/apperrors"
	"strings"
	"sync/atomic"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	assoc "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"resty.dev/v3"
)

const (
	DefaultRPCRetries = 2
)

type Client struct {
	RPC                   solanarpc.SolanaRPC
	restyClient           *resty.Client
	jitoClient            *jito.Client
	campaignRepository    *repository.SwapCampaignRepository
	transactionRepository *repository.SwapTransactionRepository
	URL                   string
	logger                *zap.Logger
}

func NewClient(
	rpc solanarpc.SolanaRPC,
	client *resty.Client,
	jitoClient *jito.Client,
	cfg *config.Config,
	campaignRepository *repository.SwapCampaignRepository,
	transactionRepository *repository.SwapTransactionRepository,
	logger *zap.Logger,
) *Client {
	return &Client{
		RPC:                   rpc,
		restyClient:           client,
		jitoClient:            jitoClient,
		URL:                   cfg.App.KucoinBaseUrl,
		campaignRepository:    campaignRepository,
		transactionRepository: transactionRepository,
		logger:                logger,
	}
}

func (c *Client) Swap(
	ctx context.Context,
	campaignID uuid.UUID,
	remainingBudget *atomic.Pointer[big.Int],
	tasks []*model.AsyncSwapTask,
	baseParams []raydium.SwapParams,
	configs []raydium.TWAPConfig,
	blockHash *atomic.Pointer[solana.Hash],
) (err error) {

	if len(tasks) == 0 || len(tasks) != len(baseParams) || len(tasks) != len(configs) {
		return apperrors.BadRequest("invalid params length")
	}

	task := tasks[0]
	cfg := configs[0]
	baseParam := baseParams[0]
	logParams := swaptxlog.Params{
		PoolID:        baseParam.PoolID.String(),
		TokenMintFrom: baseParam.InputTokenMint.String(),
		TokenMintTo:   baseParam.OutputTokenMint.String(),
		AddressFrom:   baseParam.UserSourceToken.String(),
		AddressTo:     baseParam.UserDestToken.String(),
	}

	if remainingBudget.Load().Sign() <= 0 {
		err = swaptxlog.LogSwapTransaction(ctx, raydium.BudgetExceededError, campaignID, logParams, c.transactionRepository, c.logger)
		if err != nil {
			return err
		}

		return raydium.BudgetExceededError
	}

	pool, err := FetchCPMMPoolState(ctx, c.RPC, task.PoolParams)
	if err != nil {
		return apperrors.Internal("failed to fetch pool state", err)
	}

	currentPrice := calculateCPMMPrice(pool, baseParam)
	if raydium.IsTargetReached(currentPrice, task.GoalPrice, task.TaskType) {
		return raydium.PriceIsAlreadyReachedError
	}

	minTransAmount := new(big.Int).SetUint64(cfg.MinTransactionsAmount)
	maxTransAmount := new(big.Int).SetUint64(cfg.MaxTransactionsAmount)

	if remainingBudget.Load().Sign() <= 0 {
		err = swaptxlog.LogSwapTransaction(ctx, raydium.BudgetExceededError, campaignID, logParams, c.transactionRepository, c.logger)
		if err != nil {
			return err
		}

		return raydium.BudgetExceededError
	}

	var txs []*solana.Transaction
	var tx *solana.Transaction
	var index int

	defer func() {
		logErrs := make([]error, len(txs))
		for index, tx = range txs {
			logParams.TransactionHash = tx.Signatures[0].String()
			logErr := swaptxlog.LogSwapTransaction(ctx, err, campaignID, logParams, c.transactionRepository, c.logger)
			logErrs[index] = logErr
		}
		if joinedSwapErrs := errors.Join(logErrs...); joinedSwapErrs != nil {

			err = joinedSwapErrs
		}
	}()

	latestHash := blockHash.Load()

	if latestHash == nil {
		latestBlockHash, err := c.RPC.GetLatestBlockhash(ctx)
		if err != nil {
			return err
		}
		latestHash = latestBlockHash
		blockHash.Store(latestBlockHash)
	}

	keys := make([]solana.PublicKey, len(tasks))

	for index, task = range tasks {
		if task.SourceTokenMint.Equals(solana.WrappedSol) {
			keys[index] = task.PrivateKey.PublicKey()
		} else {
			keys[index] = task.SourceAddress
		}
	}

	results, err := c.RPC.GetMultipleAccountsWithNoLimits(ctx, keys...)
	if err != nil {
		return err
	}

	accounts := make([]*rpc.Account, 0, len(keys))

	for _, result := range results {
		accounts = append(accounts, result.Value...)
	}

	subBudget := new(big.Int)
	amountIn := new(big.Int)

	errs := make([]error, len(tasks))
	amounts := make([]*big.Int, len(tasks))

	for index, task = range tasks {

		amountIn, err = raydium.RandBigIntRange(
			minTransAmount,
			maxTransAmount,
		)

		if err != nil {
			errs[index] = err
			continue
		}

		subBudget.Sub(remainingBudget.Load(), amountIn)

		if subBudget.Sign() < 0 {
			errs[index] = raydium.BudgetExceededError
			continue
		}

		account := accounts[index]

		if account == nil || account.Data == nil || account.Data.GetBinary() == nil {
			continue
		}

		if task.SourceTokenMint.Equals(solana.WrappedSol) {
			if amountIn.Cmp(new(big.Int).SetUint64(account.Lamports)) > 0 {
				if minTransAmount.Cmp(new(big.Int).SetUint64(account.Lamports)) > 0 {
					continue
				}
				amountIn = new(big.Int).SetUint64(configs[index].MinTransactionsAmount)
			}
		} else {
			tokenAcc := token.Account{}
			if tokenErr := tokenAcc.UnmarshalWithDecoder(bin.NewBinDecoder(account.Data.GetBinary())); tokenErr != nil {
				continue
			}

			if amountIn.Cmp(new(big.Int).SetUint64(tokenAcc.Amount)) > 0 {
				if minTransAmount.Cmp(new(big.Int).SetUint64(tokenAcc.Amount)) > 0 {
					continue
				}
				amountIn = new(big.Int).SetUint64(configs[index].MinTransactionsAmount)
			}
		}

		tx, err = c.buildSwapTransaction(
			amountIn,
			latestHash,
			pool,
			baseParams[index],
			task,
			configs[index],
		)
		if err != nil {
			errs[index] = apperrors.Internal("failed to build tx", err)
			continue
		}
		txs = append(txs, tx)
		amounts[index] = amountIn
	}

	if raydium.IsAllErrorAre(errs, raydium.BudgetExceededError) {
		return raydium.BudgetExceededError
	} else if err = errors.Join(errs...); err != nil {
		return errors.New(err.Error())
	}

	if task.UsingJito {

		tip, err := c.jitoClient.CalculateTip(ctx, task.TransactionSpeed)
		if err != nil {
			return err
		}

		tipAccount, err := c.jitoClient.GetTipAccount(ctx)
		if err != nil {
			return err
		}

		tipTx, err := jito.BuildTipTransaction(task.PrivateKey, tip, *tipAccount, *latestHash)
		if err != nil {
			return err
		}

		bundle := append(txs, tipTx)

		err = c.jitoClient.BroadcastBundle(ctx, bundle, raydiumcpswap.ProgramID)
		if err != nil {
			return err
		}

		remainingBudget.Store(subBudget)

	} else {
		errs = make([]error, len(txs))

		for index, transaction := range txs {
			_, err = c.RPC.SendTransactionWithOpts(ctx,
				transaction,
				rpc.TransactionOpts{
					SkipPreflight:       false,
					PreflightCommitment: rpc.CommitmentProcessed,
				})

			if err != nil {
				if customErr, ok := raydiumcpswap.DecodeCustomError(err); ok {
					if customErr == raydiumcpswap.ErrExceededSlippage {
						errs[index] = fmt.Errorf("failed to send tx: %w: %w", swaperror.ErrSlippageExceeded, customErr)
						continue
					}
					errs[index] = fmt.Errorf("failed to send tx: %w: %w", swaperror.ErrCustomProgramError, customErr)
					continue
				}

				message := strings.ToLower(err.Error())
				if strings.Contains(message, "computebudgetexceeded") || strings.Contains(message, "compute budget exceeded") || strings.Contains(message, "computational budget exceeded") {
					errs[index] = fmt.Errorf("failed to send tx: %w: %w", swaperror.ErrComputeBudgetExceeded, err)
					continue
				}
				if strings.Contains(message, "429") || strings.Contains(message, "rate limit") || strings.Contains(message, "too many requests") {
					errs[index] = fmt.Errorf("failed to send tx: %w: %w", swaperror.ErrRateLimit, err)
					continue
				}
				if strings.Contains(message, "gateway timeout") || strings.Contains(message, "context deadline exceeded") || strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout") {
					errs[index] = fmt.Errorf("failed to send tx: %w: %w", swaperror.ErrGatewayTimeout, err)
					continue
				}

				errs[index] = err
				continue
			}

			budget := new(big.Int).Sub(remainingBudget.Load(), amounts[index])

			remainingBudget.Store(budget)
		}

	}

	if err = errors.Join(errs...); err != nil {
		return err
	}

	return nil
}

func (c *Client) buildSwapTransaction(
	amountIn *big.Int,
	blockHash *solana.Hash,
	pool *PoolStateWithReserve,
	params raydium.SwapParams,
	task *model.AsyncSwapTask,
	cfg raydium.TWAPConfig,
) (*solana.Transaction, error) {

	isSolToToken := task.TaskType == model.TargetUpTaskType

	params.AmmConfig = pool.PoolState.AmmConfig
	params.ObservationState = pool.PoolState.ObservationKey

	if pool.PoolState.Token0Mint.Equals(params.InputTokenMint) && pool.PoolState.Token1Mint.Equals(params.OutputTokenMint) {
		params.Token0Vault = pool.PoolState.Token0Vault
		params.Token1Vault = pool.PoolState.Token1Vault
		params.InputTokenProgramID = pool.PoolState.Token0Program
		params.OutputTokenProgramID = pool.PoolState.Token1Program
	} else {
		params.Token0Vault = pool.PoolState.Token1Vault
		params.Token1Vault = pool.PoolState.Token0Vault
		params.InputTokenProgramID = pool.PoolState.Token1Program
		params.OutputTokenProgramID = pool.PoolState.Token0Program
	}

	out, err := poolmath.ConstantProductTokenAmountOut(
		amountIn,
		pool.ReserveB,
		pool.ReserveA,
		pool.AmmConfig.TradeFeeRate,
	)
	if err != nil {
		return nil, err
	}

	minOut, err := poolmath.ConstantProductFloorWithSlippage(out, cfg.SlippageBPS)
	if err != nil {
		return nil, err
	}

	params.AmountIn = amountIn
	params.MinAmountOut = minOut

	swapInstr, err := cpmmBuildSwapInstruction(
		pool.PoolState.AmmConfig,
		params.Token0Vault,
		params.Token1Vault,
		&params,
	)
	if err != nil {
		return nil, err
	}

	instrs := []solana.Instruction{
		computebudget.NewSetComputeUnitLimitInstruction(cfg.ComputeUnitLimit).Build(),
		computebudget.NewSetComputeUnitPriceInstruction(cfg.ComputeUnitPriceMicroLamports).Build(),
	}

	instrs = append(instrs, assoc.NewCreateInstruction(params.UserWallet, params.UserWallet, solana.WrappedSol).Build())

	if isSolToToken {
		instrs = append(instrs,
			system.NewTransferInstruction(amountIn.Uint64(), params.UserWallet, params.UserSourceToken).Build(),
			token.NewSyncNativeInstruction(params.UserSourceToken).Build(),
		)

		if !task.ATAKeyCreated {
			instrs = append(instrs, assoc.NewCreateInstruction(params.UserWallet, params.UserWallet, params.OutputTokenMint).Build())
		}
	}

	instrs = append(instrs, swapInstr)

	accountToClose := params.UserSourceToken
	if !isSolToToken {
		accountToClose = params.UserDestToken
	}

	instrs = append(instrs, token.NewCloseAccountInstruction(
		accountToClose,
		params.UserWallet,
		params.UserWallet,
		[]solana.PublicKey{},
	).Build())

	tx, err := solana.NewTransaction(instrs, *blockHash, solana.TransactionPayer(task.PrivateKey.PublicKey()))
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(task.PrivateKey.PublicKey()) {
			return &task.PrivateKey
		}
		return nil
	})

	return tx, nil
}

func FetchCPMMPoolState(ctx context.Context, rpc solanarpc.SolanaRPC, params *model.PoolParams) (*PoolStateWithReserve, error) {
	rpcWithRetry := solanarpc.WithRetries(rpc, DefaultRPCRetries)

	accounts, err := rpcWithRetry.GetMultipleAccounts(ctx, params.InputTokenVault, params.OutputTokenVault, params.PoolID, params.AmmConfig)
	if err != nil {
		return nil, err
	}

	if len(accounts.Value) != 4 {
		return nil, apperrors.Internal(fmt.Sprintf("failed to fetch token vaults: %d accounts returned", len(accounts.Value)))
	}

	for i, acc := range accounts.Value {
		if acc == nil || acc.Data == nil {
			return nil, apperrors.Internal(fmt.Sprintf("account at index %d is nil", i))
		}
	}

	rawPoolData, err := raydiumcpswap.ParseAccount_PoolState(accounts.Value[2].Data.GetBinary())
	if err != nil {
		return nil, apperrors.Internal("failed to parse pool data", err)
	}

	poolState := &PoolStateWithReserve{
		PoolState: rawPoolData,
	}

	token0Data := accounts.Value[0].Data.GetBinary()
	token1Data := accounts.Value[1].Data.GetBinary()

	token0 := token.Account{}
	token1 := token.Account{}

	if err := token0.UnmarshalWithDecoder(bin.NewBinDecoder(token0Data)); err != nil {
		return nil, apperrors.Internal("failed to parse token0 vault", err)
	}
	if err := token1.UnmarshalWithDecoder(bin.NewBinDecoder(token1Data)); err != nil {
		return nil, apperrors.Internal("failed to parse token1 vault", err)
	}

	poolState.ReserveA = new(big.Int).SetUint64(token0.Amount)
	poolState.ReserveB = new(big.Int).SetUint64(token1.Amount)

	configData, err := raydiumcpswap.ParseAccount_AmmConfig(accounts.Value[3].Data.GetBinary())
	if err != nil {
		return nil, apperrors.Internal("failed to parse AmmConfig: %w", err)
	}

	poolState.AmmConfig = configData

	return poolState, nil
}

func calculateCPMMPrice(pool *PoolStateWithReserve, params raydium.SwapParams) *big.Rat {
	if pool.PoolState.Token0Mint.Equals(params.InputTokenMint) {
		return poolmath.ConstantProductCalculatePrice(pool.ReserveA, pool.ReserveB, uint64(pool.PoolState.Mint0Decimals), uint64(pool.PoolState.Mint1Decimals))
	}
	return poolmath.ConstantProductCalculatePrice(pool.ReserveB, pool.ReserveA, uint64(pool.PoolState.Mint1Decimals), uint64(pool.PoolState.Mint0Decimals))
}
