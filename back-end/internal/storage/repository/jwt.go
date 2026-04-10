package repository

import (
	"context"
	"database/sql"
	"errors"
	"mm/internal/model"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"mm/pkg/repository"

	"github.com/google/uuid"
)

const userSessionsLimit = 5

type JWTRepository interface {
	Save(ctx context.Context, token *auth.TokenExpiration, claims auth.TokenClaims) error
	GetUserSession(ctx context.Context, userID uint64, sessionID uuid.UUID) (string, error)
	DeleteUserSession(ctx context.Context, userID uint64, sessionID uuid.UUID) error
}

func NewJWTRepository(genericRepository repository.Generic[model.Session, uuid.UUID]) JWTRepository {
	return &jwtRepository{Generic: genericRepository}
}

type jwtRepository struct {
	repository.Generic[model.Session, uuid.UUID]
}

func (r *jwtRepository) Save(ctx context.Context, token *auth.TokenExpiration, claims auth.TokenClaims) error {
	count, err := r.CountWithOptions(ctx, &repository.Options{})
	if err != nil {
		return apperrors.Internal("failed to check user session count", err)
	}
	if count >= userSessionsLimit {
		subquery := r.DB.NewSelect().
			Model((*model.Session)(nil)).
			Column("id").
			Where("user_id = ?", claims.UserID).
			Order("created_at ASC").
			Limit(1)
		_, err = r.DB.NewDelete().
			Model((*model.Session)(nil)).
			Where("id = (?)", subquery).
			Exec(ctx)

		if err != nil {
			return apperrors.Internal("failed to delete old session", err)
		}
	}
	session := model.NewSession(claims.SessionID, claims.UserID, token.Token, token.ExpiresAt)

	err = r.Create(ctx, session)
	if err != nil {
		return apperrors.Internal("failed to add user session", err)
	}
	return nil
}

func (r *jwtRepository) GetUserSession(ctx context.Context, userID uint64, sessionID uuid.UUID) (string, error) {
	session := new(model.Session)
	err := r.DB.NewSelect().
		Model(session).
		Where("user_id = ? AND id = ?", userID, sessionID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", apperrors.NotFound("user session not found")
		}
		return "", apperrors.Internal("failed to get user session", err)
	}
	return session.Token, nil

}

func (r *jwtRepository) DeleteUserSession(ctx context.Context, userID uint64, sessionID uuid.UUID) error {
	_, err := r.DB.NewDelete().
		Model((*model.Session)(nil)).
		Where("user_id = ? AND id = ?", userID, sessionID).
		Exec(ctx)
	if err != nil {
		return apperrors.Internal("failed to delete user session", err)
	}

	return nil
}
