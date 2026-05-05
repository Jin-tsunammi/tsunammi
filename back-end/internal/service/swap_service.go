package service

import (
	"context"
	"errors"
	"math"
	"math/big"
	"time"

	"mm/config"
	"mm/internal/client/helius"
	"mm/internal/client/jito"
	"mm/internal/client/solanarpc"
	"mm/internal/common"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/internal/storage/secret"
	"mm/internal/swapbudget"
	"mm/internal/worker/swaptarget"
	"mm/pkg/apperrors"
	"mm/pkg/mtype"
	"mm/pkg/solutil"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	MaxWalletsPerRequest = 300
	MinDustBalance       = 0.002
)

type SwapService struct {
	dexProviders           map[model.SwapProviderID]model.DexProvider
	solanaRPC              solanarpc.SolanaRPC
	projectRepository      *repository.ProjectRepository
	swapCampaignRepository *repository.SwapCampaignRepository
	manager                *swaptarget.SwapTargetManager
	keyStorage             *secret.KeyStorage
	jitoClient             *jito.Client
	heliusClient           *helius.Client
	logger                 *zap.Logger
	computeUnixLimit       uint32
}

func NewSwapService(
	dexProviders map[model.SwapProviderID]model.DexProvider,
	rpc solanarpc.SolanaRPC,
	projectRepository *repository.ProjectRepository,
	swapCampaignRepository *repository.SwapCampaignRepository,
	manager *swaptarget.SwapTargetManager,
	keyStorage *secret.KeyStorage,
	jitoClient *jito.Client,
	heliusClient *helius.Client,
	logger *zap.Logger,
	cfg *config.Config,
) *SwapService {
	return &SwapService{
		dexProviders:           dexProviders,
		solanaRPC:              rpc,
		projectRepository:      projectRepository,
		swapCampaignRepository: swapCampaignRepository,
		manager:                manager,
		keyStorage:             keyStorage,
		jitoClient:             jitoClient,
		heliusClient:           heliusClient,
		logger:                 logger,
		computeUnixLimit:       cfg.App.ComputeUnitLimit,
	}
}

func (s *SwapService) CreatePullUpCampaign(ctx context.Context, req *model.TargetPullUpRequest, userID uint64) (*model.TargetPullResponse, error) {
	wallets, totalBalance, err := s.fetchFundedWallets(ctx, req.ProjectID, userID, 0, solana.WrappedSol)
	if err != nil {
		return nil, err
	}

	if req.BudgetPercent > 100 {
		return nil, apperrors.BadRequest("budget percent must be less than 100", nil)
	}

	if req.BudgetPercent > 0 {
		req.Budget = totalBalance * (req.BudgetPercent / 100)
	}

	wallets, _, err = s.filterWalletsForCampaign(wallets, req.MinTransactionsBudget, req.Budget, func(wallet model.Wallet) float64 {
		return wallet.BalanceSOL
	})
	if err != nil {
		return nil, err
	}

	ataAddresses, ataAccounts, err := s.fetchATAInfo(ctx, wallets, req.DestTokenMint)
	if err != nil {
		return nil, err
	}

	microlamportsPerCU := uint64(math.Ceil(
		(req.PriorityFee * 1_000_000_000 * 1_000_000) / float64(s.computeUnixLimit),
	))
	wallets, ataAddresses, ataAccounts, _, err = s.filterWalletsBySolReserve(ctx, wallets, ataAddresses, ataAccounts, solReserveFilter{
		MinTransactionBudget:       req.MinTransactionsBudget,
		CampaignBudget:             req.Budget,
		PriorityFeeMicroLamports:   microlamportsPerCU,
		IncludeMissingTokenATARent: true,
		DeductReserveFromSource:    true,
		BalanceFn: func(wallet model.Wallet) float64 {
			return wallet.BalanceSOL
		},
	})
	if err != nil {
		return nil, err
	}

	parallelTransactionsAmount := req.ParallelTransactionsAmount
	if parallelTransactionsAmount > len(wallets) {
		parallelTransactionsAmount = len(wallets)
	}

	privateKeys, err := s.fetchPrivateKeys(ctx, userID, wallets)
	if err != nil {
		return nil, err
	}

	provider, err := s.getDEXProvider(req.ProviderID)
	if err != nil {
		return nil, err
	}

	poolResult, startedPrice, err := provider.PreparePool(ctx, solana.WrappedSol, req.DestTokenMint)
	if err != nil {
		return nil, err
	}

	goalPrice := CalculateGoalPrice(startedPrice, req.GoalPercentageChange, true)
	campaignID := uuid.New()

	params, err := provider.FetchPoolParams(ctx, poolResult.PoolID)
	if err != nil {
		return nil, err
	}

	tasks := make([]*model.AsyncSwapTask, len(wallets))
	errs := make([]error, len(wallets))

	for i, wallet := range wallets {
		walletKey, wErr := solana.PublicKeyFromBase58(wallet.PublicKey)
		if wErr != nil {
			errs[i] = apperrors.Internal("invalid wallet address", wErr)
			continue
		}

		wSolATA, _, wErr := solutil.FindAssociatedTokenAddressWithProgram(walletKey, solana.WrappedSol, solana.TokenProgramID)
		if wErr != nil {
			errs[i] = apperrors.Internal("cant find ata address", wErr)
			continue
		}

		tasks[i] = &model.AsyncSwapTask{
			SwapCampaignID:        campaignID,
			GoalPrice:             goalPrice,
			MinTransactionsAmount: req.MinTransactionsBudget,
			MaxTransactionsAmount: req.MaxTransactionsBudget,
			Slippage:              percentageToBasicPoints(req.Slippage),
			PoolID:                poolResult.PoolID,
			PoolProgramID:         poolResult.PoolProgramID,
			SourceAddress:         wSolATA,
			SourceTokenMint:       solana.WrappedSol,
			SourceTokenDecimals:   poolResult.SourceTokenDecimals,
			DestTokenDecimals:     poolResult.DestTokenDecimals,
			DestTokenMint:         req.DestTokenMint,
			DestAddress:           ataAddresses[i],
			PrivateKey:            privateKeys[i],
			TaskType:              model.TargetUpTaskType,
			TransactionSpeed:      req.TransactionSpeed,
			ATAKeyCreated:         ataAccounts[i] != nil,
			PoolParams:            params,
			UsingJito:             req.UsingJito,
			PriorityFeeMLP:        microlamportsPerCU,
		}
	}

	if err = errors.Join(errs...); err != nil {
		return nil, err
	}

	campaign := model.SwapCampaign{
		ID:                         campaignID,
		CampaignTypeID:             1,
		UserID:                     userID,
		ProviderID:                 uint64(req.ProviderID),
		ProjectID:                  req.ProjectID,
		PoolID:                     poolResult.PoolID.String(),
		TokenMintFrom:              solana.WrappedSol.String(),
		TokenMintTo:                req.DestTokenMint.String(),
		Budget:                     req.Budget,
		SlippageBPS:                percentageToBasicPoints(req.Slippage),
		StartedPrice:               mtype.NewDBBigRat(startedPrice),
		GoalPrice:                  mtype.NewDBBigRat(goalPrice),
		Status:                     model.StatusInUse,
		ParallelTransactionsAmount: parallelTransactionsAmount,
		MinTransactionsBudget:      req.MinTransactionsBudget,
		MaxTransactionsBudget:      req.MaxTransactionsBudget,
		MinTimeBetweenTransactions: req.MinTimeBetweenTransactions,
		MaxTimeBetweenTransactions: req.MaxTimeBetweenTransactions,
		TransactionSpeed:           req.TransactionSpeed,
		GoalBPSChange:              percentageToBasicPoints(req.GoalPercentageChange),
		UsingJito:                  req.UsingJito,
		CreatedAt:                  time.Now().UTC(),
		UpdatedAt:                  time.Now().UTC(),
		PriorityFee:                req.PriorityFee,
	}

	if err := s.swapCampaignRepository.Create(ctx, &campaign); err != nil {
		return nil, apperrors.Internal("failed to create swap campaign", err)
	}

	/* budget := &atomic.Pointer[big.Int]{}

	budgetInAtomicUnits := solanarpc.ToAtomicUnit(req.Budget, poolResult.SourceTokenDecimals)

	budget.Store(new(big.Int).SetUint64(budgetInAtomicUnits)) */

	budgetAtomic := solanarpc.ToAtomicUnit(req.Budget, poolResult.SourceTokenDecimals)
	budget := swapbudget.NewSwapBudget(new(big.Int).SetUint64(budgetAtomic))

	err = s.manager.AddTarget(ctx,
		req.MinTimeBetweenTransactions,
		req.MaxTimeBetweenTransactions,
		campaignID,
		parallelTransactionsAmount,
		budget,
		tasks,
	)
	if err != nil {
		return nil, apperrors.Internal("campaign created but worker failed to start", err)
	}

	return &model.TargetPullResponse{CampaignID: campaignID}, nil
}

func (s *SwapService) CreatePullDownCampaign(ctx context.Context, req *model.TargetPullDownRequest, userID uint64) (*model.TargetPullResponse, error) {
	wallets, totalBalance, err := s.fetchFundedWallets(ctx, req.ProjectID, userID, 0, req.SourceTokenMint)

	if err != nil {
		return nil, err
	}

	if req.BudgetPercent > 100 {
		return nil, apperrors.BadRequest("budget percent must be less than 100", nil)
	}

	if req.BudgetPercent > 0 {
		req.Budget = totalBalance * (req.BudgetPercent / 100)
	}

	wallets, _, err = s.filterWalletsForCampaign(wallets, req.MinTransactionsBudget, req.Budget, func(wallet model.Wallet) float64 {
		return wallet.BalanceToken
	})
	if err != nil {
		return nil, err
	}

	ataAddresses, ataAccounts, err := s.fetchATAInfo(ctx, wallets, req.SourceTokenMint)
	if err != nil {
		return nil, err
	}

	microlamportsPerCU := uint64(math.Ceil(
		(req.PriorityFee * 1_000_000_000 * 1_000_000) / float64(s.computeUnixLimit),
	))
	wallets, ataAddresses, ataAccounts, _, err = s.filterWalletsBySolReserve(ctx, wallets, ataAddresses, ataAccounts, solReserveFilter{
		MinTransactionBudget:       req.MinTransactionsBudget,
		CampaignBudget:             req.Budget,
		PriorityFeeMicroLamports:   microlamportsPerCU,
		IncludeMissingTokenATARent: false,
		DeductReserveFromSource:    false,
		BalanceFn: func(wallet model.Wallet) float64 {
			return wallet.BalanceToken
		},
	})
	if err != nil {
		return nil, err
	}

	// return nil, nil
	parallelTransactionsAmount := req.ParallelTransactionsAmount
	if parallelTransactionsAmount > len(wallets) {
		parallelTransactionsAmount = len(wallets)
	}

	privateKeys, err := s.fetchPrivateKeys(ctx, userID, wallets)
	if err != nil {
		return nil, err
	}

	provider, err := s.getDEXProvider(req.ProviderID)
	if err != nil {
		return nil, err
	}

	poolResult, startedPrice, err := provider.PreparePool(ctx, req.SourceTokenMint, solana.WrappedSol)
	if err != nil {
		return nil, err
	}
	goalPrice := CalculateGoalPrice(startedPrice, req.GoalPercentageChange, false)
	campaignID := uuid.New()

	params, err := provider.FetchPoolParams(ctx, poolResult.PoolID)
	if err != nil {
		return nil, err
	}

	tasks := make([]*model.AsyncSwapTask, len(wallets))
	errs := make([]error, len(wallets))

	for i, wallet := range wallets {
		walletKey, wErr := solana.PublicKeyFromBase58(wallet.PublicKey)
		if wErr != nil {
			errs[i] = apperrors.Internal("invalid wallet address", wErr)
			continue
		}

		wSolATA, _, wErr := solutil.FindAssociatedTokenAddressWithProgram(walletKey, solana.WrappedSol, solana.TokenProgramID)
		if wErr != nil {
			errs[i] = apperrors.Internal("cant find ata address", wErr)
			continue
		}

		tasks[i] = &model.AsyncSwapTask{
			SwapCampaignID:        campaignID,
			GoalPrice:             goalPrice,
			MinTransactionsAmount: req.MinTransactionsBudget,
			MaxTransactionsAmount: req.MaxTransactionsBudget,
			Slippage:              percentageToBasicPoints(req.Slippage),
			PoolID:                poolResult.PoolID,
			PoolProgramID:         poolResult.PoolProgramID,
			SourceAddress:         ataAddresses[i],
			SourceTokenMint:       req.SourceTokenMint,
			DestTokenMint:         solana.WrappedSol,
			DestAddress:           wSolATA,
			PrivateKey:            privateKeys[i],
			TaskType:              model.TargetDownTaskType,
			TransactionSpeed:      req.TransactionSpeed,
			ATAKeyCreated:         ataAccounts[i] != nil,
			PoolParams:            params,
			SourceTokenDecimals:   poolResult.SourceTokenDecimals,
			DestTokenDecimals:     poolResult.DestTokenDecimals,
			UsingJito:             req.UsingJito,
			PriorityFeeMLP:        microlamportsPerCU,
		}
	}

	if err = errors.Join(errs...); err != nil {
		return nil, err
	}

	campaign := model.SwapCampaign{
		ID:                         campaignID,
		CampaignTypeID:             2,
		UserID:                     userID,
		ProviderID:                 uint64(req.ProviderID),
		ProjectID:                  req.ProjectID,
		PoolID:                     poolResult.PoolID.String(),
		TokenMintFrom:              req.SourceTokenMint.String(),
		TokenMintTo:                solana.WrappedSol.String(),
		Budget:                     req.Budget,
		SlippageBPS:                percentageToBasicPoints(req.Slippage),
		StartedPrice:               mtype.NewDBBigRat(startedPrice),
		GoalPrice:                  mtype.NewDBBigRat(goalPrice),
		Status:                     model.StatusInUse,
		ParallelTransactionsAmount: parallelTransactionsAmount,
		MinTransactionsBudget:      req.MinTransactionsBudget,
		MaxTransactionsBudget:      req.MaxTransactionsBudget,
		MinTimeBetweenTransactions: req.MinTimeBetweenTransactions,
		MaxTimeBetweenTransactions: req.MaxTimeBetweenTransactions,
		TransactionSpeed:           req.TransactionSpeed,
		GoalBPSChange:              percentageToBasicPoints(req.GoalPercentageChange),
		UsingJito:                  req.UsingJito,
		CreatedAt:                  time.Now().UTC(),
		UpdatedAt:                  time.Now().UTC(),
		PriorityFee:                req.PriorityFee,
	}

	if err = s.swapCampaignRepository.Create(ctx, &campaign); err != nil {
		return nil, apperrors.Internal("failed to create swap campaign", err)
	}

	// budget := &atomic.Pointer[big.Int]{}

	budgetAtomic := solanarpc.ToAtomicUnit(req.Budget, poolResult.SourceTokenDecimals)
	budget := swapbudget.NewSwapBudget(new(big.Int).SetUint64(budgetAtomic))

	// budget.Store(new(big.Int).SetUint64(budgetInAtomicUnits))

	err = s.manager.AddTarget(ctx,
		req.MinTimeBetweenTransactions,
		req.MaxTimeBetweenTransactions,
		campaignID,
		parallelTransactionsAmount,
		budget,
		tasks,
	)
	if err != nil {
		return nil, apperrors.Internal("campaign created but worker failed to start", err)
	}

	return &model.TargetPullResponse{CampaignID: campaignID}, nil
}

func (s *SwapService) EstimateSwapCost(ctx context.Context, req *model.EstimatePullRequest, userID uint64) (*model.TargetPullEstimateResponse, error) {
	provider, err := s.getDEXProvider(req.ProviderID)
	if err != nil {
		return nil, err
	}

	pool, err := provider.FindPoolByMints(ctx, req.SourceTokenMint, req.DestTokenMint)
	if err != nil {
		return nil, apperrors.BadRequest("cannot find pool for swap", err)
	}

	wallets, _, err := s.fetchFundedWallets(ctx, req.ProjectID, userID, 0, req.SourceTokenMint)
	if err != nil {
		return nil, err
	}

	mintToCheck := req.DestTokenMint
	if !solutil.IsSOLLikeMint(req.SourceTokenMint) {
		mintToCheck = req.SourceTokenMint
	}

	_, ataAccounts, err := s.fetchATAInfo(ctx, wallets, mintToCheck)
	if err != nil {
		return nil, err
	}

	rent, err := s.solanaRPC.GetATARentExemption(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get ata rent exemption", err)
	}

	rentSol := solanarpc.FromAtomicUnit(rent, solana.SolDecimals)

	ataRentSol := 0.0
	for _, acc := range ataAccounts {
		if acc == nil {
			ataRentSol += rentSol
		}
	}

	tipFloor, err := s.jitoClient.GetTipFloor(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get tip floor", err)
	}

	jitoTip, err := jito.GetTipByTransactionSpeed(ctx, tipFloor, req.TransactionSpeed)
	if err != nil {
		return nil, apperrors.Internal("failed to get tip by transaction speed", err)
	}

	tipSOL := jitoTip*float64(len(wallets)) + pool.FeeRate*req.Budget

	accounts := []string{
		wallets[0].PublicKey,
		pool.ProgramID,
		pool.ID,
		mintToCheck.String(),
	}

	feeLevels, err := s.heliusClient.GetPriorityFeeEstimate(accounts)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch priority fee estimate", err)
	}

	lowTotalSOL := calculatePriorityFee(feeLevels.Medium, s.computeUnixLimit)
	mediumTotalSOL := calculatePriorityFee(feeLevels.High, s.computeUnixLimit)
	veryHighTotalSOL := calculatePriorityFee(feeLevels.VeryHigh, s.computeUnixLimit)

	return &model.TargetPullEstimateResponse{
		BudgetSOL: req.Budget,
		TipSOL:    tipSOL,
		RentSOl:   ataRentSol,
		PriorityFees: model.PriorityFees{
			Low:    lowTotalSOL,
			Medium: mediumTotalSOL,
			High:   veryHighTotalSOL,
		},
	}, nil
}

func calculatePriorityFee(lamportsPerCU float64, cuLimit uint32) float64 {
	totalMicroLamports := uint64(lamportsPerCU) * uint64(cuLimit)
	totalLamports := totalMicroLamports / 1_000_000
	totalSOL := float64(totalLamports) / 1_000_000_000
	return totalSOL
}

func (s *SwapService) getDEXProvider(providerID model.SwapProviderID) (model.DexProvider, error) {
	provider, ok := s.dexProviders[providerID]
	if !ok {
		return nil, apperrors.BadRequest("unsupported provider id")
	}

	return provider, nil
}

func (s *SwapService) fetchFundedWallets(ctx context.Context, projectID, userID uint64, minWallets int, mint solana.PublicKey) ([]model.Wallet, float64, error) {
	decimals := solana.SolDecimals
	tokenProgram := solana.TokenProgramID

	if !solutil.IsSOLLikeMint(mint) {
		info, err := s.solanaRPC.GetAccountInfo(ctx, mint)
		if err != nil {
			return nil, 0, err
		}

		if info == nil || info.Value == nil {
			return nil, 0, apperrors.BadRequest("mint not found", nil)
		}

		data := info.GetBinary()

		if data == nil {
			return nil, 0, apperrors.BadRequest("mint not found", nil)
		}

		tokenMint := token.Mint{}

		if err = tokenMint.UnmarshalWithDecoder(bin.NewBinDecoder(data)); err != nil {
			s.logger.Error("failed to parse mint in fetchFundedWallets",
				zap.Uint64("project_id", projectID),
				zap.Uint64("user_id", userID),
				zap.String("mint", mint.String()),
				zap.Int("data_len", len(data)),
				zap.Error(err),
			)
			return nil, 0, err
		}

		decimals = tokenMint.Decimals
		tokenProgram = info.Value.Owner
	}

	project, err := s.projectRepository.FetchProjectWithWalletsByID(ctx, projectID, userID)
	if err != nil {
		return nil, 0, apperrors.Internal("failed to fetch project", err)
	}

	if len(project.Wallets) > MaxWalletsPerRequest {
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
			address, _, err := solutil.FindAssociatedTokenAddressWithProgram(pk, mint, tokenProgram)
			if err != nil {
				return nil, 0, err
			}
			pubKeys[i] = address
		}
	}

	response, err := s.solanaRPC.GetMultipleAccountsWithNoLimits(ctx, pubKeys...)
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
				s.logger.Error("failed to parse token account in fetchFundedWallets",
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

		if balance < MinDustBalance {
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

func (s *SwapService) fetchPrivateKeys(ctx context.Context, userID uint64, wallets []model.Wallet) ([]solana.PrivateKey, error) {
	eg, errctx := errgroup.WithContext(ctx)
	eg.SetLimit(10)
	keys := make([]solana.PrivateKey, len(wallets))

	for i, w := range wallets {
		idx, wallet := i, w
		eg.Go(func() error {
			keyStr, err := s.keyStorage.Get(errctx, userID, wallet.PublicKey)
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

func (s *SwapService) fetchATAInfo(ctx context.Context, wallets []model.Wallet, mint solana.PublicKey) ([]solana.PublicKey, []*rpc.Account, error) {
	tokenProgram := solana.TokenProgramID
	if !solutil.IsSOLLikeMint(mint) {
		info, err := s.solanaRPC.GetAccountInfo(ctx, mint)
		if err != nil {
			return nil, nil, apperrors.Internal("failed to fetch mint account", err)
		}
		if info == nil || info.Value == nil || info.Value.Data == nil {
			return nil, nil, apperrors.BadRequest("mint not found", nil)
		}
		tokenProgram = info.Value.Owner
	}

	addresses := make([]solana.PublicKey, len(wallets))
	for i, w := range wallets {
		owner, err := solana.PublicKeyFromBase58(w.PublicKey)
		if err != nil {
			return nil, nil, apperrors.Internal("invalid wallet address", err)
		}
		addr, _, err := solutil.FindAssociatedTokenAddressWithProgram(owner, mint, tokenProgram)
		if err != nil {
			return nil, nil, apperrors.Internal("failed to get ata address", err)
		}
		addresses[i] = addr
	}

	accountsResp, err := s.solanaRPC.GetMultipleAccountsWithNoLimits(ctx, addresses...)
	if err != nil {
		return nil, nil, apperrors.Internal("failed to fetch ATA accounts", err)
	}

	accounts := make([]*rpc.Account, 0, len(addresses))
	for _, a := range accountsResp {
		accounts = append(accounts, a.Value...)
	}

	return addresses, accounts, nil
}

type solReserveFilter struct {
	MinTransactionBudget       float64
	CampaignBudget             float64
	PriorityFeeMicroLamports   uint64
	IncludeMissingTokenATARent bool
	DeductReserveFromSource    bool
	BalanceFn                  func(model.Wallet) float64
}

func (s *SwapService) filterWalletsBySolReserve(
	ctx context.Context,
	wallets []model.Wallet,
	ataAddresses []solana.PublicKey,
	ataAccounts []*rpc.Account,
	filter solReserveFilter,
) ([]model.Wallet, []solana.PublicKey, []*rpc.Account, float64, error) {
	if len(wallets) != len(ataAddresses) || len(wallets) != len(ataAccounts) {
		return nil, nil, nil, 0, apperrors.Internal("wallet and ATA info length mismatch")
	}

	owners := make([]solana.PublicKey, len(wallets))
	for i, wallet := range wallets {
		owner, err := solana.PublicKeyFromBase58(wallet.PublicKey)
		if err != nil {
			return nil, nil, nil, 0, apperrors.Internal("invalid wallet address", err)
		}
		owners[i] = owner
	}

	solBalances, err := s.solanaRPC.GetMulltipyWalletBalance(ctx, owners)
	if err != nil {
		return nil, nil, nil, 0, apperrors.Internal("failed to get wallet SOL balances", err)
	}
	if len(solBalances) != len(wallets) {
		return nil, nil, nil, 0, apperrors.Internal("wallet SOL balance length mismatch")
	}

	ataRentLamports, err := s.solanaRPC.GetATARentExemption(ctx)
	if err != nil {
		return nil, nil, nil, 0, apperrors.Internal("failed to get ata rent exemption", err)
	}

	eligibleWallets := make([]model.Wallet, 0, len(wallets))
	eligibleATAAddresses := make([]solana.PublicKey, 0, len(ataAddresses))
	eligibleATAAccounts := make([]*rpc.Account, 0, len(ataAccounts))
	totalEligibleBalance := 0.0

	for i, wallet := range wallets {
		createTokenATA := filter.IncludeMissingTokenATARent && ataAccounts[i] == nil
		reserveLamports := common.SolPayerReserveLamports(
			createTokenATA,
			ataRentLamports,
			s.computeUnixLimit,
			filter.PriorityFeeMicroLamports,
		)
		solBalanceLamports := solanarpc.SOLToLamports(solBalances[i])
		if solBalanceLamports <= reserveLamports {
			s.logger.Info(
				"wallet filtered by insufficient SOL reserve",
				zap.String("wallet", wallet.PublicKey),
				zap.Uint64("sol_balance_lamports", solBalanceLamports),
				zap.Uint64("required_reserve_lamports", reserveLamports),
				zap.Bool("create_token_ata", createTokenATA),
			)
			continue
		}

		wallet.BalanceSOL = solBalances[i]
		balance := filter.BalanceFn(wallet)
		if filter.DeductReserveFromSource {
			balance = solanarpc.FromAtomicUnit(solBalanceLamports-reserveLamports, solana.SolDecimals)
			wallet.BalanceSOL = balance
		}

		if balance < filter.MinTransactionBudget {
			s.logger.Info(
				"wallet filtered by insufficient spendable balance",
				zap.String("wallet", wallet.PublicKey),
				zap.Float64("spendable_balance", balance),
				zap.Float64("min_transaction_budget", filter.MinTransactionBudget),
				zap.Uint64("required_reserve_lamports", reserveLamports),
				zap.Bool("reserve_deducted_from_source", filter.DeductReserveFromSource),
			)
			continue
		}

		eligibleWallets = append(eligibleWallets, wallet)
		eligibleATAAddresses = append(eligibleATAAddresses, ataAddresses[i])
		eligibleATAAccounts = append(eligibleATAAccounts, ataAccounts[i])
		totalEligibleBalance += balance
	}

	if len(eligibleWallets) == 0 {
		return nil, nil, nil, 0, apperrors.BadRequest("no eligible wallets with enough SOL for rent and fees", nil)
	}

	if totalEligibleBalance < filter.CampaignBudget {
		return nil, nil, nil, 0, apperrors.BadRequest("not enough funds on eligible wallets after SOL reserve", nil)
	}

	return eligibleWallets, eligibleATAAddresses, eligibleATAAccounts, totalEligibleBalance, nil
}

func (s *SwapService) filterWalletsForCampaign(
	wallets []model.Wallet,
	minTransactionBudget, campaignBudget float64,
	balanceFn func(model.Wallet) float64,
) ([]model.Wallet, float64, error) {
	eligibleWallets := make([]model.Wallet, 0, len(wallets))
	totalEligibleBalance := 0.0

	for _, wallet := range wallets {
		balance := balanceFn(wallet)
		if balance < minTransactionBudget {
			s.logger.Info(
				"wallet filtered by insufficient campaign balance",
				zap.String("wallet", wallet.PublicKey),
				zap.Float64("balance", balance),
				zap.Float64("min_transaction_budget", minTransactionBudget),
				zap.Float64("campaign_budget", campaignBudget),
			)
			continue
		}

		eligibleWallets = append(eligibleWallets, wallet)
		totalEligibleBalance += balance
	}

	s.logger.Info(
		"wallet campaign balance filter completed",
		zap.Int("total_wallets", len(wallets)),
		zap.Int("eligible_wallets", len(eligibleWallets)),
		zap.Float64("total_eligible_balance", totalEligibleBalance),
		zap.Float64("campaign_budget", campaignBudget),
		zap.Float64("min_transaction_budget", minTransactionBudget),
	)

	if len(eligibleWallets) == 0 {
		return nil, 0, apperrors.BadRequest("no eligible wallets found for campaign start", nil)
	}

	if totalEligibleBalance < campaignBudget {
		return nil, 0, apperrors.BadRequest("not enough funds on eligible wallets for campaign budget", nil)
	}

	return eligibleWallets, totalEligibleBalance, nil
}
