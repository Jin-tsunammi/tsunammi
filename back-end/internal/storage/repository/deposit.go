package repository

import (
	"context"
	"mm/internal/model"
	"mm/pkg/mtype"
	"mm/pkg/repository"
)

type DepositRepository struct {
	repository.Generic[model.Deposit, uint64]
}

func NewDepositRepository(genericRepository repository.Generic[model.Deposit, uint64]) *DepositRepository {
	return &DepositRepository{Generic: genericRepository}
}

func (r *DepositRepository) GetAllBySum(ctx context.Context, sum mtype.BigRat) ([]model.Deposit, error) {
	deposits := make([]model.Deposit, 0)

	err := r.DB.NewSelect().
		Model(&deposits).
		Where("amount = ?", sum).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return deposits, nil
}

func (r *DepositRepository) UpdateAll(ctx context.Context, deposits []model.Deposit) error {
	if len(deposits) == 0 {
		return nil
	}

	values := r.DB.NewValues(&deposits)

	_, err := r.DB.NewUpdate().
		With("_data", values).
		Model((*model.Deposit)(nil)).
		ModelTableExpr("deposits AS d").
		TableExpr("_data").
		Set("amount = _data.amount").
		Set("status = _data.status").
		Where("d.id = _data.id").
		Exec(ctx)

	return err
}

func (r *DepositRepository) CreateAll(ctx context.Context, deposits []model.Deposit) error {
	if len(deposits) == 0 {
		return nil
	}

	_, err := r.DB.NewInsert().Model(&deposits).Exec(ctx)
	return err
}

type DepositOrderRepository struct {
	repository.Generic[model.DepositOrder, uint64]
}

func NewDepositOrderRepository(genericRepository repository.Generic[model.DepositOrder, uint64]) *DepositOrderRepository {
	return &DepositOrderRepository{Generic: genericRepository}
}

func (r *DepositOrderRepository) Save(ctx context.Context, order *model.DepositOrder) (*model.DepositOrder, error) {
	_, err := r.DB.NewInsert().
		Model(order).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *DepositOrderRepository) GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.DepositOrder, error) {
	order := new(model.DepositOrder)

	err := r.DB.NewSelect().
		Model(order).
		Join("JOIN projects AS p ON p.id = wo.project_id").
		Where("p.user_id = ?", userID).
		Where("wo.id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return order, nil
}
