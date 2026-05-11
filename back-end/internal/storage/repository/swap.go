package repository

import (
	"context"
	"mm/internal/model"
	"mm/pkg/repository"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SwapTransactionRepository struct {
	repository.Generic[model.SwapTransaction, uint64]
}

func NewSwapTransactionRepository(genericRepository repository.Generic[model.SwapTransaction, uint64]) *SwapTransactionRepository {
	return &SwapTransactionRepository{Generic: genericRepository}
}

func (r *SwapTransactionRepository) FindAllByCampaignID(ctx context.Context, campaignID uuid.UUID, page, pageSize int) ([]model.SwapTransaction, int, error) {
	transactions := make([]model.SwapTransaction, 0)
	query := r.DB.NewSelect().
		Model(&transactions).
		Where("campaign_id = ?", campaignID)

	total, err := query.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if page > 0 && pageSize > 0 {
		query = query.Limit(pageSize).Offset((page - 1) * pageSize)
	}

	err = query.Scan(ctx)

	if err != nil {
		return nil, 0, err
	}
	return transactions, total, nil
}

func (r *SwapTransactionRepository) FindAllByStatus(ctx context.Context, status string) ([]model.SwapTransaction, error) {
	campaigns := make([]model.SwapTransaction, 0)

	err := r.DB.NewSelect().
		Model(&campaigns).
		Where("status = ?", status).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return campaigns, nil
}

func (r *SwapTransactionRepository) UpdateAll(ctx context.Context, transactions []model.SwapTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	values := r.DB.NewValues(&transactions)

	_, err := r.DB.NewUpdate().
		With("_data", values).
		Model((*model.SwapTransaction)(nil)).
		TableExpr("_data").
		Set("amount_token_from = (_data.amount_token_from #>> '{}')::numeric").
		Set("amount_token_to = (_data.amount_token_to #>> '{}')::numeric").
		Set("status = _data.status").
		Set("message = _data.message").
		Set("debug_message = _data.debug_message").
		Where("swapt.id = _data.id").
		Exec(ctx)

	return err
}

func (r *SwapTransactionRepository) Save(ctx context.Context, transaction *model.SwapTransaction) error {
	exists, err := r.DB.NewSelect().
		Model((*model.SwapTransaction)(nil)).
		Where("transaction_hash = ?", transaction.TransactionHash).
		Exists(ctx)

	if transaction.Status == "Wrong parameters" && transaction.Message == "budget exceeded" {
		exists = false
	}

	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return r.Create(ctx, transaction)

}

type SwapCampaignRepository struct {
	repository.Generic[model.SwapCampaign, uint64]
}

func NewSwapCampaignRepository(genericRepository repository.Generic[model.SwapCampaign, uint64]) *SwapCampaignRepository {
	return &SwapCampaignRepository{Generic: genericRepository}
}

func (r *SwapCampaignRepository) WithTx(tx bun.Tx) *SwapCampaignRepository {
	return &SwapCampaignRepository{Generic: r.Generic.WithTx(tx)}
}

func (r *SwapCampaignRepository) FindAllByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]model.SwapCampaign, int, error) {
	campaigns := make([]model.SwapCampaign, 0)

	query := r.DB.NewSelect().
		Model(&campaigns).
		Relation("CampaignType").
		Where("swapc.user_id = ?", userID).
		Order("created_at DESC")

	if page > 0 && pageSize > 0 {
		query = query.Limit(pageSize).Offset((page - 1) * pageSize)
	}

	err := query.Scan(ctx)

	if err != nil {
		return nil, 0, err
	}

	total, err := r.DB.NewSelect().
		Model((*model.SwapCampaign)(nil)).
		Where("user_id = ?", userID).
		Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

func (r *SwapCampaignRepository) GetByIDAndUserID(ctx context.Context, campaignID uuid.UUID, userID uint64) (*model.SwapCampaign, error) {
	campaign := new(model.SwapCampaign)
	err := r.DB.NewSelect().
		Model(campaign).
		Relation("CampaignType").
		Where("swapc.id = ? AND swapc.user_id = ?", campaignID, userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (r *SwapCampaignRepository) UpdateStatusByID(ctx context.Context, status model.SwapStatus, campaignID uuid.UUID) error {
	_, err := r.DB.NewUpdate().
		Model((*model.SwapCampaign)(nil)).
		Where("id = ?", campaignID).
		Set("status = ?", status).
		Exec(ctx)
	return err
}

func (r *SwapCampaignRepository) UpdateDoneIfNoPendingTransactions(ctx context.Context, campaignID uuid.UUID) (bool, error) {
	result, err := r.DB.NewUpdate().
		Model((*model.SwapCampaign)(nil)).
		Where("id = ?", campaignID).
		Where("status = ?", model.SwapStatusActive).
		Where(`NOT EXISTS (
			SELECT 1
			FROM swap_transactions AS st
			WHERE st.campaign_id = ?
			  AND st.status = 'Pending'
		)`, campaignID).
		Set("status = ?", model.SwapStatusDone).
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

func (r *SwapCampaignRepository) GetCampaignsSummary(ctx context.Context, page, pageSize int, userID uint64, campaignType string, status model.SwapStatus) ([]model.CampaignSummary, int, error) {
	var summaries []model.CampaignSummary
	query := r.DB.NewSelect().
		TableExpr("swap_campaigns sc").
		ColumnExpr("sc.id as campaign_id, sc.status, sc.budget, sc.token_mint_from, sc.token_mint_to, sc.project_id, swapct.name as type_name").
		ColumnExpr("COALESCE(SUM(CASE WHEN st.status = 'Success' THEN st.amount_token_from ELSE 0 END), 0) as spent_budget").
		Join("LEFT JOIN swap_transactions st ON st.campaign_id = sc.id").
		Join("LEFT JOIN swap_campaign_types swapct ON swapct.id = sc.type_id").
		Where("sc.user_id = ?", userID)

	switch campaignType {
	case model.TargetUpTaskType:
		query = query.Where("swapct.name = 'PULL UP'")
	case model.TargetDownTaskType:
		query = query.Where("swapct.name = 'PULL DOWN'")
	}

	if status != "" {
		query = query.Where("sc.status = ?", status)
	}

	query = query.
		Group("sc.id", "sc.status", "sc.budget", "sc.token_mint_from", "sc.token_mint_to", "sc.project_id", "swapct.name").
		Order("sc.created_at DESC")

	if page > 0 && pageSize > 0 {
		query = query.Limit(pageSize).Offset((page - 1) * pageSize)
	}

	err := query.Scan(ctx, &summaries)

	if err != nil {
		return nil, 0, err
	}

	total, err := r.DB.NewSelect().
		Model((*model.SwapCampaign)(nil)).
		Where("user_id = ?", userID).
		Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}
