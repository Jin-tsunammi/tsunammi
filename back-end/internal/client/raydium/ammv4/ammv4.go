package ammv4

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"mm/internal/client/jito"
	"mm/internal/client/raydium"
	raydiumamm "mm/internal/client/raydium/ammv4/ammv4_client"
	openbookv1 "mm/internal/client/raydium/ammv4/openbook/openbook_v1_client"
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
) error {
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
		err := swaptxlog.LogSwapTransaction(ctx, raydium.BudgetExceededError, campaignID, logParams, c.transactionRepository, c.logger)
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

	if remainingBudget.Load().Sign() <= 0 {
		err = swaptxlog.LogSwapTransaction(ctx, raydium.BudgetExceededError, campaignID, logParams, c.transactionRepository, c.logger)
		if err != nil {
			return err
		}

		return raydium.BudgetExceededError
	}

	var index int
	var txs []*solana.Transaction
	var tx *solana.Transaction

	defer func() {
		logErrs := make([]error, len(txs))

		c.logger.Error("Log swap trans", zap.Any("txs", txs))

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

	c.logger.Info("Latest block hash", zap.Any("hash", latestHash.String()))

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
				if customErr, ok := raydiumamm.DecodeCustomError(err); ok {
					if customErr == raydiumamm.ErrExceededSlippage {
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
	pool *AMMInfoWithReservers,
	params raydium.SwapParams,
	task *model.AsyncSwapTask,
	cfg raydium.TWAPConfig,
) (*solana.Transaction, error) {

	isSolToToken := task.TaskType == model.TargetUpTaskType

	tradeFee, err := numeratorWithDenominatorToHundredOfBSP(pool.PoolState.Fees.TradeFeeNumerator, pool.PoolState.Fees.TradeFeeDenominator)
	if err != nil {
		return nil, err
	}

	out, err := poolmath.ConstantProductTokenAmountOut(amountIn, pool.ReserveA, pool.ReserveB, tradeFee)
	if err != nil {
		return nil, err
	}

	minOut, err := poolmath.ConstantProductFloorWithSlippage(out, cfg.SlippageBPS)
	if err != nil {
		return nil, err
	}

	params.AmountIn = amountIn
	params.MinAmountOut = minOut

	c.logger.Info("Trade fee", zap.Any("tradeFee", tradeFee))
	c.logger.Info("Pool reserve A", zap.Any("reserveA", pool.ReserveA))
	c.logger.Info("Pool reserve B", zap.Any("reserveB", pool.ReserveB))
	c.logger.Info("Amount in", zap.Any("amountIn", amountIn))
	c.logger.Info("Amount out", zap.Any("amountOut", minOut))
	c.logger.Info("Slippage", zap.Any("slippage", cfg.SlippageBPS))

	swapInstr, err := ammBuildSwapInstruction(pool, params)
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

func FetchAMMPoolState(ctx context.Context, rpc solanarpc.SolanaRPC, params *model.PoolParams, inputTokenMint, outputTokenMint solana.PublicKey) (*AMMInfoWithReservers, error) {

	accounts, err := rpc.GetMultipleAccounts(ctx, params.InputTokenVault, params.OutputTokenVault, params.Market, params.OpenOrders, params.PoolID)
	if err != nil {
		return nil, err
	}

	if len(accounts.Value) < 5 {
		return nil, apperrors.Internal(fmt.Sprintf("failed to fetch token vaults: %d accounts returned", len(accounts.Value)))
	}

	poolData := accounts.Value[4].Data.GetBinary()

	rawPoolData := raydiumamm.AmmInfo{}
	err = rawPoolData.Unmarshal(poolData)

	if err != nil {
		return nil, fmt.Errorf("failed to parse pool data: %w", err)
	}

	poolState := &AMMInfoWithReservers{
		PoolState: rawPoolData,
	}

	tokenCoinData := accounts.Value[0].Data.GetBinary()
	tokenPcData := accounts.Value[1].Data.GetBinary()

	if tokenCoinData == nil || tokenPcData == nil {
		return nil, apperrors.Internal("failed to fetch token vaults: token coin or token pc data is nil")
	}

	tokenCoin := token.Account{}
	coinErr := tokenCoin.UnmarshalWithDecoder(bin.NewBorshDecoder(tokenCoinData))

	tokenPc := token.Account{}
	pcErr := tokenPc.UnmarshalWithDecoder(bin.NewBorshDecoder(tokenPcData))

	if err = errors.Join(coinErr, pcErr); err != nil {
		return nil, apperrors.Internal("failed to parse token vaults", err)
	}

	poolState.ReserveA = new(big.Int).SetUint64(tokenCoin.Amount)
	poolState.ReserveB = new(big.Int).SetUint64(tokenPc.Amount)

	marketData := accounts.Value[2].Data.GetBinary()
	openOrdersData := accounts.Value[3].Data.GetBinary()

	if marketData == nil || openOrdersData == nil {
		return nil, apperrors.Internal("failed to fetch market or open orders data")
	}

	market, err := openbookv1.ParseMarketState(marketData)
	if err != nil {
		return nil, apperrors.Internal("failed to parse market state", err)
	}

	openOrders, err := openbookv1.ParseOpenOrders(openOrdersData)
	if err != nil {
		return nil, apperrors.Internal("failed to parse open orders", err)
	}

	poolState.Market = *market
	poolState.OpenOrders = *openOrders

	coinTotalValue, pcTotalValue, err := ammCalcTotalWithoutTakePNL(poolState)

	if err != nil {
		return nil, err
	}

	if poolState.PoolState.CoinMint.Equals(inputTokenMint) && poolState.PoolState.PcMint.Equals(outputTokenMint) {

		poolState.ReserveA = coinTotalValue
		poolState.ReserveB = pcTotalValue

	} else {
		poolState.ReserveB = coinTotalValue
		poolState.ReserveA = pcTotalValue
	}

	return poolState, nil
}

func calculateAMMPrice(pool *AMMInfoWithReservers, params raydium.SwapParams) *big.Rat {
	if pool.PoolState.CoinMint.Equals(params.OutputTokenMint) && pool.PoolState.PcMint.Equals(params.InputTokenMint) {
		return poolmath.ConstantProductCalculatePrice(pool.ReserveA, pool.ReserveB, pool.PoolState.CoinDecimals, pool.PoolState.PcDecimals)
	}
	return poolmath.ConstantProductCalculatePrice(pool.ReserveB, pool.ReserveA, pool.PoolState.PcDecimals, pool.PoolState.CoinDecimals)
}
