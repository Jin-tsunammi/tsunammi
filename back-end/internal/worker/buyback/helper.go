package buyback

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	pump_amm "mm/internal/client/pumpfun/amm"
	pumpBonding "mm/internal/client/pumpfun/bonding"
	"mm/internal/client/raydium"
	raydium_amm "mm/internal/client/raydium/ammv4/ammv4_client"
	raydium_cp_swap "mm/internal/client/raydium/cpmm/cpmm_client"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/swapbudget"
	"mm/internal/swaperror"
	"mm/internal/swaptxlog"
	"mm/pkg/apperrors"
	"mm/pkg/solutil"
	"sync"
	"sync/atomic"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var errBuybackNoFunds = errors.New("buyback no funds")

func atomicUnitsToRat(amount *big.Int, decimals uint8) *big.Rat {
	if amount == nil {
		return new(big.Rat)
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	return new(big.Rat).SetFrac(new(big.Int).Set(amount), divisor)
}

func chunkTargetTxBatch(
	tasks []*model.AsyncSwapTask,
	configs []raydium.TWAPConfig,
	params []raydium.SwapParams,
	batchSize int,
) ([][]*model.AsyncSwapTask, [][]raydium.TWAPConfig, [][]raydium.SwapParams, error) {
	if len(tasks) == 0 || len(tasks) != len(configs) || len(tasks) != len(params) {
		return nil, nil, nil, apperrors.BadRequest("invalid buyback batch lengths")
	}
	if batchSize <= 0 {
		batchSize = len(tasks)
	}

	taskChunks := make([][]*model.AsyncSwapTask, 0, len(tasks)/batchSize+1)
	configChunks := make([][]raydium.TWAPConfig, 0, len(tasks)/batchSize+1)
	paramChunks := make([][]raydium.SwapParams, 0, len(tasks)/batchSize+1)

	for start := 0; start < len(tasks); start += batchSize {
		end := min(start+batchSize, len(tasks))
		taskChunks = append(taskChunks, tasks[start:end])
		configChunks = append(configChunks, configs[start:end])
		paramChunks = append(paramChunks, params[start:end])
	}

	return taskChunks, configChunks, paramChunks, nil
}

func (m *CampaignManager) selectTargets(campaign *model.SmartBuybackCampaignWithTargets, now time.Time) []model.SmartBuybackCampaignTarget {
	activeTypes := map[model.BuybackCampaignTargetType]struct{}{}
	for i := range campaign.Targets {
		t := campaign.Targets[i]
		if t.StartAt.After(now) {
			continue
		}
		if t.Status != model.BuybackStatusActive && t.Status != model.BuybackStatusScheduled {
			continue
		}
		activeTypes[t.Type] = struct{}{}
	}

	selected := make([]model.SmartBuybackCampaignTarget, 0, len(activeTypes))

	for tType := range activeTypes {
		tokenIn, tokenOut := getTargetDirectionMints(campaign.TokenMint, tType)
		tracked, ok, err := m.priceProvider.GetPrice(campaign.PoolID, tokenIn, tokenOut)
		if err != nil || !ok {
			continue
		}

		var best *model.SmartBuybackCampaignTarget
		var bestDist *big.Rat
		bestPrice := ""

		for i := range campaign.Targets {
			t := campaign.Targets[i]
			if t.Type != tType || t.StartAt.After(now) {
				continue
			}
			if t.Status != model.BuybackStatusActive && t.Status != model.BuybackStatusScheduled {
				continue
			}

			p, ok := new(big.Rat).SetString(tracked.Price)
			if !ok {
				continue
			}
			normalized := p.RatString()
			satisfied := targetSatisfied(t, normalized)
			m.logger.Info("buyback target price iteration",
				zap.String("campaign_id", campaign.ID.String()),
				zap.String("target_id", t.ID.String()),
				zap.String("direction", string(t.Type)),
				zap.String("current_price_sol_per_token", p.FloatString(10)),
				zap.String("target_price_sol_per_token", t.TargetPrice.GetBigRat().FloatString(10)),
				zap.Bool("target_satisfied", satisfied),
				zap.String("raw_tracked_price", tracked.Price),
			)
			if !satisfied {
				continue
			}

			d := targetDistance(normalized, t.TargetPrice.GetBigRat())
			if d == nil || (bestDist != nil && d.Cmp(bestDist) >= 0) {
				continue
			}

			copyTarget := t
			best = &copyTarget
			bestDist = d
			bestPrice = normalized
		}

		if best != nil {
			reason := "price higher/equal than target"
			if best.Type == model.BuybackCampaignTargetTypeBuy {
				reason = "price lower/equal than target"
			}
			m.logger.Info("buyback target selected",
				zap.String("campaign_id", campaign.ID.String()),
				zap.String("target_id", best.ID.String()),
				zap.String("direction", string(best.Type)),
				zap.String("reason", reason),
				zap.String("current_price_sol_per_token", bestPrice),
				zap.String("target_price_sol_per_token", best.TargetPrice.GetBigRat().RatString()),
				zap.String("raw_tracked_price", tracked.Price),
			)
			selected = append(selected, *best)
		}
	}

	return selected
}

func (m *CampaignManager) buildTargetTxBatch(ctx context.Context, campaign *model.SmartBuybackCampaign, target *model.SmartBuybackCampaignTarget) ([]*model.AsyncSwapTask, []raydium.TWAPConfig, []raydium.SwapParams, error) {
	provider, err := m.getDEXProvider(campaign.ProviderID)
	if err != nil {
		return nil, nil, nil, err
	}

	var srcMint, destMint solana.PublicKey
	taskType := model.TargetUpTaskType
	switch target.Type {
	case model.BuybackCampaignTargetTypeBuy:
		dest, err := solana.PublicKeyFromBase58(campaign.TokenMint)
		if err != nil {
			return nil, nil, nil, err
		}
		srcMint, destMint = solana.SolMint, dest
	case model.BuybackCampaignTargetTypeSell:
		src, err := solana.PublicKeyFromBase58(campaign.TokenMint)
		if err != nil {
			return nil, nil, nil, err
		}
		srcMint, destMint = src, solana.SolMint
		taskType = model.TargetDownTaskType
	default:
		return nil, nil, nil, fmt.Errorf("unknown target type: %s", target.Type)
	}

	wallets, totalBalance, err := m.fetchFundedWallets(ctx, campaign.ProjectID, campaign.UserID, 0, srcMint)
	if err != nil {
		return nil, nil, nil, err
	}

	minAmount, _ := target.MinTransactionAmount.Float64()
	budget, _ := target.RemainingBudget.Float64()
	m.logger.Info("buyback: funded wallets fetched",
		zap.String("campaign_id", campaign.ID.String()),
		zap.String("target_id", target.ID.String()),
		zap.Int("funded_wallets", len(wallets)),
		zap.Float64("total_balance", totalBalance),
		zap.Float64("min_tx_amount", minAmount),
		zap.Float64("remaining_budget", budget),
	)
	wallets, _, err = filterWalletsForCampaign(wallets, minAmount, budget, func(wallet model.Wallet) float64 {
		switch target.Type {
		case model.BuybackCampaignTargetTypeBuy:
			return wallet.BalanceSOL
		default:
			return wallet.BalanceToken
		}
	})
	if err != nil {
		return nil, nil, nil, err
	}

	privateKeys, err := m.fetchPrivateKeys(ctx, campaign.UserID, wallets)
	if err != nil {
		return nil, nil, nil, err
	}

	ataMint := destMint
	if target.Type == model.BuybackCampaignTargetTypeSell {
		ataMint = srcMint
	}

	ataAddresses, ataAccounts, err := m.fetchATAInfo(ctx, wallets, ataMint)
	if err != nil {
		return nil, nil, nil, err
	}

	poolResult, _, err := provider.PreparePool(ctx, srcMint, destMint)
	if err != nil {
		return nil, nil, nil, err
	}

	poolID, err := solana.PublicKeyFromBase58(campaign.PoolID)
	if err != nil {
		return nil, nil, nil, err
	}

	params, err := provider.FetchPoolParams(ctx, poolID)
	if err != nil {
		return nil, nil, nil, err
	}

	tasks := make([]*model.AsyncSwapTask, len(wallets))
	configs := make([]raydium.TWAPConfig, len(wallets))
	params_ := make([]raydium.SwapParams, len(wallets))
	errs := make([]error, len(wallets))
	for i, wallet := range wallets {
		walletKey, wErr := solana.PublicKeyFromBase58(wallet.PublicKey)
		if wErr != nil {
			errs[i] = apperrors.Internal("invalid wallet address", wErr)
			continue
		}

		wSolATA, _, wErr := solana.FindAssociatedTokenAddress(walletKey, solana.WrappedSol)
		if wErr != nil {
			errs[i] = apperrors.Internal("cant find ata address", wErr)
			continue
		}

		sourceAddress := wSolATA
		destAddress := ataAddresses[i]
		if target.Type == model.BuybackCampaignTargetTypeSell {
			sourceAddress = ataAddresses[i]
			destAddress = wSolATA
		}

		fee, _ := target.PriorityFee.Float64()
		microlamportsPerCU := uint64(math.Ceil(
			(fee * 1_000_000_000 * 1_000_000) / float64(200000),
		))

		goalPrice := new(big.Rat)
		switch target.Type {
		case model.BuybackCampaignTargetTypeBuy:
			taskType = model.TargetUpTaskType
			goalPrice = target.TargetPrice.GetBigRat()

		case model.BuybackCampaignTargetTypeSell:
			taskType = model.TargetDownTaskType
			goalPrice = target.TargetPrice.GetBigRat()
		}

		tasks[i] = &model.AsyncSwapTask{
			SwapCampaignID:        campaign.ID,
			GoalPrice:             goalPrice,
			MinTransactionsAmount: func() float64 { t, _ := target.MinTransactionAmount.Float64(); return t }(),
			MaxTransactionsAmount: func() float64 { t, _ := target.MaxTransactionAmount.Float64(); return t }(),
			Slippage:              uint64(math.Round(float64(target.Slippage) * 100)),
			PoolID:                poolResult.PoolID,
			PoolProgramID:         poolResult.PoolProgramID,
			SourceAddress:         sourceAddress,
			SourceTokenMint:       srcMint,
			SourceTokenDecimals:   poolResult.SourceTokenDecimals,
			DestTokenDecimals:     poolResult.DestTokenDecimals,
			DestTokenMint:         destMint,
			DestAddress:           destAddress,
			PrivateKey:            privateKeys[i],
			TaskType:              taskType,
			TransactionSpeed:      target.TransactionSpeed,
			ATAKeyCreated:         ataAccounts[i] != nil,
			PoolParams:            params,
			UsingJito:             target.UsingJito,
			PriorityFeeMLP:        microlamportsPerCU,
		}

		configs[i] = raydium.TWAPConfig{
			MinTransactionsAmount:         solanarpc.ToAtomicUnit(tasks[i].MinTransactionsAmount, tasks[i].SourceTokenDecimals),
			MaxTransactionsAmount:         solanarpc.ToAtomicUnit(tasks[i].MaxTransactionsAmount, tasks[i].SourceTokenDecimals),
			SlippageBPS:                   tasks[i].Slippage,
			ComputeUnitLimit:              200_000, // TODO: from config
			ComputeUnitPriceMicroLamports: microlamportsPerCU,
		}

		params_[i] = raydium.SwapParams{
			UserWallet:      tasks[i].PrivateKey.PublicKey(),
			PoolID:          tasks[i].PoolID,
			InputTokenMint:  tasks[i].SourceTokenMint,
			OutputTokenMint: tasks[i].DestTokenMint,
			UserSourceToken: tasks[i].SourceAddress,
			UserDestToken:   tasks[i].DestAddress,
		}
	}

	filteredTasks := make([]*model.AsyncSwapTask, 0, len(wallets))
	filteredConfigs := make([]raydium.TWAPConfig, 0, len(wallets))
	filteredParams := make([]raydium.SwapParams, 0, len(wallets))
	for i, t := range tasks {
		if t == nil {
			if errs[i] != nil {
				m.logger.Warn("buyback: skip wallet in batch", zap.Int("index", i), zap.Error(errs[i]))
			}
			continue
		}
		filteredTasks = append(filteredTasks, t)
		filteredConfigs = append(filteredConfigs, configs[i])
		filteredParams = append(filteredParams, params_[i])
	}

	if len(filteredTasks) == 0 {
		return nil, nil, nil, fmt.Errorf("%w: all wallets failed to build swap task", errBuybackNoFunds)
	}

	return filteredTasks, filteredConfigs, filteredParams, nil
}

func (m *CampaignManager) fetchFundedWallets(ctx context.Context, projectID, userID uint64, minWallets int, mint solana.PublicKey) ([]model.Wallet, float64, error) {
	decimals := solana.SolDecimals

	if !solutil.IsSOLLikeMint(mint) {
		info, err := m.solanaRPC.GetAccountInfo(ctx, mint)
		if err != nil {
			return nil, 0, err
		}

		data := info.GetBinary()
		if data == nil {
			return nil, 0, apperrors.BadRequest("mint not found", nil)
		}

		tokenMint := token.Mint{}
		if err = tokenMint.UnmarshalWithDecoder(bin.NewBinDecoder(data)); err != nil {
			m.logger.Error("failed to parse mint in fetchFundedWallets",
				zap.Uint64("project_id", projectID),
				zap.Uint64("user_id", userID),
				zap.String("mint", mint.String()),
				zap.Int("data_len", len(data)),
				zap.Error(err),
			)
			return nil, 0, err
		}

		decimals = tokenMint.Decimals
	}

	fmt.Println(projectID, userID)
	project, err := m.projectRepo.FetchProjectWithWalletsByID(ctx, projectID, userID)
	if err != nil {
		return nil, 0, apperrors.Internal("failed to fetch project", err)
	}
	fmt.Printf("project: %+v\n", project)
	if len(project.Wallets) > 300 {
		return nil, 0, apperrors.BadRequest("too many wallets in project", nil)
	}
	if len(project.Wallets) == 0 {
		return nil, 0, apperrors.BadRequest("no wallets found", nil)
	}

	pubKeys := make([]solana.PublicKey, len(project.Wallets))
	for i, w := range project.Wallets {
		pk, err := solana.PublicKeyFromBase58(w.PublicKey)
		if err != nil {
			return nil, 0, apperrors.Internal("invalid wallet address", err)
		}
		if solutil.IsSOLLikeMint(mint) {
			pubKeys[i] = pk
		} else {
			address, _, err := solana.FindAssociatedTokenAddress(pk, mint)
			if err != nil {
				return nil, 0, err
			}
			pubKeys[i] = address
		}
	}

	response, err := m.solanaRPC.GetMultipleAccountsWithNoLimits(ctx, pubKeys...)
	if err != nil {
		return nil, 0, apperrors.Internal("failed to get balances", err)
	}

	accounts := make([]*rpc.Account, 0, len(pubKeys))
	for _, res := range response {
		accounts = append(accounts, res.Value...)
	}

	fundedWallets := make([]model.Wallet, 0, len(project.Wallets))
	totalBalance := 0.0
	for i, account := range accounts {
		var balance float64
		if account == nil || account.Data == nil || account.Data.GetBinary() == nil {
			continue
		}

		wallet := &project.Wallets[i]
		if solutil.IsSOLLikeMint(mint) {
			balance = solanarpc.FromAtomicUnit(account.Lamports, decimals)
			wallet.BalanceSOL = balance
		} else {
			tokenAccount := token.Account{}
			err = tokenAccount.UnmarshalWithDecoder(bin.NewBinDecoder(account.Data.GetBinary()))
			if err != nil {
				m.logger.Error("failed to parse token account in fetchFundedWallets",
					zap.Uint64("project_id", projectID),
					zap.Uint64("user_id", userID),
					zap.String("mint", mint.String()),
					zap.String("wallet", wallet.PublicKey),
					zap.Int("account_index", i),
					zap.Int("data_len", len(account.Data.GetBinary())),
					zap.Uint64("lamports", account.Lamports),
					zap.Any("owner", account.Owner),
					zap.Error(err),
				)
				return nil, 0, err
			}

			balance = solanarpc.FromAtomicUnit(tokenAccount.Amount, decimals)
			wallet.BalanceToken = balance
		}

		if balance < 0.002 {
			continue
		}

		fundedWallets = append(fundedWallets, project.Wallets[i])
		totalBalance += balance
	}

	if len(fundedWallets) == 0 {
		return nil, 0, apperrors.BadRequest("all selected wallets have insufficient balance", nil)
	}
	if minWallets > 0 && len(fundedWallets) < minWallets {
		return nil, 0, apperrors.BadRequest("not enough active wallets", nil)
	}

	return fundedWallets, totalBalance, nil
}

func (m *CampaignManager) fetchPrivateKeys(ctx context.Context, userID uint64, wallets []model.Wallet) ([]solana.PrivateKey, error) {
	eg, errctx := errgroup.WithContext(ctx)
	eg.SetLimit(10)
	keys := make([]solana.PrivateKey, len(wallets))

	for i, w := range wallets {
		idx, wallet := i, w
		eg.Go(func() error {
			keyStr, err := m.keyStorage.Get(errctx, userID, wallet.PublicKey)
			if err != nil {
				return err
			}
			key, err := solana.PrivateKeyFromBase58(keyStr)
			if err != nil {
				return err
			}
			keys[idx] = key
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, apperrors.Internal("failed to fetch keys", err)
	}
	return keys, nil
}

func (m *CampaignManager) fetchATAInfo(ctx context.Context, wallets []model.Wallet, mint solana.PublicKey) ([]solana.PublicKey, []*rpc.Account, error) {
	addresses := make([]solana.PublicKey, len(wallets))
	for i, w := range wallets {
		owner, err := solana.PublicKeyFromBase58(w.PublicKey)
		if err != nil {
			return nil, nil, apperrors.Internal("invalid wallet address", err)
		}
		addr, _, err := solana.FindAssociatedTokenAddress(owner, mint)
		if err != nil {
			return nil, nil, apperrors.Internal("failed to get ata address", err)
		}
		addresses[i] = addr
	}

	accountsResp, err := m.solanaRPC.GetMultipleAccountsWithNoLimits(ctx, addresses...)
	if err != nil {
		return nil, nil, apperrors.Internal("failed to fetch ATA accounts", err)
	}

	accounts := make([]*rpc.Account, 0, len(addresses))
	for _, a := range accountsResp {
		accounts = append(accounts, a.Value...)
	}

	return addresses, accounts, nil
}

func filterWalletsForCampaign(
	wallets []model.Wallet,
	minTransactionBudget, campaignBudget float64,
	balanceFn func(model.Wallet) float64,
) ([]model.Wallet, float64, error) {
	eligibleWallets := make([]model.Wallet, 0, len(wallets))
	totalEligibleBalance := 0.0

	for _, wallet := range wallets {
		balance := balanceFn(wallet)
		if balance < minTransactionBudget {
			continue
		}

		eligibleWallets = append(eligibleWallets, wallet)
		totalEligibleBalance += balance
	}

	if len(eligibleWallets) == 0 {
		return nil, 0, fmt.Errorf("%w: no eligible wallets found for campaign start", errBuybackNoFunds)
	}
	if totalEligibleBalance < campaignBudget {
		return nil, 0, fmt.Errorf("%w: not enough funds on eligible wallets for campaign budget", errBuybackNoFunds)
	}

	return eligibleWallets, totalEligibleBalance, nil
}

func (m *CampaignManager) getDEXProvider(providerID model.SwapProviderID) (model.DexProvider, error) {
	provider, ok := m.dexProviders[providerID]
	if !ok {
		return nil, apperrors.BadRequest("unsupported provider id")
	}

	return provider, nil
}

func (m *CampaignManager) dispatchBatch(ctx context.Context, campaign model.SmartBuybackCampaign, target model.SmartBuybackCampaignTarget) error {
	tasks, configs, params, err := m.buildTargetTxBatch(ctx, &campaign, &target)
	if err != nil {
		if errors.Is(err, errBuybackNoFunds) {
			if updateErr := m.buybackRepo.UpdateCampaignError(ctx, campaign.ID); updateErr != nil {
				return updateErr
			}
			m.logger.Warn("buyback campaign moved to error: no funds",
				zap.String("campaign_id", campaign.ID.String()),
				zap.String("target_id", target.ID.String()),
				zap.Error(err),
			)
			return nil
		}
		return err
	}

	taskChunks, configChunks, paramChunks, err := chunkTargetTxBatch(tasks, configs, params, int(max(target.ParallelTransactionsAmount, 1)))
	if err != nil {
		return err
	}

	// b, _ := target.RemainingBudget.Float64()
	// budgetAtomic := solanarpc.ToAtomicUnit(b, tasks[0].SourceTokenDecimals)
	budget := swapbudget.NewSwapBudget(
		target.RemainingBudget.ToAtomicUnits(tasks[0].SourceTokenDecimals),
	)

	m.logger.Info("buyback dispatch: budget and batch ready",
		zap.String("campaign_id", campaign.ID.String()),
		zap.String("target_id", target.ID.String()),
		zap.String("remaining_budget_atomic", target.RemainingBudget.ToAtomicUnits(tasks[0].SourceTokenDecimals).String()),
		zap.Int("wallets", len(tasks)),
		zap.Uint8("src_decimals", tasks[0].SourceTokenDecimals),
	)

	latestBlockHash, err := m.solanaRPC.GetLatestBlockhash(ctx)
	if err != nil {
		return err
	}

	hash := &atomic.Pointer[solana.Hash]{}
	hash.Store(latestBlockHash)

	errs := make([]error, len(taskChunks))
	wg := &sync.WaitGroup{}
	semaphore := make(chan struct{}, 10)

	wg.Add(len(taskChunks))

	for i := range taskChunks {
		go func(idx int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			var localErr error
			var results []swaptxlog.Result
			targetID := target.ID

			switch taskChunks[idx][0].PoolProgramID {
			case raydium_cp_swap.ProgramID:
				results, localErr = m.raydiumCPMMClient.Swap(ctx, campaign.ID, &targetID, budget, taskChunks[idx], paramChunks[idx], configChunks[idx], hash)
			case raydium_amm.ProgramID:
				results, localErr = m.raydiumAMMClient.Swap(ctx, campaign.ID, &targetID, budget, taskChunks[idx], paramChunks[idx], configChunks[idx], hash)
			case pump_amm.ProgramID:
				results, localErr = m.pumpAMMClient.Swap(ctx, campaign.ID, &targetID, budget, taskChunks[idx], paramChunks[idx], configChunks[idx], hash)
			case pumpBonding.ProgramID:
				results, localErr = m.pumpBondingClient.Swap(ctx, campaign.ID, &targetID, budget, taskChunks[idx], paramChunks[idx], configChunks[idx], hash)
			default:
				localErr = fmt.Errorf("unsupported pool program id: %s", taskChunks[idx][0].PoolProgramID.String())
			}

			for _, result := range results {
				if logErr := swaptxlog.LogBuybackTransaction(ctx, result.Err, campaign.ID, &targetID, result.Params, m.buybackTxRepo, m.logger); logErr != nil {
					localErr = errors.Join(localErr, logErr)
				}
			}

			if localErr != nil {
				m.logger.Warn("buyback swap chunk failed",
					zap.String("campaign_id", campaign.ID.String()),
					zap.String("target_id", targetID.String()),
					zap.Int("chunk_idx", idx),
					zap.Error(localErr),
				)
				errs[idx] = localErr
			}
		}(i)
	}

	wg.Wait()

	err = resolveBuybackBatchError(errs)
	m.logger.Info("buyback batch resolved",
		zap.String("campaign_id", campaign.ID.String()),
		zap.String("target_id", target.ID.String()),
		zap.String("budget_remaining_atomic", budget.Remaining().String()),
		zap.NamedError("resolved_err", err),
	)
	if err == nil {
		if joinedErr := errors.Join(errs...); joinedErr != nil {
			m.logger.Error("buyback swap chunks failed",
				zap.String("campaign_id", campaign.ID.String()),
				zap.String("target_id", target.ID.String()),
				zap.Error(joinedErr),
			)
			err = apperrors.Internal("failed to swap", errors.New("failed to swap"), joinedErr)
		}
	}

	if updateErr := m.persistTargetBudget(ctx, campaign.ID, target.ID, budget, tasks[0].SourceTokenDecimals, err); updateErr != nil {
		m.logger.Warn("failed to persist buyback target budget",
			zap.String("campaign_id", campaign.ID.String()),
			zap.String("target_id", target.ID.String()),
			zap.Error(updateErr),
		)
		if err == nil {
			err = updateErr
		}
	}

	m.logger.Info("buyback batch dispatched",
		zap.String("campaign_id", campaign.ID.String()),
		zap.String("target_id", target.ID.String()),
		zap.String("target_type", string(target.Type)),
		zap.Int("tasks_total", len(tasks)),
		zap.Int("chunks_total", len(taskChunks)),
	)

	if isTerminalBuybackTargetError(err) {
		return nil
	}

	return err
}

func (m *CampaignManager) persistTargetBudget(ctx context.Context, campaignID, targetID uuid.UUID, budget *swapbudget.SwapBudget, decimals uint8, swapErr error) error {
	// remainingAtomic := budget.Load()
	// if remainingAtomic == nil {
	// 	remainingAtomic = new(big.Int)
	// }

	remainingAtomic := budget.Remaining()

	remaining := atomicUnitsToRat(remainingAtomic, decimals)
	status := ""
	if remainingAtomic.Sign() <= 0 ||
		errors.Is(swapErr, swaperror.BudgetExceededError) ||
		errors.Is(swapErr, pumpBonding.NotEnoughTokensToSellError) {
		status = string(model.BuybackStatusDone)
	}

	if err := m.buybackRepo.UpdateTargetRemainingBudget(ctx, targetID, remaining, status); err != nil {
		return err
	}
	if status == "" {
		return nil
	}

	done, err := m.buybackRepo.UpdateDoneIfNoActiveTargets(ctx, campaignID)
	if err != nil {
		return err
	}
	if done {
		m.logger.Info("buyback campaign completed",
			zap.String("campaign_id", campaignID.String()),
		)
	}

	return nil
}

func resolveBuybackBatchError(errs []error) error {
	for _, err := range errs {
		if errors.Is(err, swaperror.BudgetExceededError) || errors.Is(err, pumpBonding.NotEnoughTokensToSellError) {
			return swaperror.BudgetExceededError
		}
	}
	return nil
}

func isTerminalBuybackTargetError(err error) bool {
	return errors.Is(err, swaperror.BudgetExceededError) ||
		errors.Is(err, pumpBonding.NotEnoughTokensToSellError)
}

func targetSatisfied(target model.SmartBuybackCampaignTarget, trackedPrice string) bool {
	price, ok := new(big.Rat).SetString(trackedPrice)
	if !ok {
		return false
	}

	targetPrice := target.TargetPrice.GetBigRat()
	if targetPrice == nil {
		return false
	}

	if target.Type == model.BuybackCampaignTargetTypeBuy {
		return price.Cmp(targetPrice) <= 0
	}

	return price.Cmp(targetPrice) >= 0
}

func targetDistance(trackedPrice string, targetPrice *big.Rat) *big.Rat {
	if targetPrice == nil {
		return nil
	}

	price, ok := new(big.Rat).SetString(trackedPrice)
	if !ok {
		return nil
	}

	d := new(big.Rat).Sub(price, targetPrice)
	if d.Sign() < 0 {
		d.Neg(d)
	}
	return d
}

func getTargetDirectionMints(tokenMint string, targetType model.BuybackCampaignTargetType) (tokenIn, tokenOut string) {
	if targetType == model.BuybackCampaignTargetTypeBuy {
		return solana.WrappedSol.String(), tokenMint
	}
	return tokenMint, solana.WrappedSol.String()
}
