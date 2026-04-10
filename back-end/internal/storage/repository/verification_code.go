package repository

import (
	"context"
	"errors"
	"mm/config"
	"mm/internal/model"
	"mm/pkg/repository"
	"time"
)

type CodeRepository interface {
	Save(ctx context.Context, email, code string) error
	GetCodeByEmail(ctx context.Context, email string) (string, error)
	DeleteCodeByEmail(ctx context.Context, email string) error
}

type codeRepository struct {
	repository.Generic[model.EmailVerificationCode, uint64]
	cfg *config.Config
}

func NewCodeRepository(
	genericRepository repository.Generic[model.EmailVerificationCode, uint64],
	cfg *config.Config,
) CodeRepository {
	return &codeRepository{Generic: genericRepository, cfg: cfg}
}

func (r *codeRepository) Save(ctx context.Context, email, code string) error {
	codeEntity := model.NewVerificationCode(email, code)
	codeEntity.ExpiresAt = codeEntity.CreatedAt.Add(r.cfg.Auth.VerificationCodeTTL)

	return r.Create(ctx, codeEntity)

}

func (r *codeRepository) GetCodeByEmail(ctx context.Context, email string) (string, error) {
	code := new(model.EmailVerificationCode)
	err := r.DB.NewSelect().
		Model(code).
		Where("email = ?", email).
		Order("expires_at DESC").
		Scan(ctx)

	if err != nil {
		return "", err
	}

	if code.ExpiresAt.Before(time.Now()) {
		return "", errors.New("code expired")
	}

	return code.Code, nil
}

func (r *codeRepository) DeleteCodeByEmail(ctx context.Context, email string) error {
	_, err := r.DB.NewDelete().
		Model((*model.EmailVerificationCode)(nil)).
		Where("email = ?", email).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}
