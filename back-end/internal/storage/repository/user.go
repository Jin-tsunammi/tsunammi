package repository

import (
	"context"
	"mm/internal/model"
	"mm/pkg/mtype"
	"mm/pkg/repository"
)

type UserRepository struct {
	repository.Generic[model.User, uint64]
}

func NewUserRepository(
	genericRepository repository.Generic[model.User, uint64],
) *UserRepository {
	return &UserRepository{
		Generic: genericRepository,
	}
}

func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email mtype.Email,
) (*model.User, error) {
	var user = new(model.User)

	err := r.DB.NewSelect().
		Model(user).
		Where("email = ?", email.String()).
		Scan(ctx)
	if repository.IsErrNoRows(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) FindByPublicAddress(
	ctx context.Context,
	address string,
) (*model.User, error) {
	var user = new(model.User)

	err := r.DB.NewSelect().
		Model(user).
		Where("public_key = ?", address).
		Scan(ctx)
	if repository.IsErrNoRows(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) CountUsers(
	ctx context.Context,
) (int, error) {
	userCount, err := r.DB.NewSelect().
		Model((*model.User)(nil)).
		Count(ctx)

	return userCount, err
}

func (r *UserRepository) CreateWithPublicAddress(
	ctx context.Context,
	user *model.User,
) (*model.User, error) {
	_, err := r.DB.NewInsert().
		Model(user).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UpdateEmail(ctx context.Context, email mtype.Email, userID uint64) error {
	_, err := r.DB.NewUpdate().
		Model((*model.User)(nil)).
		Where("id = ?", userID).
		Set("email = ?", email.String()).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}
