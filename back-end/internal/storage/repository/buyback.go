package repository

import (
	"context"
	"database/sql"
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

func (r *BuybackRepository) CreateTarget(ctx context.Context, t *model.SmartBuybackCampaignTarget) (*model.SmartBuybackCampaignTarget, error) {
	_, err := r.DB.NewInsert().
		Model(t).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
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
		Where("bc.status = ?", model.BuybackStatusActive).
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
			string(model.BuybackStatusActive),
			string(model.BuybackStatusScheduled),
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
		Set("status = ?", model.BuybackStatusError).
		Set("updated_at = ?", time.Now().UTC()).
		Exec(ctx)
	return err
}

func (r *BuybackRepository) UpdateTargetStatus(ctx context.Context, id uuid.UUID, userID uint64, status model.BuybackStatus) error {
	res, err := r.DB.NewUpdate().
		TableExpr("buyback_targets AS bt").
		Set("status = ?", status).
		Set("updated_at = ?", time.Now().UTC()).
		Where("bt.id = ?", id).
		Where("EXISTS (SELECT 1 FROM buyback_campaigns AS bc WHERE bc.id = bt.campaign_id AND bc.user_id = ?)", userID).
		Exec(ctx)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *BuybackRepository) GetTarget(ctx context.Context, id uuid.UUID, userID uint64) (*model.SmartBuybackCampaignTarget, error) {
	var res model.SmartBuybackCampaignTarget

	err := r.DB.NewSelect().
		Model(&res).
		ModelTableExpr("buyback_targets AS bt").
		Join("JOIN buyback_campaigns AS bc ON bc.id = bt.campaign_id").
		Where("bt.id = ?", id).
		Where("bc.user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *BuybackRepository) UpdateDoneIfNoActiveTargets(ctx context.Context, campaignID uuid.UUID) (bool, error) {
	result, err := r.DB.NewUpdate().
		Model((*model.SmartBuybackCampaign)(nil)).
		Where("id = ?", campaignID).
		Where("status = ?", model.BuybackStatusActive).
		Where(`NOT EXISTS (
			SELECT 1
			FROM buyback_targets AS bt
			WHERE bt.campaign_id = ?
			  AND bt.status IN (?)
		)`, campaignID, bun.In([]string{
			string(model.BuybackStatusActive),
			string(model.BuybackStatusScheduled),
		})).
		Set("status = ?", model.BuybackStatusDone).
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

func (r *BuybackRepository) UpdateTarget(ctx context.Context, target *model.SmartBuybackCampaignTarget) (*model.SmartBuybackCampaignTarget, error) {
	target.UpdatedAt = time.Now().UTC()

	_, err := r.DB.NewUpdate().Model(target).WherePK().Returning("*").Exec(ctx)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (r *BuybackTransactionRepository) FindAllByStatus(ctx context.Context, status string) ([]model.BuybackTransaction, error) {
	transactions := make([]model.BuybackTransaction, 0)

	err := r.DB.NewSelect().
		Model(&transactions).
		Where("status = ?", status).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *BuybackTransactionRepository) UpdateAll(ctx context.Context, transactions []model.BuybackTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	values := r.DB.NewValues(&transactions)

	_, err := r.DB.NewUpdate().
		With("_data", values).
		Model((*model.BuybackTransaction)(nil)).
		TableExpr("_data").
		Set("amount_token_from = (_data.amount_token_from #>> '{}')::numeric").
		Set("amount_token_to = (_data.amount_token_to #>> '{}')::numeric").
		Set("status = _data.status").
		Set("message = _data.message").
		Set("debug_message = _data.debug_message").
		Where("bbtx.id = _data.id").
		Exec(ctx)

	return err
}
