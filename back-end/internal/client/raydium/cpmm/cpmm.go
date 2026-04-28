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
	"mm/internal/common"
	"mm/internal/model"
	"mm/internal/swapbudget"
	"mm/internal/swaperror"
	"mm/internal/swaptxlog"
	"mm/pkg/apperrors"
	"mm/pkg/solutil"
	"strings"
	"sync/atomic"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
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
	RPC         solanarpc.SolanaRPC
	restyClient *resty.Client
	jitoClient  *jito.Client
	URL         string
	logger      *zap.Logger
}

func NewClient(
	rpc solanarpc.SolanaRPC,
	client *resty.Client,
	jitoClient *jito.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *Client {
	return &Client{
		RPC:         rpc,
		restyClient: client,
		jitoClient:  jitoClient,
		URL:         cfg.App.KucoinBaseUrl,
		logger:      logger,
	}
}

type reservedTx struct {
	tx     *solana.Transaction
	amount *big.Int
	params swaptxlog.Params
}

func releaseAll(b *swapbudget.SwapBudget, items []reservedTx) {
	for _, item := range items {
		if item.amount != nil {
			b.Release(item.amount)
		}
	}
}

func (c *Client) Swap(
	ctx context.Context,
	campaignID uuid.UUID,
	targetID *uuid.UUID,
	remainingBudget *swapbudget.SwapBudget,
	tasks []*model.AsyncSwapTask,
	baseParams []raydium.SwapParams,
	configs []raydium.TWAPConfig,
	blockHash *atomic.Pointer[solana.Hash],
) (results []swaptxlog.Result, err error) {

	if len(tasks) == 0 || len(tasks) != len(baseParams) || len(tasks) != len(configs) {
		return nil, apperrors.BadRequest("invalid params length")
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

	// if remainingBudget.Load().Sign() <= 0 {
	// 	return []swaptxlog.Result{{Params: logParams, Err: swaperror.BudgetExceededError}}, swaperror.BudgetExceededError
	// }

	if remainingBudget.Remaining().Sign() <= 0 {
		return []swaptxlog.Result{{Params: logParams, Err: swaperror.BudgetExceededError}}, swaperror.BudgetExceededError
	}

	pool, err := FetchCPMMPoolState(ctx, c.RPC, task.PoolParams)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch pool state", err)
	}

	currentPrice := calculateCPMMPrice(pool, baseParam)
	if raydium.IsTargetReached(currentPrice, task.GoalPrice, task.TaskType) {
		return nil, raydium.PriceIsAlreadyReachedError
	}

	minTransAmount := new(big.Int).SetUint64(cfg.MinTransactionsAmount)
	maxTransAmount := new(big.Int).SetUint64(cfg.MaxTransactionsAmount)

	// if remainingBudget.Load().Sign() <= 0 {
	// 	return []swaptxlog.Result{{Params: logParams, Err: swaperror.BudgetExceededError}}, swaperror.BudgetExceededError
	// }

	txs := make([]reservedTx, 0, len(tasks))

	defer func() {
		for _, item := range txs {
			p := item.params
			p.TransactionHash = item.tx.Signatures[0].String()
			results = append(results, swaptxlog.Result{Params: p, Err: err})
		}
	}()

	latestHash := blockHash.Load()
	if latestHash == nil {
		latestBlockHash, err := c.RPC.GetLatestBlockhash(ctx)
		if err != nil {
			return nil, err
		}
		latestHash = latestBlockHash
		blockHash.Store(latestBlockHash)
	}

	keys := make([]solana.PublicKey, len(tasks))

	for index, task := range tasks {
		if task.SourceTokenMint.Equals(solana.WrappedSol) {
			keys[index] = task.PrivateKey.PublicKey()
		} else {
			keys[index] = task.SourceAddress
		}
	}

	rpcResults, err := c.RPC.GetMultipleAccountsWithNoLimits(ctx, keys...)
	if err != nil {
		return nil, err
	}

	accounts := make([]*rpc.Account, 0, len(keys))

	for _, result := range rpcResults {
		accounts = append(accounts, result.Value...)
	}

	ataRentLamports, err := c.RPC.GetATARentExemption(ctx)
	if err != nil {
		return nil, err
	}

	errs := make([]error, len(tasks))

	for index, task := range tasks {
		amountIn, err := remainingBudget.Reserve(minTransAmount, maxTransAmount)
		if err != nil {
			errs[index] = err
			continue
		}
		reservedAmount := new(big.Int).Set(amountIn)

		account := accounts[index]
		if account == nil || account.Data == nil || account.Data.GetBinary() == nil {
			remainingBudget.Release(reservedAmount)
			continue
		}

		if task.SourceTokenMint.Equals(solana.WrappedSol) {
			reserveLamports := common.SolPayerReserveLamports(
				!task.ATAKeyCreated,
				ataRentLamports,
				configs[index].ComputeUnitLimit,
				configs[index].ComputeUnitPriceMicroLamports,
			)
			if account.Lamports <= reserveLamports {
				errs[index] = fmt.Errorf(
					"insufficient SOL reserve wallet %s balance %d reserve %d: %w",
					task.PrivateKey.PublicKey().String(),
					account.Lamports,
					reserveLamports,
					swaperror.ErrInsufficientFunds,
				)
				remainingBudget.Release(reservedAmount)
				continue
			}

			spendableLamports := new(big.Int).SetUint64(account.Lamports - reserveLamports)
			if amountIn.Cmp(spendableLamports) > 0 {
				if minTransAmount.Cmp(spendableLamports) > 0 {
					errs[index] = fmt.Errorf(
						"insufficient SOL for min tx amount wallet %s spendable %d needed %d: %w",
						task.PrivateKey.PublicKey().String(),
						spendableLamports,
						minTransAmount,
						swaperror.ErrInsufficientFunds,
					)
					remainingBudget.Release(reservedAmount)
					continue
				}
				amountIn = spendableLamports
			}
		} else {
			tokenAcc := token.Account{}
			if tokenErr := tokenAcc.UnmarshalWithDecoder(bin.NewBinDecoder(account.Data.GetBinary())); tokenErr != nil {
				remainingBudget.Release(reservedAmount)
				continue
			}

			if amountIn.Cmp(new(big.Int).SetUint64(tokenAcc.Amount)) > 0 {
				if minTransAmount.Cmp(new(big.Int).SetUint64(tokenAcc.Amount)) > 0 {
					remainingBudget.Release(reservedAmount)
					continue
				}
				amountIn = new(big.Int).SetUint64(configs[index].MinTransactionsAmount)
			}
		}
		// if reserved 10 and amountIn reduced to 4, need to release diff (6) now
		if amountIn.Cmp(reservedAmount) < 0 {
			diff := new(big.Int).Sub(reservedAmount, amountIn)
			remainingBudget.Release(diff)
			reservedAmount = new(big.Int).Set(amountIn)
		}

		tx, err := c.buildSwapTransaction(
			amountIn,
			latestHash,
			pool,
			baseParams[index],
			task,
			configs[index],
		)
		if err != nil {
			errs[index] = apperrors.Internal("failed to build tx", err)
			remainingBudget.Release(reservedAmount)
			continue
		}

		txs = append(txs, reservedTx{
			tx:     tx,
			amount: reservedAmount,
			params: swaptxlog.Params{
				PoolID:        baseParams[index].PoolID.String(),
				TokenMintFrom: baseParams[index].InputTokenMint.String(),
				TokenMintTo:   baseParams[index].OutputTokenMint.String(),
				AddressFrom:   baseParams[index].UserSourceToken.String(),
				AddressTo:     baseParams[index].UserDestToken.String(),
			},
		})
	}

	// do not return and cancel valid transactions if some returned an error
	buildErr := errors.Join(errs...)
	if len(txs) == 0 {
		if raydium.IsAllErrorAre(errs, swaperror.BudgetExceededError) {
			return nil, swaperror.BudgetExceededError
		}
		if buildErr != nil {
			return nil, errors.New(buildErr.Error())
		}
		return nil, nil
	}
	if buildErr != nil {
		c.logger.Warn("some swap txs were skipped", zap.Error(buildErr))
	}

	for _, a := range txs {
		c.logger.Info("buyback tx publish disabled",
			zap.String("campaign_id", campaignID.String()),
			zap.String("target_id", func() string {
				if targetID == nil {
					niluuid := uuid.Nil
					targetID = &niluuid
				}
				return targetID.String()
			}()),
			zap.String("direction", fmt.Sprintf("%s -> %s", task.SourceTokenMint.String(), task.DestTokenMint.String())),
			zap.String("tx", a.tx.Signatures[0].String()),
			zap.String("amount_atomic", a.amount.String()),
			zap.Float64("amount", solanarpc.FromAtomicUnit(a.amount.Uint64(), task.SourceTokenDecimals)),
			zap.String("current_price", currentPrice.RatString()),
			zap.String("goal_price", task.GoalPrice.RatString()),
			zap.String("remaining_budget", remainingBudget.Remaining().String()),
		)
	}
	// dry run
	return nil, nil
	if task.UsingJito {
		tip, err := c.jitoClient.CalculateTip(ctx, task.TransactionSpeed)
		if err != nil {
			releaseAll(remainingBudget, txs)
			return nil, err
		}

		tipAccount, err := c.jitoClient.GetTipAccount(ctx)
		if err != nil {
			releaseAll(remainingBudget, txs)
			return nil, err
		}

		tipTx, err := jito.BuildTipTransaction(task.PrivateKey, tip, *tipAccount, *latestHash)
		if err != nil {
			releaseAll(remainingBudget, txs)
			return nil, err
		}

		// bundle := append(txs, tipTx)
		bundle := make([]*solana.Transaction, 0, len(txs)+1)
		for _, item := range txs {
			bundle = append(bundle, item.tx)
		}
		bundle = append(bundle, tipTx)

		err = c.jitoClient.BroadcastBundle(ctx, bundle, raydiumcpswap.ProgramID)
		if err != nil {
			if errors.Is(err, swaperror.ErrBundleRejected) {
				releaseAll(remainingBudget, txs)
			}
			return nil, err
		}

	} else {
		errs = make([]error, len(txs))

		for index, item := range txs {
			_, err = c.RPC.SendTransactionWithOpts(ctx,
				item.tx,
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
		}
	}

	if err = errors.Join(errs...); err != nil {
		return nil, err
	}

	return nil, nil
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

	createWSOLATAInstr, err := solutil.NewCreateAssociatedTokenAccountInstruction(
		params.UserWallet,
		params.UserWallet,
		solana.WrappedSol,
		solana.TokenProgramID,
	)
	if err != nil {
		return nil, err
	}
	instrs = append(instrs, createWSOLATAInstr)

	if isSolToToken {
		instrs = append(instrs,
			system.NewTransferInstruction(amountIn.Uint64(), params.UserWallet, params.UserSourceToken).Build(),
			token.NewSyncNativeInstruction(params.UserSourceToken).Build(),
		)

		if !task.ATAKeyCreated {
			createOutputATAInstr, err := solutil.NewCreateAssociatedTokenAccountInstruction(
				params.UserWallet,
				params.UserWallet,
				params.OutputTokenMint,
				params.OutputTokenProgramID,
			)
			if err != nil {
				return nil, err
			}
			instrs = append(instrs, createOutputATAInstr)
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

func calculateCPMMPrice(pool *PoolStateWithReserve, _ raydium.SwapParams) *big.Rat {
	if solutil.IsSOLLikeMint(pool.PoolState.Token0Mint) {
		// Token0=SOL: ReserveA=SOL, ReserveB=token -> SOL/token
		return poolmath.ConstantProductCalculatePrice(pool.ReserveB, pool.ReserveA, uint64(pool.PoolState.Mint1Decimals), uint64(pool.PoolState.Mint0Decimals))
	}
	// Token0=token: ReserveA=token, ReserveB=SOL -> SOL/token
	return poolmath.ConstantProductCalculatePrice(pool.ReserveA, pool.ReserveB, uint64(pool.PoolState.Mint0Decimals), uint64(pool.PoolState.Mint1Decimals))
}
