package service

import (
	"context"
	"database/sql"
	"errors"
	"math/big"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/pkg/apperrors"
	"mm/pkg/mtype"
	repo "mm/pkg/repository"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
)

type SmartBuybackService struct {
	dexProviders  map[model.SwapProviderID]model.DexProvider
	buybackRepo   *repository.BuybackRepository
	buybackTxRepo *repository.BuybackTransactionRepository
	projectRepo   *repository.ProjectRepository
}

func NewSmartBuybackService(
	dexProviders map[model.SwapProviderID]model.DexProvider,
	buybackRepo *repository.BuybackRepository,
	buybackTxRepo *repository.BuybackTransactionRepository,
	projectRepo *repository.ProjectRepository,
) *SmartBuybackService {
	return &SmartBuybackService{
		dexProviders:  dexProviders,
		buybackRepo:   buybackRepo,
		buybackTxRepo: buybackTxRepo,
		projectRepo:   projectRepo,
	}
}

func (s *SmartBuybackService) CreateCampaign(ctx context.Context, userID uint64, req *model.CreateSmartBuybackCampaignRequest) (*model.SmartBuybackCampaignWithTargets, error) {
	project, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, apperrors.Internal("failed to get project", err)
	}

	if project.UserID != userID {
		return nil, apperrors.BadRequest("unknown project")
	}

	provider, err := s.getDEXProvider(model.SwapProviderID(req.ProviderID))
	if err != nil {
		return nil, apperrors.BadRequest("invalid provider_id", err)
	}

	tokenMint, err := solana.PublicKeyFromBase58(req.TokenMint)
	if err != nil {
		return nil, apperrors.BadRequest("invalid token mint", err)
	}

	poolID, err := provider.FindPoolByMints(ctx, solana.SolMint, tokenMint)
	if err != nil {
		return nil, apperrors.Internal("failed to get pool", err)
	}
	if poolID == nil {
		return nil, apperrors.BadRequest("pool not found for given token mint")
	}

	campaign := &model.SmartBuybackCampaignWithTargets{
		SmartBuybackCampaign: model.SmartBuybackCampaign{
			ID:            uuid.Must(uuid.NewV7()),
			UserID:        userID,
			ProviderID:    req.ProviderID,
			ProjectID:     req.ProjectID,
			TokenMint:     tokenMint.String(),
			PoolID:        poolID.ID,
			PoolProgramID: poolID.ProgramID,
			Status:        model.SmartBuybackCampaignStatusActive,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		Targets: make([]model.SmartBuybackCampaignTarget, len(req.Targets)),
	}

	for i, rt := range req.Targets {
		b, ok := new(big.Rat).SetString(rt.Budget)
		if !ok {
			return nil, apperrors.BadRequest("invalid budget")
		}

		tp, ok := new(big.Rat).SetString(rt.TargetPrice)
		if !ok {
			return nil, apperrors.BadRequest("invalid target price")
		}

		minTx, ok := new(big.Rat).SetString(rt.MinTransactionAmount)
		if !ok {
			return nil, apperrors.BadRequest("invalid min tx amount")
		}

		maxTx, ok := new(big.Rat).SetString(rt.MaxTransactionAmount)
		if !ok {
			return nil, apperrors.BadRequest("invalid max tx amount")
		}

		fee, ok := new(big.Rat).SetString(rt.PriorityFee)
		if !ok {
			return nil, apperrors.BadRequest("invalid priority fee")
		}

		t := model.SmartBuybackCampaignTarget{
			ID:                         uuid.Must(uuid.NewV7()),
			CampaignID:                 campaign.ID,
			Type:                       rt.Type,
			TargetPrice:                mtype.NewDBBigRat(tp),
			Budget:                     mtype.NewDBBigRat(b),
			RemainingBudget:            mtype.NewDBBigRat(b),
			Slippage:                   rt.Slippage,
			MinTransactionAmount:       mtype.NewDBBigRat(minTx),
			MaxTransactionAmount:       mtype.NewDBBigRat(maxTx),
			ParallelTransactionsAmount: rt.ParallelTransactionsAmount,
			MinTimeBetweenTransactions: rt.MinTimeBetweenTransactions,
			MaxTimeBetweenTransactions: rt.MaxTimeBetweenTransactions,
			TransactionSpeed:           rt.TransactionSpeed,
			UsingJito:                  rt.UsingJito,
			PriorityFee:                mtype.NewDBBigRat(fee),
			CreatedAt:                  time.Now().UTC(),
			UpdatedAt:                  time.Now().UTC(),
			StartAt:                    time.Unix(rt.StartAt, 0).UTC(),
		}

		if t.StartAt.After(time.Now()) {
			t.Status = model.SmartBuybackTargetStatusScheduled
		} else {
			t.Status = model.SmartBuybackTargetStatusActive
		}

		campaign.Targets[i] = t
	}

	createdCampaign, err := s.buybackRepo.CreateWithTargets(ctx, campaign)
	if err != nil {
		return nil, apperrors.Internal("failed to create campaign", err)
	}

	return createdCampaign, nil
}

func (c *SmartBuybackService) GetCampaigns(ctx context.Context, userID uint64) ([]model.SmartBuybackCampaign, error) {
	res, err := c.buybackRepo.FindAllWithOptions(ctx, &repo.Options{
		Filters: []repo.Filter{{
			Column:   "user_id",
			Operator: "=",
			Value:    strconv.FormatUint(userID, 10),
		}},
	})
	if err != nil {
		return nil, apperrors.Internal("failed to fetch campaigns", err)
	}

	return res, nil
}

func (c *SmartBuybackService) GetByID(ctx context.Context, userID uint64, id uuid.UUID) (*model.SmartBuybackCampaignWithTargets, error) {
	res, err := c.buybackRepo.GetWithTargets(ctx, id, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch campaign", err)
	}

	return res, nil
}

func (c *SmartBuybackService) GetTransactions(ctx context.Context, id uuid.UUID, userID uint64, targetID uuid.UUID, page, size int) ([]model.BuybackTransaction, error) {
	campaign, err := c.buybackRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.BadRequest("campaign not found", err)
		}
		return nil, apperrors.Internal("failed to fetch campaigns", err)
	}
	if campaign.UserID != userID {
		return nil, apperrors.BadRequest("campaign not found", err)
	}

	options := &repo.Options{
		Filters: []repo.Filter{
			{
				Column:   "campaign_id",
				Operator: "=",
				Value:    campaign.ID.String(),
			},
		},
		Order: repo.Order{
			OrderBy:   "created_at",
			OrderType: repo.OrderDesc,
		},
		Pagination: repo.Pagination{
			PageSize: size,
			PageNum:  page,
		},
	}

	if targetID != uuid.Nil {
		options.Filters = append(options.Filters, repo.Filter{
			Column:   "target_id",
			Operator: "=",
			Value:    targetID.String(),
		})
	}

	res, err := c.buybackTxRepo.FindAllWithOptions(ctx, options)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch transactions", err)
	}

	return res, nil
}

func (s *SmartBuybackService) StopCampaign(ctx context.Context, userID uint64, id uuid.UUID) error {
	campaign, err := s.buybackRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.BadRequest("campaign not found")
		}
		return apperrors.Internal("failed to fetch campaign", err)
	}
	if campaign.UserID != userID {
		return apperrors.BadRequest("campaign not found")
	}
	if campaign.Status != model.SmartBuybackCampaignStatusActive {
		return apperrors.BadRequest("campaign is not active")
	}

	_, err = s.buybackRepo.DB.NewUpdate().
		Model((*model.SmartBuybackCampaign)(nil)).
		Where("id = ?", id).
		Set("status = ?", model.SmartBuybackCampaignStatusDone).
		Set("updated_at = NOW()").
		Exec(ctx)
	return err
}

func (s *SmartBuybackService) getDEXProvider(providerID model.SwapProviderID) (model.DexProvider, error) {
	provider, ok := s.dexProviders[providerID]
	if !ok {
		return nil, apperrors.BadRequest("unsupported provider id")
	}

	return provider, nil
}
