package bonding

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"mm/internal/client/jito"
	pump_bonding "mm/internal/client/pumpfun/bonding/bonding_client"
	poolmath "mm/internal/client/pumpfun/math"
	"mm/internal/client/raydium"
	"mm/internal/client/solanarpc"
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
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Client struct {
	RPC        solanarpc.SolanaRPC
	jitoClient *jito.Client
	logger     *zap.Logger
}

func NewClient(
	rpc solanarpc.SolanaRPC,
	jitoClient *jito.Client,
	logger *zap.Logger,
) *Client {
	return &Client{
		RPC:        rpc,
		jitoClient: jitoClient,
		logger:     logger,
	}
}

type reservedTx struct {
	tx     *solana.Transaction
	amount *big.Int
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
	if len(tasks) == 0 || len(tasks) != len(baseParams) {
		return nil, apperrors.BadRequest("invalid params length")
	}

	task := tasks[0]
	baseParam := baseParams[0]
	cfg := configs[0]
	logParams := swaptxlog.Params{
		PoolID:        baseParam.PoolID.String(),
		TokenMintFrom: baseParam.InputTokenMint.String(),
		TokenMintTo:   baseParam.OutputTokenMint.String(),
		AddressFrom:   baseParam.UserSourceToken.String(),
		AddressTo:     baseParam.UserDestToken.String(),
	}

	if remainingBudget.Remaining().Sign() <= 0 {
		return []swaptxlog.Result{{Params: logParams, Err: swaperror.BudgetExceededError}}, swaperror.BudgetExceededError
	}

	pool, err := FetchBondingCurveState(ctx, c.RPC, task.PoolParams, baseParam.InputTokenMint, baseParam.OutputTokenMint)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch bonding curve state", err)
	}

	currentPrice, err := calculateBondingCurvePrice(pool, baseParam.InputTokenMint)
	if err != nil {
		return nil, apperrors.Internal("failed to calculate bonding price", err)
	}

	if raydium.IsTargetReached(currentPrice, task.GoalPrice, task.TaskType) {
		return nil, raydium.PriceIsAlreadyReachedError
	}

	minTransAmount := new(big.Int).SetUint64(cfg.MinTransactionsAmount)
	maxTransAmount := new(big.Int).SetUint64(cfg.MaxTransactionsAmount)

	txs := make([]reservedTx, 0, len(tasks))
	defer func() {
		for _, item := range txs {
			params := logParams
			params.TransactionHash = item.tx.Signatures[0].String()
			results = append(results, swaptxlog.Result{Params: params, Err: err})
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

	errs := make([]error, len(tasks))
	tokenMint := baseParam.InputTokenMint
	if solutil.IsSOLLikeMint(tokenMint) {
		tokenMint = baseParam.OutputTokenMint
	}

	for index, task := range tasks {
		amountIn, err := remainingBudget.Reserve(minTransAmount, maxTransAmount)
		if err != nil {
			errs[index] = err
			continue
		}
		reservedAmount := new(big.Int).Set(amountIn)

		account := accounts[index]
		if account == nil {
			remainingBudget.Release(reservedAmount)
			continue
		}

		if task.SourceTokenMint.Equals(solana.WrappedSol) {
			if amountIn.Cmp(new(big.Int).SetUint64(account.Lamports)) > 0 {
				if minTransAmount.Cmp(new(big.Int).SetUint64(account.Lamports)) > 0 {
					remainingBudget.Release(reservedAmount)
					continue
				}
				amountIn = new(big.Int).SetUint64(configs[index].MinTransactionsAmount)
			}
		} else {
			if account.Data == nil || account.Data.GetBinary() == nil {
				remainingBudget.Release(reservedAmount)
				continue
			}

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

		swapParams, sErr := FetchBondingSwapParams(ctx, c.RPC, task.PoolID, task.PrivateKey.PublicKey(), tokenMint)
		if sErr != nil {
			errs[index] = apperrors.Internal("failed to fetch pump bonding swap params", sErr)
			remainingBudget.Release(reservedAmount)
			continue
		}

		tx, err := c.buildSwapTransaction(
			amountIn,
			latestHash,
			pool,
			swapParams,
			task,
			configs[index],
		)
		if err != nil {
			errs[index] = apperrors.Internal("failed to build tx", err)
			remainingBudget.Release(reservedAmount)
			continue
		}

		simulated, sErr := c.RPC.SimulateTransaction(ctx, tx, &rpc.SimulateTransactionOpts{
			SigVerify:              false,
			Commitment:             rpc.CommitmentProcessed,
			ReplaceRecentBlockhash: true,
		})
		if sErr != nil {
			message := strings.ToLower(sErr.Error())
			if strings.Contains(message, "429") || strings.Contains(message, "rate limit") || strings.Contains(message, "too many requests") {
				errs[index] = fmt.Errorf("failed to simulate tx: %w: %w", swaperror.ErrRateLimit, sErr)
				remainingBudget.Release(reservedAmount)
				continue
			}
			if strings.Contains(message, "gateway timeout") || strings.Contains(message, "context deadline exceeded") || strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout") {
				errs[index] = fmt.Errorf("failed to simulate tx: %w: %w", swaperror.ErrGatewayTimeout, sErr)
				remainingBudget.Release(reservedAmount)
				continue
			}
			errs[index] = fmt.Errorf("failed to simulate tx: %w: %w", swaperror.ErrSimulationError, sErr)
			remainingBudget.Release(reservedAmount)
			continue
		}
		if simulated != nil && simulated.Value.Err != nil {
			simErr := fmt.Errorf("simulation failed: %v, logs: %v", simulated.Value.Err, simulated.Value.Logs)
			message := strings.ToLower(simErr.Error())
			if strings.Contains(message, "slippage") {
				errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrSlippageExceeded, simErr)
				remainingBudget.Release(reservedAmount)
				continue
			}
			if strings.Contains(message, "insufficient funds") || strings.Contains(message, "empty funds") || strings.Contains(message, "input token account empty") {
				errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrInsufficientFunds, simErr)
				remainingBudget.Release(reservedAmount)
				continue
			}
			if strings.Contains(message, "custom program error") {
				errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrCustomProgramError, simErr)
				remainingBudget.Release(reservedAmount)
				continue
			}
			if strings.Contains(message, "compute budget exceeded") || strings.Contains(message, "computebudgetexceeded") || strings.Contains(message, "computational budget exceeded") || strings.Contains(message, "exceeded the maximum number of instructions allowed") {
				errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrComputeBudgetExceeded, simErr)
				remainingBudget.Release(reservedAmount)
				continue
			}
			errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrSimulationError, simErr)
			remainingBudget.Release(reservedAmount)
			continue
		}

		txs = append(txs, reservedTx{
			tx:     tx,
			amount: reservedAmount,
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
	// return nil, nil
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

		bundle := make([]*solana.Transaction, 0, len(txs)+1)
		for _, item := range txs {
			bundle = append(bundle, item.tx)
		}
		bundle = append(bundle, tipTx)

		if err = c.jitoClient.BroadcastBundle(ctx, bundle, ProgramID); err != nil {
			if errors.Is(err, swaperror.ErrBundleRejected) {
				releaseAll(remainingBudget, txs)
			}
			return nil, err
		}

		return nil, nil
	}
	errs = make([]error, len(txs))
	for index, item := range txs {
		_, err = c.RPC.SendTransactionWithOpts(ctx,
			item.tx,
			rpc.TransactionOpts{
				SkipPreflight:       false,
				PreflightCommitment: rpc.CommitmentProcessed,
			})
		if err != nil {
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

	return nil, errors.Join(errs...)
}

func (c *Client) buildSwapTransaction(
	amountIn *big.Int,
	blockHash *solana.Hash,
	pool *PoolStateWithReserve,
	params *SwapParams,
	task *model.AsyncSwapTask,
	cfg raydium.TWAPConfig,
) (*solana.Transaction, error) {
	isSolToToken := task.TaskType == model.TargetUpTaskType

	out, err := poolmath.ConstantProductTokenAmountOut(
		amountIn,
		pool.ReserveA,
		pool.ReserveB,
		0,
	)
	if err != nil {
		return nil, err
	}

	minOut, err := poolmath.ConstantProductFloorWithSlippage(out, cfg.SlippageBPS)
	if err != nil {
		return nil, err
	}

	amountInU64, err := raydium.SafeUint64(amountIn)
	if err != nil {
		return nil, err
	}

	minOutU64, err := raydium.SafeUint64(minOut)
	if err != nil {
		return nil, err
	}

	if isSolToToken {
		if !params.AssociatedUserAccount.Equals(task.DestAddress) {
			return nil, fmt.Errorf(
				"dest address mismatch: expected associated user account %s, got %s",
				params.AssociatedUserAccount,
				task.DestAddress,
			)
		}
	} else if !params.AssociatedUserAccount.Equals(task.SourceAddress) {
		return nil, fmt.Errorf(
			"source address mismatch: expected associated user account %s, got %s",
			params.AssociatedUserAccount,
			task.SourceAddress,
		)
	}

	swapInstr, err := buildSwapInstruction(
		params,
		task.PrivateKey.PublicKey(),
		task.SourceTokenMint,
		task.DestTokenMint,
		amountInU64,
		minOutU64,
		pump_bonding.OptionBool{V0: false},
	)
	if err != nil {
		return nil, err
	}

	instrs := make([]solana.Instruction, 0, 2)
	if isSolToToken && !task.ATAKeyCreated {
		createDestATAInstr, err := solutil.NewCreateAssociatedTokenAccountInstruction(
			task.PrivateKey.PublicKey(),
			task.PrivateKey.PublicKey(),
			task.DestTokenMint,
			params.TokenProgram,
		)
		if err != nil {
			return nil, err
		}
		instrs = append(instrs, createDestATAInstr)
	}
	instrs = append(instrs, swapInstr)

	tx, err := solana.NewTransaction(
		instrs,
		*blockHash,
		solana.TransactionPayer(task.PrivateKey.PublicKey()),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(task.PrivateKey.PublicKey()) {
			return &task.PrivateKey
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func buildSwapInstruction(
	params *SwapParams,
	user solana.PublicKey,
	inputTokenMint solana.PublicKey,
	outputTokenMint solana.PublicKey,
	amountIn uint64,
	minOut uint64,
	trackVolume pump_bonding.OptionBool,
) (solana.Instruction, error) {
	isBuy := solutil.IsSOLLikeMint(inputTokenMint) && params.Mint.Equals(outputTokenMint)
	isSell := params.Mint.Equals(inputTokenMint) && solutil.IsSOLLikeMint(outputTokenMint)

	switch {
	case isBuy:
		return buildBuyExactSolInInstruction(
			params,
			user,
			amountIn,
			minOut,
			trackVolume,
		)
	case isSell:
		return buildSellInstruction(
			params,
			user,
			amountIn,
			minOut,
		)
	default:
		return nil, fmt.Errorf("input/output mints do not match bonding curve")
	}
}

func buildBuyExactSolInInstruction(
	params *SwapParams,
	user solana.PublicKey,
	spendableSolIn uint64,
	minTokensOut uint64,
	trackVolume pump_bonding.OptionBool,
) (solana.Instruction, error) {
	instruction, err := pump_bonding.NewBuyExactSolInInstruction(
		spendableSolIn,
		minTokensOut,
		trackVolume,

		params.GlobalID,
		params.FeeRecipient,
		params.Mint,
		params.BondingCurveID,
		params.AssociatedBondingCurve,
		params.AssociatedUserAccount,
		user,
		params.SystemProgram,
		params.TokenProgram,
		params.CreatorVault,
		params.EventAuthority,
		params.Program,
		params.GlobalVolumeAccumulator,
		params.UserVolumeAccumulator,
		params.FeeConfig,
		params.FeeProgram,
	)
	if err != nil {
		return nil, err
	}

	if genericInstruction, ok := instruction.(*solana.GenericInstruction); ok {
		genericInstruction.AccountValues = append(
			genericInstruction.AccountValues,
			solana.Meta(params.BondingCurveV2),
		)
		genericInstruction.AccountValues = append(
			genericInstruction.AccountValues,
			solana.Meta(params.FeeRecipientNew).WRITE(),
		)
	}

	return instruction, nil
}

func buildSellInstruction(
	params *SwapParams,
	user solana.PublicKey,
	amount uint64,
	minSolOutput uint64,
) (solana.Instruction, error) {
	instruction, err := pump_bonding.NewSellInstruction(
		amount,
		minSolOutput,

		params.GlobalID,
		params.FeeRecipient,
		params.Mint,
		params.BondingCurveID,
		params.AssociatedBondingCurve,
		params.AssociatedUserAccount,
		user,
		params.SystemProgram,
		params.CreatorVault,
		params.TokenProgram,
		params.EventAuthority,
		params.Program,
		params.FeeConfig,
		params.FeeProgram,
	)
	if err != nil {
		return nil, err
	}

	if genericInstruction, ok := instruction.(*solana.GenericInstruction); ok {
		genericInstruction.AccountValues = append(
			genericInstruction.AccountValues,
			solana.Meta(params.BondingCurveV2),
		)
		genericInstruction.AccountValues = append(
			genericInstruction.AccountValues,
			solana.Meta(params.FeeRecipientNew).WRITE(),
		)
	}

	return instruction, nil
}

func calculateBondingCurvePrice(pool *PoolStateWithReserve, inputTokenMint solana.PublicKey) (*big.Rat, error) {
	if solutil.IsSOLLikeMint(inputTokenMint) {
		// pullup: ReserveA=SOL, ReserveB=token -> SOL/token
		return poolmath.ConstantProductCalculatePrice(
			pool.ReserveB,
			pool.ReserveA,
			uint64(pool.BaseMintDecimals),
			uint64(pool.QuoteMintDecimals),
		), nil
	}
	// pulldown: ReserveA=token, ReserveB=SOL -> SOL/token
	return poolmath.ConstantProductCalculatePrice(
		pool.ReserveA,
		pool.ReserveB,
		uint64(pool.BaseMintDecimals),
		uint64(solana.SolDecimals),
	), nil
}
