package repository

import (
	"context"
	"mm/internal/model"
	"mm/pkg/repository"

	"github.com/google/uuid"
)

type PumpfunLaunchRepository struct {
	repository.Generic[model.PumpfunLaunch, uuid.UUID]
}

func NewPumpfunLaunchRepository(genericRepo repository.Generic[model.PumpfunLaunch, uuid.UUID]) *PumpfunLaunchRepository {
	return &PumpfunLaunchRepository{Generic: genericRepo}
}

func (r *PumpfunLaunchRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.PumpfunLaunchStatus) error {
	_, err := r.DB.NewUpdate().
		Model((*model.PumpfunLaunch)(nil)).
		Set("status = ?", status).
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx)
	return err
}
