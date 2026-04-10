package pump_amm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"mm/internal/client/jito"
	pumpammclient "mm/internal/client/pumpfun/amm/amm_client"
	"mm/internal/client/raydium"
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
)

type Client struct {
	RPC                   solanarpc.SolanaRPC
	jitoClient            *jito.Client
	campaignRepository    *repository.SwapCampaignRepository
	transactionRepository *repository.SwapTransactionRepository
	logger                *zap.Logger
}

func NewClient(
	rpc solanarpc.SolanaRPC,
	jitoClient *jito.Client,
	campaignRepository *repository.SwapCampaignRepository,
	transactionRepository *repository.SwapTransactionRepository,
	logger *zap.Logger,
) *Client {
	return &Client{
		RPC:                   rpc,
		jitoClient:            jitoClient,
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
	if len(tasks) == 0 || len(tasks) != len(baseParams) {
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

	pool, err := FetchAMMPoolState(ctx, c.RPC, task.PoolParams, baseParam.InputTokenMint, baseParam.OutputTokenMint)
	if err != nil {
		return apperrors.Internal("failed to fetch pool state", err)
	}

	currentPrice := calculateAMMPrice(pool, baseParam)
	if raydium.IsTargetReached(currentPrice, task.GoalPrice, task.TaskType) {
		return raydium.PriceIsAlreadyReachedError
	}

	minTransAmount := new(big.Int).SetUint64(cfg.MinTransactionsAmount)
	maxTransAmount := new(big.Int).SetUint64(cfg.MaxTransactionsAmount)

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
		amountIn, err = raydium.RandBigIntRange(minTransAmount, maxTransAmount)
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
		if account == nil {
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
			if account.Data == nil || account.Data.GetBinary() == nil {
				continue
			}

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

		swapParams, sErr := FetchAMMSwapParams(ctx, c.RPC, task.PoolID, task.PrivateKey.PublicKey())
		if sErr != nil {
			errs[index] = apperrors.Internal("failed to fetch pump amm swap params", sErr)
			continue
		}

		tx, err = c.buildSwapTransaction(
			amountIn,
			latestHash,
			pool,
			swapParams,
			task,
			configs[index],
		)
		if err != nil {
			errs[index] = apperrors.Internal("failed to build tx", err)
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
				continue
			}
			if strings.Contains(message, "gateway timeout") || strings.Contains(message, "context deadline exceeded") || strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout") {
				errs[index] = fmt.Errorf("failed to simulate tx: %w: %w", swaperror.ErrGatewayTimeout, sErr)
				continue
			}
			errs[index] = fmt.Errorf("failed to simulate tx: %w: %w", swaperror.ErrSimulationError, sErr)
			continue
		}
		if simulated != nil && simulated.Value.Err != nil {
			simErr := fmt.Errorf("simulation failed: %v, logs: %v", simulated.Value.Err, simulated.Value.Logs)
			message := strings.ToLower(simErr.Error())
			if strings.Contains(message, "slippage") {
				errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrSlippageExceeded, simErr)
				continue
			}
			if strings.Contains(message, "insufficient funds") || strings.Contains(message, "empty funds") || strings.Contains(message, "input token account empty") {
				errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrInsufficientFunds, simErr)
				continue
			}
			if strings.Contains(message, "custom program error") {
				errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrCustomProgramError, simErr)
				continue
			}
			if strings.Contains(message, "compute budget exceeded") || strings.Contains(message, "computebudgetexceeded") || strings.Contains(message, "computational budget exceeded") || strings.Contains(message, "exceeded the maximum number of instructions allowed") {
				errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrComputeBudgetExceeded, simErr)
				continue
			}
			errs[index] = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrSimulationError, simErr)
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
		if err = c.jitoClient.BroadcastBundle(ctx, bundle, ProgramID); err != nil {
			return err
		}

		remainingBudget.Store(subBudget)
		return nil
	}

	errs = make([]error, len(txs))
	for index, transaction := range txs {
		_, err = c.RPC.SendTransactionWithOpts(ctx,
			transaction,
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

		budget := new(big.Int).Sub(remainingBudget.Load(), amounts[index])
		remainingBudget.Store(budget)
	}

	return errors.Join(errs...)
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
		params.GlobalConfig.LpFeeBasisPoints,
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

	swapInstr, err := buildSwapInstruction(
		params,
		task.PrivateKey.PublicKey(),
		task.SourceTokenMint,
		task.DestTokenMint,
		amountInU64,
		minOutU64,
		pumpammclient.OptionBool{V0: false},
	)
	if err != nil {
		return nil, err
	}

	instrs := []solana.Instruction{
		computebudget.NewSetComputeUnitLimitInstruction(cfg.ComputeUnitLimit).Build(),
		computebudget.NewSetComputeUnitPriceInstruction(cfg.ComputeUnitPriceMicroLamports).Build(),
	}

	if isSolToToken { // BUY token
		expectedWSOLATA, _, err := solana.FindAssociatedTokenAddress(
			task.PrivateKey.PublicKey(),
			solana.WrappedSol,
		)
		if err != nil {
			return nil, err
		}

		if !expectedWSOLATA.Equals(task.SourceAddress) {
			return nil, fmt.Errorf(
				"source address mismatch: expected WSOL ATA %s, got %s",
				expectedWSOLATA,
				task.SourceAddress,
			)
		}
		instrs = append(instrs, assoc.NewCreateInstruction(task.PrivateKey.PublicKey(), task.PrivateKey.PublicKey(), solana.WrappedSol).Build())

		instrs = append(instrs,
			system.NewTransferInstruction(amountIn.Uint64(), task.PrivateKey.PublicKey(), task.SourceAddress).Build(),
			token.NewSyncNativeInstruction(task.SourceAddress).Build(),
		)

		if !task.ATAKeyCreated {
			instrs = append(instrs, assoc.NewCreateInstruction(task.PrivateKey.PublicKey(), task.PrivateKey.PublicKey(), task.DestTokenMint).Build())
		}
	} else { // SELL token
		expectedWSOLATA, _, err := solana.FindAssociatedTokenAddress(
			task.PrivateKey.PublicKey(),
			solana.WrappedSol,
		)
		if err != nil {
			return nil, err
		}

		if !expectedWSOLATA.Equals(task.DestAddress) {
			return nil, fmt.Errorf(
				"dest address mismatch: expected WSOL ATA %s, got %s",
				expectedWSOLATA,
				task.DestAddress,
			)
		}

		instrs = append(instrs,
			assoc.NewCreateInstruction(
				task.PrivateKey.PublicKey(),
				task.PrivateKey.PublicKey(),
				solana.WrappedSol,
			).Build(),
		)
	}

	instrs = append(instrs, swapInstr)

	accountToClose := task.SourceAddress // SOL -> token
	if !isSolToToken {                   // token -> SOL
		accountToClose = task.DestAddress
	}

	instrs = append(instrs, token.NewCloseAccountInstruction(
		accountToClose,
		task.PrivateKey.PublicKey(),
		task.PrivateKey.PublicKey(),
		[]solana.PublicKey{},
	).Build())

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

func calculateAMMPrice(pool *PoolStateWithReserve, params raydium.SwapParams) *big.Rat {
	if pool.PoolState.BaseMint.Equals(params.OutputTokenMint) && pool.PoolState.QuoteMint.Equals(params.InputTokenMint) {
		return poolmath.ConstantProductCalculatePrice(pool.ReserveA, pool.ReserveB, uint64(pool.BaseMintDecimals), uint64(pool.QuoteMintDecimals))
	}

	return poolmath.ConstantProductCalculatePrice(pool.ReserveB, pool.ReserveA, uint64(pool.QuoteMintDecimals), uint64(pool.BaseMintDecimals))
}
