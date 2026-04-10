package bonding

import (
	"context"
	"fmt"
	"math/big"
	"mm/internal/client/jito"
	pump_bonding "mm/internal/client/pumpfun/bonding/bonding_client"
	poolmath "mm/internal/client/pumpfun/math"
	"mm/internal/client/raydium"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/internal/swaperror"
	"mm/internal/swaptxlog"
	"mm/pkg/apperrors"
	"mm/pkg/solutil"
	"strings"
	"sync/atomic"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	assoc "github.com/gagliardetto/solana-go/programs/associated-token-account"
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
	baseParam := baseParams[0]
	cfg := configs[0]
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

	pool, err := FetchBondingCurveState(ctx, c.RPC, task.PoolParams, baseParam.InputTokenMint, baseParam.OutputTokenMint)
	if err != nil {
		return apperrors.Internal("failed to fetch bonding curve state", err)
	}

	currentPrice, err := calculateBondingCurvePrice(pool)
	if err != nil {
		return apperrors.Internal("failed to calculate bonding price", err)
	}

	if raydium.IsTargetReached(currentPrice, task.GoalPrice, task.TaskType) {
		return raydium.PriceIsAlreadyReachedError
	}

	latestHash, err := c.RPC.GetLatestBlockhash(ctx)
	if err != nil {
		return err
	}
	blockHash.Store(latestHash)

	key := task.SourceAddress
	if solutil.IsSOLLikeMint(task.SourceTokenMint) {
		key = task.PrivateKey.PublicKey()
	}

	account, err := c.RPC.GetAccountInfo(ctx, key)
	if err != nil {
		return err
	}

	minTransAmount := new(big.Int).SetUint64(cfg.MinTransactionsAmount)
	maxTransAmount := new(big.Int).SetUint64(cfg.MaxTransactionsAmount)

	amountIn, err := raydium.RandBigIntRange(minTransAmount, maxTransAmount)
	if err != nil {
		return err
	}

	subBudget := new(big.Int).Sub(remainingBudget.Load(), amountIn)
	if subBudget.Sign() < 0 {
		return raydium.BudgetExceededError
	}

	if account != nil && account.Value != nil {
		if solutil.IsSOLLikeMint(task.SourceTokenMint) {
			if amountIn.Cmp(new(big.Int).SetUint64(account.Value.Lamports)) > 0 {
				if minTransAmount.Cmp(new(big.Int).SetUint64(account.Value.Lamports)) > 0 {
					return fmt.Errorf("not enough source balance: %w", swaperror.ErrInsufficientFunds)
				}
				amountIn = new(big.Int).SetUint64(cfg.MinTransactionsAmount)
			}
		} else if account.Value.Data != nil && account.Value.Data.GetBinary() != nil {
			tokenAcc := token.Account{}
			if tokenErr := tokenAcc.UnmarshalWithDecoder(bin.NewBinDecoder(account.Value.Data.GetBinary())); tokenErr == nil {
				if amountIn.Cmp(new(big.Int).SetUint64(tokenAcc.Amount)) > 0 {
					if minTransAmount.Cmp(new(big.Int).SetUint64(tokenAcc.Amount)) > 0 {
						return fmt.Errorf("not enough source balance: %w", swaperror.ErrInsufficientFunds)
					}
					amountIn = new(big.Int).SetUint64(cfg.MinTransactionsAmount)
				}
			}
		}
	}

	tokenMint := baseParam.InputTokenMint
	if solutil.IsSOLLikeMint(tokenMint) {
		tokenMint = baseParam.OutputTokenMint
	}

	swapParams, err := FetchBondingSwapParams(ctx, c.RPC, task.PoolID, task.PrivateKey.PublicKey(), tokenMint)
	if err != nil {
		return apperrors.Internal("failed to fetch bonding swap params", err)
	}

	tx, err := c.buildSwapTransaction(
		amountIn,
		latestHash,
		pool,
		swapParams,
		task,
		cfg,
	)
	if err != nil {
		return apperrors.Internal("failed to build tx", err)
	}

	simulated, err := c.RPC.SimulateTransaction(ctx, tx, &rpc.SimulateTransactionOpts{
		SigVerify:              false,
		Commitment:             rpc.CommitmentProcessed,
		ReplaceRecentBlockhash: true,
	})
	if err != nil {
		logParams.TransactionHash = tx.Signatures[0].String()
		if logErr := swaptxlog.LogSwapTransaction(ctx, err, campaignID, logParams, c.transactionRepository, c.logger); logErr != nil {
			return logErr
		}
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "429") || strings.Contains(message, "rate limit") || strings.Contains(message, "too many requests") {
			return fmt.Errorf("failed to simulate tx: %w: %w", swaperror.ErrRateLimit, err)
		}
		if strings.Contains(message, "gateway timeout") || strings.Contains(message, "context deadline exceeded") || strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout") {
			return fmt.Errorf("failed to simulate tx: %w: %w", swaperror.ErrGatewayTimeout, err)
		}
		return fmt.Errorf("failed to simulate tx: %w: %w", swaperror.ErrSimulationError, err)
	}
	if simulated != nil && simulated.Value.Err != nil {
		if strings.Contains(fmt.Sprint(simulated.Value.Logs), "NotEnoughTokensToSell") {
			return fmt.Errorf("not enough tokens to sell: %w: %w", swaperror.ErrInsufficientFunds, NotEnoughTokensToSellError)
		}
		err = fmt.Errorf("simulation failed: %v, logs: %v", simulated.Value.Err, simulated.Value.Logs)
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "slippage") {
			err = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrSlippageExceeded, err)
		} else if strings.Contains(message, "insufficient funds") || strings.Contains(message, "empty funds") || strings.Contains(message, "input token account empty") {
			err = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrInsufficientFunds, err)
		} else if strings.Contains(message, "custom program error") {
			err = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrCustomProgramError, err)
		} else if strings.Contains(message, "compute budget exceeded") || strings.Contains(message, "computebudgetexceeded") || strings.Contains(message, "computational budget exceeded") || strings.Contains(message, "exceeded the maximum number of instructions allowed") {
			err = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrComputeBudgetExceeded, err)
		} else {
			err = fmt.Errorf("simulation failed: %w: %w", swaperror.ErrSimulationError, err)
		}
		logParams.TransactionHash = tx.Signatures[0].String()
		if logErr := swaptxlog.LogSwapTransaction(ctx, err, campaignID, logParams, c.transactionRepository, c.logger); logErr != nil {
			return logErr
		}
		return err
	}

	if task.UsingJito {
		tip, tipErr := c.jitoClient.CalculateTip(ctx, task.TransactionSpeed)
		if tipErr != nil {
			err = tipErr
			logParams.TransactionHash = tx.Signatures[0].String()
			if logErr := swaptxlog.LogSwapTransaction(ctx, err, campaignID, logParams, c.transactionRepository, c.logger); logErr != nil {
				return logErr
			}
			return err
		}

		tipAccount, tipAccErr := c.jitoClient.GetTipAccount(ctx)
		if tipAccErr != nil {
			err = tipAccErr
			logParams.TransactionHash = tx.Signatures[0].String()
			if logErr := swaptxlog.LogSwapTransaction(ctx, err, campaignID, logParams, c.transactionRepository, c.logger); logErr != nil {
				return logErr
			}
			return err
		}

		tipTx, tipTxErr := jito.BuildTipTransaction(task.PrivateKey, tip, *tipAccount, *latestHash)
		if tipTxErr != nil {
			err = tipTxErr
			logParams.TransactionHash = tx.Signatures[0].String()
			if logErr := swaptxlog.LogSwapTransaction(ctx, err, campaignID, logParams, c.transactionRepository, c.logger); logErr != nil {
				return logErr
			}
			return err
		}

		if broadcastErr := c.jitoClient.BroadcastBundle(ctx, []*solana.Transaction{tx, tipTx}, ProgramID); broadcastErr != nil {
			err = broadcastErr
			logParams.TransactionHash = tx.Signatures[0].String()
			if logErr := swaptxlog.LogSwapTransaction(ctx, err, campaignID, logParams, c.transactionRepository, c.logger); logErr != nil {
				return logErr
			}
			return err
		}
	} else {
		if _, sendErr := c.RPC.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentProcessed,
		}); sendErr != nil {
			message := strings.ToLower(sendErr.Error())
			if strings.Contains(message, "computebudgetexceeded") || strings.Contains(message, "compute budget exceeded") || strings.Contains(message, "computational budget exceeded") {
				err = fmt.Errorf("failed to send tx: %w: %w", swaperror.ErrComputeBudgetExceeded, sendErr)
			} else if strings.Contains(message, "429") || strings.Contains(message, "rate limit") || strings.Contains(message, "too many requests") {
				err = fmt.Errorf("failed to send tx: %w: %w", swaperror.ErrRateLimit, sendErr)
			} else if strings.Contains(message, "gateway timeout") || strings.Contains(message, "context deadline exceeded") || strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout") {
				err = fmt.Errorf("failed to send tx: %w: %w", swaperror.ErrGatewayTimeout, sendErr)
			} else {
				err = sendErr
			}
			logParams.TransactionHash = tx.Signatures[0].String()
			if logErr := swaptxlog.LogSwapTransaction(ctx, err, campaignID, logParams, c.transactionRepository, c.logger); logErr != nil {
				return logErr
			}
			return err
		}
	}

	remainingBudget.Store(subBudget)
	logParams.TransactionHash = tx.Signatures[0].String()
	if logErr := swaptxlog.LogSwapTransaction(ctx, err, campaignID, logParams, c.transactionRepository, c.logger); logErr != nil {
		return logErr
	}
	c.logger.Info("bonding swap completed",
		zap.String("campaign_id", campaignID.String()),
		zap.String("pool_id", task.PoolID.String()),
		zap.String("source_mint", task.SourceTokenMint.String()),
		zap.String("dest_mint", task.DestTokenMint.String()),
		zap.String("tx", tx.Signatures[0].String()),
	)
	return nil
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
		instrs = append(instrs, assoc.NewCreateInstruction(
			task.PrivateKey.PublicKey(),
			task.PrivateKey.PublicKey(),
			task.DestTokenMint,
		).Build())
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
	}

	return instruction, nil
}

func calculateBondingCurvePrice(pool *PoolStateWithReserve) (*big.Rat, error) {
	return poolmath.ConstantProductCalculatePrice(
		pool.ReserveA,
		pool.ReserveB,
		uint64(pool.BaseMintDecimals),
		uint64(pool.QuoteMintDecimals),
	), nil
}
