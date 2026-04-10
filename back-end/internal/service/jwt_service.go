package service

import (
	"context"
	"mm/internal/storage/repository"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"strings"

	"github.com/google/uuid"
)

type JWTService struct {
	JwtRepository repository.JWTRepository
	JWT           auth.JWTAuthenticator
}

func NewJWTService(jwt auth.JWTAuthenticator, jwtRepository repository.JWTRepository) *JWTService {
	return &JWTService{
		JwtRepository: jwtRepository,
		JWT:           jwt,
	}
}

func (s *JWTService) GenerateTokenPair(ctx context.Context, claims auth.TokenClaims) (*auth.TokenPair, error) {
	tokenPair, err := s.JWT.GenerateTokenPair(claims)
	if err != nil {
		return nil, err
	}

	err = s.JwtRepository.Save(ctx, tokenPair.RefreshToken, claims)
	if err != nil {
		return nil, err
	}

	return &auth.TokenPair{
		AccessToken:  tokenPair.AccessToken.Token,
		RefreshToken: tokenPair.RefreshToken.Token,
	}, nil
}

func (s *JWTService) RefreshTokenPair(
	ctx context.Context,
	claims auth.TokenClaims,
	prevSessionID uuid.UUID,
) (*auth.TokenPair, error) {
	err := s.JwtRepository.DeleteUserSession(ctx, claims.UserID, prevSessionID)

	if err != nil {
		return nil, err
	}

	tokenPair, err := s.JWT.GenerateTokenPair(claims)
	if err != nil {
		return nil, err
	}

	err = s.JwtRepository.Save(ctx, tokenPair.RefreshToken, claims)
	err = error(nil)
	if err != nil {
		return nil, err
	}

	return &auth.TokenPair{
		AccessToken:  tokenPair.AccessToken.Token,
		RefreshToken: tokenPair.RefreshToken.Token,
	}, nil
}

func (s *JWTService) GetRefreshByUserID(
	ctx context.Context,
	userID uint64,
	sessionID uuid.UUID,
) (string, error) {
	refreshToken, err := s.JwtRepository.GetUserSession(ctx, userID, sessionID)

	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *JWTService) RefreshSession(
	ctx context.Context,
	token string,
) (*auth.TokenPair, error) {
	claims, err := s.ParseToken(token, 1)
	if err != nil {
		return nil, err
	}

	if claims == nil {
		return nil, apperrors.Unauthorized("invalid token claims")
	}

	storedRefresh, err := s.GetRefreshByUserID(ctx, claims.UserID, claims.SessionID)
	if err != nil {
		return nil, err
	}

	token = strings.TrimPrefix(token, "Bearer ")
	if token != storedRefresh {
		return nil, apperrors.Unauthorized("refresh token does not match with stored one")
	}

	prevSessionID := claims.SessionID
	claims.RefreshSessionID()

	tokenPair, err := s.RefreshTokenPair(ctx, *claims, prevSessionID)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (s *JWTService) ParseToken(token string, tokenType int) (*auth.TokenClaims, error) {
	return s.JWT.ParseToken(token, tokenType)
}
