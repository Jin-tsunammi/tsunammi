package auth

import (
	"github.com/google/uuid"
)

var (
	ErrGenerateToken        = "failed to generate token"
	ErrInvalidSigningMethod = "unexpected signing method"
	ErrInvalidToken         = "token is not valid"
	ErrInvalidTokenClaims   = "invalid token claims"
	ErrInvalidUserIDClaim   = "invalid user_id claim"
)

type JWTAuthenticator interface {
	GenerateTokenPair(options TokenClaims) (*TokenPairWithExpiration, error)
	ParseToken(accessToken string, tokenType int) (*TokenClaims, error)
}

type TokenPairWithExpiration struct {
	AccessToken  *TokenExpiration
	RefreshToken *TokenExpiration
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenClaims struct {
	SessionID uuid.UUID `json:"session_id"`
	UserID    uint64    `json:"user_id"`
}

func (c *TokenClaims) RefreshSessionID() bool {
	if c == nil {
		return false
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return false
	}

	c.SessionID = sessionID
	return true
}

func NewTokenClaims(userID uint64) TokenClaims {
	sessionID, _ := uuid.NewV7()

	return TokenClaims{
		UserID:    userID,
		SessionID: sessionID,
	}
}
