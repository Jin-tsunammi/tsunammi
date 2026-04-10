package repository

import (
	"context"
	"mm/internal/model"

	"github.com/uptrace/bun"
)

type depositHistoryRow struct {
	AmountSOL         float64             `json:"-" bun:"deposit_order_amount"`
	PublicKey         string              `json:"-" bun:"wallet_public_key"`
	TransactionID     string              `json:"-" bun:"transaction_hash"`
	TransactionStatus model.DepositStatus `json:"-" bun:"transaction_status"`

	ID          uint64              `json:"id" bun:"deposit_order_id,pk"`
	ProjectID   uint64              `json:"project_id" bun:"project_id"`
	Name        string              `json:"name" bun:"project_name"`
	TotalSumSOL float64             `json:"total_sum_sol" bun:""`
	Status      model.DepositStatus `json:"status" bun:"deposit_order_status"`
	CreatedAt   string              `json:"created_at" bun:"deposit_order_created_at"`
}

func (r *DepositRepository) GetDepositHistory(ctx context.Context, userID uint64, page, pageSize int) ([]model.DepositHistoryResponse, int, error) {

	rows := make([]depositHistoryRow, 0)
	query := r.newDepositHistoryQuery(userID)

	total, err := query.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if page != 0 && pageSize != 0 {
		subq := r.DB.NewSelect().
			ColumnExpr("DISTINCT wo.id").
			Table("deposits").
			Join("JOIN deposit_orders AS wo ON wo.id = deposits.deposit_order_id").
			Join("JOIN wallets AS w ON w.id = deposits.wallet_id").
			Join("JOIN project_wallets AS pw ON pw.wallet_id = w.id").
			Join("JOIN projects AS p ON p.id = pw.project_id").
			Where("p.user_id = ?", userID).
			Order("wo.id asc").
			Limit(pageSize).
			Offset((page - 1) * pageSize)

		query = query.Where("wo.id IN (?)", subq)
	}

	query = query.Order("wo.id asc")

	err = query.Scan(ctx, &rows)

	if err != nil {
		return nil, 0, err
	}

	history := depositHistoryRowsToModels(rows)

	return history, total, nil
}
func (r *DepositRepository) GetDepositHistoryByProjectID(ctx context.Context, projectID, userID uint64) (*model.DepositHistoryResponse, error) {

	rows := make([]depositHistoryRow, 0)

	err := r.newDepositHistoryQuery(userID).
		Where("p.id = ?", projectID).
		Scan(ctx, &rows)

	if err != nil {
		return nil, err
	}

	depositHistoryResponse := depositHistoryRowsToModel(rows)

	return depositHistoryResponse, nil
}

func (r *DepositRepository) newDepositHistoryQuery(userID uint64) *bun.SelectQuery {
	query := r.DB.NewSelect().
		Table("deposits").
		ColumnExpr("p.name AS project_name").
		ColumnExpr("p.id AS project_id").
		ColumnExpr("wo.id AS deposit_order_id").
		ColumnExpr("deposits.amount AS deposit_order_amount").
		ColumnExpr("wo.status AS deposit_order_status").
		ColumnExpr("wo.created_at AS deposit_order_created_at").
		ColumnExpr("w.public_key AS wallet_public_key").
		ColumnExpr("deposits.transaction_id AS transaction_hash").
		ColumnExpr("deposits.status AS transaction_status").
		Join("JOIN deposit_orders AS wo ON wo.id = deposits.deposit_order_id").
		Join("JOIN wallets AS w ON w.id = deposits.wallet_id").
		Join("JOIN project_wallets AS pw ON pw.wallet_id = w.id").
		Join("JOIN projects AS p ON p.id = pw.project_id").
		Where("p.user_id = ?", userID).
		Order("wo.id")

	return query
}

func depositHistoryRowsToModel(flatRows []depositHistoryRow) *model.DepositHistoryResponse {
	history := depositHistoryRowsToModels(flatRows)

	if len(history) == 0 {
		return nil
	}

	return &history[0]
}

func depositHistoryRowsToModels(flatRows []depositHistoryRow) []model.DepositHistoryResponse {

	if len(flatRows) == 0 {
		return []model.DepositHistoryResponse{}
	}

	historyMap := make(map[uint64]*model.DepositHistoryResponse, len(flatRows))

	orderIDs := make([]uint64, 0, len(flatRows))

	for _, row := range flatRows {

		transaction := model.Transaction{
			PublicKey:     row.PublicKey,
			TransactionID: row.TransactionID,
			Status:        row.TransactionStatus,
			SumSOL:        row.AmountSOL,
			BalanceSOL:    0,
		}

		if _, exists := historyMap[row.ID]; !exists {
			orderIDs = append(orderIDs, row.ID)

			historyMap[row.ID] = &model.DepositHistoryResponse{
				ID:           row.ID,
				ProjectID:    row.ProjectID,
				ProjectName:  row.Name,
				Status:       row.Status,
				CreatedAt:    row.CreatedAt,
				Transactions: make([]model.Transaction, 0, len(flatRows)),
				TotalSumSOL:  0,
			}
		}

		historyMap[row.ID].Transactions = append(
			historyMap[row.ID].Transactions,
			transaction,
		)

	}

	finalHistory := make([]model.DepositHistoryResponse, len(orderIDs))
	for i, id := range orderIDs {
		finalHistory[i] = *historyMap[id]
	}

	return finalHistory
}
