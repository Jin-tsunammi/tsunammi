package repository

import (
	"context"
	"mm/internal/model"
	"mm/pkg/repository"
	"time"

	"github.com/uptrace/bun"
)

type UserHistoryRepository struct {
	repository.Generic[model.UserHistory, uint64]
}

func NewUserHistoryRepository(genericRepository repository.Generic[model.UserHistory, uint64]) *UserHistoryRepository {
	return &UserHistoryRepository{Generic: genericRepository}
}

func (r *UserHistoryRepository) WithTx(tx bun.Tx) *UserHistoryRepository {
	return &UserHistoryRepository{Generic: r.Generic.WithTx(tx)}
}

func (r *UserHistoryRepository) Save(ctx context.Context, event *model.UserHistory) (*model.UserHistory, error) {
	_, err := r.DB.NewInsert().
		Model(event).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	return event, nil
}

func (r *UserHistoryRepository) FetchAllByUserID(ctx context.Context, userID uint64, page, pageSize int, from, to time.Time) ([]model.UserHistory, int, error) {
	events := make([]model.UserHistory, 0)

	query := r.DB.NewSelect().
		Model(&events).
		Where("user_id = ?", userID)

	if !from.IsZero() && !to.IsZero() {
		query = query.Where("created_at BETWEEN ? AND ?", from, to)
	}

	total, err := query.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}

	query = query.Order("created_at DESC")

	err = query.Scan(ctx)

	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *UserHistoryRepository) UpdateAll(ctx context.Context, events []model.UserHistory) error {
	_, err := r.DB.NewUpdate().Model(&events).Exec(ctx)
	return err
}

func (r *UserHistoryRepository) CreateAll(ctx context.Context, events []model.UserHistory) error {
	_, err := r.DB.NewInsert().Model(&events).Exec(ctx)
	return err
}
