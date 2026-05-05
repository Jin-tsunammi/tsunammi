package repository

import (
	"context"
	"math/big"
	"mm/internal/model"
	"mm/pkg/mtype"
	"mm/pkg/repository"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type BuybackRepository struct {
	repository.Generic[model.SmartBuybackCampaign, uuid.UUID]
}

type BuybackTransactionRepository struct {
	repository.Generic[model.BuybackTransaction, uint64]
}

func NewBuybackRepository(genericRepo repository.Generic[model.SmartBuybackCampaign, uuid.UUID]) *BuybackRepository {
	return &BuybackRepository{Generic: genericRepo}
}

func NewBuybackTransactionRepository(genericRepo repository.Generic[model.BuybackTransaction, uint64]) *BuybackTransactionRepository {
	return &BuybackTransactionRepository{Generic: genericRepo}
}

func (r *BuybackRepository) CreateWithTargets(ctx context.Context, campaign *model.SmartBuybackCampaignWithTargets) (_ *model.SmartBuybackCampaignWithTargets, err error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.NewInsert().
		Model(&campaign.SmartBuybackCampaign).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	if len(campaign.Targets) > 0 {
		for i := range campaign.Targets {
			target := campaign.Targets[i]
			_, err = tx.NewInsert().
				Model(&target).
				Returning("*").
				Exec(ctx)
			if err != nil {
				return nil, err
			}
			campaign.Targets[i] = target
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return campaign, nil
}

func (r *BuybackRepository) GetActiveWithTargetsByID(ctx context.Context, campaignID uuid.UUID) (*model.SmartBuybackCampaignWithTargets, error) {
	campaign := new(model.SmartBuybackCampaignWithTargets)

	if err := r.DB.NewSelect().
		Model(&campaign.SmartBuybackCampaign).
		Where("bc.id = ?", campaignID).
		Scan(ctx); err != nil {
		return nil, err
	}

	targets := make([]model.SmartBuybackCampaignTarget, 0)
	if err := r.DB.NewSelect().
		Model(&targets).
		ModelTableExpr("buyback_targets AS bt").
		Where("bt.campaign_id = ?", campaignID).
		Order("bt.created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	campaign.Targets = targets

	return campaign, nil
}

func (r *BuybackRepository) GetActiveWithTargets(ctx context.Context) ([]model.SmartBuybackCampaignWithTargets, error) {
	campaigns := make([]model.SmartBuybackCampaign, 0)
	if err := r.DB.NewSelect().
		Model(&campaigns).
		ModelTableExpr("buyback_campaigns AS bc").
		Where("bc.status = ?", model.SmartBuybackCampaignStatusActive).
		Order("bc.created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	if len(campaigns) == 0 {
		return []model.SmartBuybackCampaignWithTargets{}, nil
	}

	campaignIDs := make([]uuid.UUID, 0, len(campaigns))
	for _, campaign := range campaigns {
		campaignIDs = append(campaignIDs, campaign.ID)
	}

	targets := make([]model.SmartBuybackCampaignTarget, 0)
	if err := r.DB.NewSelect().
		Model(&targets).
		ModelTableExpr("buyback_targets AS bt").
		Where("bt.campaign_id IN (?)", bun.In(campaignIDs)).
		Where("bt.status IN (?)", bun.In([]string{
			model.SmartBuybackTargetStatusActive,
			model.SmartBuybackTargetStatusScheduled,
		})).
		Order("bt.created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	targetsByCampaignID := make(map[uuid.UUID][]model.SmartBuybackCampaignTarget, len(campaigns))
	for _, target := range targets {
		targetsByCampaignID[target.CampaignID] = append(targetsByCampaignID[target.CampaignID], target)
	}

	result := make([]model.SmartBuybackCampaignWithTargets, 0, len(campaigns))
	for _, campaign := range campaigns {
		result = append(result, model.SmartBuybackCampaignWithTargets{
			SmartBuybackCampaign: campaign,
			Targets:              targetsByCampaignID[campaign.ID],
		})
	}

	return result, nil
}

func (r *BuybackRepository) GetWithTargets(ctx context.Context, id uuid.UUID, userID uint64) (*model.SmartBuybackCampaignWithTargets, error) {
	var campaign model.SmartBuybackCampaign
	if err := r.DB.NewSelect().
		Model(&campaign).
		ModelTableExpr("buyback_campaigns AS bc").
		Where("bc.id = ?", id).
		Where("bc.user_id = ?", userID).
		// Where("bc.status = ?", model.SmartBuybackCampaignStatusActive).
		Order("bc.created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	/* 	campaignIDs := make([]uuid.UUID, 0, len(campaigns))
	   	for _, campaign := range campaigns {
	   		campaignIDs = append(campaignIDs, campaign.ID)
	   	} */

	targets := make([]model.SmartBuybackCampaignTarget, 0)
	if err := r.DB.NewSelect().
		Model(&targets).
		ModelTableExpr("buyback_targets AS bt").
		Where("bt.campaign_id = ?", campaign.ID).
		// Where("bt.status IN (?)", bun.In([]string{
		// 	model.SmartBuybackTargetStatusActive,
		// 	model.SmartBuybackTargetStatusScheduled,
		// })).
		Order("bt.created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	result := new(model.SmartBuybackCampaignWithTargets)
	result.SmartBuybackCampaign = campaign
	result.Targets = targets

	return result, nil
}

func (r *BuybackRepository) UpdateTargetRemainingBudget(ctx context.Context, targetID uuid.UUID, remaining *big.Rat, status string) error {
	query := r.DB.NewUpdate().
		Model((*model.SmartBuybackCampaignTarget)(nil)).
		Where("id = ?", targetID).
		Set("remaining_budget = ?", mtype.NewDBBigRat(remaining)).
		Set("updated_at = ?", time.Now().UTC())

	if status != "" {
		query = query.Set("status = ?", status)
	}

	_, err := query.Exec(ctx)
	return err
}

func (r *BuybackRepository) UpdateCampaignError(ctx context.Context, campaignID uuid.UUID) error {
	_, err := r.DB.NewUpdate().
		Model((*model.SmartBuybackCampaign)(nil)).
		Where("id = ?", campaignID).
		Set("status = ?", model.SmartBuybackCampaignStatusError).
		Set("updated_at = ?", time.Now().UTC()).
		Exec(ctx)
	return err
}

func (r *BuybackRepository) UpdateDoneIfNoActiveTargets(ctx context.Context, campaignID uuid.UUID) (bool, error) {
	result, err := r.DB.NewUpdate().
		Model((*model.SmartBuybackCampaign)(nil)).
		Where("id = ?", campaignID).
		Where("status = ?", model.SmartBuybackCampaignStatusActive).
		Where(`NOT EXISTS (
			SELECT 1
			FROM buyback_targets AS bt
			WHERE bt.campaign_id = ?
			  AND bt.status IN (?)
		)`, campaignID, bun.In([]string{
			model.SmartBuybackTargetStatusActive,
			model.SmartBuybackTargetStatusScheduled,
		})).
		Set("status = ?", model.SmartBuybackCampaignStatusDone).
		Set("updated_at = ?", time.Now().UTC()).
		Exec(ctx)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
