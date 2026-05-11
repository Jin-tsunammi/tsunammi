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
	"strings"
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
			Status:        model.BuybackStatusActive,
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
			t.Status = model.BuybackStatusScheduled
		} else {
			t.Status = model.BuybackStatusActive
		}

		campaign.Targets[i] = t
	}

	createdCampaign, err := s.buybackRepo.CreateWithTargets(ctx, campaign)
	if err != nil {
		return nil, apperrors.Internal("failed to create campaign", err)
	}

	return createdCampaign, nil
}

func (c *SmartBuybackService) GetCampaigns(ctx context.Context, userID uint64, req *model.GetAllBuybackCampaignsRequest) (*model.GetSmartBuybackCampaignsResponse, error) {
	params := &repo.Options{
		Filters: []repo.Filter{
			{
				Column:   "user_id",
				Operator: "=",
				Value:    strconv.FormatUint(userID, 10),
			},
		},
		Order: repo.Order{
			OrderBy:   "created_at",
			OrderType: repo.OrderDesc,
		},
		Pagination: repo.Pagination{
			PageSize: req.Size,
			PageNum:  req.Page,
		},
	}

	if len(req.Status) != 0 {
		params.Filters = append(params.Filters, repo.Filter{
			Column:   "status",
			Operator: "in",
			Value:    strings.Join(req.Status, ","),
		})
	}

	campaigns, err := c.buybackRepo.FindAllWithOptions(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch campaigns", err)
	}

	count, err := c.buybackRepo.CountWithOptions(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch count", err)
	}

	return &model.GetSmartBuybackCampaignsResponse{
		Page:      req.Page,
		PageSize:  req.Size,
		Total:     count,
		Campaigns: campaigns,
	}, nil
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
	if campaign.Status != model.BuybackStatusActive {
		return apperrors.BadRequest("campaign is not active")
	}

	_, err = s.buybackRepo.DB.NewUpdate().
		Model((*model.SmartBuybackCampaign)(nil)).
		Where("id = ?", id).
		Set("status = ?", model.BuybackStatusStop).
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

func (s *SmartBuybackService) CreateTarget(ctx context.Context, userID uint64, req *model.CreateSmartBuybackTargetRequest) (*model.SmartBuybackCampaignTarget, error) {
	campaign, err := s.buybackRepo.GetWithTargets(ctx, req.CampaignID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("campaign not found")
		}
		return nil, apperrors.Internal("failed to get campaign", err)
	}

	if len(campaign.Targets) >= 5 {
		return nil, apperrors.UnprocessableEntity("limit of 5 targets reached")
	}

	b, ok := new(big.Rat).SetString(req.Budget)
	if !ok {
		return nil, apperrors.BadRequest("invalid budget")
	}

	tp, ok := new(big.Rat).SetString(req.TargetPrice)
	if !ok {
		return nil, apperrors.BadRequest("invalid target price")
	}

	minTx, ok := new(big.Rat).SetString(req.MinTransactionAmount)
	if !ok {
		return nil, apperrors.BadRequest("invalid min tx amount")
	}

	maxTx, ok := new(big.Rat).SetString(req.MaxTransactionAmount)
	if !ok {
		return nil, apperrors.BadRequest("invalid max tx amount")
	}

	fee, ok := new(big.Rat).SetString(req.PriorityFee)
	if !ok {
		return nil, apperrors.BadRequest("invalid priority fee")
	}

	t := &model.SmartBuybackCampaignTarget{
		ID:                         uuid.Must(uuid.NewV7()),
		CampaignID:                 campaign.ID,
		Type:                       req.Type,
		TargetPrice:                mtype.NewDBBigRat(tp),
		Budget:                     mtype.NewDBBigRat(b),
		RemainingBudget:            mtype.NewDBBigRat(b),
		Slippage:                   req.Slippage,
		MinTransactionAmount:       mtype.NewDBBigRat(minTx),
		MaxTransactionAmount:       mtype.NewDBBigRat(maxTx),
		ParallelTransactionsAmount: req.ParallelTransactionsAmount,
		MinTimeBetweenTransactions: req.MinTimeBetweenTransactions,
		MaxTimeBetweenTransactions: req.MaxTimeBetweenTransactions,
		TransactionSpeed:           req.TransactionSpeed,
		UsingJito:                  req.UsingJito,
		PriorityFee:                mtype.NewDBBigRat(fee),
		CreatedAt:                  time.Now().UTC(),
		UpdatedAt:                  time.Now().UTC(),
		StartAt:                    time.Unix(req.StartAt, 0).UTC(),
	}

	if t.StartAt.After(time.Now()) {
		t.Status = model.BuybackStatusScheduled
	} else {
		t.Status = model.BuybackStatusActive
	}

	t, err = s.buybackRepo.CreateTarget(ctx, t)
	if err != nil {
		return nil, apperrors.Internal("failed to create target", err)
	}

	return t, nil
}

func (s *SmartBuybackService) StopTarget(ctx context.Context, userID uint64, id uuid.UUID) error {
	err := s.buybackRepo.UpdateTargetStatus(ctx, id, userID, model.BuybackStatusStop)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound("target not found")
		}
		return apperrors.Internal("failed to stop target", err)
	}

	return nil
}

func (s *SmartBuybackService) UpdateTarget(ctx context.Context, userID uint64, req *model.UpdateSmartBuybackTargetRequest) (*model.SmartBuybackCampaignTarget, error) {
	target, err := s.buybackRepo.GetTarget(ctx, req.ID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.BadRequest("target not found", err)
		}
		return nil, apperrors.Internal("failed to get target", err)
	}

	if req.TargetPrice != "" {
		tp, ok := new(big.Rat).SetString(req.TargetPrice)
		if !ok || tp.Sign() <= 0 {
			return nil, apperrors.BadRequest("invalid target price")
		}
		target.TargetPrice = mtype.NewDBBigRat(tp)
	}

	if req.Budget != "" {
		b, ok := new(big.Rat).SetString(req.Budget)
		if !ok || b.Sign() <= 0 {
			return nil, apperrors.BadRequest("invalid budget")
		}
		spent := new(big.Rat).Sub(target.Budget.GetBigRat(), target.RemainingBudget.GetBigRat())
		if b.Cmp(spent) < 0 {
			return nil, apperrors.BadRequest("budget cannot be less than already spent amount")
		}
		target.Budget = mtype.NewDBBigRat(b)
		target.RemainingBudget = mtype.NewDBBigRat(new(big.Rat).Sub(b, spent))
	}

	if req.PriorityFee != "" {
		fee, ok := new(big.Rat).SetString(req.PriorityFee)
		if !ok || fee.Sign() < 0 {
			return nil, apperrors.BadRequest("invalid priority fee")
		}
		target.PriorityFee = mtype.NewDBBigRat(fee)
	}

	if req.MinTransactionAmount != "" {
		minTx, ok := new(big.Rat).SetString(req.MinTransactionAmount)
		if !ok || minTx.Sign() <= 0 {
			return nil, apperrors.BadRequest("invalid min tx amount")
		}
		target.MinTransactionAmount = mtype.NewDBBigRat(minTx)
	}

	if req.MaxTransactionAmount != "" {
		maxTx, ok := new(big.Rat).SetString(req.MaxTransactionAmount)
		if !ok || maxTx.Sign() <= 0 {
			return nil, apperrors.BadRequest("invalid max tx amount")
		}
		target.MaxTransactionAmount = mtype.NewDBBigRat(maxTx)
	}

	if target.MinTransactionAmount.GetBigRat().Cmp(target.MaxTransactionAmount.GetBigRat()) > 0 {
		return nil, apperrors.BadRequest("min tx amount cannot be greater than max tx amount")
	}

	if req.MinTimeBetweenTransactions != nil {
		if *req.MinTimeBetweenTransactions < 0 {
			return nil, apperrors.BadRequest("invalid min time between transactions")
		}
		target.MinTimeBetweenTransactions = *req.MinTimeBetweenTransactions
	}

	if req.MaxTimeBetweenTransactions != nil {
		if *req.MaxTimeBetweenTransactions < 0 {
			return nil, apperrors.BadRequest("invalid max time between transactions")
		}
		target.MaxTimeBetweenTransactions = *req.MaxTimeBetweenTransactions
	}

	if target.MinTimeBetweenTransactions > target.MaxTimeBetweenTransactions {
		return nil, apperrors.BadRequest("min time between transactions cannot be greater than max")
	}

	if req.TransactionSpeed != "" {
		if req.TransactionSpeed != model.Default && req.TransactionSpeed != model.Fast && req.TransactionSpeed != model.Extra {
			return nil, apperrors.BadRequest("invalid transaction speed")
		}
		target.TransactionSpeed = req.TransactionSpeed
	}

	if req.Slippage != nil {
		target.Slippage = *req.Slippage
	}

	if req.ParallelTransactionsAmount != nil {
		if *req.ParallelTransactionsAmount == 0 {
			return nil, apperrors.BadRequest("parallel transactions amount must be greater than 0")
		}
		target.ParallelTransactionsAmount = *req.ParallelTransactionsAmount
	}

	if req.UsingJito != nil {
		target.UsingJito = *req.UsingJito
	}

	if req.StartAt != nil {
		t := time.Unix(*req.StartAt, 0)
		if t.After(time.Now()) {
			target.Status = model.BuybackStatusScheduled
		} else if t.Before(time.Now()) || t.Equal(time.Now()) {
			target.Status = model.BuybackStatusActive
		}
		target.StartAt = t
	}

	target, err = s.buybackRepo.UpdateTarget(ctx, target)
	if err != nil {
		return nil, apperrors.Internal("failed to update target", err)
	}

	return target, nil
}
