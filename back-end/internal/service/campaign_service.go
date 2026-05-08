package service

import (
	"context"
	"errors"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/pricing"
	"mm/internal/storage/repository"
	"mm/internal/worker/swaptarget"
	"mm/pkg/apperrors"
	"mm/pkg/mtype"
	repo "mm/pkg/repository"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CampaignService struct {
	CampaignRepository    *repository.SwapCampaignRepository
	TransactionRepository *repository.SwapTransactionRepository
	TaskManager           *swaptarget.SwapTargetManager
	TransactionManager    *repo.TransactionManager
	solanaRPC             solanarpc.SolanaRPC
}

func NewCampaignService(
	campaignRepository *repository.SwapCampaignRepository,
	transactionRepository *repository.SwapTransactionRepository,
	taskManager *swaptarget.SwapTargetManager,
	transactionManager *repo.TransactionManager,
	solanaRPC solanarpc.SolanaRPC,
) *CampaignService {
	return &CampaignService{
		CampaignRepository:    campaignRepository,
		TransactionRepository: transactionRepository,
		TaskManager:           taskManager,
		TransactionManager:    transactionManager,
		solanaRPC:             solanaRPC,
	}
}

func (s *CampaignService) GetCampaigns(ctx context.Context, userID uint64, page, pageSize int) (*model.CampaignsWithPaginationResponse, error) {
	campaigns, total, err := s.CampaignRepository.FindAllByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, apperrors.Internal("failed to get campaigns", err)
	}

	if len(campaigns) == 0 {
		return &model.CampaignsWithPaginationResponse{
			Campaigns: make([]model.SwapCampaign, 0),
			PageSize:  pageSize,
			Page:      page,
			Total:     total,
		}, nil
	}

	var poolID solana.PublicKey

	poolIDs := make([]solana.PublicKey, len(campaigns))
	errs := make([]error, len(campaigns))

	for index, campaign := range campaigns {
		poolID, err = solana.PublicKeyFromBase58(campaign.PoolID)
		if err != nil {
			errs[index] = err
			continue
		}

		poolIDs[index] = poolID
	}

	if err = errors.Join(errs...); err != nil {
		return nil, err
	}

	res, err := s.solanaRPC.GetMultipleAccountsWithNoLimits(ctx, poolIDs...)

	if err != nil {
		return nil, apperrors.Internal("failed to get pools", err)
	}

	accounts := make([]*rpc.Account, 0, len(campaigns))

	for _, r := range res {
		accounts = append(accounts, r.Value...)
	}

	errs = make([]error, len(campaigns))

	for index := range campaigns {
		campaign := &campaigns[index]
		account := accounts[index]

		if account == nil {
			errs[index] = apperrors.NotFound("pool not found")
			continue
		}

		tokeMintFrom, fErr := solana.PublicKeyFromBase58(campaign.TokenMintFrom)
		tokeMintTo, tErr := solana.PublicKeyFromBase58(campaign.TokenMintTo)

		if err = errors.Join(fErr, tErr); err != nil {
			errs[index] = err
			continue
		}

		currentPrice, err := pricing.CalculatePoolPrice(ctx, s.solanaRPC, account, poolIDs[index], tokeMintFrom, tokeMintTo)

		if err != nil {
			errs[index] = err
			continue
		}

		campaign.CurrentPrice = mtype.NewDBBigRat(currentPrice)
		campaign.GoalPercentChange = basicPointToBasicPoints(campaign.GoalBPSChange)
	}
	return &model.CampaignsWithPaginationResponse{
		Campaigns: campaigns,
		PageSize:  pageSize,
		Page:      page,
		Total:     total,
	}, nil
}

func (s *CampaignService) StopCampaign(ctx context.Context, campaignID uuid.UUID, userID uint64) error {
	err := s.TransactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		campaign, err := s.CampaignRepository.WithTx(tx).GetByIDAndUserID(ctx, campaignID, userID)

		if err != nil {
			return apperrors.Internal("failed to get campaign", err)
		}

		campaign.Status = model.SwapStatusStop

		if err = s.CampaignRepository.WithTx(tx).Update(ctx, campaign); err != nil {
			return apperrors.Internal("failed to stop campaign", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = s.TaskManager.DeleteTarget(ctx, campaignID)

	if err != nil {
		return apperrors.Internal("failed to delete target", err)
	}

	return nil
}

func (s *CampaignService) UpdateCampaign(ctx context.Context, campaignID uuid.UUID, userID uint64, req *model.CampaignRequest) error {

	err := s.TransactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		campaign, err := s.CampaignRepository.WithTx(tx).GetByIDAndUserID(ctx, campaignID, userID)
		if err != nil {
			return apperrors.Internal("failed to get campaign", err)
		}

		if req.Budget != nil {
			campaign.Budget = *req.Budget
		}

		if req.GoalPercentageChange != nil {
			newGoalPrice := CalculateGoalPrice(&campaign.StartedPrice.Rat, *req.GoalPercentageChange, campaign.CampaignType.Name == model.TargetUpTaskType)
			campaign.GoalPrice = mtype.NewDBBigRat(newGoalPrice)
			campaign.GoalBPSChange = percentageToBasicPoints(*req.GoalPercentageChange)
		}

		if req.Slippage != nil {
			campaign.SlippageBPS = percentageToBasicPoints(*req.Slippage)
		}

		if req.ParallelTransactionsAmount != nil {
			campaign.ParallelTransactionsAmount = *req.ParallelTransactionsAmount
		}

		if req.MinTransactionsAmount != nil {
			campaign.MinTransactionsBudget = *req.MinTransactionsAmount
		}
		if req.MaxTransactionsAmount != nil {
			campaign.MaxTransactionsBudget = *req.MaxTransactionsAmount
		}

		if req.MinTimeBetweenTransactions != nil {
			campaign.MinTimeBetweenTransactions = *req.MinTimeBetweenTransactions
		}
		if req.MaxTimeBetweenTransactions != nil {
			campaign.MaxTimeBetweenTransactions = *req.MaxTimeBetweenTransactions
		}

		if req.TransactionSpeed != nil {
			campaign.TransactionSpeed = *req.TransactionSpeed
		}

		if campaign.MinTransactionsBudget > campaign.MaxTransactionsBudget {
			return apperrors.BadRequest("min_transactions_amount cannot be greater than max_transactions_amount", nil)
		}

		if campaign.MinTimeBetweenTransactions > campaign.MaxTimeBetweenTransactions {
			return apperrors.BadRequest("min_time_between_transactions cannot be greater than max_time_between_transactions", nil)
		}

		campaign.UpdatedAt = time.Now().UTC()

		err = s.CampaignRepository.WithTx(tx).Update(ctx, campaign)

		if err != nil {
			return apperrors.Internal("failed to update campaign", err)
		}

		err = s.TaskManager.UpdateTarget(campaignID, campaign)

		if err != nil {
			return apperrors.Internal("failed to update target", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *CampaignService) GetCampaignTransactions(ctx context.Context, campaignID uuid.UUID, userID uint64, page, pageSize int) (*model.SwapTransactionsWithPaginationResponse, error) {

	transactions, total, err := s.TransactionRepository.FindAllByCampaignID(ctx, campaignID, page, pageSize)
	if err != nil {
		return nil, apperrors.Internal("failed to get campaign transactions", err)
	}

	return &model.SwapTransactionsWithPaginationResponse{
		Transactions: transactions,
		Page:         page,
		PageSize:     pageSize,
		Total:        total,
	}, nil
}

func (s *CampaignService) GetCampaignsSummary(ctx context.Context, parsedPage, parsedPageSize int, userID uint64, campaignType string, status model.SwapStatus) (*model.CampaignSummaryWithPagination, error) {
	summary, total, err := s.CampaignRepository.GetCampaignsSummary(ctx, parsedPage, parsedPageSize, userID, campaignType, status)
	if err != nil {
		return nil, apperrors.Internal("failed to get campaigns summary", err)
	}

	return &model.CampaignSummaryWithPagination{
		CampaignSummary: summary,
		Total:           total,
		PageSize:        parsedPageSize,
		Page:            parsedPage,
	}, nil
}

func (s *CampaignService) GetCampaignByID(ctx context.Context, campaignID uuid.UUID, userID uint64) (*model.SwapCampaign, error) {
	campaign, err := s.CampaignRepository.GetByIDAndUserID(ctx, campaignID, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to get campaign", err)
	}

	poolID, err := solana.PublicKeyFromBase58(campaign.PoolID)
	if err != nil {
		return nil, apperrors.Internal("failed to parse pool ID", err)
	}

	info, err := s.solanaRPC.GetAccountInfo(ctx, poolID)
	if err != nil {
		return nil, apperrors.Internal("failed to get pool", err)
	}

	if info == nil || info.Value == nil {
		return nil, apperrors.NotFound("pool not found")
	}

	account := info.Value

	tokenMintFrom, err := solana.PublicKeyFromBase58(campaign.TokenMintFrom)
	if err != nil {
		return nil, apperrors.Internal("failed to parse token mint from", err)
	}

	tokenMintTo, err := solana.PublicKeyFromBase58(campaign.TokenMintTo)
	if err != nil {
		return nil, apperrors.Internal("failed to parse token mint to", err)
	}

	currentPrice, err := pricing.CalculatePoolPrice(ctx, s.solanaRPC, account, poolID, tokenMintFrom, tokenMintTo)
	if err != nil {
		return nil, apperrors.Internal("failed to calculate pool price", err)
	}

	campaign.CurrentPrice = mtype.NewDBBigRat(currentPrice)
	campaign.GoalPercentChange = basicPointToBasicPoints(campaign.GoalBPSChange)

	return campaign, nil
}
